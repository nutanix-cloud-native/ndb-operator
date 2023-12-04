package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// The handleSync function synchronizes the database CR with the database info object in the
// NDBServer CR (which fetches it from NDB). It handles the transition from EMPTY (initial state) => WAITING => PROVISIONING => RUNNING
// and updates the status accordingly. The update() triggers an implicit requeue of the reconcile request.
func (r *SnapshotReconciler) handleSync(ctx context.Context, snapshot *ndbv1alpha1.Snapshot, ndbClient *ndb_client.NDBClient, req ctrl.Request, ndbServer *ndbv1alpha1.NDBServer) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered snapshot_conxtroller_helper.handleSync")

	snapshotStatus := snapshot.Status.DeepCopy()

	// Take a snapshot
	if snapshotStatus.Status == "" {
		// DB Status.Status is empty => Provision a DB
		requestBody := ndb_api.GenerateTakeSnapshotRequest(snapshot)
		taskResponse, err := ndb_api.TakeSnapshot(ctx, ndbClient, requestBody, snapshot.Spec.TimeMachineID)
		if err != nil {
			errStatement := "Failed to create snapshot of database"
			log.Error(err, errStatement)
			r.recorder.Eventf(snapshot, "Warning", EVENT_NDB_REQUEST_FAILED, "Error: %s. %s", errStatement, err.Error())
			return requeueOnErr(err)
		}
		log.Info(fmt.Sprintf("Updating Snapshot CR to Status: CREATING, id: %s and creationOperationId: %s", taskResponse.EntityId, taskResponse.OperationId))

		snapshotStatus.Status = common.DATABASE_CR_STATUS_CREATING
		snapshotStatus.OperationID = taskResponse.OperationId
		r.recorder.Event(snapshot, "Normal", EVENT_CREATION_STARTED, "Snapshot creation initiated")
	}

	isUnderDeletion := !snapshot.ObjectMeta.DeletionTimestamp.IsZero()
	if isUnderDeletion {
		if snapshotStatus.Status != common.DATABASE_CR_STATUS_DELETING {
			snapshots, err := ndb_api.GetAllSnapshots(ctx, ndbClient)
			if err != nil {
				log.Error(err, "Unable to get snapshots")
				r.recorder.Eventf(snapshot, "Warning", EVENT_NDB_REQUEST_FAILED, "Error:", "Unable to get snapshots", err.Error())
				return requeueOnErr(err)
			}
			for _, snap := range snapshots {
				if snap.LcmConfig != nil {
					var lcmConfig LcmConfig
					err = json.Unmarshal(snap.LcmConfig, &lcmConfig)
					if err != nil {
						log.Error(err, "Unmarshalling error")
						r.recorder.Eventf(snapshot, "Warning", EVENT_NDB_REQUEST_FAILED, "Error:", "Unmarshalling error", err.Error())
						return requeueOnErr(err)
					}
					if snap.Name == snapshot.Spec.Name && lcmConfig.ExpiryDateTimezone == snapshot.ExpiryDateTimezone && lcmConfig.ExpiryInDays == snapshot.ExpiryInDays {
						snapshotStatus.Id = snap.Id
						snapshotStatus.Status = common.DATABASE_CR_STATUS_DELETING
						log.Info(fmt.Sprintf("Snap %s with id %s", snap.Name, snap.Id))
						break
					}
				}
			}
		}
	} else if snapshotStatus.Status == common.DATABASE_CR_STATUS_CREATING {
		creationOp, err := ndb_api.GetOperationById(ctx, ndbClient, snapshotStatus.OperationID)
		if err != nil {
			message := fmt.Sprintf("NDB API to fetch operation by id failed. OperationId: %s:, error: %s", creationOp.Id, err.Error())
			r.recorder.Event(snapshot, "Warning", EVENT_NDB_REQUEST_FAILED, message)
		} else {
			switch ndb_api.GetOperationStatus(creationOp) {
			case ndb_api.OPERATION_STATUS_FAILED:
				snapshotStatus.Status = common.DATABASE_CR_STATUS_CREATION_ERROR
				err = fmt.Errorf("creation operation terminated. status: %s, message: %s, operationId: %s", creationOp.Status, creationOp.Message, creationOp.Id)
				log.Error(err, "Database Creation Failed")
				r.recorder.Event(snapshot, "Warning", EVENT_CREATION_FAILED, "Take Snapshot operation failed with error: "+err.Error())
			case ndb_api.OPERATION_STATUS_PASSED:
				snapshotStatus.Status = common.DATABASE_CR_STATUS_READY
				r.recorder.Event(snapshot, "Normal", EVENT_CREATION_COMPLETED, "Take Snapshot operation passed")
			default:
				// Do nothing, we do not care about other statuses
			}
		}
	} else {
		log.Info("Snapshot missing from NDB CR")
		snapshotStatus.Status = common.DATABASE_CR_STATUS_NOT_FOUND
	}

	if !reflect.DeepEqual(snapshot.Status, *snapshotStatus) {
		snapshot.Status = *snapshotStatus
		err := r.Status().Update(ctx, snapshot)
		if err != nil {
			errStatement := "Failed to update status of snapshot custom resource"
			log.Error(err, errStatement)
			r.recorder.Eventf(snapshot, "Warning", EVENT_CR_STATUS_UPDATE_FAILED, "Error: %s. %s.", err.Error())
			return requeueOnErr(err)
		}
	}

	switch snapshotStatus.Status {
	case common.DATABASE_CR_STATUS_READY:
		if !isUnderDeletion {
			if !controllerutil.ContainsFinalizer(snapshot, common.FINALIZER_INSTANCE) {
				return r.addFinalizer(ctx, req, common.FINALIZER_INSTANCE, snapshot)
			}
		}
	case common.DATABASE_CR_STATUS_DELETING:
		return r.handleDelete(ctx, snapshot, ndbClient)
	case common.DATABASE_CR_STATUS_NOT_FOUND:
		r.recorder.Eventf(snapshot, "Warning", EVENT_EXTERNAL_DELETE, "Error: Resource not found on NDB")
	case common.DATABASE_CR_STATUS_CREATION_ERROR:
		return doNotRequeue()
	default:
		// No-Op
	}

	return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
}

