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
	"strings"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
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

// updateStatusWithRetry applies applyFn to database and persists the status subresource,
// retrying exactly once on failure. On the retry it re-fetches the latest object first
// (resolving the 409 Conflict / stale resourceVersion that causes most update failures),
// then calls applyFn again before the second attempt.
// applyFn must be idempotent — it will be called before each attempt.
// If both attempts fail, criticalMsg is logged as an error before returning.
func (r *DatabaseReconciler) updateStatusWithRetry(ctx context.Context, database *ndbv1alpha1.Database, applyFn func(), criticalMsg string) error {
	log := ctrllog.FromContext(ctx)
	applyFn()
	if err := r.Status().Update(ctx, database); err == nil {
		return nil
	}
	log.Info("Status update conflict, re-fetching and retrying once")
	// r.Get overwrites *database in place with the latest version from the API server,
	// giving us a fresh resourceVersion but also discarding the changes applyFn made above.
	// applyFn must be called again to re-apply those changes on top of the refreshed object.
	if fetchErr := r.Get(ctx, types.NamespacedName{Name: database.Name, Namespace: database.Namespace}, database); fetchErr != nil {
		return fetchErr
	}
	applyFn()
	if retryErr := r.Status().Update(ctx, database); retryErr != nil {
		log.Error(retryErr, criticalMsg)
		return retryErr
	}
	return nil
}

// persistHADeletionOpIds persists the HA DPC deletion operation IDs into status.
// Uses updateStatusWithRetry because these IDs must survive across reconcile cycles;
// losing them would cause the next reconcile to re-fire the destructive DELETE /dpcs call.
func (r *DatabaseReconciler) persistHADeletionOpIds(ctx context.Context, database *ndbv1alpha1.Database, opIds []string) error {
	return r.updateStatusWithRetry(ctx, database,
		func() { database.Status.DBServerDeletionOperationIds = opIds },
		"CRITICAL: HA DPC deletion was fired but operation IDs could not be persisted after retry; next reconcile may re-fire the delete",
	)
}

// persistDeregistrationOpId persists the deregistration operation ID into status.
// Uses updateStatusWithRetry because losing the ID would cause the next reconcile to
// re-fire the deregistration call against an already-deregistering database instance.
func (r *DatabaseReconciler) persistDeregistrationOpId(ctx context.Context, database *ndbv1alpha1.Database, opId string) error {
	return r.updateStatusWithRetry(ctx, database,
		func() { database.Status.DeregistrationOperationId = opId },
		"CRITICAL: deregistration was triggered but the operation ID could not be persisted after retry; next reconcile may re-fire deregistration",
	)
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
	return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
}

// handleDelete function handles the deletion of
//
//	a. Database instance
//	b. Database server
func (r *DatabaseReconciler) handleDelete(ctx context.Context, req ctrl.Request, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Database CR is being deleted")
	instanceManager := getInstanceManager(*database)
	if controllerutil.ContainsFinalizer(database, common.FINALIZER_INSTANCE) {
		// Check if the deregistration operation id (database.Status.DeregistrationOperationId) is empty
		// If so, then make a deprovisionDatabase API call to NDB
		// else proceed check for the operation completion before removing finalizer.
		deregistrationOperationId := database.Status.DeregistrationOperationId
		if deregistrationOperationId == "" {
			deregistrationOp, err := instanceManager.deregister(ctx, r, ndbClient, database)
			if err != nil {
				// Not logging here, already done in the deregister function
				return requeueOnErr(err)
			}
			if err := r.persistDeregistrationOpId(ctx, database, deregistrationOp.OperationId); err != nil {
				return requeueOnErr(err)
			}
		} else {
			deregistrationOp, err := ndb_api.GetOperationById(ctx, ndbClient, deregistrationOperationId)
			if err != nil {
				message := fmt.Sprintf("NDB API to fetch operation by id failed. OperationId: %s:, error: %s", deregistrationOperationId, err.Error())
				r.recorder.Event(database, "Warning", EVENT_NDB_REQUEST_FAILED, message)
			} else {
				switch ndb_api.GetOperationStatus(deregistrationOp) {
				case ndb_api.OPERATION_STATUS_FAILED:
					err := fmt.Errorf("deregistration operation terminated. status: %s, message: %s, operationId: %s", deregistrationOp.Status, deregistrationOp.Message, deregistrationOperationId)
					log.Error(err, "Deregistration Failed")
					r.recorder.Event(database, "Warning", "OPERATION FAILED", "Database creation operation failed with error: "+err.Error())
				case ndb_api.OPERATION_STATUS_PASSED:
					r.recorder.Eventf(database, "Normal", EVENT_DEREGISTRATION_COMPLETED, "Database deprovisioned from NDB.")
					log.Info("Removing Finalizer " + common.FINALIZER_INSTANCE)
					controllerutil.RemoveFinalizer(database, common.FINALIZER_INSTANCE)
					if err := r.Update(ctx, database); err != nil {
						return requeueOnErr(err)
					}
					log.Info("Removed Finalizer " + common.FINALIZER_INSTANCE)
				default:
					// Do nothing, we do not care about other statuses
				}
			}
		}

	} else if controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_SERVER) {
		done, err := instanceManager.deleteDatabaseServer(ctx, r, ndbClient, database)
		if err != nil {
			return requeueOnErr(err)
		}
		if !done {
			return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
		}

		// deleteDatabaseServer signalled done: remove finalizer.
		log.Info("Removing Finalizer " + common.FINALIZER_DATABASE_SERVER)
		controllerutil.RemoveFinalizer(database, common.FINALIZER_DATABASE_SERVER)
		if err := r.Update(ctx, database); err != nil {
			return requeueOnErr(err)
		}
		log.Info("Removed Finalizer " + common.FINALIZER_DATABASE_SERVER)
		r.recorder.Event(database, "Normal", EVENT_CR_DELETED, "Database Custom Resource has been deleted from the k8s cluster")
		return requeue()

	} else {
		// Both database instance and database server finalizers have been removed, no need to requeue
		// CR will be deleted.
		return doNotRequeue()
	}
	// Requeue the request while waiting for the database instance to be deleted from NDB.
	return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
}

