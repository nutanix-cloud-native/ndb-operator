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

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/controller_adapters"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

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
			return doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Database")
		return requeueOnErr(err)
	}
	log.Info("Database CR fetched. Adding finalizer " + finalizer)
	controllerutil.AddFinalizer(database, finalizer)
	if err := r.Update(ctx, database); err != nil {
		return requeueOnErr(err)
	} else {
		log.Info("Added finalizer " + finalizer)
	}
	//Not requeuing as a successful update automatically triggers a reconcile.
	return doNotRequeue()
}

// handleDelete function handles the deletion of
//
//	a. Database instance
//	b. Database server
func (r *DatabaseReconciler) handleDelete(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Database CR is being deleted")
	if controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_INSTANCE) {
		// Check if the database instance id (database.Status.Id) is present in the status
		// If present, then make a deprovisionDatabase API call to NDB
		// else proceed with removing finalizer as database instance provisioning wasn't successful earlier.
		if database.Status.Id != "" {
			infoStatement := "Deprovisioning database instance from NDB."
			log.Info(infoStatement)
			r.recorder.Event(database, "Normal", EVENT_DEPROVISIONING_STARTED, infoStatement)

			_, err := ndb_api.DeprovisionDatabase(ctx, ndbClient, database.Status.Id, *ndb_api.GenerateDeprovisionDatabaseRequest())
			if err != nil {
				errStatement := "Deprovisioning database instance request failed."
				log.Error(err, errStatement)
				r.recorder.Eventf(database, "Warning", EVENT_DEPROVISIONING_FAILED, "Error: %s. %s", errStatement, err.Error())
				return requeueOnErr(err)
			}
		}

		log.Info("Removing Finalizer " + common.FINALIZER_DATABASE_INSTANCE)
		controllerutil.RemoveFinalizer(database, common.FINALIZER_DATABASE_INSTANCE)
		if err := r.Update(ctx, database); err != nil {
			return requeueOnErr(err)
		}
		log.Info("Removed Finalizer " + common.FINALIZER_DATABASE_INSTANCE)

	} else if controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_SERVER) {
		// Checking if the database instance still exists in NDB. (It might take some time for the delete db instance operation to complete)
		// Proceed to delete the database server vm only after the database instance has been deleted.
		log.Info("Checking if database instance exists")
		allDatabases, err := ndb_api.GetAllDatabases(ctx, ndbClient)
		if err != nil {
			errStatement := "Error fetching all databases from NDB"
			log.Error(err, errStatement)
			r.recorder.Eventf(database, "Warning", EVENT_RESOURCE_LOOKUP_ERROR, "Error: %s. %s", errStatement, err.Error())
			return requeueOnErr(err)
		}
		if len(util.Filter(allDatabases, func(d ndb_api.DatabaseResponse) bool { return d.Id == database.Status.Id })) == 0 {
			// Could not find the database with the given database id => database instance has been deleted
			log.Info("Database instance not found, attempting to remove database server.")
			r.recorder.Eventf(database, "Normal", EVENT_DEPROVISIONING_COMPLETED, "Database deprovisioned from NDB.")
			r.recorder.Eventf(database, "Normal", EVENT_DEPROVISIONING_STARTED, "Deprovisioning database server from NDB.")
			databaseServerId := database.Status.DatabaseServerId
			// Make a dbserver deprovisioning request to NDB only if the serverId is present in status
			if databaseServerId != "" {
				_, err := ndb_api.DeprovisionDatabaseServer(ctx, ndbClient, databaseServerId, *ndb_api.GenerateDeprovisionDatabaseServerRequest())
				if err != nil {
					errStament := fmt.Sprintf("Deprovisioning database server request failed for id: %s", databaseServerId)
					log.Error(err, errStament)
					r.recorder.Eventf(database, "Warning", EVENT_DEPROVISIONING_FAILED, "Error: %s. %s", errStament, err.Error())
					return requeueOnErr(err)
				}
			} else {
				// Database and server has been deprovisioned
				r.recorder.Event(database, "Normal", EVENT_DEPROVISIONING_COMPLETED, "Database Server has been deprovisioned from NDB.")
				log.Info("Database server id was not found on the database CR, removing finalizers and deleting the CR.")
			}
			// remove our finalizer from the list and update it.
			log.Info("Removing Finalizer " + common.FINALIZER_DATABASE_SERVER)
			controllerutil.RemoveFinalizer(database, common.FINALIZER_DATABASE_SERVER)
			if err := r.Update(ctx, database); err != nil {
				return requeueOnErr(err)
			}
			log.Info("Removed Finalizer " + common.FINALIZER_DATABASE_SERVER)
			r.recorder.Event(database, "Normal", EVENT_CR_DELETED, "Database Custom Resource has been deleted from the k8s cluster")
			return requeue()
		}
	} else {
		// Both database instance and database server finalizers have been removed, no need to requeue
		// CR will be deleted.
		return doNotRequeue()
	}
	// Requeue the request while waiting for the database instance to be deleted from NDB.
	return requeueWithTimeout(15)
}

