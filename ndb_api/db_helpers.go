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
	"strconv"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB
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

	schedule, err := database.GetInstanceTMSchedule()
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
				Properties: make([]string, 0),
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
	appender, err := GetRequestAppender(database.GetInstanceType())
	if err != nil {
		log.Error(err, "Error while appending provisioning request")
		return
	}
	requestBody = appender.appendProvisioningRequest(requestBody, database, reqData)

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

func (a *MSSQLRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	req.DatabaseName = string(database.GetInstanceDatabaseNames())
	adminPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	profileMap := reqData[common.PROFILE_MAP_PARAM].(map[string]ProfileResponse)
	dbParamInstanceProfile := profileMap[common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE]

	actionArgs := []ActionArgument{
		{
			Name:  "working_dir",
			Value: "C:\\temp",
		},
		{
			Name:  "sql_user_name",
			Value: "sa",
		},
		{
			Name:  "authentication_mode",
			Value: "windows",
		},
		{
			Name:  "delete_vm_on_failure",
			Value: "false",
		},
		{
			Name:  "is_gmsa_sql_service_account",
			Value: "false",
		},
		{
			Name:  "provision_from_backup",
			Value: "false",
		},
		{
			Name:  "distribute_database_data",
			Value: "true",
		},
		{
			Name:  "retain_database_in_restoring_mode",
			Value: "false",
		},
		{
			Name:  "dbserver_name",
			Value: database.GetName(),
		},
		{
			Name:  "server_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "database_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "dbParameterProfileIdInstance",
			Value: dbParamInstanceProfile.Id,
		},
		{
			Name:  "vm_dbserver_admin_password",
			Value: adminPassword,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)

	return req
}

func (a *MongoDbRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey
	actionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "27017",
		},
		{
			Name:  "log_size",
			Value: "100",
		},
		{
			Name:  "journal_size",
			Value: "100",
		},
		{
			Name:  "restart_mongod",
			Value: "true",
		},
		{
			Name:  "working_dir",
			Value: "/tmp",
		},
		{
			Name:  "db_user",
			Value: "admin",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
		{
			Name:  "database_names",
			Value: databaseNames,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req
}

func (a *PostgresRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey
	actionArgs := []ActionArgument{
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "listener_port",
			Value: "5432",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "false",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
		{
			Name:  "database_names",
			Value: databaseNames,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req
}

func (a *MySqlRequestAppender) appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey
	actionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "3306",
		},
		{
			Name:  "db_password",
			Value: dbPassword,
		},
		{
			Name:  "database_names",
			Value: databaseNames,
		},
	}

	req.ActionArguments = append(req.ActionArguments, actionArgs...)
	return req
}
