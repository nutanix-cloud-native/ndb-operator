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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/util"
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
	log := log.FromContext(ctx)

	// Fetch the Memcached instance
	// The purpose is check if the Custom Resource for the Kind Memcached
	// is applied on the cluster if not we return nil to stop the reconciliation
	dataProtection := &v1alpha1.DataProtection{}
	err := r.Get(ctx, req.NamespacedName, dataProtection)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then, it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("memcached resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get memcached")
		return ctrl.Result{}, err
	}

	log.Info("Database CR Status: " + util.ToString(dataProtection.Status))

	// Examine DeletionTimestamp to determine if object is under deletion
	if dataProtection.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted,
		// if it does not have our finalizer then add the finalizer(s) and update the object.
		if !controllerutil.ContainsFinalizer(dataProtection, ndbv1alpha1.FINALIZER_DATAPROTECTION_INSTANCE) {
			return r.addFinalizer(ctx, req, ndbv1alpha1.FINALIZER_DATAPROTECTION_INSTANCE, dataProtection)
		}
	} else {
		// The object is under deletion. Perform deletion based on the finalizers we've added.
		return r.handleDelete(ctx, database, ndbClient)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataProtectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.DataProtection{}).
		Complete(r)
}
