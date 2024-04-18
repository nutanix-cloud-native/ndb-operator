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
	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
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
				Properties: make([]map[string]string, 0),
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

	// Appending request body based on database type
	appender, err := GetRequestAppender(database.GetInstanceType(), database.GetInstanceIsHighAvailability())
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

	// Default action arguments
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

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}

func setNodesParameters(req *DatabaseProvisionRequest, database DatabaseInterface) (nodeErrors error) {
	// Clear the original req.Nodes array
	req.Nodes = []Node{}

	// Validate node counts
	nodesRequested := database.GetInstanceNodes()
	nodeCount := len(nodesRequested)
	if nodeCount == 0 {
		nodeCount = 5
		nodesRequested = createDefaultNodes(database)
	}
	databaseNodeCount := 0
	proxyNodeCount := 0
	req.NodeCount = nodeCount
	primaryNodeCount := getPrimaryNodeCount(nodesRequested)
	if primaryNodeCount > 1 {
		return fmt.Errorf("invalid nodes: HA instance can only have one primary node")
	}

	for i := 0; i < nodeCount; i++ {
		currentNode := nodesRequested[i]

		if currentNode.Properties.NodeType != "database" && currentNode.Properties.NodeType != "haproxy" {
			return fmt.Errorf("invalid node type: %s", currentNode.Properties.NodeType)
		}
		if currentNode.Properties.NodeType == "database" {
			if databaseNodeCount == 0 && primaryNodeCount == 0 && currentNode.Properties.Role == "" {
				currentNode.Properties.Role = "Primary"
			}
			databaseNodeCount++
			if currentNode.VmName == "" {
				defaultDatabaseName := database.GetAdditionalArguments()["cluster_name"] + "-" + strconv.Itoa(databaseNodeCount+1)
				currentNode.VmName = defaultDatabaseName
			}
			if currentNode.Properties.Role == "" {
				currentNode.Properties.Role = "Secondary"
			}
		}
		if currentNode.Properties.NodeType == "haproxy" {
			proxyNodeCount++
			if currentNode.VmName == "" {
				defaultDatabaseName := database.GetAdditionalArguments()["cluster_name"] + "_haproxy" + strconv.Itoa(proxyNodeCount+1)
				currentNode.VmName = defaultDatabaseName
			}
		}
		isPrimaryNode := currentNode.Properties.NodeType == "database" && currentNode.Properties.Role == "Primary"
		if isPrimaryNode {
			primaryNodeCount += 1
		}

		props := make([]map[string]string, 4)
		props[0] = map[string]string{
			"name":  "role",
			"value": currentNode.Properties.Role,
		}
		props[1] = map[string]string{
			"name":  "failover_mode",
			"value": currentNode.Properties.FailoverMode,
		}
		props[2] = map[string]string{
			"name":  "node_type",
			"value": currentNode.Properties.NodeType,
		}
		props[3] = map[string]string{
			"name":  "remove_archive_destination",
			"value": database.GetAdditionalArguments()["remove_archive_destination"],
		}
		req.Nodes = append(req.Nodes, Node{
			Properties:       props,
			VmName:           currentNode.VmName,
			NxClusterId:      database.GetClusterId(),
			NetworkProfileId: req.NetworkProfileId,
			ComputeProfileId: req.ComputeProfileId,
		})
	}
	const MinReqDatabaseNodes = 3
	if nodeCount < MinReqDatabaseNodes {
		return fmt.Errorf("invalid node count: HA instance needs at least %d nodes, given: %d", MinReqDatabaseNodes, nodeCount)
	}

	return nil
}

func createDefaultNodes(database DatabaseInterface) []*v1alpha1.Node {
	nodes := make([]*v1alpha1.Node, 0)
	nodes = append(nodes, &v1alpha1.Node{
		VmName: database.GetAdditionalArguments()["cluster_name"] + "_haproxy1",
		Properties: v1alpha1.NodeProperties{
			NodeType: "haproxy",
		},
	})
	nodes = append(nodes, &v1alpha1.Node{
		VmName: database.GetAdditionalArguments()["cluster_name"] + "_haproxy2",
		Properties: v1alpha1.NodeProperties{
			NodeType: "haproxy",
		},
	})
	nodes = append(nodes, &v1alpha1.Node{
		VmName: database.GetAdditionalArguments()["cluster_name"] + "-1",
		Properties: v1alpha1.NodeProperties{
			NodeType:     "database",
			Role:         "Primary",
			FailoverMode: "Automatic",
		},
	})
	nodes = append(nodes, &v1alpha1.Node{
		VmName: database.GetAdditionalArguments()["cluster_name"] + "-2",
		Properties: v1alpha1.NodeProperties{
			NodeType:     "database",
			Role:         "Secondary",
			FailoverMode: "Automatic",
		},
	})
	nodes = append(nodes, &v1alpha1.Node{
		VmName: database.GetAdditionalArguments()["cluster_name"] + "-3",
		Properties: v1alpha1.NodeProperties{
			NodeType:     "database",
			Role:         "Secondary",
			FailoverMode: "Automatic",
		},
	})
	return nodes
}

func getPrimaryNodeCount(nodesRequested []*v1alpha1.Node) int {
	count := 0
	for _, node := range nodesRequested {
		if node.Properties.Role == "Primary" {
			count++
		}
	}
	return count
}

func (a *PostgresHARequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	req.SSHPublicKey = reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	// Set the number of nodes to 5, 3 Postgres nodes + 2 HA Proxy nodes
	err := setNodesParameters(req, database)
	if err != nil {
		return nil, err
	}

	req.Clustered = true

	// Default action arguments
	actionArguments := defaultActionArgumentsforHAProvisioning(database, dbPassword, databaseNames)

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}
	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}

func defaultActionArgumentsforHAProvisioning(database DatabaseInterface, dbPassword string, databaseNames string) map[string]string {
	defaults := map[string]string{
		"failover_mode":           "Automatic",
		"proxy_read_port":         "5001",
		"listener_port":           "5432",
		"proxy_write_port":        "5000",
		"enable_synchronous_mode": "true",
		"auto_tune_staging_drive": "true",
		"backup_policy":           "primary_only",
		"provision_virtual_ip":    "true",
		"deploy_haproxy":          "true",
		"node_type":               "database",
		"allocate_pg_hugepage":    "false",
		"cluster_database":        "false",
		"archive_wal_expire_days": "-1",
		"enable_peer_auth":        "false",
		"cluster_name":            "psqlcluster",
		"patroni_cluster_name":    "patroni",
	}

	additionalArguments := map[string]string{
		"db_password":    dbPassword,
		"database_names": databaseNames,
	}
	originalAdditionalArguments := database.GetAdditionalArguments()
	for key, defaultValue := range defaults {
		value := originalAdditionalArguments[key]
		if value == "" {
			additionalArguments[key] = defaultValue
		} else {
			additionalArguments[key] = value
		}
	}

	return additionalArguments
}

func (a *MySqlRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error) {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)
	req.SSHPublicKey = SSHPublicKey

	// Default action arguments
	actionArguments := map[string]string{
		"listener_port":           "3306",
		"db_password":             dbPassword,
		"database_names":          databaseNames,
		"auto_tune_staging_drive": "true",
	}

	// Appending/overwriting database actionArguments to actionArguments
	if err := setConfiguredActionArguments(database, actionArguments); err != nil {
		return nil, err
	}

	// Converting action arguments map to list and appending to req.ActionArguments
	req.ActionArguments = append(req.ActionArguments, convertMapToActionArguments(actionArguments)...)

	return req, nil
}
