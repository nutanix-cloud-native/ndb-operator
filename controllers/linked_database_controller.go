/*
Copyright 2022-2023 Nutanix, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

type linkedDatabaseNDBClientFactory func(username, password, server, caCert string, skipVerify bool) ndb_client.NDBClientHTTPInterface

// LinkedDatabaseReconciler reconciles a LinkedDatabase object.
type LinkedDatabaseReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	recorder         record.EventRecorder
	ndbClientFactory linkedDatabaseNDBClientFactory
}

// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="core",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=ndb.nutanix.com,resources=ndbservers,verbs=get;list;watch
// +kubebuilder:rbac:groups=ndb.nutanix.com,resources=linkeddatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ndb.nutanix.com,resources=linkeddatabases/status,verbs=get;update;patch

func (r *LinkedDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("LinkedDatabase reconcile started")

	linkedDatabase := &ndbv1alpha1.LinkedDatabase{}
	if err := r.Get(ctx, req.NamespacedName, linkedDatabase); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("LinkedDatabase resource not found. Ignoring since object must be deleted")
			return doNotRequeue()
		}
		log.Error(err, "Failed to get LinkedDatabase")
		return requeueOnErr(err)
	}

	if !linkedDatabase.ObjectMeta.DeletionTimestamp.IsZero() {
		return doNotRequeue()
	}

	ndbServer := &ndbv1alpha1.NDBServer{}
	if err := r.Get(ctx, types.NamespacedName{Name: linkedDatabase.Spec.NDBRef}, ndbServer); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("NDBServer resource not found. Ignoring until it exists", "ndbRef", linkedDatabase.Spec.NDBRef)
			return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
		}
		log.Error(err, "Failed to get NDBServer")
		return requeueOnErr(err)
	}

	username, password, caCert, err := getNDBCredentialsFromSecret(ctx, r.Client, ndbServer.Spec.CredentialSecretRef)
	if err != nil {
		r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_INVALID_CREDENTIALS, "Error: %s", err.Error())
		return requeueOnErr(err)
	}

	ndbClient := r.newNDBClient(username, password, ndbServer.Spec.Server, caCert, ndbServer.Spec.SkipCertificateVerification)
	return r.handleLinkedDatabaseSync(ctx, linkedDatabase, ndbServer, ndbClient)
}

func (r *LinkedDatabaseReconciler) newNDBClient(username, password, server, caCert string, skipVerify bool) ndb_client.NDBClientHTTPInterface {
	if r.ndbClientFactory != nil {
		return r.ndbClientFactory(username, password, server, caCert, skipVerify)
	}
	return ndb_client.NewNDBClient(username, password, server, caCert, skipVerify)
}

func (r *LinkedDatabaseReconciler) handleLinkedDatabaseSync(ctx context.Context, linkedDatabase *ndbv1alpha1.LinkedDatabase, ndbServer *ndbv1alpha1.NDBServer, ndbClient ndb_client.NDBClientHTTPInterface) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	status := linkedDatabase.Status

	if status.Status == "" && status.CreationOperationId == "" {
		sourceDatabaseId, err := resolveLinkedDatabaseSourceDatabaseId(linkedDatabase, ndbServer)
		if err != nil {
			r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_RESOURCE_LOOKUP_ERROR, "Error: %s", err.Error())
			return requeueOnErr(err)
		}

		existingLinkedDatabase, err := ndb_api.GetLinkedDatabaseByName(ctx, ndbClient, sourceDatabaseId, linkedDatabase.Spec.DatabaseName)
		if err != nil {
			r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_NDB_REQUEST_FAILED, "Error checking linked database existence: %s", err.Error())
			return requeueOnErr(err)
		}
		if existingLinkedDatabase != nil {
			status.Status = common.DATABASE_CR_STATUS_READY
			status.SourceDatabaseId = sourceDatabaseId
			status.LinkedDatabaseId = existingLinkedDatabase.Id
			status.Message = "Linked database already exists on NDB"
			r.recordEvent(linkedDatabase, corev1.EventTypeNormal, EVENT_CREATION_COMPLETED, "Linked database already exists on NDB")
		} else {
			request, err := ndb_api.GenerateLinkedDatabaseProvisionRequest(linkedDatabase.Spec.DatabaseName)
			if err != nil {
				r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_REQUEST_GENERATION_FAILURE, "Error: %s", err.Error())
				return requeueOnErr(err)
			}
			r.recordEvent(linkedDatabase, corev1.EventTypeNormal, EVENT_REQUEST_GENERATION, "Generated linked database provisioning request")

			task, err := ndb_api.ProvisionLinkedDatabase(ctx, ndbClient, sourceDatabaseId, request)
			if err != nil {
				r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_NDB_REQUEST_FAILED, "Error: %s", err.Error())
				return requeueOnErr(err)
			}

			status.Status = common.DATABASE_CR_STATUS_CREATING
			status.SourceDatabaseId = sourceDatabaseId
			status.CreationOperationId = task.OperationId
			r.recordEvent(linkedDatabase, corev1.EventTypeNormal, EVENT_CREATION_STARTED, "Linked database creation initiated on NDB")
		}
	} else if status.Status == common.DATABASE_CR_STATUS_CREATING {
		operation, err := ndb_api.GetOperationById(ctx, ndbClient, status.CreationOperationId)
		if err != nil {
			r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_NDB_REQUEST_FAILED, "Error fetching operation %s: %s", status.CreationOperationId, err.Error())
		} else {
			switch ndb_api.GetOperationStatus(operation) {
			case ndb_api.OPERATION_STATUS_FAILED:
				status.Status = common.DATABASE_CR_STATUS_CREATION_ERROR
				status.Message = operation.Message
				r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_CREATION_FAILED, "Linked database creation failed: %s", operation.Message)
			case ndb_api.OPERATION_STATUS_PASSED:
				status.Status = common.DATABASE_CR_STATUS_READY
				status.Message = operation.Message
				linkedDatabaseResponse, lookupErr := ndb_api.GetLinkedDatabaseByName(ctx, ndbClient, status.SourceDatabaseId, linkedDatabase.Spec.DatabaseName)
				if lookupErr != nil {
					r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_NDB_REQUEST_FAILED, "Error fetching linked database after creation: %s", lookupErr.Error())
				} else if linkedDatabaseResponse != nil {
					status.LinkedDatabaseId = linkedDatabaseResponse.Id
				}
				r.recordEvent(linkedDatabase, corev1.EventTypeNormal, EVENT_CREATION_COMPLETED, "Linked database creation operation passed")
			default:
				log.Info("Linked database creation operation is still running", "operationId", operation.Id, "status", operation.Status)
			}
		}
	}

	if !reflect.DeepEqual(linkedDatabase.Status, status) {
		if err := r.updateLinkedDatabaseStatusWithRetry(ctx, linkedDatabase, status); err != nil {
			r.recordEvent(linkedDatabase, corev1.EventTypeWarning, EVENT_CR_STATUS_UPDATE_FAILED, "Error: %s", err.Error())
			return requeueOnErr(err)
		}
	}

	switch status.Status {
	case common.DATABASE_CR_STATUS_READY, common.DATABASE_CR_STATUS_CREATION_ERROR:
		return doNotRequeue()
	default:
		return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
	}
}

func resolveLinkedDatabaseSourceDatabaseId(linkedDatabase *ndbv1alpha1.LinkedDatabase, ndbServer *ndbv1alpha1.NDBServer) (string, error) {
	if linkedDatabase.Spec.SourceDatabaseId != "" {
		return linkedDatabase.Spec.SourceDatabaseId, nil
	}
	if linkedDatabase.Spec.SourceDatabaseName == "" {
		return "", fmt.Errorf("either sourceDatabaseId or sourceDatabaseName must be provided")
	}
	for _, database := range ndbServer.Status.Databases {
		if database.Name == linkedDatabase.Spec.SourceDatabaseName {
			return database.Id, nil
		}
	}
	return "", fmt.Errorf("source database %q was not found in NDBServer status", linkedDatabase.Spec.SourceDatabaseName)
}

func (r *LinkedDatabaseReconciler) updateLinkedDatabaseStatusWithRetry(ctx context.Context, linkedDatabase *ndbv1alpha1.LinkedDatabase, status ndbv1alpha1.LinkedDatabaseStatus) error {
	linkedDatabase.Status = status
	if err := r.Status().Update(ctx, linkedDatabase); err == nil {
		return nil
	}
	if err := r.Get(ctx, types.NamespacedName{Name: linkedDatabase.Name, Namespace: linkedDatabase.Namespace}, linkedDatabase); err != nil {
		return err
	}
	linkedDatabase.Status = status
	return r.Status().Update(ctx, linkedDatabase)
}

func (r *LinkedDatabaseReconciler) recordEvent(linkedDatabase *ndbv1alpha1.LinkedDatabase, eventtype, reason, messageFmt string, args ...interface{}) {
	if r.recorder == nil {
		return
	}
	r.recorder.Eventf(linkedDatabase, eventtype, reason, messageFmt, args...)
}

// SetupWithManager sets up the controller with the Manager.
func (r *LinkedDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("linked-database-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.LinkedDatabase{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
