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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// DataProtectionReconciler reconciles a DataProtection object
type DataProtectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=dataprotections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=dataprotections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=dataprotections/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DataProtection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DataProtectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("<==============================Reconcile Started=============================>")
	// Fetch the data protection resource from the namespace
	dataprotection := &ndbv1alpha1.DataProtection{}
	err := r.Get(ctx, req.NamespacedName, dataprotection)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("Data Protection resource not found. Ignoring since object must be deleted")
			return r.doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Database")
		return r.requeueOnErr(err)
	}

	log.Info("DataProtection CR Status: " + util.ToString(dataprotection.Status))

	NDBInfo := dataprotection.Spec.NDB
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

	// Synchronize the database CR with the database instance on NDB.
	return r.handleSync(ctx, dataprotection, ndbClient, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataProtectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.DataProtection{}).
		Complete(r)
}
