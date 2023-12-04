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

package ndb_api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns a request to delete a clone instance
func GenerateDeprovisionCloneRequest() (req *CloneDeprovisionRequest) {
	req = &CloneDeprovisionRequest{
		SoftRemove:           false,
		Remove:               false,
		Delete:               true,
		Forced:               true,
		DeleteDataDrives:     true,
		DeleteLogicalCluster: true,
		RemoveLogicalCluster: false,
		DeleteTimeMachine:    true,
	}
	return

}

// This function generates and returns a request for cloning a database on NDB
func GenerateCloningRequest(ctx context.Context, ndb_client *ndb_client.NDBClient, database DatabaseInterface, reqData map[string]interface{}) (requestBody *DatabaseCloneRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GenerateCloningRequest", "database name", database.GetName())

	sourceDatabase, _ := GetDatabaseById(ctx, ndb_client, database.GetCloneSourceDBId())
	databaseType := GetDatabaseTypeFromEngine(sourceDatabase.Type)
	// database.
	// Fetch the required profiles for the database
	profilesMap, err := ResolveProfiles(ctx, ndb_client, databaseType, database.GetProfileResolvers())
	if err != nil {
		log.Error(err, "Error occurred while getting required profiles", "database name", database.GetName(), "isClone", database.IsClone())
		return
	}
	// Required for dbParameterProfileIdInstance in MSSQL action args
	reqData[common.PROFILE_MAP_PARAM] = profilesMap

	// Creating a provisioning request based on the database type
	requestBody = &DatabaseCloneRequest{
		Name:           database.GetName(),
		Description:    database.GetDescription(),
		CreateDbServer: true,
		Clustered:      false,
		NxClusterId:    database.GetClusterId(),
		// SSHPublicKey populated by request appenders for non mssql dbs
		DbServerId:               "",
		DbServerClusterId:        "",
		DbserverLogicalClusterId: "",
		TimeMachineId:            sourceDatabase.TimeMachineId,
		SnapshotId:               database.GetCloneSnapshotId(),
		UserPitrTimestamp:        "",
		TimeZone:                 database.GetTimeZone(),
		LatestSnapshot:           false,
		NodeCount:                1,
		Nodes: []Node{
			{
				VmName:              database.GetName() + "_CLONE_VM",
				ComputeProfileId:    profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
				NetworkProfileId:    profilesMap[common.PROFILE_TYPE_NETWORK].Id,
				NewDbServerTimeZone: "",
				NxClusterId:         database.GetClusterId(),
				Properties:          make([]map[string]string, 0),
			},
		},
		// Added by request appenders as per the engine
		ActionArguments:            []ActionArgument{},
		Tags:                       make([]interface{}, 0),
		VmPassword:                 "",
		ComputeProfileId:           profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
		NetworkProfileId:           profilesMap[common.PROFILE_TYPE_NETWORK].Id,
		DatabaseParameterProfileId: profilesMap[common.PROFILE_TYPE_DATABASE_PARAMETER].Id,
	}

	// boolean for high availability
	isHighAvailability := false

	// Appending request body based on database type
	appender, err := GetRequestAppender(databaseType, isHighAvailability)
	if err != nil {
		log.Error(err, "Error while getting a request appender")
		return
	}
	requestBody, err = appender.appendCloningRequest(requestBody, database, reqData)
	if err != nil {
		log.Error(err, "Error while appending clone request")
		return
	}

	log.Info("Database Cloning", "requestBody", requestBody)
	log.Info("Returning from ndb_api.GenerateCloningRequest", "database name", database.GetName())
	return
}