// The handleSync function synchronizes the database CR's with the database instance in NDB
// It handles the transition from EMPTY (initial state) => PROVISIONING => RUNNING
// and updates the status accordingly. The update() triggers an implicit requeue of the reconcile request.
func (r *DatabaseReconciler) handleSync(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient, req ctrl.Request, ndbServer *ndbv1alpha1.NDBServer) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleSync")

	databaseStatus := database.Status.DeepCopy()
	databaseStatus.Type = database.Spec.Instance.Type

	// Provision the database if it has not been provisioned earlier
	if databaseStatus.Status == "" || databaseStatus.Id == "" {
		// DB Status.Status is empty => Provision a DB
		infoStatement := "Provisioning a database instance on NDB."
		log.Info(infoStatement)

		dbPassword, sshPublicKey, err := r.getDatabaseInstanceCredentials(ctx, database.Spec.Instance.CredentialSecret, req.Namespace)
		if err != nil || dbPassword == "" || sshPublicKey == "" {
			var errStatement string
			if err == nil {
				errStatement = "Database instance password and ssh key cannot be empty"
				err = fmt.Errorf("empty DB instance credentials")
			} else {
				errStatement = "An error occured while fetching the DB Instance Secrets"
			}
			log.Error(err, errStatement)
			r.recorder.Eventf(database, "Warning", EVENT_INVALID_CREDENTIALS, "Error: %s", errStatement)
			return requeueOnErr(err)
		}

		reqData := map[string]interface{}{
			common.NDB_PARAM_PASSWORD:       dbPassword,
			common.NDB_PARAM_SSH_PUBLIC_KEY: sshPublicKey,
		}

		databaseAdapter := &controller_adapters.Database{Database: *database}
		generatedReq, err := ndb_api.GenerateProvisioningRequest(ctx, ndbClient, databaseAdapter, reqData)
		if err != nil {
			errStatement := "Could not generate database provisioning request"
			log.Error(err, errStatement)
			r.recorder.Eventf(database, "Warning", EVENT_REQUEST_GENERATION_FAILURE, "Error: %s. %s", errStatement, err.Error())
			return requeueOnErr(err)
		}
		r.recorder.Event(database, "Normal", EVENT_REQUEST_GENERATION, "Generated database provisiong request")

		taskResponse, err := ndb_api.ProvisionDatabase(ctx, ndbClient, generatedReq)
		if err != nil {
			errStatement := "Failed to make database provisioning request to NDB"
			log.Error(err, errStatement)
			r.recorder.Eventf(database, "Warning", EVENT_NDB_REQUEST_FAILED, "Error: %s. %s", errStatement, err.Error())
			return requeueOnErr(err)
		}
		// log.Info(fmt.Sprintf("Provisioning response from NDB: %+v", taskResponse))

		log.Info("Setting database CR status to provisioning and id as " + taskResponse.EntityId)

		databaseStatus.Status = common.DATABASE_CR_STATUS_PROVISIONING
		databaseStatus.Id = taskResponse.EntityId
		databaseStatus.ProvisioningOperationId = taskResponse.OperationId

		r.recorder.Event(database, "Normal", EVENT_PROVISIONING_STARTED, "Database provisioning initiated on NDB")

	}

	// Handle External Sync
	dbInfo := ndbServer.Status.Databases[databaseStatus.Id]

	// Not found in NDB CR
	if dbInfo == (ndbv1alpha1.NDBServerDatabaseInfo{}) {
		if databaseStatus.Status == common.DATABASE_CR_STATUS_PROVISIONING {
			log.Info("Database not found in NDB CR yet")
			r.recorder.Event(database, "Normal", EVENT_WAITING_FOR_NDB_RECONCILE, "Waiting for NDB server resource to reconcile")
			// NDB CR might not have been updated / reconciled yet OR
			// The operation to provision might have been aborted externally
			// So we will have to reconcile using operation Id and set the status accordingly
		} else {
			log.Info("Database might have been deleted externally")
			databaseStatus.Status = "EXTERNALLY DELETED"
			// It is not in provisioning state, which means it must have passed this
			// state in earlier reconciles and must have reached one of the next states.
			// The absence of the DB in the NDB CR indicates an external deletion
			// Change the status to indicate external deletion.
			// databaseStatus.Status = dbInfo.Status
		}
		databaseStatus.Status = "NOT FOUND ON NDB CR"
		// Using this as a placeholder until operation tracking is added
	} else {
		databaseStatus.Id = dbInfo.Id
		databaseStatus.IPAddress = dbInfo.IPAddress
		databaseStatus.Status = dbInfo.Status
		databaseStatus.DatabaseServerId = dbInfo.DBServerId
	}

	// Handle Internal Sync
	switch databaseStatus.Status {

	case common.DATABASE_CR_STATUS_READY:
		r.setupConnectivity(ctx, database, req)

	case "EXTERNALLY DELETED":
		log.Info("Externally deleted in internal sync")
	default:
		// Do Nothing
	}

	if !reflect.DeepEqual(database.Status, *databaseStatus) {
		database.Status = *databaseStatus
		err := r.Status().Update(ctx, database)
		if err != nil {
			errStatement := "Failed to update status of database custom resource"
			log.Error(err, errStatement)
			r.recorder.Eventf(database, "Warning", EVENT_CR_STATUS_UPDATE_FAILED, "Error: %s. %s.", err.Error())
			return requeueOnErr(err)
		}
	}

	return requeueWithTimeout(15)
}