// The handleSync function synchronizes the database CR with the database info object in the
// NDBServer CR (which fetches it from NDB). It handles the transition from EMPTY (initial state) => WAITING => PROVISIONING => RUNNING
// and updates the status accordingly. The update() triggers an implicit requeue of the reconcile request.
func (r *DatabaseReconciler) handleSync(ctx context.Context, database *ndbv1alpha1.Database, ndbClient *ndb_client.NDBClient, req ctrl.Request, ndbServer *ndbv1alpha1.NDBServer) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleSync")

	databaseStatus := database.Status.DeepCopy()

	instanceManager := getInstanceManager(*database)

	// Provision the database if it has not been provisioned earlier
	if databaseStatus.Status == "" && databaseStatus.Id == "" {
		// DB Status.Status is empty => Provision a DB
		taskResponse, err := instanceManager.create(ctx, r, ndbClient, database, req.Namespace)
		if err != nil {
			errStatement := "Failed to create database on NDB"
			log.Error(err, errStatement)
			r.recorder.Eventf(database, "Warning", EVENT_NDB_REQUEST_FAILED, "Error: %s. %s", errStatement, err.Error())
			return requeueOnErr(err)
		}
		log.Info(fmt.Sprintf("Updating Database CR to Status: CREATING, id: %s and creationOperationId: %s", taskResponse.EntityId, taskResponse.OperationId))

		databaseStatus.Status = common.DATABASE_CR_STATUS_CREATING
		databaseStatus.Id = taskResponse.EntityId
		databaseStatus.CreationOperationId = taskResponse.OperationId
		r.recorder.Event(database, "Normal", EVENT_CREATION_STARTED, "Database creation initiated on NDB")
	}

	// Handle External Sync
	dbInfo := ndbServer.Status.Databases[databaseStatus.Id]
	isUnderDeletion := !database.ObjectMeta.DeletionTimestamp.IsZero()
	if isUnderDeletion {
		databaseStatus.Status = common.DATABASE_CR_STATUS_DELETING
	} else if databaseStatus.Status == common.DATABASE_CR_STATUS_CREATING {
		creationOp, err := ndb_api.GetOperationById(ctx, ndbClient, databaseStatus.CreationOperationId)
		if err != nil {
			message := fmt.Sprintf("NDB API to fetch operation by id failed. OperationId: %s:, error: %s", creationOp.Id, err.Error())
			r.recorder.Event(database, "Warning", EVENT_NDB_REQUEST_FAILED, message)
		} else {
			switch ndb_api.GetOperationStatus(creationOp) {
			case ndb_api.OPERATION_STATUS_FAILED:
				databaseStatus.Status = common.DATABASE_CR_STATUS_CREATION_ERROR
				err = fmt.Errorf("creation operation terminated. status: %s, message: %s, operationId: %s", creationOp.Status, creationOp.Message, creationOp.Id)
				log.Error(err, "Database Creation Failed")
				r.recorder.Event(database, "Warning", EVENT_CREATION_FAILED, "Database creation operation failed with error: "+err.Error())
			case ndb_api.OPERATION_STATUS_PASSED:
				databaseStatus.Status = common.DATABASE_CR_STATUS_READY
				r.recorder.Event(database, "Normal", EVENT_CREATION_COMPLETED, "Database creation operation passed")
			default:
				// Do nothing, we do not care about other statuses
			}
		}
	} else if dbInfo.Id != "" {
		databaseStatus.Status = dbInfo.Status
		databaseStatus.Id = dbInfo.Id
		databaseStatus.IPAddress = dbInfo.IPAddress
		databaseStatus.DatabaseServerId = dbInfo.DBServerId
		databaseStatus.Type = ndb_api.GetDatabaseTypeFromEngine(dbInfo.Type)
	} else {
		log.Info("Database missing from NDB CR")
		databaseStatus.Status = common.DATABASE_CR_STATUS_NOT_FOUND
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

	// Handle Internal Sync -
	// [READY]
	// Add finalizers only when the database is in ready state so that if
	// any failure occurrs before reaching the ready state, the failure
	// would not cause the deletion to block the terminal.
	// Also, setup and create network services.
	// [DELETING]
	// Delete the database instance and the VM as per the finalizers
	// [NOT FOUND]
	// Record an event and then do not requeue since the resource has been deleted externally
	// or was not found on NDB
	switch databaseStatus.Status {
	case common.DATABASE_CR_STATUS_READY:
		if !isUnderDeletion {
			if !controllerutil.ContainsFinalizer(database, common.FINALIZER_INSTANCE) {
				return r.addFinalizer(ctx, req, common.FINALIZER_INSTANCE, database)
			}
			if !controllerutil.ContainsFinalizer(database, common.FINALIZER_DATABASE_SERVER) {
				return r.addFinalizer(ctx, req, common.FINALIZER_DATABASE_SERVER, database)
			}
		}
		if databaseStatus.IPAddress != "" {
			r.setupConnectivity(ctx, database)
		} else {
			// The database is in "READY" state on NDB, but the API responses sometimes do not have
			// an IP address in the response right after reaching the READY state. We only setup connectivity
			// once we have a non-empty IP Address. Just logging and raising an event to notify the user.
			message := fmt.Sprintf("Empty IP Address for Database %s, will setup connectivity once the IP address is assigned", database.Name)
			log.Info(message)
			r.recorder.Event(database, "Warning", EVENT_WAITING_FOR_IP_ADDRESS, message)
		}
	case common.DATABASE_CR_STATUS_DELETING:
		return r.handleDelete(ctx, req, database, ndbClient)
	case common.DATABASE_CR_STATUS_NOT_FOUND:
		r.recorder.Eventf(database, "Warning", EVENT_EXTERNAL_DELETE, "Error: Resource not found on NDB")
	case common.DATABASE_CR_STATUS_CREATION_ERROR:
		return doNotRequeue()
	default:
		// No-Op
	}

	return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
}

