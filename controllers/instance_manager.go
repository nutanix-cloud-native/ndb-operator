package controllers

import (
	"context"
	"fmt"

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
	create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (task ndb_api.TaskInfoSummaryResponse, err error)
	deregister(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error)
	deleteVM(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error)
}

type DatabaseManager struct{}

type CloneManager struct{}

func (dm *DatabaseManager) create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (taskResponse ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Provisioning a database on NDB")
	dbPassword, sshPublicKey, err := r.getDatabaseCredentials(ctx, database.Spec.Instance.CredentialSecret, namespace)
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

func (dm *DatabaseManager) deregister(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	infoStatement := "Deregistering Database Instance from NDB."
	log.Info(infoStatement)
	r.recorder.Event(database, "Normal", EVENT_DEREGISTRATION_STARTED, infoStatement)
	task, err = ndb_api.DeprovisionDatabase(ctx, ndbClient, database.Status.Id, *ndb_api.GenerateDeprovisionDatabaseRequest())
	if err != nil {
		errStatement := "Deregistering instance API call failed."
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED, "Error: %s. %s", errStatement, err.Error())
	}
	return
}

func (dm *DatabaseManager) deleteVM(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error) {
	return deleteVM(ctx, r, ndbClient, database)
}

func (cm *CloneManager) create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (taskResponse ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Cloning a database on NDB")
	databaseAdapter := &controller_adapters.Database{Database: *database}
	dbPassword, sshPublicKey, err := r.getDatabaseCredentials(ctx, databaseAdapter.GetCredentialSecret(), namespace)
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

func (cm *CloneManager) deregister(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	infoStatement := "Deregistering Clone Instance from NDB."
	log.Info(infoStatement)
	r.recorder.Event(database, "Normal", EVENT_DEREGISTRATION_STARTED, infoStatement)
	task, err = ndb_api.DeprovisionClone(ctx, ndbClient, database.Status.Id, *ndb_api.GenerateDeprovisionCloneRequest())
	if err != nil {
		errStatement := "Deregistering instance API call failed."
		log.Error(err, errStatement)
		r.recorder.Eventf(database, "Warning", EVENT_DEREGISTRATION_FAILED, "Error: %s. %s", errStatement, err.Error())
	}
	return
}

func (cm *CloneManager) deleteVM(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error) {
	return deleteVM(ctx, r, ndbClient, database)
}

func deleteVM(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) (task ndb_api.TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	databaseServerId := database.Status.DatabaseServerId
	// Make a dbserver deprovisioning request to NDB only if the serverId is present in status
	if databaseServerId != "" {
		r.recorder.Eventf(database, "Normal", EVENT_DEREGISTRATION_STARTED, "Deprovisioning database server from NDB.")
		task, err = ndb_api.DeprovisionDatabaseServer(ctx, ndbClient, databaseServerId, *ndb_api.GenerateDeprovisionDatabaseServerRequest())
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