// handleDelete function handles the deletion of
//
//		a. Snapshot
//	 b. Snapshot Finalizer
func (r *SnapshotReconciler) handleDelete(ctx context.Context, snapshot *ndbv1alpha1.Snapshot, ndbClient *ndb_client.NDBClient) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info(fmt.Sprintf("Snapshot CR is being deleted with id %s", snapshot.Status.Id))
	log.Info(snapshot.ResourceVersion)
	if controllerutil.ContainsFinalizer(snapshot, common.FINALIZER_INSTANCE) {
		// Check if the deregistration operation id (database.Status.DeregistrationOperationId) is empty
		// If so, then make a deprovisionDatabase API call to NDB
		// else proceed check for the operation completion before removing finalizer.
		deletionOperationId := snapshot.Status.DeletionOperationID
		if deletionOperationId == "" {
			deletionOp, err := ndb_api.DeleteSnapshot(ctx, ndbClient, snapshot.Status.Id)
			if err != nil {
				// Not logging here, already done in the deregister function
				return requeueOnErr(err)
			}
			snapshot.Status.DeletionOperationID = deletionOp.OperationId
			if err := r.Status().Update(ctx, snapshot); err != nil {
				log.Error(err, "An error occurred while updating the CR.")
				return requeueOnErr(err)
			}
		} else {
			deletionOp, err := ndb_api.GetOperationById(ctx, ndbClient, deletionOperationId)
			if err != nil {
				message := fmt.Sprintf("NDB API to fetch operation by id failed. OperationId: %s:, error: %s", deletionOperationId, err.Error())
				r.recorder.Event(snapshot, "Warning", EVENT_NDB_REQUEST_FAILED, message)
			} else {
				switch ndb_api.GetOperationStatus(deletionOp) {
				case ndb_api.OPERATION_STATUS_FAILED:
					err := fmt.Errorf("Deletion operation terminated. status: %s, message: %s, operationId: %s", deletionOp.Status, deletionOp.Message, deletionOperationId)
					log.Error(err, "Deletion Failed")
					r.recorder.Event(snapshot, "Warning", "OPERATION FAILED", "Snapshot deletion operation failed with error: "+err.Error())
				case ndb_api.OPERATION_STATUS_PASSED:
					r.recorder.Eventf(snapshot, "Normal", EVENT_DEREGISTRATION_COMPLETED, "Snapshot deleted from NDB.")
					log.Info("Removing Finalizer " + common.FINALIZER_INSTANCE)
					controllerutil.RemoveFinalizer(snapshot, common.FINALIZER_INSTANCE)
					if err := r.Update(ctx, snapshot); err != nil {
						return requeueOnErr(err)
					}
					log.Info("Removed Finalizer " + common.FINALIZER_INSTANCE)
				default:
					// Do nothing, we do not care about other statuses
				}
			}
		}
	} else {
		// Finalizer has been removed, no need to requeue
		// CR will be deleted.
		return doNotRequeue()
	}
	// Requeue the request while waiting for the database instance to be deleted from NDB.
	return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
}

func (r *SnapshotReconciler) addFinalizer(ctx context.Context, req ctrl.Request, finalizer string, snapshot *ndbv1alpha1.Snapshot) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Fetching the most recent version of the Snapshot CR")
	err := r.Get(ctx, req.NamespacedName, snapshot)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Snapshot resource not found. Ignoring since object must be deleted")
			return doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Snapshot")
		return requeueOnErr(err)
	}
	log.Info("Snapshot CR fetched. Adding finalizer " + finalizer)
	controllerutil.AddFinalizer(snapshot, finalizer)
	if err := r.Update(ctx, snapshot); err != nil {
		return requeueOnErr(err)
	} else {
		log.Info("Added finalizer " + finalizer)
	}
	//Not requeuing as a successful update automatically triggers a reconcile.
	return requeueWithTimeout(common.DATABASE_RECONCILE_INTERVAL_SECONDS)
}
