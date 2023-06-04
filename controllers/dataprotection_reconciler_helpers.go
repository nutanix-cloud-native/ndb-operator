package controllers

import (
	"context"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
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

// The handleSync function synchronizes the database CR's with the database instance in NDB
// It handles the transition from EMPTY (initial state) => PROVISIONING => RUNNING
// and updates the status accordingly. The update() triggers an implicit requeue of the reconcile request.
func (r *DataProtectionReconciler) handleSync(ctx context.Context, dataprotection *ndbv1alpha1.DataProtection, ndbClient *ndb_client.NDBClient, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered database_reconciler_helpers.handleSync")

	switch dataprotection.Status.Status {

	case common.DATAPROTECTION_CR_STATUS_EMPTY:
		// DP Status.Status is empty => Restore a DB
		log.Info("Restpring a database instance with NDB.")

	case common.DATABASE_CR_STATUS_PROVISIONING:

	case common.DATABASE_CR_STATUS_READY:

	default:
		// Do Nothing
	}

	return r.doNotRequeue()
}
