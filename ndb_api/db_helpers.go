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
	"errors"
	"fmt"
	"strconv"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB
// The database provisioned has a NONE time machine SLA attached to it, and uses the default OOB profiles
func GenerateProvisioningRequest(ctx context.Context, ndb_client *ndb_client.NDBClient, database DatabaseInterface, reqData map[string]interface{}) (requestBody *DatabaseProvisionRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GenerateProvisioningRequest", "database name", database.GetName(), "database type", database.GetInstanceType())

	// Fetching the TM details
	tmName, tmDescription, slaName := database.GetInstanceTMDetails()
	// NDB rejects any SLA other than "NONE" for MySQL HA.
	//TODO: Remove this once NDB 2.11+ is released (time machine support for MySQL HA)
	if database.IsMysqlHA() {
		slaName = common.SLA_NAME_NONE
	}
	// Fetching the SLA for the TM by name
	sla, err := GetSLAByName(ctx, ndb_client, slaName)
	if err != nil {
		log.Error(err, "Error occurred while getting TM SLA", "SLA Name", slaName)
		return
	}

	schedule, err := database.GetTMScheduleForInstance()
	if err != nil {
		log.Error(err, "Error occurred while generating the Time Machine Schedule")
		return
	}

	// Fetch the required profiles for the database
	profilesMap, err := ResolveProfiles(ctx, ndb_client, database.GetInstanceType(), database.GetProfileResolvers())
	if err != nil {
		log.Error(err, "Error occurred while getting required profiles", "database name", database.GetName(), "database type", database.GetInstanceType())
		return
	}
	// Required for dbParameterProfileIdInstance in MSSQL action args
	reqData[common.PROFILE_MAP_PARAM] = profilesMap

	// Validate request data
	err = validateReqData(ctx, database.GetInstanceType(), reqData)
	if err != nil {
		log.Error(err, "Error occurred while validating reqData", "reqData", reqData)
		return
	}

	// Creating a provisioning request based on the database type
	requestBody = &DatabaseProvisionRequest{
		DatabaseType:             GetDatabaseEngineName(database.GetInstanceType()),
		Name:                     database.GetName(),
		DatabaseDescription:      database.GetDescription(),
		SoftwareProfileId:        profilesMap[common.PROFILE_TYPE_SOFTWARE].Id,
		SoftwareProfileVersionId: profilesMap[common.PROFILE_TYPE_SOFTWARE].LatestVersionId,
		ComputeProfileId:         profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
		NetworkProfileId:         profilesMap[common.PROFILE_TYPE_NETWORK].Id,
		DbParameterProfileId:     profilesMap[common.PROFILE_TYPE_DATABASE_PARAMETER].Id,
		NewDbServerTimeZone:      database.GetTimeZone(),
		CreateDbServer:           true,
		NodeCount:                1,
		NxClusterId:              database.GetClusterId(),
		Clustered:                false,
		AutoTuneStagingDrive:     true,

		TimeMachineInfo: TimeMachineInfo{
			Name:             tmName,
			Description:      tmDescription,
			SlaId:            sla.Id,
			Schedule:         schedule,
			Tags:             make([]string, 0),
			AutoTuneLogDrive: true,
		},
		Nodes: []Node{
			{
				Properties: make([]NodeProperty, 0),
				VmName:     database.GetName() + "_VM",
			},
		},
		ActionArguments: []ActionArgument{
			{
				Name:  "dbserver_description",
				Value: "dbserver for " + database.GetName(),
			},
			{
				Name:  "database_size",
				Value: strconv.Itoa(database.GetInstanceSize()),
			},
		},
	}

	// Override request fields for HA postgres instances
	if database.IsPostgresHA() {
		haConfig := database.GetInstanceHAConfig()
		requestBody.Clustered = true
		requestBody.NodeCount = len(haConfig.Nodes)

		// Build the per-node entries from the HA config
		haNodes := make([]Node, 0, len(haConfig.Nodes))
		for _, n := range haConfig.Nodes {
			node := Node{
				VmName:      n.VmName,
				NxClusterId: n.ClusterId,
			}
			if n.NodeType == common.HA_NODE_TYPE_HAPROXY {
				node.Properties = []NodeProperty{
					{Name: "node_type", Value: common.HA_NODE_TYPE_HAPROXY},
				}
			} else {
				// database node
				node.NetworkProfileId = profilesMap[common.PROFILE_TYPE_NETWORK].Id
				node.ComputeProfileId = profilesMap[common.PROFILE_TYPE_COMPUTE].Id
				node.Properties = []NodeProperty{
					{Name: "role", Value: n.Role},
					{Name: "failover_mode", Value: n.FailoverMode},
					{Name: "node_type", Value: "database"},
					{Name: "remote_archive_destination", Value: ""},
				}
			}
			haNodes = append(haNodes, node)
		}
		requestBody.Nodes = haNodes

		// Collect unique cluster IDs across all HA nodes for the TM SLA details
		clusterIdSet := make(map[string]struct{})
		for _, n := range haConfig.Nodes {
			if n.ClusterId != "" {
				clusterIdSet[n.ClusterId] = struct{}{}
			}
		}
		clusterIds := make([]string, 0, len(clusterIdSet))
		for id := range clusterIdSet {
			clusterIds = append(clusterIds, id)
		}
		requestBody.TimeMachineInfo.SlaId = ""
		requestBody.TimeMachineInfo.SlaDetails = &SlaDetails{
			PrimarySla: PrimarySlaDetails{
				SlaId:        sla.Id,
				NxClusterIds: clusterIds,
			},
		}
	}

	// Appending request body based on database type
	appender, err := GetRequestAppender(database.GetInstanceType())
	if err != nil {
		log.Error(err, "Error while appending provisioning request")
		return
	}

	requestBody, err = appender.appendProvisioningRequest(requestBody, database, reqData)
	if err != nil {
		log.Error(err, "Error while appending provisioning request")
	}

	log.Info("Database Provisioning", "requestBody", requestBody)
	log.Info("Returning from ndb_api.GenerateProvisioningRequest", "database name", database.GetName(), "database type", database.GetInstanceType())
	return
}

