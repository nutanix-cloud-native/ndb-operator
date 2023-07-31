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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// NDBServerReconciler reconciles a NDBServer object
type NDBServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=ndbservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=ndbservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=ndbservers/finalizers,verbs=update

/*
Reconciles the NDBServer custom resources by
1. Checks for deletion
2. Verify credentials and connectivity
3. Take actions based on current status.status, fetch data
4. Update the status if any changes are observed (excluding counter)
*/
func (r *NDBServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("NDBServer reconcile started")
	// 1. Checks for deletion
	// Fetch the NDBServer resource from the namespace
	ndbServer := &ndbv1alpha1.NDBServer{}
	err := r.Get(ctx, req.NamespacedName, ndbServer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("NDBServer resource not found. Ignoring since object must be deleted")
			return doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get NDBServer")
		return requeueOnErr(err)
	}

	// Copy the status, this will be updated by the end of the reconcile
	status := ndbServer.Status.DeepCopy()
	// Initialize maps if they do not exist
	if status.Databases == nil {
		status.Databases = make(map[string]ndbv1alpha1.NDBServerDatabaseInfo)
	}

	// log.Info("NDBServer CR Status: " + util.ToString(status))

	// 2. Verify credentials and connectivity
	// Fetch credentials and check Authentication
	username, password, caCert, err := getNDBCredentialsFromSecret(ctx, r.Client, ndbServer.Spec.CredentialSecret, req.Namespace)
	ndbClient := ndb_client.NewNDBClient(username, password, ndbServer.Spec.Server, caCert, ndbServer.Spec.SkipCertificateVerification)
	if err != nil {
		log.Error(err, "Credential Error: error while fetching credentials from CredentialSecret", "secret name", ndbServer.Spec.CredentialSecret)
		status.Status = common.NDB_CR_STATUS_CREDENTIAL_ERROR
	} else {
		authResponse, err := ndb_api.AuthValidate(ctx, ndbClient)
		if err != nil || authResponse.Status != "success" {
			log.Error(err, "Authentication Error: Could not verify connectivity / auth credentials for NDB")
			status.Status = common.NDB_CR_STATUS_AUTHENTICATION_ERROR
		} else {
			status.Status = common.NDB_CR_STATUS_OK
		}
	}

	// 3. Take actions based on current status.status
	switch status.Status {
	case common.NDB_CR_STATUS_CREDENTIAL_ERROR, common.NDB_CR_STATUS_AUTHENTICATION_ERROR:
		// no-op
	case common.NDB_CR_STATUS_OK:
		// Get Status (check and perform data fetching, update counters)
		status = getNDBServerStatus(ctx, status, ndbClient)
	default:
		// no-op
		return doNotRequeue()
	}

	// 4. Update the status if any changes are observed (excluding counter)
	if !util.DeepEqualWithException(ndbServer.Status, *status, "Counter") {
		log.Info("Status Changed, updating lastUpdated time")
		status.LastUpdated = time.Now().Format(time.DateTime)
	}
	ndbServer.Status = *status
	err = r.Status().Update(ctx, ndbServer)
	if err != nil {
		log.Error(err, "Failed to update ndbServer status")
		return requeueOnErr(err)
	}

	return requeueWithTimeout(common.NDB_RECONCILE_INTERVAL_SECONDS)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NDBServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.NDBServer{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
