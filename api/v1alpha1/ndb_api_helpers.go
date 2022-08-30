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
	"strconv"
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	"github.com/nutanix-cloud-native/ndb-operator/util"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func GenerateProvisioningRequest(ctx context.Context, ndbclient *ndbclient.NDBClient, dbSpec DatabaseSpec) (req *DatabaseProvisionRequest, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api_helpers.GenerateProvisioningRequest", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)

	sla, err := GetNoneTimeMachineSLA(ctx, ndbclient)
	if err != nil {
		log.Error(err, "Error occured while getting NONE TM SLA", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	profilesMap, err := GetOOBProfiles(ctx, ndbclient, dbSpec.Instance.Type)
	if err != nil {
		log.Error(err, "Error occured while getting OOB profiles", "database name", dbSpec.Instance.DatabaseInstanceName, "database type", dbSpec.Instance.Type)
		return
	}

	database_names := strings.Join(dbSpec.Instance.DatabaseNames, ",")

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
		SSHPublicKey:             dbSpec.NDB.Credentials.SSHPublicKey,
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
				Value: dbSpec.Instance.Password,
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

func GetOOBProfiles(ctx context.Context, ndbclient *ndbclient.NDBClient, dbType string) (profileMap map[string]ProfileResponse, err error) {
	profileMap = make(map[string]ProfileResponse)

	profiles, err := GetAllProfiles(ctx, ndbclient)

	if err != nil {
		return
	}

	genericProfiles := util.Filter(profiles, func(p ProfileResponse) bool { return p.EngineType == DATABASE_ENGINE_TYPE_GENERIC })
	dbEngineSpecificProfiles := util.Filter(profiles, func(p ProfileResponse) bool { return p.EngineType == GetDatabaseEngineName(dbType) })

	computeProfiles := util.Filter(genericProfiles, func(p ProfileResponse) bool {
		return p.Type == PROFILE_TYPE_COMPUTE && strings.Contains(strings.ToLower(p.Name), "small")
	})
	storageProfiles := util.Filter(genericProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_STORAGE })
	softwareProfiles := util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_SOFTWARE && p.Topology == TOPOLOGY_SINGLE })
	networkProfiles := util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_NETWORK })
	dbParamProfiles := util.Filter(dbEngineSpecificProfiles, func(p ProfileResponse) bool { return p.Type == PROFILE_TYPE_DATABASE_PARAMETER })

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

func GetDatabaseEngineName(dbType string) string {
	switch dbType {
	case "postgres":
		return DATABASE_ENGINE_TYPE_POSTGRES
	case "mysql":
		return DATABASE_ENGINE_TYPE_MYSQL
	case "mongodb":
		return DATABASE_ENGINE_TYPE_MONGODB
	case "generic":
		return DATABASE_ENGINE_TYPE_GENERIC
	default:
		return ""
	}
}

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