// Returns a request to delete a database instance
func GenerateDeprovisionDatabaseRequest() (req *DatabaseDeprovisionRequest) {
	req = &DatabaseDeprovisionRequest{
		Delete:               true,
		Remove:               false,
		SoftRemove:           false,
		Forced:               false,
		DeleteTimeMachine:    true,
		DeleteLogicalCluster: true,
	}
	return

}

func validateReqData(ctx context.Context, databaseInstanceType string, reqData map[string]interface{}) (err error) {
	log := ctrllog.FromContext(ctx)
	dbPassword, ok := reqData[common.NDB_PARAM_PASSWORD].(string)
	// Type Assertion for dbPassword
	if !ok || dbPassword == "" {
		err = errors.New("invalid database password")
		var errStatement string
		if !ok {
			errStatement = "Type assertion failed for database password. Expected a string value"
		} else {
			errStatement = "Empty database password"
		}
		log.Error(err, errStatement)
		return
	}

	// Type Assertion for SSHKey
	if databaseInstanceType != common.DATABASE_TYPE_MSSQL {
		SSHPublicKey, ok := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
		if !ok || SSHPublicKey == "" {
			err = errors.New("invalid ssh public key")
			var errStatement string
			if !ok {
				errStatement = "Type assertion failed for SSHPublicKey. Expected a string value"
			} else {
				errStatement = "Empty SSHPublicKey"
			}
			log.Error(err, errStatement)
			return
		}
	}
	return
}

// Converts a map to an action arguments list
func convertMapToActionArguments(myMap map[string]string) []ActionArgument {
	actionArgs := []ActionArgument{}
	for name, value := range myMap {
		actionArgs = append(actionArgs, ActionArgument{Name: name, Value: value})
	}
	return actionArgs
}

