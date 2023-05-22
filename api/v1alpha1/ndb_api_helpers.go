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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	"github.com/nutanix-cloud-native/ndb-operator/util"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB
// The database provisioned has a NONE time machine SLA attached to it, and uses the default OOB profiles
func GenerateProvisioningRequest(ctx context.Context, ndbclient *ndbclient.NDBClient, dbSpec DatabaseSpec, reqData map[string]interface{}) (requestBody *DatabaseProvisionRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api_helpers.GenerateProvisioningRequest", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)

	// Fetching the NONE TM SLA
	sla, err := GetNoneTimeMachineSLA(ctx, ndbclient)
	if err != nil {
		log.Error(err, "Error occurred while getting NONE TM SLA", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	// Fetch the OOB profiles for the database
	profilesMap, err := GetProfiles(ctx, ndbclient, dbSpec.Instance)
	if err != nil {
		log.Error(err, "Error occurred while getting OOB profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	database_names := strings.Join(dbSpec.Instance.DatabaseNames, ",")

	// Type assertion
	dbPassword, ok := reqData[NDB_PARAM_PASSWORD].(string)
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
	SSHPublicKey, ok := reqData[NDB_PARAM_SSH_PUBLIC_KEY].(string)
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
		SoftwareProfileId:        profilesMap[PROFILE_TYPE_SOFTWARE].Id,
		SoftwareProfileVersionId: profilesMap[PROFILE_TYPE_SOFTWARE].LatestVersionId,
		ComputeProfileId:         profilesMap[PROFILE_TYPE_COMPUTE].Id,
		NetworkProfileId:         profilesMap[PROFILE_TYPE_NETWORK].Id,
		DbParameterProfileId:     profilesMap[PROFILE_TYPE_DATABASE_PARAMETER].Id,
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

	requestBody.ActionArguments = append(requestBody.ActionArguments, dbTypeActionArgs.GetActionArguments(dbSpec)...)

	log.Info("Database Provisioning", "requestBody", requestBody)
	log.Info("Returning from ndb_api_helpers.GenerateProvisioningRequest", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
	return
}

// Fetches all the SLAs from the ndb and returns the NONE TM SLA.
// Returns an error if not found.
func GetNoneTimeMachineSLA(ctx context.Context, ndbclient *ndbclient.NDBClient) (sla SLAResponse, err error) {
	slas, err := GetAllSLAs(ctx, ndbclient)
	if err != nil {
		return
	}
	for _, s := range slas {
		if s.Name == SLA_NAME_NONE {
			sla = s
			return
		}
	}
	return sla, fmt.Errorf("NONE TimeMachine not found")
}

// +kubebuilder:object:generate:=false
type ProfileResolver interface {
	Resolve(ctx context.Context, allProfiles []ProfileResponse, pType string, dbEngine string, filter func(p ProfileResponse) bool) (profile ProfileResponse, err error)
}

func (inputProfile *Profile) Resolve(ctx context.Context, allProfiles []ProfileResponse, pType string, dbType string, filter func(p ProfileResponse) bool) (profile ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api_helpers.resolve", "input profile", inputProfile)

	name, id := inputProfile.Name, inputProfile.Id
	isNameProvided, isIdProvided := name != "", id != ""

	var profileByName, profileById ProfileResponse

	// validation of software profile for closed-source db engines
	isClosedSourceEngine := (dbType == DATABASE_TYPE_ORACLE) || (dbType == DATABASE_TYPE_SQLSERVER)

	if pType == PROFILE_TYPE_SOFTWARE && isClosedSourceEngine {
		if !isNameProvided && !isIdProvided {
			log.Error(errors.New("software profile not provided for closed-source database"), "Provide software profile info", "dbType", dbType)
			return ProfileResponse{}, fmt.Errorf("software profile is a mandatory input for %s type of database", dbType)
		}
	}

	// resolve the profile based on provided input, return an error if not resolved
	if isNameProvided {
		profileByName, err = util.FindFirst(allProfiles, func(pr ProfileResponse) bool {
			return pr.Name == name
		})
	}

	if isIdProvided && err == nil {
		profileById, err = util.FindFirst(allProfiles, func(pr ProfileResponse) bool {
			return pr.Id == id
		})
	}

	if err != nil {
		log.Error(err, "could not resolve profile by id or name", "id", id, "name", name)
		return ProfileResponse{}, fmt.Errorf("could not resolve profile by id=%v or name=%v", id, name)
	}

	/*
		1. if both name & id NOT provided -> resolve the OOB profile
		2. if both name & id provided -> Resolve by both & ensure that both resolved profiles are same
		3. if only id provided -> resolve by id
		4. if only name provided -> resolve by name
		5. else throw an error
	*/
	if !isNameProvided && !isIdProvided { // OOB
		log.Info("Attempting to resolve the OOB profile, no id or name provided in the spec", "Profile", pType)
		oobProfile, err := util.FindFirst(allProfiles, filter)

		if err != nil {
			log.Error(err, "Error resolving OOB Profile", "type", pType)
			return ProfileResponse{}, fmt.Errorf("no OOB profile found of type=%v", pType)
		}
		return oobProfile, nil

	} else if isNameProvided && isIdProvided { // verify that both resolved profiles (by id and name) are one and the same
		if !reflect.DeepEqual(profileById, profileByName) {
			log.Error(err, "profile matching both the given name & id does not exist. Retry with correct inputs")
			return ProfileResponse{}, fmt.Errorf("profiles returned by id & name resolve to different profiles")
		}
		return profileById, nil

	} else if isIdProvided {
		return profileById, nil

	} else if isNameProvided {
		return profileByName, nil
	}

	return ProfileResponse{}, fmt.Errorf("could not resolve the profile by Name or Id, err=%v", err)
}

// TODO: Once the database_types refactoring is over, move out below methods to  profile_helper.go
var ComputeOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == PROFILE_TYPE_COMPUTE &&
		strings.EqualFold(p.Name, DEFAULT_OOB_SMALL_COMPUTE)
}

var SoftwareOOBProfileResolverForSingleInstance = func(p ProfileResponse) bool {
	return p.Type == PROFILE_TYPE_SOFTWARE && p.Topology == TOPOLOGY_SINGLE
}

var NetworkOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == PROFILE_TYPE_NETWORK
}

var DbParamOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == PROFILE_TYPE_DATABASE_PARAMETER
}

// Fetches all the profiles and returns a map of profiles
// Returns an error if any profile is not found
func GetProfiles(ctx context.Context, ndbclient *ndbclient.NDBClient, instanceSpec Instance) (profilesMap map[string]ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)
	inputProfiles := instanceSpec.Profiles
	log.Info("Entered ndb_api_helpers.GetProfiles", "Input profiles", inputProfiles)

	allProfiles, err := GetAllProfiles(ctx, ndbclient)
	activeProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Status == PROFILE_STATUS_READY })
	if err != nil {
		log.Error(err, "Profiles could not be fetched")
		return
	}

	dbEngineSpecific := util.Filter(activeProfiles, func(p ProfileResponse) bool { return p.EngineType == GetDatabaseEngineName(instanceSpec.Type) })

	// Compute Profile
	compute, err := inputProfiles.Compute.Resolve(ctx, activeProfiles, PROFILE_TYPE_COMPUTE, instanceSpec.Type, ComputeOOBProfileResolver)
	if err != nil {
		log.Error(err, "Compute Profile could not be resolved", "Input Profile", inputProfiles.Compute)
		return nil, err
	}

	// Software Profile
	software, err := inputProfiles.Software.Resolve(ctx, dbEngineSpecific, PROFILE_TYPE_SOFTWARE, instanceSpec.Type, SoftwareOOBProfileResolverForSingleInstance)
	if err != nil {
		log.Error(err, "Software Profile could not be resolved or is not in READY state", "Input Profile", inputProfiles.Software)
		return nil, err
	}

	// Network Profile
	network, err := inputProfiles.Network.Resolve(ctx, dbEngineSpecific, PROFILE_TYPE_NETWORK, instanceSpec.Type, NetworkOOBProfileResolver)
	if err != nil {
		log.Error(err, "Network Profile could not be resolved", "Input Profile", inputProfiles.Network)
		return nil, err
	}

	// DB Param Profile
	dbParam, err := inputProfiles.DbParam.Resolve(ctx, dbEngineSpecific, PROFILE_TYPE_DATABASE_PARAMETER, instanceSpec.Type, DbParamOOBProfileResolver)
	if err != nil {
		log.Error(err, "DbParam Profile could not be resolved", "Input Profile", inputProfiles.DbParam)
		return nil, err
	}

	profilesMap = map[string]ProfileResponse{
		PROFILE_TYPE_COMPUTE:            compute,
		PROFILE_TYPE_SOFTWARE:           software,
		PROFILE_TYPE_NETWORK:            network,
		PROFILE_TYPE_DATABASE_PARAMETER: dbParam,
	}

	log.Info("Generated", "profiles map", profilesMap)
	return
}

