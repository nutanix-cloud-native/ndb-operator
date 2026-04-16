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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Fetches all the profiles and returns a map of profiles
// Returns an error if any profile is not found
func ResolveProfiles(ctx context.Context, ndb_client ndb_client.NDBClientHTTPInterface, databaseType string, profileResolvers ProfileResolvers) (profilesMap map[string]ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)

	log.Info("Entered ndb_api.GetProfiles", "Input profiles", profileResolvers)

	allProfiles, err := GetAllProfiles(ctx, ndb_client)
	if err != nil {
		log.Error(err, "Profiles could not be fetched")
		return
	}

	// profiles need to be in the ready state
	activeProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Status == common.PROFILE_STATUS_READY })

	expectedEngine := GetDatabaseEngineName(databaseType)
	dbEngineSpecific := util.Filter(activeProfiles, func(p ProfileResponse) bool {
		return p.EngineType == expectedEngine
	})

	// #region agent log - H8: Check profile filtering
	func() {
		f, _ := os.OpenFile("/Users/sasikanth.masini/ndb-operator/.cursor/debug-8a3458.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			softwareProfiles := []map[string]string{}
			for _, p := range activeProfiles {
				if p.Type == common.PROFILE_TYPE_SOFTWARE {
					softwareProfiles = append(softwareProfiles, map[string]string{
						"id": p.Id, "name": p.Name, "engineType": p.EngineType, "status": p.Status,
					})
				}
			}
			filteredProfiles := []map[string]string{}
			for _, p := range dbEngineSpecific {
				if p.Type == common.PROFILE_TYPE_SOFTWARE {
					filteredProfiles = append(filteredProfiles, map[string]string{
						"id": p.Id, "name": p.Name, "engineType": p.EngineType,
					})
				}
			}
			payload := map[string]interface{}{
				"sessionId": "8a3458",
				"location":  "profile_helpers.go:profileFiltering",
				"message":   "Profile filtering for database type",
				"data": map[string]interface{}{
					"databaseType":             databaseType,
					"expectedEngineType":       expectedEngine,
					"totalActiveProfiles":      len(activeProfiles),
					"filteredProfiles":         len(dbEngineSpecific),
					"allSoftwareProfiles":      softwareProfiles,
					"filteredSoftwareProfiles": filteredProfiles,
				},
				"timestamp":    time.Now().UnixMilli(),
				"hypothesisId": "H8",
			}
			if b, e := json.Marshal(payload); e == nil {
				f.Write(b)
				f.WriteString("\n")
			}
		}
	}()
	// #endregion

	computeProfileResolver := profileResolvers[common.PROFILE_TYPE_COMPUTE]
	softwareProfileResolver := profileResolvers[common.PROFILE_TYPE_SOFTWARE]
	networkProfileResolver := profileResolvers[common.PROFILE_TYPE_NETWORK]
	dbParamProfileResolver := profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER]
	dbParamInstanceProfileResolver := profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE]

	// Compute Profile
	compute, err := computeProfileResolver.Resolve(ctx, activeProfiles, ComputeOOBProfileResolver)
	if err != nil {
		log.Error(err, "Compute Profile could not be resolved", "Input Profile", computeProfileResolver)
		return
	}

	// Software Profile
	// validation of software profile for closed-source db engines
	isClosedSourceEngine := (databaseType == common.DATABASE_TYPE_ORACLE) || (databaseType == common.DATABASE_TYPE_MSSQL)
	// #region agent log
	func() {
		f, _ := os.OpenFile("/Users/sasikanth.masini/ndb-operator/.cursor/debug-8a3458.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			defer f.Close()
			payload := map[string]interface{}{"sessionId": "8a3458", "location": "profile_helpers.go:softwareProfileCheck", "message": "Software profile validation", "data": map[string]interface{}{"databaseType": databaseType, "isClosedSource": isClosedSourceEngine, "profileId": softwareProfileResolver.GetId(), "profileName": softwareProfileResolver.GetName()}, "timestamp": time.Now().UnixMilli(), "hypothesisId": "H2"}
			if b, e := json.Marshal(payload); e == nil {
				f.Write(b)
				f.WriteString("\n")
			}
		}
	}()
	// #endregion
	if isClosedSourceEngine {
		if softwareProfileResolver.GetId() == "" && softwareProfileResolver.GetName() == "" {
			log.Error(errors.New("software profile not provided"), "Provide software profile info", "dbType", databaseType)
			err = fmt.Errorf("software profile is a mandatory input for %s database", databaseType)
			return
		}
	}

	software, err := softwareProfileResolver.Resolve(ctx, dbEngineSpecific, SoftwareOOBProfileResolverForSingleInstance)
	if err != nil {
		// #region agent log
		func() {
			f, _ := os.OpenFile("/Users/sasikanth.masini/ndb-operator/.cursor/debug-8a3458.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				defer f.Close()
				payload := map[string]interface{}{"sessionId": "8a3458", "location": "profile_helpers.go:softwareProfileError", "message": "Software profile resolution failed", "data": map[string]interface{}{"err": err.Error(), "profileName": softwareProfileResolver.GetName()}, "timestamp": time.Now().UnixMilli(), "hypothesisId": "H2"}
				if b, e := json.Marshal(payload); e == nil {
					f.Write(b)
					f.WriteString("\n")
				}
			}
		}()
		// #endregion
		log.Error(err, "Software Profile could not be resolved or is not in READY state", "Input Profile", softwareProfileResolver)
		return
	}

	// Network Profile
	network, err := networkProfileResolver.Resolve(ctx, dbEngineSpecific, NetworkOOBProfileResolver)
	if err != nil {
		log.Error(err, "Network Profile could not be resolved", "Input Profile", networkProfileResolver)
		return
	}

	// DB Param Profile
	dbParam, err := dbParamProfileResolver.Resolve(ctx, dbEngineSpecific, DbParamOOBProfileResolver)
	if err != nil {
		log.Error(err, "DbParam Profile could not be resolved", "Input Profile", dbParamProfileResolver)
		return
	}

	var dbParamInstance ProfileResponse
	// DB Param Instance Profile should only be resolved for mssql engine
	if databaseType == common.DATABASE_TYPE_MSSQL {
		dbParamInstance, err = dbParamInstanceProfileResolver.Resolve(ctx, dbEngineSpecific, DbParamInstanceOOBProfileResolver)
		if err != nil {
			log.Error(err, "Db Param Instance Profile could not be resolved", "Input Profile", dbParamInstanceProfileResolver)
			return
		}
	}

	profilesMap = map[string]ProfileResponse{
		common.PROFILE_TYPE_COMPUTE:                     compute,
		common.PROFILE_TYPE_SOFTWARE:                    software,
		common.PROFILE_TYPE_NETWORK:                     network,
		common.PROFILE_TYPE_DATABASE_PARAMETER:          dbParam,
		common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: dbParamInstance,
	}

	log.Info("Returning from ndb_api.GetProfiles", "profiles map", profilesMap)
	return
}

var ComputeOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == common.PROFILE_TYPE_COMPUTE && p.SystemProfile &&
		strings.EqualFold(p.Name, common.PROFILE_DEFAULT_OOB_SMALL_COMPUTE)
}

var SoftwareOOBProfileResolverForSingleInstance = func(p ProfileResponse) bool {
	return p.Type == common.PROFILE_TYPE_SOFTWARE && p.SystemProfile && p.Topology == common.TOPOLOGY_SINGLE
}

var NetworkOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == common.PROFILE_TYPE_NETWORK
}

var DbParamOOBProfileResolver = func(p ProfileResponse) bool {
	return p.SystemProfile && p.Type == common.PROFILE_TYPE_DATABASE_PARAMETER
}

var DbParamInstanceOOBProfileResolver = func(p ProfileResponse) bool {
	// The DB Instance profile has the topology type as "instance"
	return p.SystemProfile && p.Type == common.PROFILE_TYPE_DATABASE_PARAMETER && p.Topology == common.TOPOLOGY_INSTANCE
}