// Sets up a kubernetes networking service (Without selectors)
// Then sets up an endpoint with the same name as the service
// to map to an external endpoint (NDB database instance in our scenario).
func (r *DatabaseReconciler) setupConnectivity(ctx context.Context, database *ndbv1alpha1.Database, req ctrl.Request) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.setupConnectivity")
	// The 'service' and 'endpoint' objects should have the
	// same name for the service to map to the enpoint.
	commonMetadata := metav1.ObjectMeta{
		Name:      database.Name + "-svc",
		Namespace: req.Namespace,
	}
	commonNamespacedName := types.NamespacedName{
		Name:      database.Name + "-svc",
		Namespace: req.Namespace,
	}
	targetPort := ndb_api.GetDatabasePortByType(database.Spec.Instance.Type)

	err = r.setupService(ctx, database, commonNamespacedName, commonMetadata, targetPort)
	if err != nil {
		errStatement := "Failed to setup kubernetes service for database custom resource"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_SERVICE_SETUP_FAILED, "Error: %s.", errStatement, err.Error())
		return
	}
	err = r.setupEndpoints(ctx, database, commonNamespacedName, commonMetadata, targetPort)
	if err != nil {
		errStatement := "Failed to setup kubernetes endpoints for database custom resource"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_ENDPOINT_SETUP_FAILED, "Error: %s. %s", errStatement, err.Error())
		return
	}
	log.Info("Returning from database_reconciler_helpers.setupConnectivity")
	return
}