// Overwrites and appends actionArguments from database.additionalArguments to actionArguments
func setConfiguredActionArguments(database DatabaseInterface, actionArguments map[string]string) error {
	errMsgRoot := "Setting configured action arguments failed"
	if actionArguments == nil {
		return fmt.Errorf("%s! Action arguments cannot be nil", errMsgRoot)
	}

	allowedAdditionalArguments, err := util.GetAllowedAdditionalArguments(database.IsClone(), database.GetInstanceType())
	if err != nil {
		return fmt.Errorf("%s! %s", errMsgRoot, err.Error())
	}

	if len(database.GetAdditionalArguments()) > len(allowedAdditionalArguments) {
		return fmt.Errorf("%s! Length of specified action arguments is greater then allowed additional arguments", errMsgRoot)
	}

	// Rewrite or add actionArguments from additionalArgument list if it is an actionArgument
	for name, value := range database.GetAdditionalArguments() {
		isActionArgument, isPresent := allowedAdditionalArguments[name]
		if !isPresent {
			return fmt.Errorf("%s! %s is not an allowed additional argument", errMsgRoot, name)
		} else if isPresent && isActionArgument {
			actionArguments[name] = value
		}
	}

	return nil
}

func (a *MSSQLRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	req.DatabaseName = string(database.GetInstanceDatabaseNames())
	adminPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	profileMap := reqData[common.PROFILE_MAP_PARAM].(map[string]ProfileResponse)
	dbParamInstanceProfile := profileMap[common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE]

	// Default action arguments
	actionArguments := map[string]string{
		"working_dir":                       "C:\\temp",
		"sql_user_name":                     "sa",
		"authentication_mode":               "windows",
		"delete_vm_on_failure":              "false",
		"is_gmsa_sql_service_account":       "false",
		"provision_from_backup":             "false",
		"distribute_database_data":          "true",
		"retain_database_in_restoring_mode": "false",
		"dbserver_name":                     database.GetName(),
		"server_collation":                  "SQL_Latin1_General_CP1_CI_AS",
		"database_collation":                "SQL_Latin1_General_CP1_CI_AS",
		"dbParameterProfileIdInstance":      dbParamInstanceProfile.Id,
		"vm_dbserver_admin_password":        adminPassword,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}

func (a *MongoDbRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	req.SSHPublicKey = SSHPublicKey

	// Default action arguments
	actionArguments := map[string]string{
		"listener_port":  "27017",
		"log_size":       "100",
		"journal_size":   "100",
		"restart_mongod": "true",
		"working_dir":    "/tmp",
		"db_user":        "admin",
		"backup_policy":  "primary_only",
		"db_password":    dbPassword,
		"database_names": databaseNames,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}

func (a *PostgresRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	req.SSHPublicKey = SSHPublicKey

	// Default action arguments — write/read ports are overridden below for HA instances
	actionArguments := map[string]string{
		"proxy_read_port":         "5001",
		"listener_port":           "5432",
		"proxy_write_port":        "5000",
		"enable_synchronous_mode": "false",
		"auto_tune_staging_drive": "true",
		"backup_policy":           "primary_only",
		"db_password":             dbPassword,
		"database_names":          databaseNames,
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// HA-specific action arguments (override defaults where needed)
	if database.IsPostgresHA() {
		haConfig := database.GetInstanceHAConfig()
		syncMode := "false"
		if haConfig.EnableSynchronousMode {
			syncMode = "true"
		}
		provisionVIP := "false"
		if haConfig.ProvisionVirtualIP {
			provisionVIP = "true"
		}
		haActionArguments := map[string]string{
			"proxy_write_port":            fmt.Sprintf("%d", haConfig.WritePort),
			"proxy_read_port":             fmt.Sprintf("%d", haConfig.ReadPort),
			"provision_virtual_ip":        provisionVIP,
			"deploy_haproxy":              "true",
			"failover_mode":               common.HA_NODE_FAILOVER_MODE_AUTOMATIC,
			"enable_synchronous_mode":     syncMode,
			"patroni_cluster_name":        haConfig.PatroniClusterName,
			"cluster_name":                haConfig.ClusterName,
			"allocate_pg_hugepage":        "false",
			"cluster_database":            "false",
			"cte_intent":                  "false",
			"archive_wal_expire_days":     "2",
			"enable_peer_auth":            "false",
			"ensure_vm_host_distribution": "false",
			"db_user":                     "postgres",
		}
		for k, v := range haActionArguments {
			actionArguments[k] = v
		}
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}

func (a *MySqlRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	req.SSHPublicKey = SSHPublicKey

	// Default action arguments — shared by SI and HA
	actionArguments := map[string]string{
		"listener_port":           "3306",
		"db_password":             dbPassword,
		"database_names":          databaseNames,
		"auto_tune_staging_drive": "true",
	}

	// HA-specific request overrides (override req fields and merge HA action arguments).
	// This runs before setConfiguredActionArguments so that user-provided additionalArguments
	// take final precedence over the HA defaults below.
	if database.IsMysqlHA() {
		haConfig := database.GetInstanceHAConfig()

		req.Clustered = true
		req.NodeCount = len(haConfig.Nodes)

		// Build per-node entries from the HA config.
		profilesMap := reqData[common.PROFILE_MAP_PARAM].(map[string]ProfileResponse)
		haNodes := make([]Node, 0, len(haConfig.Nodes))
		for _, n := range haConfig.Nodes {
			node := Node{
				VmName:      n.VmName,
				NxClusterId: n.ClusterId,
			}

			node.NetworkProfileId = profilesMap[common.PROFILE_TYPE_NETWORK].Id
			node.ComputeProfileId = profilesMap[common.PROFILE_TYPE_COMPUTE].Id
			if n.NodeType == common.HA_NODE_TYPE_MYSQLROUTER {
				node.Properties = []NodeProperty{
					{Name: "node_type", Value: common.HA_NODE_TYPE_MYSQLROUTER},
				}
			} else {
				node.Properties = []NodeProperty{
					{Name: "role", Value: n.Role},
					{Name: "node_type", Value: common.HA_NODE_TYPE_DATABASE},
				}
			}
			haNodes = append(haNodes, node)
		}
		req.Nodes = haNodes

		deployRouter := "false"
		if haConfig.DeployMySQLRouter {
			deployRouter = "true"
		}

		haActionArguments := map[string]string{
			"cluster_name":                haConfig.ClusterName,
			"innodb_cluster_name":         haConfig.InnoDBClusterName,
			"mysql_cluster_username":      haConfig.MySQLClusterUsername,
			"mysql_cluster_password":      dbPassword,
			"replication_user":            haConfig.ReplicationUser,
			"replication_password":        dbPassword,
			"deploy_mysqlrouter":          deployRouter,
			"router_rw_port":              fmt.Sprintf("%d", haConfig.RouterRWPort),
			"router_ro_port":              fmt.Sprintf("%d", haConfig.RouterROPort),
			"allocate_mysql_hugepage":     "true",
			"ensure_vm_host_distribution": "false",
		}
		for k, v := range haActionArguments {
			actionArguments[k] = v
		}
	}

	// Apply user-provided additionalArguments last so they override both the base
	// defaults and the HA defaults merged above.
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}

func (a *OracleRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	// Oracle uses req.DatabaseName for global database name
	databaseNames := database.GetInstanceDatabaseNames()
	req.DatabaseName = databaseNames
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	req.SSHPublicKey = SSHPublicKey

	// Oracle REQUIRES extensive action arguments (verified from working manual API call)
	actionArguments := map[string]string{
		// Basic configuration
		"listener_port":           "1521",
		"auto_tune_staging_drive": "true",
		"working_dir":             "/tmp",

		// Database server naming
		"dbserver_name": database.GetName() + "_VM",

		// Oracle-specific identifiers (all required)
		"oracle_sid":           databaseNames, // Same as database name
		"global_database_name": databaseNames, // Same as database name
		"db_unique_name":       databaseNames, // Same as database name

		// Passwords (required for SYS and SYSTEM users)
		"sys_password":    dbPassword,
		"system_password": dbPassword,

		// Character sets
		"db_character_set":       "AL32UTF8",
		"national_character_set": "AL16UTF16",

		// Storage configuration
		"database_fra_size": fmt.Sprintf("%d", database.GetInstanceSize()), // FRA = Fast Recovery Area

		// Feature flags
		"enable_cdb": "false", // Container Database (we're provisioning SI, not multitenant)
		"enable_tde": "false", // Transparent Data Encryption
		"enable_ha":  "false", // High Availability

		// Cleanup
		"delete_logs_older_than":      "0",
		"ensure_vm_host_distribution": "false",
		"pre_create_script":           "",
		"post_create_script":          "",
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}
