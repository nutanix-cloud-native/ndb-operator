package controllers

import (
	"context"
	"fmt"
	"strings"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/controller_adapters"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func getInstanceManager(database ndbv1alpha1.Database) (instanceManager InstanceManager) {
	if database.Spec.IsClone {
		instanceManager = &CloneManager{}
	} else {
		instanceManager = &DatabaseManager{}
	}
	return
}

type InstanceManager interface {
	create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (task *ndb_api.TaskInfoSummaryResponse, err error)
	deregister(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task *ndb_api.TaskInfoSummaryResponse, err error)
	// deleteDatabaseServer deprovisions the underlying VM(s) for a database.
	// Returns done=true when the caller should proceed to remove the finalizer,
	// or done=false when the operation is still in progress and the caller should requeue.
	deleteDatabaseServer(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (done bool, err error)
}

type DatabaseManager struct{}

type CloneManager struct{}

// resolveNamesToUUIDs is a helper function that resolves name fields to UUID fields
// This should be called before generating NDB API requests
func resolveNamesToUUIDs(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) error {
	log := ctrllog.FromContext(ctx)
	if err := controller_adapters.ResolveNamesToUUIDs(ctx, ndbClient, database); err != nil {
		errStatement := "Failed to resolve names to UUIDs"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_REQUEST_GENERATION_FAILURE, "Error: %s. %s", errStatement, err.Error())
		return err
	}
	return nil
}

func (dm *DatabaseManager) create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (taskResponse *ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Provisioning a database on NDB")
	dbPassword, sshPublicKey, dbUsername, err := r.getDatabaseCredentials(ctx, database.Spec.Instance.CredentialSecret, namespace)
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
		return
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD:       dbPassword,
		common.NDB_PARAM_SSH_PUBLIC_KEY: sshPublicKey,
		common.NDB_PARAM_USERNAME:       dbUsername,
	}

	// Resolve names to UUIDs before generating request
	if err = resolveNamesToUUIDs(ctx, r, ndbClient, database); err != nil {
		return
	}

	databaseAdapter := &controller_adapters.Database{Database: *database}
	generatedReq, err := ndb_api.GenerateProvisioningRequest(ctx, ndbClient, databaseAdapter, reqData)
	if err != nil {
		errStatement := "Could not generate database provisioning request"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_REQUEST_GENERATION_FAILURE, "Error: %s. %s", errStatement, err.Error())
		return
	}
	r.recorder.Event(database, "Normal", EVENT_REQUEST_GENERATION, "Generated database provisiong request")

	taskResponse, err = ndb_api.ProvisionDatabase(ctx, ndbClient, generatedReq)
	if err != nil {
		errStatement := "Failed to make database provisioning request to NDB"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_NDB_REQUEST_FAILED, "Error: %s. %s", errStatement, err.Error())
		return
	}
	return
}

func (dm *DatabaseManager) deregister(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task *ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	infoStatement := "Deregistering Database Instance from NDB."
	log.Info(infoStatement)
	r.recorder.Event(database, "Normal", EVENT_DEREGISTRATION_STARTED, infoStatement)
	task, err = ndb_api.DeprovisionDatabase(ctx, ndbClient, database.Status.Id, ndb_api.GenerateDeprovisionDatabaseRequest())
	if err != nil {
		errStatement := "Deregistering instance API call failed."
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED, "Error: %s. %s", errStatement, err.Error())
	}
	return
}

func (dm *DatabaseManager) deleteDatabaseServer(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (done bool, err error) {
	log := ctrllog.FromContext(ctx)

	if database.Spec.Instance != nil && database.Spec.Instance.HAConfig != nil {
		// HA path: two-phase start → poll to avoid re-firing a destructive API call
		// on requeue. See persistHADeletionOpIds for details.
		if len(database.Status.DBServerDeletionOperationIds) == 0 {
			// Start phase: fire DELETE /dpcs/{clusterId} and persist the operation IDs.
			// IMPORTANT: opIds must be persisted before this call returns.
			// If they are lost, the next reconcile will re-fire the DELETE against an
			// already-deleting DPC. persistHADeletionOpIds retries once on conflict
			// before giving up, making it robust against the common 409-Conflict case.
			opIds, apiErr := ndb_api.DeprovisionHADatabaseServers(ctx, ndbClient, database.Status.DatabaseServerId)
			if apiErr != nil {
				// If NDB reports the DB server is already gone (e.g. cluster manually deleted
				// via NDB UI), skip deprovision and signal done so the finalizer is removed.
				if strings.Contains(apiErr.Error(), "ERA-ENT-0000001") {
					log.Info("HA DB server already removed from NDB, skipping deprovision and removing finalizer " + common.FINALIZER_DATABASE_SERVER)
					r.recorder.Eventf(database, "Normal", EVENT_DEREGISTRATION_COMPLETED, "HA DB server was already removed from NDB; skipping deprovision.")
					return true, nil
				}
				log.Error(apiErr, "Failed to deprovision one or more HA database servers")
				r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED, "Error: %s", apiErr.Error())
				return false, apiErr
			}
			if persistErr := r.persistHADeletionOpIds(ctx, database, opIds); persistErr != nil {
				return false, persistErr
			}
			return false, nil // requeue to enter poll phase
		}

		// Poll phase: check every operation ID; signal done only when all are terminal.
		allDone := true
		for _, opId := range database.Status.DBServerDeletionOperationIds {
			op, opErr := ndb_api.GetOperationById(ctx, ndbClient, opId)
			if opErr != nil {
				log.Error(opErr, "Failed to fetch HA server deletion operation", "operationId", opId)
				allDone = false
				continue
			}
			switch ndb_api.GetOperationStatus(op) {
			case ndb_api.OPERATION_STATUS_PASSED:
				log.Info("HA server deletion operation completed", "operationId", opId)
			case ndb_api.OPERATION_STATUS_FAILED:
				// The operation is terminal — setting allDone=false would cause an infinite
				// requeue since NDB will never retry a FAILED operation. Treat it as done so
				// the finalizer is removed, but surface the failure loudly for manual cleanup.
				log.Error(fmt.Errorf("HA server deletion operation failed"), "operationId", opId, "message", op.Message)
				r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED,
					"HA server deletion operation %s failed: %s — manual cleanup of NDB resources may be required", opId, op.Message)
			default:
				allDone = false
			}
		}
		return allDone, nil
	}

	// Non-HA: fire-and-forget (mirrors the existing behaviour — errors are logged
	// and recorded inside deleteDatabaseServer; the caller proceeds immediately).
	deleteDatabaseServer(ctx, r, ndbClient, database)
	return true, nil
}

