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
	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	"github.com/nutanix-cloud-native/ndb-operator/util"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"strings"
)

// This function generates and returns a request for provisioning a database (and a dbserver vm) on NDB
// The database provisioned has a NONE time machine SLA attached to it, and uses the default OOB profiles
func GenerateProvisioningRequest(ctx context.Context, ndbclient *ndbclient.NDBClient, dbSpec DatabaseSpec, reqData map[string]interface{}) (req *DatabaseProvisionRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api_helpers.GenerateProvisioningRequest", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)

	// Fetching the NONE TM SLA
	sla, err := GetNoneTimeMachineSLA(ctx, ndbclient)
	if err != nil {
		log.Error(err, "Error occurred while getting NONE TM SLA", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	// Fetch upto date profiles for the database
	profilesMap, err := GetProfiles(ctx, ndbclient, dbSpec.Instance.Type, dbSpec.Instance.Profiles)
	if err != nil {
		log.Error(err, "Error occurred while enriching and getting profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
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
	req = &DatabaseProvisionRequest{
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

	req.ActionArguments = append(req.ActionArguments, dbTypeActionArgs.GetActionArguments(dbSpec)...)

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

// This method populates the profileMap with all the values for the profileType available
func GetProfiles(ctx context.Context, ndbclient *ndbclient.NDBClient, dbType string, profiles Profiles) (profileMap map[string]ProfileResponse, err error) {

	log := ctrllog.FromContext(ctx)

	// Map of profile type to profiles
	profileMap = make(map[string]ProfileResponse)

	allProfiles, err := GetAllProfiles(ctx, ndbclient)

	err = GetOOBProfiles(profileMap, allProfiles, dbType)
	if err != nil {
		return
	}

	log.Info("Received Input Profiles = ", "Received Input Profiles", profiles)

	customProfileOptions := [...]string{PROFILE_TYPE_COMPUTE, PROFILE_TYPE_SOFTWARE, PROFILE_TYPE_NETWORK, PROFILE_TYPE_DATABASE_PARAMETER}
	for _, profileValue := range customProfileOptions {
		err = MatchAndSetProfile(ctx, profileValue, profiles.Compute, allProfiles, profileMap)
		if err != nil {
			return
		}
	}

	return

}

/*
This function does:
(a) Id level matching with profiles in allProfiles
(b) If Id level match is successful, flow proceeds to match based on versionId

	When matched, the latestVersionId is overridden with the versionId as it is this attribute while dbProvisioning which is used forprofileType versionId
*/
func MatchAndSetProfile(ctx context.Context, profileType string, profile Profile, allProfiles []ProfileResponse, profilesMap map[string]ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	if isEmptyProfile(profile) {
		log.Info("Defaulting to OOB profile population for profileType = ", "profileType", profileType)
		return
	}
	var matchedProfile []ProfileResponse
	var matchedProfileByVersionId []Version
	matchedProfile = util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Id == profile.Id && p.Type == profileType })
	if len(matchedProfile) > 0 {
		log.Info("Id Matched for profileType", "profileType", profileType)
		matchedProfileByVersionId = util.Filter(matchedProfile[0].Versions, func(versions Version) bool { return versions.Id == profile.VersionId })
		if len(matchedProfileByVersionId) > 0 {
			log.Info("VersionId Matched for profileType", "profileType", profileType)
			matchedProfile[0].LatestVersionId = profile.VersionId
		}
		profilesMap[profileType] = matchedProfile[0]
		return
	}
	err = fmt.Errorf("No matching profile found for profileType = %s", profileType)
	return err
}

// fetches and sets the default profile for each profileType
func GetOOBProfiles(profileMap map[string]ProfileResponse, profiles []ProfileResponse, dbType string) (err error) {

	genericProfiles := util.Filter(profiles, func(p ProfileResponse) bool { return p.EngineType == DATABASE_ENGINE_TYPE_GENERIC })
	dbEngineSpecificProfiles := util.Filter(profiles, func(p ProfileResponse) bool { return p.EngineType == GetDatabaseEngineName(dbType) })

	// Specifying the usage of small compute profiles
	computeProfiles := util.Filter(genericProfiles, func(p ProfileResponse) bool {
		return p.Type == PROFILE_TYPE_COMPUTE && strings.Contains(strings.ToLower(p.Name), "small")
	})
	storageProfiles := util.Filter(genericProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_STORAGE })
	// Specifying the usage of single instance topoplogy
	softwareProfiles := util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_SOFTWARE && p.Topology == TOPOLOGY_SINGLE })
	networkProfiles := util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_NETWORK })
	dbParamProfiles := util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_DATABASE_PARAMETER })

	// Some profile not found, return an error
	if len(computeProfiles) == 0 || len(softwareProfiles) == 0 || len(storageProfiles) == 0 || len(networkProfiles) == 0 || len(dbParamProfiles) == 0 {
		err = errors.New("oob profile: one or more OOB profile(s) were not found")
		return
	}

	profileMap[PROFILE_TYPE_COMPUTE] = computeProfiles[0]
	profileMap[PROFILE_TYPE_STORAGE] = storageProfiles[0]
	profileMap[PROFILE_TYPE_SOFTWARE] = softwareProfiles[0]
	profileMap[PROFILE_TYPE_NETWORK] = networkProfiles[0]
	profileMap[PROFILE_TYPE_DATABASE_PARAMETER] = dbParamProfiles[0]
	return
}

func GetProfileByType(profileType string, profiles Profiles) Profile {
	defaultEmptyProfile := Profile{}
	switch profileType {
	case PROFILE_TYPE_COMPUTE:
		return profiles.Compute
	case PROFILE_TYPE_SOFTWARE:
		return profiles.Software
	case PROFILE_TYPE_NETWORK:
		return profiles.Network
	case PROFILE_TYPE_DATABASE_PARAMETER:
		return profiles.DbParam
	default:
		return defaultEmptyProfile
	}
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

func isEmptyProfile(customProfile Profile) (isEmpty bool) {
	isEmpty = customProfile == (Profile{})
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
		return nil, errors.New("Invalid Database Type: supported values: mysql, postgres, mongodb")
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
			Value: dbSpec.Instance.DatabaseInstanceName,
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
	}
}
