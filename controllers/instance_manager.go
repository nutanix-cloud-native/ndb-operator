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
		instanceManager = &DatabaseManager{}
	} else {
		instanceManager = &CloneManager{}
	}
	return
}

type InstanceManager interface {
	create(ctx context.Context, r *DatabaseReconciler, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database, namespace string) (task ndb_api.TaskInfoSummaryResponse, err error)
	deregister() error
	deleteVM() error
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

func (dm *DatabaseManager) deregister() error {
	return nil
}

func (dm *DatabaseManager) deleteVM() error {
	return nil
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

func (cm *CloneManager) deregister() error {
	return nil
}

func (cm *CloneManager) deleteVM() error {
	return nil
}
