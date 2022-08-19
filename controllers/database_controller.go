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
		return r.handleDelete(ctx, database, ndbClient)
	}

	// In the case when a database has been provisioned through the operator an Id is assigned to the database CR.
	// If someone deletes the database externally/aborts the operation through NDB (and not through the operator), the operator should
	// create a new database. To do this, we fetch the database by the Id we have in the datbase CR's Status.Id.
	// If the database response's status is empty, we set our CR's status to be empty so that it is provisioned again.
	if database.Status.Id != "" {
		// database was provisioned earlier => sync status

		var databaseResponse ndbv1alpha1.DatabaseResponse

		allDatabases, err := ndbv1alpha1.GetAllDatabases(ctx, ndbClient)
		if err != nil {
			log.Error(err, "An error occured while trying to get all databases")
			return r.requeueOnErr(err)
		} else {
			for _, db := range allDatabases {
				if db.Id == database.Status.Id {
					databaseResponse = db
					break
				}
			}
		}

		// Update the CR status if the database response is empty so that it triggers a provision operation
		if databaseResponse.Status == ndbv1alpha1.DATABASE_CR_STATUS_EMPTY {
			database.Status.Status = ndbv1alpha1.DATABASE_CR_STATUS_EMPTY
			err = r.Status().Update(ctx, database)
			if err != nil {
				log.Error(err, "Failed to update database status")
				return r.requeueOnErr(err)
			}
		}

	}

	switch database.Status.Status {

	case ndbv1alpha1.DATABASE_CR_STATUS_EMPTY:
		// DB Status.Status is empty => Provision a DB
		log.Info("Provisioning a database instance with NDB.")

		generatedReq, err := ndbv1alpha1.GenerateProvisioningRequest(ctx, ndbClient, spec)
		if err != nil {
			log.Error(err, "Could not generate provisioning request, requeuing.")
			return r.requeueOnErr(err)
		}

		taskResponse, err := ndbv1alpha1.ProvisionDatabase(ctx, ndbClient, generatedReq)
		if err != nil {
			log.Error(err, "An error occured while trying to provision the database")
			return r.requeueOnErr(err)
		}
		// log.Info(fmt.Sprintf("Provisioning response from NDB: %+v", taskResponse))

		log.Info("Setting database CR status to provisioning and id as " + taskResponse.EntityId)
		database.Status.Status = ndbv1alpha1.DATABASE_CR_STATUS_PROVISIONING
		database.Status.Id = taskResponse.EntityId

		err = r.Status().Update(ctx, database)
		if err != nil {
			log.Error(err, "Failed to update database status")
			return r.requeueOnErr(err)
		}

	case ndbv1alpha1.DATABASE_CR_STATUS_READY:
		return r.requeueWithTimeout(15)

	case ndbv1alpha1.DATABASE_CR_STATUS_PROVISIONING:
		// Check the status of the DB
		databaseResponse, err := ndbv1alpha1.GetDatabaseById(ctx, ndbClient, database.Status.Id)
		if err != nil {
			log.Error(err, "An error occured while trying to get the database with id: "+database.Status.Id)
			r.requeueOnErr(err)
		}
		// if READY => Change status
		// log.Info("DEBUG Database Response: " + util.ToString(databaseResponse))
		if databaseResponse.Status == ndbv1alpha1.DATABASE_CR_STATUS_READY {
			log.Info("Database instance is READY, adding data to CR's status and updating the CR")
			database.Status.Status = ndbv1alpha1.DATABASE_CR_STATUS_READY
			database.Status.DatabaseServerId = databaseResponse.DatabaseNodes[0].DatabaseServerId
			for _, property := range databaseResponse.Properties {
				if property.Name == ndbv1alpha1.PROPERTY_NAME_VM_IP {
					database.Status.IPAddress = property.Value
				}
			}

			err = r.Status().Update(ctx, database)
			if err != nil {
				log.Error(err, "Failed to update database status")
				return r.requeueOnErr(err)
			}
		}
		// If database instance is not yet ready, requeue with wait
		return r.requeueWithTimeout(15)

	default:
		// Do Nothing
	}
	return r.doNotRequeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.Database{}).
		Complete(r)
}