// Sets up a kubernetes networking service (Without selectors)
// Then sets up an endpoint with the same name as the service
// to map to an external endpoint (NDB database instance in our scenario).
// Every database gets a primary -svc. HA databases additionally get engine-specific
// extra services (e.g. -ro-svc for Postgres) via the HAConnectivityManager registry.
func (r *DatabaseReconciler) setupConnectivity(ctx context.Context, database *ndbv1alpha1.Database) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.setupConnectivity")

	ns := database.Namespace

	// Determine the port for the primary -svc and whether an HA manager is needed.
	// HA databases use the engine-specific write port; non-HA use the standard DB port.
	primaryPort := ndb_api.GetDatabasePortByType(database.Status.Type)
	var haManager HAConnectivityManager
	if database.Spec.Instance != nil && database.Spec.Instance.HAConfig != nil {
		haManager = haConnectivityManagers[database.Spec.Instance.Type]
		primaryPort = haManager.PrimaryPort(database.Spec.Instance.HAConfig)
	}

	// All databases get a primary -svc.
	if err = r.setupServiceAndEndpoints(ctx, database, database.Name+"-svc", ns, primaryPort); err != nil {
		return
	}

	// HA databases additionally get engine-specific extra services (e.g. -ro-svc for Postgres).
	if haManager != nil {
		for _, svcSpec := range haManager.AdditionalServices(database.Spec.Instance.HAConfig) {
			if err = r.setupServiceAndEndpoints(ctx, database, database.Name+svcSpec.NameSuffix, ns, svcSpec.Port); err != nil {
				return
			}
		}
	}

	log.Info("Returning from database_reconciler_helpers.setupConnectivity")
	return
}