func (a *MSSQLRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	vmName := req.Name
	dbName := database.GetName()
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	// Default action arguments
	actionArguments := map[string]string{
		/* Non-Configurable */
		"quorum_witness_type":   "disk_share",
		"vm_win_lang_settings":  "en-US",
		"drives_to_mountpoints": "false",
		"cluster_db":            "false",
		/* Configurable */
		"vm_name":                    vmName,
		"database_name":              dbName,
		"vm_dbserver_admin_password": dbPassword,
		"dbserver_description":       "DB Server VM for " + database.GetName(),
		"sql_user_name":              "sa",
		"authentication_mode":        "windows",
		"instance_name":              "CDMINSTANCE",
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	// Appending LCMConfig Details if specified
	if err := appendLCMConfigDetailsToRequest(req, database.GetAdditionalArguments()); err != nil {
		return nil, err
	}

	return req, nil
}

func (a *MongoDbRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	req.Description = "DB Server VM for " + database.GetName()
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	// Default action arguments
	actionArguments := map[string]string{
		/* Non-Configurable */
		"listener_port": "27017",
		/* Configurable */
		"vm_name":     database.GetName(),
		"db_password": dbPassword,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	// Appending LCMConfig Details if specified
	if err := appendLCMConfigDetailsToRequest(req, database.GetAdditionalArguments()); err != nil {
		return nil, err
	}

	return req, nil
}

func (a *PostgresRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	// Default action arguments
	actionArguments := map[string]string{
		/* Non-Configurable from additionalArguments*/
		"vm_name":              database.GetName(),
		"dbserver_description": "DB Server VM for " + database.GetName(),
		"db_password":          dbPassword,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	// Appending LCMConfig Details if specified
	if err := appendLCMConfigDetailsToRequest(req, database.GetAdditionalArguments()); err != nil {
		return nil, err
	}

	return req, nil
}

func setCloneNodesParameters(req *DatabaseCloneRequest, database DatabaseInterface) {
	// Extract values of ComputeProfileId and NetworkProfileId
	computeProfileId := req.Nodes[0].ComputeProfileId
	networkProfileId := req.Nodes[0].NetworkProfileId
	serverTimeZone := req.Nodes[0].NewDbServerTimeZone

	// Clear the original req.Nodes array
	req.Nodes = []Node{}

	// Create node object for HA Proxy
	for i := 0; i < 2; i++ {
		// Hard coding the HA Proxy properties
		props := make([]map[string]string, 1)
		props[0] = map[string]string{
			"name":  "node_type",
			"value": "haproxy",
		}
		req.Nodes = append(req.Nodes, Node{
			Properties:  props,
			VmName:      database.GetName() + "_haproxy" + strconv.Itoa(i),
			NxClusterId: database.GetClusterId(),
		})
	}

	// Create node object for Database Instances
	for i := 0; i < 3; i++ {
		// Hard coding the DB properties
		props := make([]map[string]string, 4)
		props[0] = map[string]string{
			"name":  "role",
			"value": "Secondary",
		}
		// 1st node will be the primary node
		if i == 0 {
			props[0]["value"] = "Primary"
		}
		props[1] = map[string]string{
			"name":  "failover_mode",
			"value": "Automatic",
		}
		props[2] = map[string]string{
			"name":  "node_type",
			"value": "database",
		}
		props[3] = map[string]string{
			"name":  "remote_archive_destination",
			"value": "",
		}
		req.Nodes = append(req.Nodes, Node{
			ComputeProfileId:    computeProfileId,
			NetworkProfileId:    networkProfileId,
			NewDbServerTimeZone: serverTimeZone,
			Properties:          props,
			VmName:              database.GetName() + "-" + strconv.Itoa(i),
			NxClusterId:         database.GetClusterId(),
		})
	}
}

func (a *PostgresHARequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	// Set the number of nodes to 5, 3 Postgres nodes + 2 HA Proxy nodes
	req.NodeCount = 5
	setCloneNodesParameters(req, database)

	// Default action arguments
	actionArguments := map[string]string{
		/* Non-Configurable from additionalArguments*/
		"vm_name":              database.GetName(),
		"dbserver_description": "DB Server VM for " + database.GetName(),
		"db_password":          dbPassword,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	// Appending LCMConfig Details if specified
	if err := appendLCMConfigDetailsToRequest(req, database.GetAdditionalArguments()); err != nil {
		return nil, err
	}

	return req, nil
}

func (a *MySqlRequestAppender) appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error) {
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	// Default action arguments
	actionArguments := map[string]string{
		/* Non-Configurable */
		/* Configurable */
		"vm_name":              database.GetName(),
		"dbserver_description": "DB Server VM for " + database.GetName(),
		"db_password":          dbPassword,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	// Appending LCMConfig Details if specified
	if err := appendLCMConfigDetailsToRequest(req, database.GetAdditionalArguments()); err != nil {
		return nil, err
	}

	return req, nil
}

func appendLCMConfigDetailsToRequest(req *DatabaseCloneRequest, additionalArguments map[string]string) error {
	errMsg := "appendLCMConfigDetailsToRequest() failed!"

	// expiryDetails appender
	databaseLcmConfigProperties := []string{"expireInDays", "expiryDateTimezone", "deleteDatabase"}
	databaseLcmConfigCount := 0
	for _, property := range databaseLcmConfigProperties {
		if _, isPresent := additionalArguments[property]; isPresent {
			databaseLcmConfigCount += 1
		}
	}
	if databaseLcmConfigCount == 3 {
		req.LcmConfig.DatabaseLCMConfig = DatabaseLCMConfig{
			ExpiryDetails: ExpiryDetails{
				ExpireInDays:       additionalArguments["expireInDays"],
				ExpiryDateTimezone: additionalArguments["expiryDateTimezone"],
				DeleteDatabase:     additionalArguments["deleteDatabase"],
			},
		}
	} else if databaseLcmConfigCount != 0 {
		return fmt.Errorf("%s. Ensure expireInDays, expiryDateTimezone, and deleteDatabase are all specified. You only have %d/3 specified", errMsg, databaseLcmConfigCount)
	}

	// refreshDetails appender
	refreshDetailsProperties := []string{"refreshInDays", "refreshTime", "refreshDateTimezone"}
	refreshDetailsCount := 0
	for _, property := range refreshDetailsProperties {
		if _, isPresent := additionalArguments[property]; isPresent {
			refreshDetailsCount += 1
		}
	}
	if refreshDetailsCount == 3 {
		req.LcmConfig.DatabaseLCMConfig.RefreshDetails = RefreshDetails{
			RefreshInDays:       additionalArguments["refreshInDays"],
			RefreshTime:         additionalArguments["refreshTime"],
			RefreshDateTimezone: additionalArguments["refreshDateTimezone"],
		}
	} else if databaseLcmConfigCount != 0 {
		return fmt.Errorf("%s. Ensure refreshInDay, refreshTime, refreshDateTimezone are all specified. You only have %d/3 specified", errMsg, refreshDetailsCount)
	}

	return nil
}
