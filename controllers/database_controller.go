/*
Copyright 2021-2022 Nutanix, Inc.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	"github.com/nutanix-cloud-native/ndb-operator/util"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=databases/finalizers,verbs=update
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

	spec := database.Spec
	server := spec.NDB
	ndbClient := ndbclient.NewNDBClient(server.Credentials.LoginUser, server.Credentials.Password, server.Server)

	// log.Info(fmt.Sprintf("Finalizers: %v", database.Finalizers))

	// Examine DeletionTimestamp to determine if object is under deletion
	if database.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted,
		// if it does not have our finalizer then add the finalizer(s) and update the object.
		if !controllerutil.ContainsFinalizer(database, ndbv1alpha1.FINALIZER_DATABASE_INSTANCE) {
			return r.addFinalizer(ctx, req, ndbv1alpha1.FINALIZER_DATABASE_INSTANCE, database)
		}
		if !controllerutil.ContainsFinalizer(database, ndbv1alpha1.FINALIZER_DATABASE_SERVER) {
			return r.addFinalizer(ctx, req, ndbv1alpha1.FINALIZER_DATABASE_SERVER, database)
		}
	} else {
		// The object is under deletion. Perform deletion based on the finalizers we've added.
		return r.handleDelete(ctx, database, ndbClient)
	}

	// To check and handle the case when the database ha been deleted/aborted externally (not through the operator).
	err = r.handleExternalDelete(ctx, database, ndbClient)
	if err != nil {
		log.Error(err, "Error occured while external delete check")
		return r.requeueOnErr(err)
	}

	return r.handleSync(ctx, database, ndbClient)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.Database{}).
		Complete(r)
}