// setupServiceAndEndpoints creates (or reconciles) a single Kubernetes Service and its
// matching Endpoints object for the given database. It is the unit of work called for both
// the primary -svc and each additional HA service (e.g. -ro-svc).
func (r *DatabaseReconciler) setupServiceAndEndpoints(ctx context.Context, database *ndbv1alpha1.Database, name, namespace string, port int32) error {
	log := ctrllog.FromContext(ctx)
	nn := types.NamespacedName{Name: name, Namespace: namespace}
	meta := metav1.ObjectMeta{Name: name, Namespace: namespace}

	if err := r.setupService(ctx, database, nn, meta, port); err != nil {
		errStatement := "Failed to setup kubernetes service for database"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_SERVICE_SETUP_FAILED, "Error: %s. %s", errStatement, err.Error())
		return err
	}
	if err := r.setupEndpoints(ctx, database, nn, meta, port); err != nil {
		errStatement := "Failed to setup kubernetes endpoints for database"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_ENDPOINT_SETUP_FAILED, "Error: %s. %s", errStatement, err.Error())
		return err
	}
	return nil
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

// buildEndpointAddresses returns the list of endpoint addresses for the database.
// For HA databases ipAddress holds comma-separated HAProxy IPs; each is returned
// as a separate EndpointAddress so kube-proxy load-balances across all of them.
// For single-instance databases the single ipAddress is used directly.
func buildEndpointAddresses(database *ndbv1alpha1.Database) []corev1.EndpointAddress {
	ips := strings.Split(database.Status.IPAddress, ",")
	addrs := make([]corev1.EndpointAddress, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			addrs = append(addrs, corev1.EndpointAddress{IP: ip})
		}
	}
	return addrs
}

// endpointAddressesInSync returns true when the existing endpoint subsets already
// contain exactly the addresses we want (order-independent).
func endpointAddressesInSync(existing []corev1.EndpointSubset, desired []corev1.EndpointAddress) bool {
	existingSet := make(map[string]struct{})
	for _, subset := range existing {
		for _, addr := range subset.Addresses {
			existingSet[addr.IP] = struct{}{}
		}
	}
	if len(existingSet) != len(desired) {
		return false
	}
	for _, addr := range desired {
		if _, ok := existingSet[addr.IP]; !ok {
			return false
		}
	}
	return true
}

// Checks and creates an endpoints object for the service if it does not already exist.
// If it is already present, syncs addresses with the Database.Status IP(s) if out of sync.
// For HA databases all HAProxy IPs are registered; for single-instance the single IPAddress is used.
func (r *DatabaseReconciler) setupEndpoints(ctx context.Context, database *ndbv1alpha1.Database, namespacedName types.NamespacedName, metadata metav1.ObjectMeta, targetPort int32) (err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.setupEndpoints")

	desiredAddresses := buildEndpointAddresses(database)
	endpointSubsets := []corev1.EndpointSubset{
		{
			Addresses: desiredAddresses,
			Ports:     []corev1.EndpointPort{{Port: targetPort}},
		},
	}

	foundEndpoint := &corev1.Endpoints{}
	err = r.Get(ctx, namespacedName, foundEndpoint)
	// Create an endpoint if it does not exist.
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
		// If endpoint exists, check if the addresses have changed.
		// If changed, sync with the latest IPs from the database CR status.
		if endpointAddressesInSync(foundEndpoint.Subsets, desiredAddresses) {
			return
		}
		log.Info("Endpoint found with different addresses, updating.")
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
func (r *DatabaseReconciler) getDatabaseCredentials(ctx context.Context, name, namespace string) (password, sshPublicKey string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.getDatabaseCredentials")
	secretDataMap, err := util.GetAllDataFromSecret(ctx, r.Client, name, namespace)
	if err != nil {
		log.Error(err, "Error occured in util.GetAllDataFromSecret while fetching all database instance secrets", "Secret Name", name, "Namespace", namespace)
		return
	}
	password = string(secretDataMap[common.SECRET_DATA_KEY_PASSWORD])
	sshPublicKey = string(secretDataMap[common.SECRET_DATA_KEY_SSH_PUBLIC_KEY])
	log.Info("Returning from database_reconciler_helpers.getDatabaseCredentials")
	return
}