func GetDatabaseEngineName(dbType string) string {
	switch dbType {
	case DATABASE_TYPE_POSTGRES:
		return DATABASE_ENGINE_TYPE_POSTGRES
	case DATABASE_TYPE_MYSQL:
		return DATABASE_ENGINE_TYPE_MYSQL
	case DATABASE_TYPE_MONGODB:
		return DATABASE_ENGINE_TYPE_MONGODB
	case DATABASE_TYPE_GENERIC:
		return DATABASE_ENGINE_TYPE_GENERIC
	default:
		return ""
	}
}

func GetDatabasePortByType(dbType string) int32 {
	switch dbType {
	case DATABASE_TYPE_POSTGRES:
		return DATABASE_DEFAULT_PORT_POSTGRES
	case DATABASE_TYPE_MONGODB:
		return DATABASE_DEFAULT_PORT_MONGODB
	case DATABASE_TYPE_MYSQL:
		return DATABASE_DEFAULT_PORT_MYSQL
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
	case DATABASE_TYPE_MYSQL:
		dbTypeActionArgs = &MysqlActionArgs{}
	case DATABASE_TYPE_POSTGRES:
		dbTypeActionArgs = &PostgresActionArgs{}
	case DATABASE_TYPE_MONGODB:
		dbTypeActionArgs = &MongodbActionArgs{}
	default:
		return nil, errors.New("invalid database type: supported values: mysql, postgres, mongodb")
	}
	return dbTypeActionArgs, nil
}

func (m *MysqlActionArgs) GetActionArguments(dbSpec DatabaseSpec) []ActionArgument {
	return []ActionArgument{
		{
			Name:  "listener_port",
			Value: "3306",
		},
	}
}

func (p *PostgresActionArgs) GetActionArguments(dbSpec DatabaseSpec) []ActionArgument {
	return []ActionArgument{
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
	}
}

func (m *MongodbActionArgs) GetActionArguments(dbSpec DatabaseSpec) []ActionArgument {
	return []ActionArgument{
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
	}
}
