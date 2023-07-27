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

/*
GENERATED by operator-sdk
Changes added
*/

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ndb.nutanix.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ndb.nutanix.com,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ndb.nutanix.com,resources=databases/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=services;endpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// The Reconcile method is where the controller logic resides.
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("<==============================Reconcile Started=============================>")
	// Fetch the database resource from the namespace
	database := &ndbv1alpha1.Database{}
	err := r.Get(ctx, req.NamespacedName, database)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("Database resource not found. Ignoring since object must be deleted")
			return r.doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Database")
		return r.requeueOnErr(err)
	}

	log.Info("Database CR Status: " + util.ToString(database.Status))

	NDBInfo := database.Spec.NDB
	username, password, caCert, err := r.getNDBCredentials(ctx, NDBInfo.CredentialSecret, req.Namespace)
	if err != nil || username == "" || password == "" {
		var errStatement string
		if err == nil {
			errStatement = "NDB username or password cannot be empty"
			err = fmt.Errorf("empty NDB credentials")
		} else {
			errStatement = "An error occured while fetching the NDB Secrets"
		}
		log.Error(err, errStatement)
		return r.requeueOnErr(err)
	}
	if caCert == "" {
		log.Info("Ca-cert not found, falling back to host's HTTPs certs.")
	}
	ndbClient := ndb_client.NewNDBClient(username, password, NDBInfo.Server, caCert, NDBInfo.SkipCertificateVerification)

	// Examine DeletionTimestamp to determine if object is under deletion
	if database.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted,
		// if it does not have our finalizer then add the finalizer(s) and update the object.
		if !controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_INSTANCE) {
			return r.addFinalizer(ctx, req, common.FINALIZER_DATABASE_INSTANCE, database)
		}
		if !controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_SERVER) {
			return r.addFinalizer(ctx, req, common.FINALIZER_DATABASE_SERVER, database)
		}
	} else {
		// The object is under deletion. Perform deletion based on the finalizers we've added.
		return r.handleDelete(ctx, database, ndbClient)
	}

	// To check and handle the case when the database ha been deleted/aborted externally (not through the operator).
	err = r.handleExternalDelete(ctx, database, ndbClient)
	if err != nil {
		log.Error(err, "Error occurred while external delete check")
		return r.requeueOnErr(err)
	}
	// Synchronize the database CR with the database instance on NDB.
	return r.handleSync(ctx, database, ndbClient, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.Database{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Endpoints{}).
		Complete(r)
}
