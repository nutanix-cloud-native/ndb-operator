/*
Copyright 2021-2022 Nutanix, Inc.

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
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB
// The database provisioned has a NONE time machine SLA attached to it, and uses the default OOB profiles
func GenerateProvisioningRequest(ctx context.Context, ndb_client *ndb_client.NDBClient, dbSpec v1alpha1.DatabaseSpec, reqData map[string]interface{}) (requestBody *DatabaseProvisionRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api_helpers.GenerateProvisioningRequest", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)

	// Fetching the NONE TM SLA
	sla, err := GetNoneTimeMachineSLA(ctx, ndb_client)
	if err != nil {
		log.Error(err, "Error occurred while getting NONE TM SLA", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	// Fetch the OOB profiles for the database
	profilesMap, err := GetProfiles(ctx, ndb_client, dbSpec.Instance)
	if err != nil {
		log.Error(err, "Error occurred while getting OOB profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	database_names := strings.Join(dbSpec.Instance.DatabaseNames, ",")

	// Type assertion
	dbPassword, ok := reqData[common.NDB_PARAM_PASSWORD].(string)
	if !ok || dbPassword == "" {
		err = errors.New("invalid database password")
		var errStatement string
		if !ok {
			errStatement = "Type assertion failed for database password. Expected a string value"
		} else {
			errStatement = "Empty database password"
		}
		log.Error(err, errStatement)
	}
	// Type assertion
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
	}
	// Creating a provisioning request based on the database type
	requestBody = &DatabaseProvisionRequest{
		DatabaseType:             GetDatabaseEngineName(dbSpec.Instance.Type),
		Name:                     dbSpec.Instance.DatabaseInstanceName,
		DatabaseDescription:      "Database provisioned by ndb-operator: " + dbSpec.Instance.DatabaseInstanceName,
		SoftwareProfileId:        profilesMap[common.PROFILE_TYPE_SOFTWARE].Id,
		SoftwareProfileVersionId: profilesMap[common.PROFILE_TYPE_SOFTWARE].LatestVersionId,
		ComputeProfileId:         profilesMap[common.PROFILE_TYPE_COMPUTE].Id,
		NetworkProfileId:         profilesMap[common.PROFILE_TYPE_NETWORK].Id,
		DbParameterProfileId:     profilesMap[common.PROFILE_TYPE_DATABASE_PARAMETER].Id,
		NewDbServerTimeZone:      dbSpec.Instance.TimeZone,
		CreateDbServer:           true,
		NodeCount:                1,
		NxClusterId:              dbSpec.NDB.ClusterId,
		SSHPublicKey:             SSHPublicKey,
		Clustered:                false,
		AutoTuneStagingDrive:     true,

		TimeMachineInfo: TimeMachineInfo{
			Name:             dbSpec.Instance.DatabaseInstanceName + "_TM",
			Description:      sla.Description,
			SlaId:            sla.Id,
			Schedule:         make(map[string]string),
			Tags:             make([]string, 0),
			AutoTuneLogDrive: true,
		},
		Nodes: []Node{
			{
				Properties: make([]string, 0),
				VmName:     dbSpec.Instance.DatabaseInstanceName + "_VM",
			},
		},
		ActionArguments: []ActionArgument{
			{
				Name:  "dbserver_description",
				Value: "dbserver for " + dbSpec.Instance.DatabaseInstanceName,
			},
			{
				Name:  "database_names",
				Value: database_names,
			},
			{
				Name:  "db_password",
				Value: dbPassword,
			},
			{
				Name:  "database_size",
				Value: strconv.Itoa(dbSpec.Instance.Size),
			},
		},
	}
	// Setting action arguments based on database type
	dbTypeActionArgs, err := GetActionArgumentsByDatabaseType(dbSpec.Instance.Type)

	if err != nil {
		log.Error(err, "Error occurred while getting dbTypeActionArgs", "database type", dbSpec.Instance.Type)
		return
	}

	requestBody.ActionArguments = append(requestBody.ActionArguments, dbTypeActionArgs.Get(dbSpec)...)

	log.Info("Database Provisioning", "requestBody", requestBody)
	log.Info("Returning from ndb_api_helpers.GenerateProvisioningRequest", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
	return
}

// Fetches all the SLAs from the ndb and returns the NONE TM SLA.
// Returns an error if not found.
func GetNoneTimeMachineSLA(ctx context.Context, ndb_client *ndb_client.NDBClient) (sla SLAResponse, err error) {
	slas, err := GetAllSLAs(ctx, ndb_client)
	if err != nil {
		return
	}
	for _, s := range slas {
		if s.Name == common.SLA_NAME_NONE {
			sla = s
			return
		}
	}
	return sla, fmt.Errorf("NONE TimeMachine not found")
}

// Fetches all the profiles and returns a map of profiles
// Returns an error if any profile is not found
func GetProfiles(ctx context.Context, ndb_client *ndb_client.NDBClient, instanceSpec v1alpha1.Instance) (profilesMap map[string]ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)
	inputProfiles := instanceSpec.Profiles
	log.Info("Entered ndb_api_helpers.GetProfiles", "Input profiles", inputProfiles)

	allProfiles, err := GetAllProfiles(ctx, ndb_client)

	// profiles need to be in the ready state
	activeProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Status == common.PROFILE_STATUS_READY })
	if err != nil {
		log.Error(err, "Profiles could not be fetched")
		return
	}

	dbEngineSpecific := util.Filter(activeProfiles, func(p ProfileResponse) bool {
		return p.EngineType == GetDatabaseEngineName(instanceSpec.Type)
	})

	computeProfileInput := Profile{
		Profile:     inputProfiles.Compute,
		ProfileType: common.PROFILE_TYPE_COMPUTE,
	}
	softwareProfileInput := Profile{
		Profile:     inputProfiles.Software,
		ProfileType: common.PROFILE_TYPE_SOFTWARE,
	}
	networkProfileInput := Profile{
		Profile:     inputProfiles.Network,
		ProfileType: common.PROFILE_TYPE_NETWORK,
	}
	dbParamProfileInput := Profile{
		Profile:     inputProfiles.DbParam,
		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
	}

	// Compute Profile
	compute, err := computeProfileInput.Resolve(ctx, activeProfiles, ComputeOOBProfileResolver)
	if err != nil {
		log.Error(err, "Compute Profile could not be resolved", "Input Profile", inputProfiles.Compute)
		return nil, err
	}

	// Software Profile
	// validation of software profile for closed-source db engines
	isClosedSourceEngine := (instanceSpec.Type == common.DATABASE_TYPE_ORACLE) || (instanceSpec.Type == common.DATABASE_TYPE_SQLSERVER)
	if isClosedSourceEngine {
		if inputProfiles.Software.Id == "" && inputProfiles.Software.Name == "" {
			log.Error(errors.New("software profile not provided"), "Provide software profile info", "dbType", instanceSpec.Type)
			return nil, fmt.Errorf("software profile is a mandatory input for %s database", instanceSpec.Type)
		}
	}

	software, err := softwareProfileInput.Resolve(ctx, dbEngineSpecific, SoftwareOOBProfileResolverForSingleInstance)
	if err != nil {
		log.Error(err, "Software Profile could not be resolved or is not in READY state", "Input Profile", inputProfiles.Software)
		return nil, err
	}

	// Network Profile
	network, err := networkProfileInput.Resolve(ctx, dbEngineSpecific, NetworkOOBProfileResolver)
	if err != nil {
		log.Error(err, "Network Profile could not be resolved", "Input Profile", inputProfiles.Network)
		return nil, err
	}

	// DB Param Profile
	dbParam, err := dbParamProfileInput.Resolve(ctx, dbEngineSpecific, DbParamOOBProfileResolver)
	if err != nil {
		log.Error(err, "DbParam Profile could not be resolved", "Input Profile", inputProfiles.DbParam)
		return nil, err
	}

	profilesMap = map[string]ProfileResponse{
		common.PROFILE_TYPE_COMPUTE:            compute,
		common.PROFILE_TYPE_SOFTWARE:           software,
		common.PROFILE_TYPE_NETWORK:            network,
		common.PROFILE_TYPE_DATABASE_PARAMETER: dbParam,
	}

	log.Info("Returning from ndb_api_helpers.GetProfiles", "profiles map", profilesMap)
	return
}

func GetDatabaseEngineName(dbType string) string {
	switch dbType {
	case common.DATABASE_TYPE_POSTGRES:
		return common.DATABASE_ENGINE_TYPE_POSTGRES
	case common.DATABASE_TYPE_MYSQL:
		return common.DATABASE_ENGINE_TYPE_MYSQL
	case common.DATABASE_TYPE_MONGODB:
		return common.DATABASE_ENGINE_TYPE_MONGODB
	case common.DATABASE_TYPE_GENERIC:
		return common.DATABASE_ENGINE_TYPE_GENERIC
	default:
		return ""
	}
}

func GetDatabasePortByType(dbType string) int32 {
	switch dbType {
	case common.DATABASE_TYPE_POSTGRES:
		return common.DATABASE_DEFAULT_PORT_POSTGRES
	case common.DATABASE_TYPE_MONGODB:
		return common.DATABASE_DEFAULT_PORT_MONGODB
	case common.DATABASE_TYPE_MYSQL:
		return common.DATABASE_DEFAULT_PORT_MYSQL
	default:
		return 80
	}
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

// Returns a request to delete a database server vm
func GenerateDeprovisionDatabaseServerRequest() (req *DatabaseServerDeprovisionRequest) {
	req = &DatabaseServerDeprovisionRequest{
		Delete:            true,
		Remove:            false,
		SoftRemove:        false,
		DeleteVgs:         true,
		DeleteVmSnapshots: true,
	}
	return
}

// Returns action arguments based on the type of database
func GetActionArgumentsByDatabaseType(databaseType string) (DatabaseActionArgs, error) {
	var dbTypeActionArgs DatabaseActionArgs
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		dbTypeActionArgs = &MysqlActionArgs{}
	case common.DATABASE_TYPE_POSTGRES:
		dbTypeActionArgs = &PostgresActionArgs{}
	case common.DATABASE_TYPE_MONGODB:
		dbTypeActionArgs = &MongodbActionArgs{}
	default:
		return nil, errors.New("invalid database type: supported values: mysql, postgres, mongodb")
	}
	return dbTypeActionArgs, nil
}
