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
	"math"
	"strings"
	"time"

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
//
//	a. Database instance
//	b. Database server
func (r *DatabaseReconciler) handleDelete(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Database CR is being deleted.")
	if controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_INSTANCE) {
		// Check if the database instance id (database.Status.Id) is present in the status
		// If present, then make a deprovisionDatabase API call to NDB
		// else proceed with removing finalizer as database instance provisioning wasn't successful earlier.
		if database.Status.Id != "" {
			log.Info("Deprovisioning database instance from NDB.")
			_, err := ndb_api.DeprovisionDatabase(ctx, ndbClient, database.Status.Id, *ndb_api.GenerateDeprovisionDatabaseRequest())
			if err != nil {
				log.Error(err, "Deprovisioning database instance request failed.")
				return r.requeueOnErr(err)
			}
		}
		log.Info("Removing Finalizer " + common.FINALIZER_DATABASE_INSTANCE)
		controllerutil.RemoveFinalizer(database, common.FINALIZER_DATABASE_INSTANCE)
		if err := r.Update(ctx, database); err != nil {
			return r.requeueOnErr(err)
		}
		log.Info("Removed Finalizer " + common.FINALIZER_DATABASE_INSTANCE)

	} else if controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_SERVER) {
		// Checking if the database instance still exists in NDB. (It might take some time for the delete db instance operation to complete)
		// Proceed to delete the database server vm only after the database instance has been deleted.
		log.Info("Checking if database instance exists")
		allDatabases, err := ndb_api.GetAllDatabases(ctx, ndbClient)
		if err != nil {
			log.Error(err, "An error occurred while trying to get all databases")
			return r.requeueOnErr(err)
		}
		if len(util.Filter(allDatabases, func(d ndb_api.DatabaseResponse) bool { return d.Id == database.Status.Id })) == 0 {
			// Could not find the database with the given database id => database instance has been deleted
			log.Info("Database instance not found, attempting to remove database server.")
			databaseServerId := database.Status.DatabaseServerId
			// Make a dbserver deprovisioning request to NDB only if the serverId is present in status
			if databaseServerId != "" {
				_, err := ndb_api.DeprovisionDatabaseServer(ctx, ndbClient, databaseServerId, *ndb_api.GenerateDeprovisionDatabaseServerRequest())
				if err != nil {
					log.Error(err, "Deprovisioning database server request failed.", "database server id", databaseServerId)
					return r.requeueOnErr(err)
				}
			} else {
				log.Info("Database server id was not found on the database CR, removing finalizers and deleting the CR.")
			}
			// remove our finalizer from the list and update it.
			log.Info("Removing Finalizer " + common.FINALIZER_DATABASE_SERVER)
			controllerutil.RemoveFinalizer(database, common.FINALIZER_DATABASE_SERVER)
			if err := r.Update(ctx, database); err != nil {
				return r.requeueOnErr(err)
			}
			log.Info("Removed Finalizer " + common.FINALIZER_DATABASE_SERVER)
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

// When a database is provisioned, an id is assigned to the database CR.
// In case if someone deletes the database externally/aborts the operation through NDB (and not through the operator), the operator should
// create a new database. To do this, we fetch the database by the Id we have in the datbase CR's Status.Id.
// If the database response's status is empty, we set our CR's status to be empty so that it is provisioned again.
func (r *DatabaseReconciler) handleExternalDelete(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleExternalDelete")
	if database.Status.Id != "" {
		// database was provisioned earlier => sync status
		var databaseResponse ndb_api.DatabaseResponse
		allDatabases, err := ndb_api.GetAllDatabases(ctx, ndbClient)
		if err != nil {
			log.Error(err, "An error occurred while trying to get all databases")
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
		if databaseResponse.Status == common.DATABASE_CR_STATUS_EMPTY {
			log.Info("The database might have been deleted externally, setting an empty status so it can be re-provisioned.")
			database.Status.Status = common.DATABASE_CR_STATUS_EMPTY
			err = r.Status().Update(ctx, database)
			if err != nil {
				log.Error(err, "Failed to update database status")
				return err
			}
		}

	}
	return nil
}

type ProvisionDB struct{}
type CloneDB struct{}

type DatabaseCreator interface {
	CreateDatabase(ctx context.Context, database *ndbv1alpha1.Database,
		ndbClient *ndb_client.NDBClient, r *DatabaseReconciler, req ctrl.Request) (ctrl.Result, error)
}

func GetDatabaseCreator(database *ndbv1alpha1.Database) DatabaseCreator {
	if strings.Compare(database.Spec.ProvisionType, "clone") == 0 {
		return &CloneDB{}
	} else {
		return &ProvisionDB{}
	}
}

// Clone a new DB implementation
func (p *CloneDB) CreateDatabase(ctx context.Context, database *ndbv1alpha1.Database,
	ndbClient *ndb_client.NDBClient, r *DatabaseReconciler, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("in Clone Database operation...")

	dbPassword, sshPublicKey, err := r.getDatabaseInstanceCredentials(ctx, database.Spec.Instance.CredentialSecret, req.Namespace)
	if err != nil || dbPassword == "" || sshPublicKey == "" {
		var errStatement string
		if err == nil {
			errStatement = "Clone Database instance password and ssh key cannot be empty"
			err = fmt.Errorf("empty Clone DB instance credentials")
		} else {
			errStatement = "An error occured while fetching the Clone DB Instance Secrets"
		}
		log.Error(err, errStatement)
		return r.requeueOnErr(err)
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD:        dbPassword,
		common.NDB_PARAM_SSH_PUBLIC_KEY:  sshPublicKey,
		common.NDB_PARAM_TIME_MACHINE_ID: database.Spec.Clone.TimeMachineId,
		common.NDB_PARAM_SNAPSHOT_ID:     database.Spec.Clone.SnapshotId,
		common.NDB_PARAM_DB_PASSWORD:     database.Spec.Clone.DBPassword,
	}

	databaseAdapter := &controller_adapters.Database{Database: *database}
	// change it to clone request
	generatedReq, err := ndb_api.GenerateCloningRequest(ctx, ndbClient, databaseAdapter, reqData)
	log.Info("Clone Request Body", generatedReq)
	if err != nil {
		log.Error(err, "Could not generate cloning request, requeuing.")
		return r.requeueOnErr(err)
	}

	// send the clone request
	taskResponse, err := ndb_api.CloneDatabase(ctx, ndbClient, generatedReq)
	if err != nil {
		log.Error(err, "An error occurred while trying to provision the database")
		return r.requeueOnErr(err)
	}

	log.Info("Setting Clone database CR status to provisioning and id as " + taskResponse.EntityId)
	database.Status.Status = common.DATABASE_CR_STATUS_PROVISIONING
	database.Status.Id = taskResponse.EntityId

	// Updating the type in the Database Status based on the input
	database.Status.Type = database.Spec.Instance.Type
	database.Status.ProvisionType = database.Spec.ProvisionType

	err = r.Status().Update(ctx, database)
	if err != nil {
		log.Error(err, "Failed to update clone database status")
		return r.requeueOnErr(err)
	}

	return r.doNotRequeue()
}

// Provision a new DB implementation
func (p *ProvisionDB) CreateDatabase(ctx context.Context, database *ndbv1alpha1.Database,
	ndbClient *ndb_client.NDBClient, r *DatabaseReconciler, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("in Provision DB operation...")
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
		return r.requeueOnErr(err)
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD:       dbPassword,
		common.NDB_PARAM_SSH_PUBLIC_KEY: sshPublicKey,
	}

	databaseAdapter := &controller_adapters.Database{Database: *database}
	generatedReq, err := ndb_api.GenerateProvisioningRequest(ctx, ndbClient, databaseAdapter, reqData)
	log.Info("Provision Request Body", generatedReq)
	if err != nil {
		log.Error(err, "Could not generate provisioning request, requeuing.")
		return r.requeueOnErr(err)
	}

	taskResponse, err := ndb_api.ProvisionDatabase(ctx, ndbClient, generatedReq)
	if err != nil {
		log.Error(err, "An error occurred while trying to provision the database")
		return r.requeueOnErr(err)
	}
	// log.Info(fmt.Sprintf("Provisioning response from NDB: %+v", taskResponse))

	log.Info("Setting database CR status to provisioning and id as " + taskResponse.EntityId)
	database.Status.Status = common.DATABASE_CR_STATUS_PROVISIONING
	database.Status.Id = taskResponse.EntityId

	// Updating the type in the Database Status based on the input
	database.Status.Type = database.Spec.Instance.Type
	database.Status.ProvisionType = database.Spec.ProvisionType

	err = r.Status().Update(ctx, database)
	if err != nil {
		log.Error(err, "Failed to update database status")
		return r.requeueOnErr(err)
	}
	// If database instance is not yet ready, requeue with wait
	return r.requeueWithTimeout(15)
}

// The handleSync function synchronizes the database CR's with the database instance in NDB
// It handles the transition from EMPTY (initial state) => PROVISIONING => RUNNING
// and updates the status accordingly. The update() triggers an implicit requeue of the reconcile request.
func (r *DatabaseReconciler) handleSync(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleSync")

	switch database.Status.Status {

	case common.DATABASE_CR_STATUS_EMPTY:
		// DB/Clone Status.Status is empty => Provision a DB/Clone
		log.Info("Provisioning a database instance with NDB.")
		dbCreator := GetDatabaseCreator(database)
		dbCreator.CreateDatabase(ctx, database, ndbClient, r, req)

	case common.DATABASE_CR_STATUS_PROVISIONING:
		// Check the status of the DB
		databaseResponse, err := ndb_api.GetDatabaseById(ctx, ndbClient, database.Status.Id)
		if err != nil {
			log.Error(err, "An error occurred while trying to get the database with id: "+database.Status.Id)
			r.requeueOnErr(err)
		}

		// if READY => Change status
		// log.Info("DEBUG Database Response: " + util.ToString(databaseResponse))
		if databaseResponse.Status == common.DATABASE_CR_STATUS_READY {
			log.Info("Clone instance is READY, adding data to CR's status and updating the CR")
			database.Status.Status = common.DATABASE_CR_STATUS_READY
			database.Status.DatabaseServerId = databaseResponse.DatabaseNodes[0].DatabaseServerId
			database.Status.IPAddress = databaseResponse.DatabaseNodes[0].DbServer.IPAddresses[0]
			if database.Status.IPAddress != "" {
				err = r.Status().Update(ctx, database)
				if err != nil {
					log.Error(err, "Failed to update database status")
					return r.requeueOnErr(err)
				}
			}
		}
		// If database instance is not yet ready, requeue with wait
		return r.requeueWithTimeout(15)

	case common.DATABASE_CR_STATUS_READY:
		r.setupConnectivity(ctx, database, req)
		return r.requeueWithTimeout(15)

	default:
		// Do Nothing
	}

	return r.doNotRequeue()
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
		log.Error(err, "Error occurred while setting up the service")
		return
	}
	err = r.setupEndpoints(ctx, database, commonNamespacedName, commonMetadata, targetPort)
	if err != nil {
		log.Error(err, "Error occurred while setting up the endpoints")
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

// Returns the credentials(username, password and caCertificate) for NDB
// Returns an error if reading the secret containing credentials fails
func (r *DatabaseReconciler) getNDBCredentials(ctx context.Context, name, namespace string) (username, password, caCert string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.getNDBCredentials")
	secretDataMap, err := util.GetAllDataFromSecret(ctx, r.Client, name, namespace)
	if err != nil {
		log.Error(err, "Error occured in util.GetAllDataFromSecret while fetching all NDB secrets", "Secret Name", name, "Namespace", namespace)
		return
	}
	username = string(secretDataMap[common.SECRET_DATA_KEY_USERNAME])
	password = string(secretDataMap[common.SECRET_DATA_KEY_PASSWORD])
	caCert = string(secretDataMap[common.SECRET_DATA_KEY_CA_CERTIFICATE])
	log.Info("Returning from database_reconciler_helpers.getNDBCredentials")
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
