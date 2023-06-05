package controllers

import (
	"context"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// TODO: Move to a common controller helper as this will need to be repeated in every reconciler helper

// doNotRequeue Finished processing. No need to put back on the reconcile queue.
func (r *DataProtectionReconciler) doNotRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// requeueOnErr Failed while processing. Put back on reconcile queue and try again.
func (r *DataProtectionReconciler) requeueOnErr(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// Returns the credentials(username, password and caCertificate) for NDB
// Returns an error if reading the secret containing credentials fails
func (r *DataProtectionReconciler) getNDBCredentials(ctx context.Context, name, namespace string) (username, password, caCert string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered dataprotection_reconciler_helpers.getNDBCredentials")
	secretDataMap, err := util.GetAllDataFromSecret(ctx, r.Client, name, namespace)
	if err != nil {
		log.Error(err, "Error occured in util.GetAllDataFromSecret while fetching all NDB secrets", "Secret Name", name, "Namespace", namespace)
		return
	}
	username = string(secretDataMap[common.SECRET_DATA_KEY_USERNAME])
	password = string(secretDataMap[common.SECRET_DATA_KEY_PASSWORD])
	caCert = string(secretDataMap[common.SECRET_DATA_KEY_CA_CERTIFICATE])
	log.Info("Returning from dataprotection_reconciler_helpers.getNDBCredentials")
	return
}

// The handleSync function synchronizes the DataProtection CR's with the DP instance in NDB
// It handles the transition from EMPTY (initial state) => Unassigned => Running => Succeeded/Failed
// and updates the status accordingly. The update() triggers an implicit requeue of the reconcile request.
func (r *DataProtectionReconciler) handleSync(ctx context.Context,
	dataprotection *ndbv1alpha1.DataProtection, database *ndbv1alpha1.Database,
	ndbClient *ndb_client.NDBClient, req ctrl.Request) (ctrl.Result, error) {

	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleSync")

	switch dataprotection.Status.Status {

	case common.DPCR_STATUS_EMPTY:
		// DP Status.Status is empty => Restore a DB
		log.Info("Restoring a DB...")

		generatedReq, err := ndb_api.GenerateRestoreRequestFromSnapshot(ctx, ndbClient, dataprotection.Spec.Restore.Snapshot.Id)
		validation_error := ndb_api.ValidateDatabaseRestoreRequest(*generatedReq)

		if validation_error != nil {
			log.Error(err, "Could not generate restore request, re-queuing.")
			return r.requeueOnErr(err)
		}

		if err != nil {
			log.Error(err, "Could not generate restore request, re-queuing.")
			return r.requeueOnErr(err)
		}

		taskResponse, err := ndb_api.RestoreDatabase(ctx, ndbClient, generatedReq, database.Status.Id)
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

		err = r.Status().Update(ctx, database)
		if err != nil {
			log.Error(err, "Failed to update database status")
			return r.requeueOnErr(err)
		}

	case common.DPCR_STATUS_UNASSIGNED, common.DPCR_STATUS_RUNNING:

	case common.DPCR_STATUS_SUCCEEDED:

	case common.DPCR_STATUS_FAILED:

	default:
		// Do Nothing
	}

	return r.doNotRequeue()
}
