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
	"math"
	"time"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	"github.com/nutanix-cloud-native/ndb-operator/util"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// doNotRequeue Finished processing. No need to put back on the reconcile queue.
func (r *DatabaseReconciler) doNotRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// requeue Not finished processing. Put back on reconcile queue and continue.
func (r *DatabaseReconciler) requeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// requeueOnErr Failed while processing. Put back on reconcile queue and try again.
func (r *DatabaseReconciler) requeueOnErr(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// requeue after a timeout. Put back on reconcile queue after a timeout and continue.
func (r *DatabaseReconciler) requeueWithTimeout(t int) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: time.Second * time.Duration(math.Abs(float64(t)))}, nil
}

func (r *DatabaseReconciler) addFinalizer(ctx context.Context, req ctrl.Request, finalizer string, database *ndbv1alpha1.Database) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Fetching the most recent version of the database CR")
	err := r.Get(ctx, req.NamespacedName, database)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Database resource not found. Ignoring since object must be deleted")
			return r.doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Database")
		return r.requeueOnErr(err)
	}
	log.Info("Database CR fetched. Adding finalizer " + finalizer)
	controllerutil.AddFinalizer(database, finalizer)
	if err := r.Update(ctx, database); err != nil {
		return r.requeueOnErr(err)
	} else {
		log.Info("Added finalizer " + finalizer)
	}
	//Not requeuing as a successful update automatically triggers a reconcile.
	return r.doNotRequeue()
}

// handleDelete function handles the deletion of
// 		a. Database instance
//		b. Database server
func (r *DatabaseReconciler) handleDelete(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndbclient.NDBClient) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Database CR is being deleted.")
	if controllerutil.ContainsFinalizer(database, ndbv1alpha1.FINALIZER_DATABASE_INSTANCE) {
		// Check if the database instance id (database.Status.Id) is present in the status
		// If present, then make a deprovisionDatabase API call to NDB
		// else proceed with removing finalizer as database instance provisioning wasn't successful earlier.
		if database.Status.Id != "" {
			log.Info("Deprovisioning database instance from NDB.")
			_, err := ndbv1alpha1.DeprovisionDatabase(ctx, ndbClient, database.Status.Id, *ndbv1alpha1.GenerateDeprovisionDatabaseRequest())
			if err != nil {
				log.Error(err, "Deprovisioning database instance request failed.")
				return r.requeueOnErr(err)
			}
		}
		log.Info("Removing Finalizer " + ndbv1alpha1.FINALIZER_DATABASE_INSTANCE)
		controllerutil.RemoveFinalizer(database, ndbv1alpha1.FINALIZER_DATABASE_INSTANCE)
		if err := r.Update(ctx, database); err != nil {
			return r.requeueOnErr(err)
		}
		log.Info("Removed Finalizer " + ndbv1alpha1.FINALIZER_DATABASE_INSTANCE)

	} else if controllerutil.ContainsFinalizer(database, ndbv1alpha1.FINALIZER_DATABASE_SERVER) {
		// Checking if the database instance still exists in NDB. (It might take some time for the delete db instance operation to complete)
		// Proceed to delete the database server vm only after the database instance has been deleted.
		log.Info("Checking if database instance exists")
		allDatabases, err := ndbv1alpha1.GetAllDatabases(ctx, ndbClient)
		if err != nil {
			log.Error(err, "An error occured while trying to get all databases")
			return r.requeueOnErr(err)
		}
		if len(util.Filter(allDatabases, func(d ndbv1alpha1.DatabaseResponse) bool { return d.Id == database.Status.Id })) == 0 {
			// Could not find the database with the given database id => database instance has been deleted
			log.Info("Database instance not found, attempting to remove database server.")
			databaseServerId := database.Status.DatabaseServerId
			// Make a dbserver deprovisioning request to NDB only if the serverId is present in status
			if databaseServerId != "" {
				_, err := ndbv1alpha1.DeprovisionDatabaseServer(ctx, ndbClient, databaseServerId, *ndbv1alpha1.GenerateDeprovisionDatabaseServerRequest())
				if err != nil {
					log.Error(err, "Deprovisioning database server request failed.", "database server id", databaseServerId)
					return r.requeueOnErr(err)
				}
			} else {
				log.Info("Database server id was not found on the database CR, removing finalizers and deleting the CR.")
			}
			// remove our finalizer from the list and update it.
			log.Info("Removing Finalizer " + ndbv1alpha1.FINALIZER_DATABASE_SERVER)
			controllerutil.RemoveFinalizer(database, ndbv1alpha1.FINALIZER_DATABASE_SERVER)
			if err := r.Update(ctx, database); err != nil {
				return r.requeueOnErr(err)
			}
			log.Info("Removed Finalizer " + ndbv1alpha1.FINALIZER_DATABASE_SERVER)
			return r.requeue()
		}
	} else {
		// Both database instance and database server finalizers have been removed, no need to requeue
		// CR will be deleted.
		return r.doNotRequeue()
	}
	// Requeue the request while waiting for the database instance to be deleted from NDB.
	return r.requeueWithTimeout(15)
}

// In the case when a database has been provisioned through the operator an Id is assigned to the database CR.
// If someone deletes the database externally/aborts the operation through NDB (and not through the operator), the operator should
// create a new database. To do this, we fetch the database by the Id we have in the datbase CR's Status.Id.
// If the database response's status is empty, we set our CR's status to be empty so that it is provisioned again.
func (r *DatabaseReconciler) handleExternalDelete(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndbclient.NDBClient) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleExternalDelete")
	if database.Status.Id != "" {
		// database was provisioned earlier => sync status
		var databaseResponse ndbv1alpha1.DatabaseResponse
		allDatabases, err := ndbv1alpha1.GetAllDatabases(ctx, ndbClient)
		if err != nil {
			log.Error(err, "An error occured while trying to get all databases")
			return err
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
			log.Info("The database might have been deleted externally, setting an empty status so it can be re-provisioned.")
			database.Status.Status = ndbv1alpha1.DATABASE_CR_STATUS_EMPTY
			err = r.Status().Update(ctx, database)
			if err != nil {
				log.Error(err, "Failed to update database status")
				return err
			}
		}

	}
	return nil
}

func (r *DatabaseReconciler) handleSync(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndbclient.NDBClient) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleSync")

	switch database.Status.Status {

	case ndbv1alpha1.DATABASE_CR_STATUS_EMPTY:
		// DB Status.Status is empty => Provision a DB
		log.Info("Provisioning a database instance with NDB.")

		generatedReq, err := ndbv1alpha1.GenerateProvisioningRequest(ctx, ndbClient, database.Spec)
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

	case ndbv1alpha1.DATABASE_CR_STATUS_READY:
		return r.requeueWithTimeout(15)

	default:
		// Do Nothing
	}

	return r.doNotRequeue()
}