// Checks and creates a new service (without label selectors) if it does not exists
// and also sets up the database as the owner for the created service
func (r *DatabaseReconciler) setupService(ctx context.Context, database *ndbv1alpha1.Database, namespacedName types.NamespacedName, metadata metav1.ObjectMeta, targetPort int32) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.setupService")
	// Create a new service if it does not exists.
	foundService := &corev1.Service{}
	err = r.Get(ctx, namespacedName, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("No service found, creating a new service", "target port", targetPort)
		service := &corev1.Service{
			ObjectMeta: metadata,
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Protocol:   corev1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: targetPort},
					},
				},
			},
		}
		// Setting database as the owner of this service
		err = ctrl.SetControllerReference(database, service, r.Scheme)
		if err != nil {
			log.Error(err, "Error setting controller reference for the service")
		}
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create a new service")
			return
		}
		log.Info("Created a new service", "service name", service.GetName())
	}
	log.Info("Returning from database_reconciler_helpers.setupService")
	return
}

// Checks and creates an endpoints object for the service if it does not already exists.
// If it is already present, syncs the IP address with the Database.Status.IPAddress if out of sync.
func (r *DatabaseReconciler) setupEndpoints(ctx context.Context, database *ndbv1alpha1.Database, namespacedName types.NamespacedName, metadata metav1.ObjectMeta, targetPort int32) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.setupEndpoints")
	foundEndpoint := &corev1.Endpoints{}
	endpointSubsets := []corev1.EndpointSubset{
		{
			Addresses: []corev1.EndpointAddress{{IP: database.Status.IPAddress}},
			Ports:     []corev1.EndpointPort{{Port: targetPort}},
		},
	}
	err = r.Get(ctx, namespacedName, foundEndpoint)
	// Create an endpoint if it does not exists.
	if err != nil && errors.IsNotFound(err) {
		log.Info("No endpoint found, creating a new endpoint")
		endpoint := &corev1.Endpoints{
			ObjectMeta: metadata,
			Subsets:    endpointSubsets,
		}
		// Setting database as the owner of this endpoint
		ctrl.SetControllerReference(database, endpoint, r.Scheme)
		err = r.Create(ctx, endpoint)
		if err != nil {
			log.Error(err, "Failed to create a new ep")
			return
		}
		log.Info("Created a new endpoint", "endpoint name", endpoint.GetName())
	} else {
		// If endpoint exists, check if the IP has changed.
		// If changed, sync with the latest IP in the database CR status.
		for _, subset := range foundEndpoint.Subsets {
			for _, address := range subset.Addresses {
				if address.IP == database.Status.IPAddress {
					// IP has not changed, no need to update endpoint
					return
				}
			}
		}
		log.Info("Endpoint found with a different IP address, updating.")
		foundEndpoint.Subsets = endpointSubsets
		err = r.Update(ctx, foundEndpoint)
		if err != nil {
			log.Error(err, "Failed to update endpoint")
			return
		}
	}
	log.Info("Returning from database_reconciler_helpers.setupEndpoints")
	return
}

// Returns the credentials(password and ssh public key) for NDB
// Returns an error if reading the secret containing credentials fails
func (r *DatabaseReconciler) getDatabaseInstanceCredentials(ctx context.Context, name, namespace string) (password, sshPublicKey string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.getDatabaseInstanceCredentials")
	secretDataMap, err := util.GetAllDataFromSecret(ctx, r.Client, name, namespace)
	if err != nil {
		log.Error(err, "Error occured in util.GetAllDataFromSecret while fetching all database instance secrets", "Secret Name", name, "Namespace", namespace)
		return
	}
	password = string(secretDataMap[common.SECRET_DATA_KEY_PASSWORD])
	sshPublicKey = string(secretDataMap[common.SECRET_DATA_KEY_SSH_PUBLIC_KEY])
	log.Info("Returning from database_reconciler_helpers.getDatabaseInstanceCredentials")
	return
}