func (cm *CloneManager) create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (taskResponse *ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Cloning a database on NDB")

	// Resolve names to UUIDs before generating request
	if err = resolveNamesToUUIDs(ctx, r, ndbClient, database); err != nil {
		return
	}

	databaseAdapter := &controller_adapters.Database{Database: *database}
	dbPassword, sshPublicKey, dbUsername, err := r.getDatabaseCredentials(ctx, databaseAdapter.GetCredentialSecret(), namespace)
	if err != nil || dbPassword == "" || sshPublicKey == "" {
		var errStatement string
		if err == nil {
			errStatement = "Database clone password and ssh key cannot be empty"
			err = fmt.Errorf("empty DB clone credentials")
		} else {
			errStatement = "An error occured while fetching the DB clone Secrets"
		}
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_INVALID_CREDENTIALS, "Error: %s", errStatement)
		return
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD:       dbPassword,
		common.NDB_PARAM_SSH_PUBLIC_KEY: sshPublicKey,
		common.NDB_PARAM_USERNAME:       dbUsername,
	}

	generatedReq, err := ndb_api.GenerateCloningRequest(ctx, ndbClient, databaseAdapter, reqData)
	if err != nil {
		errStatement := "Could not generate database cloning request"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_REQUEST_GENERATION_FAILURE, "Error: %s. %s", errStatement, err.Error())
		return
	}
	r.recorder.Event(database, "Normal", EVENT_REQUEST_GENERATION, "Generated database cloning request")

	taskResponse, err = ndb_api.ProvisionClone(ctx, ndbClient, generatedReq)
	if err != nil {
		errStatement := "Failed to make database cloning request to NDB"
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_NDB_REQUEST_FAILED, "Error: %s. %s", errStatement, err.Error())
		return
	}
	return
}

func (cm *CloneManager) deregister(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task *ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	infoStatement := "Deregistering Clone Instance from NDB."
	log.Info(infoStatement)
	r.recorder.Event(database, "Normal", EVENT_DEREGISTRATION_STARTED, infoStatement)
	task, err = ndb_api.DeprovisionClone(ctx, ndbClient, database.Status.Id, ndb_api.GenerateDeprovisionCloneRequest())
	if err != nil {
		errStatement := "Deregistering instance API call failed."
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED, "Error: %s. %s", errStatement, err.Error())
	}
	return
}

func (cm *CloneManager) deleteDatabaseServer(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (done bool, err error) {
	// Clones are never HA, so this is always fire-and-forget.
	deleteDatabaseServer(ctx, r, ndbClient, database)
	return true, nil
}

func deleteDatabaseServer(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task *ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	databaseServerId := database.Status.DatabaseServerId
	// Make a dbserver deprovisioning request to NDB only if the serverId is present in status
	if databaseServerId != "" {
		r.recorder.Eventf(database, "Normal", EVENT_DEREGISTRATION_STARTED, "Deprovisioning database server from NDB.")
		task, err = ndb_api.DeprovisionDatabaseServer(ctx, ndbClient, databaseServerId, ndb_api.GenerateDeprovisionDatabaseServerRequest())
		if err != nil {
			errStament := fmt.Sprintf("Deprovisioning database server request failed for id: %s", databaseServerId)
			log.Error(err, errStament)
			r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED, "Error: %s. %s", errStament, err.Error())
			return
		}
	} else {
		// Database and server has been deprovisioned
		r.recorder.Event(database, "Normal", EVENT_DEREGISTRATION_COMPLETED, "Database Server has been deprovisioned from NDB.")
		log.Info("Database server id was not found on the database CR, removing finalizers and deleting the CR.")
	}
	return
}
