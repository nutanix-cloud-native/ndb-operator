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

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Fetches all the profiles and returns a map of profiles
// Returns an error if any profile is not found
// func GetProfiles(ctx context.Context, ndb_client *ndb_client.NDBClient, instanceSpec v1alpha1.Instance) (profilesMap map[string]ProfileResponse, err error) {
// 	log := ctrllog.FromContext(ctx)
// 	inputProfiles := instanceSpec.Profiles
// 	log.Info("Entered ndb_api_helpers.GetProfiles", "Input profiles", inputProfiles)

// 	allProfiles, err := GetAllProfiles(ctx, ndb_client)

// 	// profiles need to be in the ready state
// 	activeProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Status == common.PROFILE_STATUS_READY })
// 	if err != nil {
// 		log.Error(err, "Profiles could not be fetched")
// 		return
// 	}

// 	dbEngineSpecific := util.Filter(activeProfiles, func(p ProfileResponse) bool {
// 		return p.EngineType == GetDatabaseEngineName(instanceSpec.Type)
// 	})

// 	var computeProfileResolver ProfileResolver = &Profile{
// 		Profile:     inputProfiles.Compute,
// 		ProfileType: common.PROFILE_TYPE_COMPUTE,
// 	}
// 	var softwareProfileResolver ProfileResolver = &Profile{
// 		Profile:     inputProfiles.Software,
// 		ProfileType: common.PROFILE_TYPE_SOFTWARE,
// 	}
// 	var networkProfileResolver ProfileResolver = &Profile{
// 		Profile:     inputProfiles.Network,
// 		ProfileType: common.PROFILE_TYPE_NETWORK,
// 	}
// 	var dbParamProfileResolver ProfileResolver = &Profile{
// 		Profile:     inputProfiles.DbParam,
// 		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
// 	}

// 	// Compute Profile
// 	compute, err := computeProfileResolver.Resolve(ctx, activeProfiles, ComputeOOBProfileResolver)
// 	if err != nil {
// 		log.Error(err, "Compute Profile could not be resolved", "Input Profile", inputProfiles.Compute)
// 		return nil, err
// 	}

// 	// Software Profile
// 	// validation of software profile for closed-source db engines
// 	isClosedSourceEngine := (instanceSpec.Type == common.DATABASE_TYPE_ORACLE) || (instanceSpec.Type == common.DATABASE_TYPE_SQLSERVER)
// 	if isClosedSourceEngine {
// 		if inputProfiles.Software.Id == "" && inputProfiles.Software.Name == "" {
// 			log.Error(errors.New("software profile not provided"), "Provide software profile info", "dbType", instanceSpec.Type)
// 			return nil, fmt.Errorf("software profile is a mandatory input for %s database", instanceSpec.Type)
// 		}
// 	}

// 	software, err := softwareProfileResolver.Resolve(ctx, dbEngineSpecific, SoftwareOOBProfileResolverForSingleInstance)
// 	if err != nil {
// 		log.Error(err, "Software Profile could not be resolved or is not in READY state", "Input Profile", inputProfiles.Software)
// 		return nil, err
// 	}

// 	// Network Profile
// 	network, err := networkProfileResolver.Resolve(ctx, dbEngineSpecific, NetworkOOBProfileResolver)
// 	if err != nil {
// 		log.Error(err, "Network Profile could not be resolved", "Input Profile", inputProfiles.Network)
// 		return nil, err
// 	}

// 	// DB Param Profile
// 	dbParam, err := dbParamProfileResolver.Resolve(ctx, dbEngineSpecific, DbParamOOBProfileResolver)
// 	if err != nil {
// 		log.Error(err, "DbParam Profile could not be resolved", "Input Profile", inputProfiles.DbParam)
// 		return nil, err
// 	}

// 	profilesMap = map[string]ProfileResponse{
// 		common.PROFILE_TYPE_COMPUTE:            compute,
// 		common.PROFILE_TYPE_SOFTWARE:           software,
// 		common.PROFILE_TYPE_NETWORK:            network,
// 		common.PROFILE_TYPE_DATABASE_PARAMETER: dbParam,
// 	}

// 	log.Info("Returning from ndb_api_helpers.GetProfiles", "profiles map", profilesMap)
// 	return
// }

// Fetches all the profiles and returns a map of profiles
// Returns an error if any profile is not found
func GetProfiless(ctx context.Context, ndb_client *ndb_client.NDBClient, databaseType string, profileResolvers ProfileResolvers) (profilesMap map[string]ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)

	log.Info("Entered ndb_api_helpers.GetProfiles", "Input profiles", profileResolvers)

	allProfiles, err := GetAllProfiles(ctx, ndb_client)

	// profiles need to be in the ready state
	activeProfiles := util.Filter(allProfiles, func(p ProfileResponse) bool { return p.Status == common.PROFILE_STATUS_READY })
	if err != nil {
		log.Error(err, "Profiles could not be fetched")
		return
	}

	dbEngineSpecific := util.Filter(activeProfiles, func(p ProfileResponse) bool {
		return p.EngineType == GetDatabaseEngineName(databaseType)
	})

	computeProfileResolver := profileResolvers[common.PROFILE_TYPE_COMPUTE]
	softwareProfileResolver := profileResolvers[common.PROFILE_TYPE_SOFTWARE]
	networkProfileResolver := profileResolvers[common.PROFILE_TYPE_NETWORK]
	dbParamProfileResolver := profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER]

	// Compute Profile
	compute, err := computeProfileResolver.Resolve(ctx, activeProfiles, ComputeOOBProfileResolver)
	if err != nil {
		log.Error(err, "Compute Profile could not be resolved", "Input Profile", computeProfileResolver)
		return nil, err
	}

	// Software Profile
	// validation of software profile for closed-source db engines
	isClosedSourceEngine := (databaseType == common.DATABASE_TYPE_ORACLE) || (databaseType == common.DATABASE_TYPE_SQLSERVER)
	if isClosedSourceEngine {
		if softwareProfileResolver.GetId() == "" && softwareProfileResolver.GetName() == "" {
			log.Error(errors.New("software profile not provided"), "Provide software profile info", "dbType", databaseType)
			return nil, fmt.Errorf("software profile is a mandatory input for %s database", databaseType)
		}
	}

	software, err := softwareProfileResolver.Resolve(ctx, dbEngineSpecific, SoftwareOOBProfileResolverForSingleInstance)
	if err != nil {
		log.Error(err, "Software Profile could not be resolved or is not in READY state", "Input Profile", softwareProfileResolver)
		return nil, err
	}

	// Network Profile
	network, err := networkProfileResolver.Resolve(ctx, dbEngineSpecific, NetworkOOBProfileResolver)
	if err != nil {
		log.Error(err, "Network Profile could not be resolved", "Input Profile", networkProfileResolver)
		return nil, err
	}

	// DB Param Profile
	dbParam, err := dbParamProfileResolver.Resolve(ctx, dbEngineSpecific, DbParamOOBProfileResolver)
	if err != nil {
		log.Error(err, "DbParam Profile could not be resolved", "Input Profile", networkProfileResolver)
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
