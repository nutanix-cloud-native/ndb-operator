package controllers

import (
	"context"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Fetches all databases from the NDB API and converts them to
// NDBServerDatabaseInfo type object (to be consumed by the NDBServer CR)
func getNDBServerDatabasesInfo(ctx context.Context, ndbClient *ndb_client.NDBClient) (databases []ndbv1alpha1.NDBServerDatabaseInfo, err error) {
	log := log.FromContext(ctx)
	log.Info("Fetching and converting databases from NDB")

	dbs, err := ndb_api.GetAllDatabases(ctx, ndbClient)
	if err != nil {
		log.Error(err, "NDB API error while fetching databases")
		return
	}
	clones, err := ndb_api.GetAllClones(ctx, ndbClient)
	if err != nil {
		log.Error(err, "NDB API error while fetching clones")
		return
	}
	dbs = append(dbs, clones...)
	databases = make([]ndbv1alpha1.NDBServerDatabaseInfo, len(dbs))
	for i, db := range dbs {
		databaseInfo := ndbv1alpha1.NDBServerDatabaseInfo{
			Name:          db.Name,
			Id:            db.Id,
			Status:        db.Status,
			TimeMachineId: db.TimeMachineId,
			Type:          db.Type,
		}
		if len(db.DatabaseNodes) > 0 {
			databaseInfo.DBServerId = db.DatabaseNodes[0].DatabaseServerId
			if len(db.DatabaseNodes[0].DbServer.IPAddresses) > 0 {
				databaseInfo.IPAddress = db.DatabaseNodes[0].DbServer.IPAddresses[0]
			}
		}
		databases[i] = databaseInfo
	}
	log.Info("Returning from ndbserver_controller_helpers.getNDBServerDatabasesInfo")
	return
}

// Returns the NDBServerStatus after performing the following steps:
// 1. Checks and fetch data if dbcounter is zero (we fetch data only when counter hits 0).
// 2. TODO: Filter and set the required list of databases (we only want to store the databases managed by the operator).
// 3. Update the counter value.
func getNDBServerStatus(ctx context.Context, status *ndbv1alpha1.NDBServerStatus, ndbClient *ndb_client.NDBClient) *ndbv1alpha1.NDBServerStatus {
	log := log.FromContext(ctx)
	log.Info("Entered ndbserver_controller_helpers.getNDBServerStatus")

	dbCounter := status.ReconcileCounter.Database
	// 1. Fetch dbs only if dbcounter is 0
	if dbCounter == 0 {
		log.Info("DbCounter 0, fetching databases (NDBServerDatabaseInfo)")
		databases, err := getNDBServerDatabasesInfo(ctx, ndbClient)
		if err != nil {
			log.Error(err, "Error occurred while fetching databases (NDBServerDatabaseInfo)")
			status.Status = common.NDB_CR_STATUS_ERROR
		} else {
			/* 2. TODO: Perform filtration on the databases associated with this NDB CR
			databaseList := &ndbv1alpha1.DatabaseList{}
			err = r.List(ctx, databaseList) // Also, we'll need to filter the dbs which are solely managed by THIS NDB CR. => Manual filter OR List Opts
			if err != nil {
				status.Status = common.NDB_CR_STATUS_ERROR
			}
			log.Info(util.ToString(databaseList))
			filteredDBs := util.Filter(databaseList.Items, FILTER_FUNC )
			*/
			status.Databases, err = util.CreateMapForKey(databases, "Id")
			if err != nil {
				log.Error(err, "Error occurred while creating dbId-db map")
				status.Status = common.NDB_CR_STATUS_ERROR
			}
		}
	}

	// 3. Update counters
	status.ReconcileCounter = ndbv1alpha1.ReconcileCounter{
		Database: (dbCounter + 1) % common.NDB_RECONCILE_DATABASE_COUNTER,
	}
	log.Info("Returning from ndbserver_controller_helpers.getNDBServerStatus")
	return status
}
