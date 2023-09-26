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

	// Fetching the TM details
	tmName, tmDescription, slaName := database.GetTMDetails()
	// Fetching the SLA for the TM by name
	sla, err := GetSLAByName(ctx, ndb_client, slaName)
	if err != nil {
		log.Error(err, "Error occurred while getting TM SLA", "SLA Name", slaName)
		return
	}

	schedule, err := database.GetTMSchedule()
	if err != nil {
		log.Error(err, "Error occurred while generating the Time Machine Schedule")
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

	// Fetching and appending ActionArguments
	actionArguments := database.GetDBInstanceTypeDetails()
	actionArguments = append(actionArguments, []ActionArgument{
		{
			Name:  "dbserver_description",
			Value: "dbserver for " + database.GetDBInstanceName(),
		},
		{
			Name:  "database_size",
			Value: strconv.Itoa(database.GetDBInstanceSize()),
		},
	}...)

	// Creating a provisioning request based on the database type
	requestBody = &DatabaseProvisionRequest{
		DatabaseType:             GetDatabaseEngineName(database.GetDBInstanceType()),
		Name:                     database.GetDBInstanceName(),
		DatabaseDescription:      database.GetDBInstanceDescription(),
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
				VmName:     database.GetDBInstanceName() + "_VM",
			},
		},
		ActionArguments: actionArguments,
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

// Appends defaultAction arguments and replacableActionArgs (if they are not specified in reqData)
func appendActionArguments(req *DatabaseProvisionRequest, replacableActionArgs []ActionArgument, defaultActionArgs []ActionArgument) {
	reqActionArgs := make(map[string]string)

	// Append reqActionArgs to map so we know which replacableActionArgs we should add
	for _, arg := range req.ActionArguments {
		reqActionArgs[arg.Name] = arg.Value
	}

	// Add actionArgument to req if it is not present
	for _, arg := range replacableActionArgs {
		if _, isPresent := reqActionArgs[arg.Name]; !isPresent {
			req.ActionArguments = append(req.ActionArguments, arg)
		}
	}

	// Append defaultActionArgs
	req.ActionArguments = append(req.ActionArguments, defaultActionArgs...)
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

	appendActionArguments(req, mssqlReplacableActionArgs(dbParamInstanceProfile.Id, adminPassword), mssqlDefaultActionArgs(database.GetDBInstanceName()))

	return req
}

func (a *MongoDbProvisionRequestAppender) appendRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) *DatabaseProvisionRequest {
	dbPassword := reqData[common.NDB_PARAM_PASSWORD].(string)
	databaseNames := database.GetDBInstanceDatabaseNames()
	SSHPublicKey := reqData[common.NDB_PARAM_SSH_PUBLIC_KEY].(string)

	req.SSHPublicKey = SSHPublicKey

	appendActionArguments(req, mongoDbReplacableActionArgs(), mongoDbDefaultActionArgs(dbPassword, databaseNames))

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
