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
// The database provisioned has a NONE time machine SLA attached to it, and uses the default OOB profiles
func GenerateProvisioningRequest(ctx context.Context, ndb_client *ndb_client.NDBClient, database DatabaseInterface, reqData map[string]interface{}) (requestBody *DatabaseProvisionRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GenerateProvisioningRequest", "database name", database.GetDBInstanceName(), "database type", database.GetDBInstanceType())

	// Fetching the NONE TM SLA
	sla, err := GetNoneTimeMachineSLA(ctx, ndb_client)
	if err != nil {
		log.Error(err, "Error occurred while getting NONE TM SLA", "database name", database.GetDBInstanceName(), "database type", database.GetDBInstanceType())
		return
	}

	// Fetch the required profiles for the database
	profilesMap, err := ResolveProfiles(ctx, ndb_client, database.GetDBInstanceType(), database.GetProfileResolvers())
	if err != nil {
		log.Error(err, "Error occurred while getting required profiles", "database name", database.GetDBInstanceName(), "database type", database.GetDBInstanceType())
		return
	}
	// Required for dbParameterProfileIdInstance in MSSQL action args
	reqData[common.PROFILE_MAP_PARAM] = profilesMap

	// Validate request data
	err = validateReqData(ctx, database.GetDBInstanceType(), reqData)
	if err != nil {
		log.Error(err, "Error occurred while validating reqData", "reqData", reqData)
		return
	}
	// Creating a provisioning request based on the database type
	requestBody = &DatabaseProvisionRequest{
		DatabaseType:             GetDatabaseEngineName(database.GetDBInstanceType()),
		Name:                     database.GetDBInstanceName(),
		DatabaseDescription:      "Database provisioned by ndb-operator: " + database.GetDBInstanceName(),
		SoftwareProfileId:        profilesMap[common.PROFILE_TYPE_SOFTWARE].Id,
		SoftwareProfileVersionId: profilesMap[common.PROFILE_TYPE_SOFTWARE].LatestVersionId,
		ComputeProfileId:         profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
		NetworkProfileId:         profilesMap[common.PROFILE_TYPE_NETWORK].Id,
		DbParameterProfileId:     profilesMap[common.PROFILE_TYPE_DATABASE_PARAMETER].Id,
		NewDbServerTimeZone:      database.GetDBInstanceTimeZone(),
		CreateDbServer:           true,
		NodeCount:                1,
		NxClusterId:              database.GetNDBClusterId(),
		Clustered:                false,
		AutoTuneStagingDrive:     true,

		TimeMachineInfo: TimeMachineInfo{
			Name:             database.GetDBInstanceName() + "_TM",
			Description:      sla.Description,
			SlaId:            sla.Id,
			Schedule:         make(map[string]string),
			Tags:             make([]string, 0),
			AutoTuneLogDrive: true,
		},
		Nodes: []Node{
			{
				Properties: make([]string, 0),
				VmName:     database.GetDBInstanceName() + "_VM",
			},
		},
		ActionArguments: []ActionArgument{
			{
				Name:  "dbserver_description",
				Value: "dbserver for " + database.GetDBInstanceName(),
			},
			{
				Name:  "database_size",
				Value: strconv.Itoa(database.GetDBInstanceSize()),
			},
		},
	}
	// Appending request body based on database type
	appender, err := GetDbProvRequestAppender(database.GetDBInstanceType())
	if err != nil {
		log.Error(err, "Error while appending provisioning request")
		return
	}
	requestBody = appender.appendRequest(requestBody, database, reqData)

	log.Info("Database Provisioning", "requestBody", requestBody)
	log.Info("Returning from ndb_api.GenerateProvisioningRequest", "database name", database.GetDBInstanceName(), "database type", database.GetDBInstanceType())
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

// Appends request based on database type
type DBProvisionRequestAppender interface {
	appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest
}

type MSSQLProvisionRequestAppender struct{}

type MongoDbProvisionRequestAppender struct{}

type PostgresProvisionRequestAppender struct{}

type MySqlProvisionRequestAppender struct{}

func (a *MSSQLProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	req.DatabaseName = string(database.GetDBInstanceDatabaseNames())
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
			Value: database.GetDBInstanceName(),
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

func (a *MongoDbProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
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

func (a *PostgresProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
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

func (a *MySqlProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
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

// Get specific implementation of the DBProvisionRequestAppender interface based on the provided databaseType
func GetDbProvRequestAppender(databaseType string) (requestAppender DBProvisionRequestAppender, err error) {
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		requestAppender = &MySqlProvisionRequestAppender{}
	case common.DATABASE_TYPE_POSTGRES:
		requestAppender = &PostgresProvisionRequestAppender{}
	case common.DATABASE_TYPE_MONGODB:
		requestAppender = &MongoDbProvisionRequestAppender{}
	case common.DATABASE_TYPE_MSSQL:
		requestAppender = &MSSQLProvisionRequestAppender{}
	default:
		return nil, errors.New("invalid database type: supported values: mssql, mysql, postgres, mongodb")
	}
	return
}
