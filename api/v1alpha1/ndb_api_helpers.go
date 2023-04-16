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
	profilesMap, err := EnrichAndGetProfiles(ctx, ndbclient, dbSpec.Instance.Type, dbSpec.Instance.Profiles)
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
		ActionArguments: []ActionArgument{
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
				Name:  "database_size",
				Value: strconv.Itoa(dbSpec.Instance.Size),
			},
			{
				Name:  "auto_tune_staging_drive",
				Value: "true",
			},
			{
				Name:  "enable_synchronous_mode",
				Value: "false",
			},
			{
				Name:  "backup_policy",
				Value: "primary_only",
			},
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
		},
		Nodes: []Node{
			{
				Properties: make([]string, 0),
				VmName:     dbSpec.Instance.DatabaseInstanceName + "_VM",
			},
		},
	}
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

/*
	 Fetches all the profiles and returns a map of OOB (default) profiles and overrided them with custom profiles if provided
	 Returns an error if:
		 (a) one or more OOB profiles (default) is not found
	 	 (b) custom profile provided does not match any of the profiles from GetAllProfiles()
*/
func EnrichAndGetProfiles(ctx context.Context, ndbclient *ndbclient.NDBClient, dbType string, profiles Profiles) (profileMap map[string]ProfileResponse, err error) {

	//log := ctrllog.FromContext(ctx)

	// Map of profile type to profiles
	profileMap = make(map[string]ProfileResponse)

	allProfiles, err := GetAllProfiles(ctx, ndbclient)

	if err != nil {
		return
	}

	genericProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.EngineType == DATABASE_ENGINE_TYPE_GENERIC })
	dbEngineSpecificProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.EngineType == GetDatabaseEngineName(dbType) })

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

	// performs overriding of default OOB profiles based on the profiles obtained through YAML
	err = EnrichProfilesMap(ctx, profiles, allProfiles, genericProfiles, dbEngineSpecificProfiles, profileMap)

	return
}

/*
This function takes customProfiles, generic & dbEngineSpecific profiles & validates the entry of custom profile provided through YAML.
If checks pass, the custom profile values are used for overriding the default behaviour/profile value which are used for database provisioning
*/
func EnrichProfilesMap(ctx context.Context, profiles Profiles, allProfiles []ProfileResponse, genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse, profilesMap map[string]ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	if profiles == (Profiles{}) {
		log.Info("Defaulting to using OOB Profiles as no custom profiles received. Returning from enrichingProfilesMap")
		return
	}
	log.Info("Received Custom Profiles => ", "Received Custom Profiles", profiles)
	customProfileOptions := [...]string{PROFILE_TYPE_COMPUTE, PROFILE_TYPE_SOFTWARE, PROFILE_TYPE_NETWORK, PROFILE_TYPE_DATABASE_PARAMETER}
	for _, profileStringValue := range customProfileOptions {
		err = PerformProfileMatchingAndEnrichProfiles(ctx, profileStringValue, profiles, allProfiles, genericProfiles, dbEngineSpecificProfiles, profilesMap)
		if err != nil {
			return
		}
	}
	return
}

/*
Based on compute or (software, network & dbParam), generic or dbEngineSpecific profiles are used for matching the input customProfile
Furthermore, based on whether matched or not matched, delegation is performed to override the default profile values
*/
func PerformProfileMatchingAndEnrichProfiles(ctx context.Context, profileType string, profiles Profiles, allProfiles []ProfileResponse, genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse, profilesMap map[string]ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	var matchedProfile []ProfileResponse
	var versionMatchedArr []Version
	profile := GetProfileForType(profileType, profiles)
	if !isEmptyProfile(profile) {
		//profileToUseForMatching := GetAppropriateProfileForType(profileType, genericProfiles, dbEngineSpecificProfiles)
		log.Info("Performing profile matching for profileType => ", "profileType", profileType)
		matchedProfile = util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Id == profile.Id })
		finalProfile := matchedProfile[0]
		versions := finalProfile.Versions
		versionMatchedArr = util.Filter(versions, func(versions Version) bool { return versions.Id == profile.VersionId })
		if len(versionMatchedArr) > 0 {
			log.Info("Id and Version matched for profileType", "profileType", profileType)
			fmt.Println("*****************************************************************************")
			fmt.Println(versionMatchedArr)
			fmt.Println("*****************************************************************************")
			finalProfile.LatestVersionId = profile.VersionId
		}
	}
	err = EnrichProfileMapForProfileType(ctx, profilesMap, profileType, matchedProfile)
	return
}

/*
This function checks the correctness of the profile (response) passed as the parameter
and overrides the profilesMap for the custom profile type specified if the custom profile provided passes the checks
*/
func EnrichProfileMapForProfileType(ctx context.Context, profileMap map[string]ProfileResponse, profileType string, response []ProfileResponse) (err error) {
	log := ctrllog.FromContext(ctx)
	if len(response) == 0 {
		err = fmt.Errorf("No matching profile found for profileType = %s", profileType)
		log.Info("Error Occurred. No enrichment performed for profile => ", "profileType", profileType)
		return
	}
	log.Info("Profile Matching succeeded for profileType => ", "profileType", profileType)
	log.Info("Going to perform custom profile enrichment performed for => ", "profileType", profileType)
	profileMap[profileType] = response[0]
	return
}

// This functions gives either generic or dbEngine specific profiles based upon the profile type to be filtered upon
func GetAppropriateProfileForType(profileType string, genericProfiles []ProfileResponse, dbEngineSpecificProfiles []ProfileResponse) (profiles []ProfileResponse) {
	if profileType == PROFILE_TYPE_COMPUTE {
		return genericProfiles
	} else {
		return dbEngineSpecificProfiles
	}
}

/*
Providing the least costly topology based on each profile type
NOTE: Values need to verified / Can be changes after having further discussions!
*/
func GetTopologyForProfileType(profileType string) string {
	switch profileType {
	case PROFILE_TYPE_COMPUTE:
		return TOPOLOGY_ALL
	case PROFILE_TYPE_SOFTWARE:
		return TOPOLOGY_SINGLE
	case PROFILE_TYPE_NETWORK:
		return TOPOLOGY_ALL
	case PROFILE_TYPE_DATABASE_PARAMETER:
		return TOPOLOGY_INSTANCE
	default:
		return ""
	}
}

func GetProfileForType(profileType string, profiles Profiles) Profile {
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
