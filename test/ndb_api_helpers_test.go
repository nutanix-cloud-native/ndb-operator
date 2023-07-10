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

package test

import (
	"context"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/controller_adapters"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/stretchr/testify/assert"
)

func GetProfileResolvers(d v1alpha1.Database) ndb_api.ProfileResolvers {
	profileResolvers := make(ndb_api.ProfileResolvers)

	profileResolvers[common.PROFILE_TYPE_COMPUTE] = &controller_adapters.Profile{
		Profile:     d.Spec.Instance.Profiles.Compute,
		ProfileType: common.PROFILE_TYPE_COMPUTE,
	}
	profileResolvers[common.PROFILE_TYPE_SOFTWARE] = &controller_adapters.Profile{
		Profile:     d.Spec.Instance.Profiles.Software,
		ProfileType: common.PROFILE_TYPE_SOFTWARE,
	}
	profileResolvers[common.PROFILE_TYPE_NETWORK] = &controller_adapters.Profile{
		Profile:     d.Spec.Instance.Profiles.Network,
		ProfileType: common.PROFILE_TYPE_NETWORK,
	}
	profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER] = &controller_adapters.Profile{
		Profile:     d.Spec.Instance.Profiles.DbParam,
		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
	}

	profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE] = &controller_adapters.Profile{
		Profile:     d.Spec.Instance.Profiles.DbParamInstance,
		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
	}

	return profileResolvers

}

func profilesListGenerator() []ndb_api.ProfileResponse {

	oob_compute_profile := ndb_api.ProfileResponse{
		Id:              "DEFAULT_OOB_SMALL_COMPUTE",
		Name:            "DEFAULT_OOB_SMALL_COMPUTE",
		Type:            "Compute",
		EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
		LatestVersionId: "cp-vid-",
		Topology:        "test topology",
		SystemProfile:   true,
		Status:          "READY",
	}

	custom_generic_compute := ndb_api.ProfileResponse{
		Id:              "cp-id-1",
		Name:            "Compute_Profile_1",
		Type:            "Compute",
		EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
		LatestVersionId: "cp-vid-1",
		Topology:        "test topology",
		SystemProfile:   false,
		Status:          "READY",
	}

	oob_generic_compute := ndb_api.ProfileResponse{
		Id:              "cp-id-2",
		Name:            "small",
		Type:            "Compute",
		EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
		LatestVersionId: "cp-vid-2",
		Topology:        "test topology",
		SystemProfile:   true,
		Status:          "READY",
	}

	oob_oracle_software := ndb_api.ProfileResponse{
		Id:              "sw-id-5",
		Name:            "Software_Profile_5",
		Type:            "Software",
		EngineType:      common.DATABASE_ENGINE_TYPE_ORACLE,
		LatestVersionId: "sw-vid-5",
		Topology:        "single",
		SystemProfile:   true,
		Status:          "READY",
	}

	oob_postgres_software := ndb_api.ProfileResponse{
		Id:              "sw-id-1",
		Name:            "Software_Profile_1",
		Type:            "Software",
		EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
		LatestVersionId: "sw-vid-1",
		Topology:        "single",
		SystemProfile:   true,
		Status:          "READY",
	}

	oob_postgres_software_not_ready := ndb_api.ProfileResponse{
		Id:              "sw-id-2",
		Name:            "Software_Profile_Not_READY",
		Type:            "Software",
		EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
		LatestVersionId: "sw-vid-2",
		Topology:        "single",
		SystemProfile:   false,
		Status:          "NOT_YET_CREATED",
	}

	oob_postgres_network := ndb_api.ProfileResponse{
		Id:              "nw-id-1",
		Name:            "Network_Profile_1",
		Type:            "Network",
		EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
		LatestVersionId: "nw-vid-1",
		Topology:        "test topology",
		SystemProfile:   true,
		Status:          "READY",
	}

	oob_postgres_dbparam := ndb_api.ProfileResponse{
		Id:              "dbp-id-1",
		Name:            "DBParam_Profile_1",
		Type:            "DBParam",
		EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
		LatestVersionId: "dbp-vid-1",
		Topology:        "test topology",
		SystemProfile:   true,
		Status:          "READY",
	}

	allProfiles := [10]ndb_api.ProfileResponse{
		oob_compute_profile,
		custom_generic_compute,
		oob_generic_compute,
		oob_oracle_software,
		oob_postgres_software,
		oob_postgres_software_not_ready,
		oob_postgres_network,
		oob_postgres_dbparam,
	}

	return allProfiles[:]
}

func TestGetProfilesOOBComputeProfileResolved(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()

	inputProfile := GetProfileResolvers(v1alpha1.Database{})[common.PROFILE_TYPE_COMPUTE]

	resolvedComputeProfile, err := inputProfile.Resolve(ctx,
		allProfiles,
		ndb_api.ComputeOOBProfileResolver)

	assert.Nil(t, err)
	// assert that its OOB profile
	assert.True(t, resolvedComputeProfile.SystemProfile)
}

func TestResolveOOBSoftwareProfile_ByEmptyNameAndID_ResolvesOk(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()
	pgSpecificProfiles := util.Filter(allProfiles, func(p ndb_api.ProfileResponse) bool {
		return p.EngineType == common.DATABASE_ENGINE_TYPE_POSTGRES
	})

	inputProfile := GetProfileResolvers(v1alpha1.Database{})[common.PROFILE_TYPE_SOFTWARE]

	resolvedSoftwareProfile, err := inputProfile.Resolve(ctx,
		pgSpecificProfiles,
		ndb_api.SoftwareOOBProfileResolverForSingleInstance)

	assert.Nil(t, err)
	// assert that its OOB profile
	assert.True(t, resolvedSoftwareProfile.SystemProfile)
}

func TestResolveSoftwareProfileByName_ByName_ResolvesOk(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()
	pgSpecificProfiles := util.Filter(allProfiles, func(p ndb_api.ProfileResponse) bool {
		return p.EngineType == common.DATABASE_ENGINE_TYPE_POSTGRES
	})

	inputProfile := GetProfileResolvers(v1alpha1.Database{
		Spec: v1alpha1.DatabaseSpec{
			Instance: v1alpha1.Instance{
				Profiles: v1alpha1.Profiles{
					Software: v1alpha1.Profile{
						Name: "Software_Profile_1",
					},
				},
			},
		},
	})[common.PROFILE_TYPE_SOFTWARE]

	resolvedSoftwareProfile, err := inputProfile.Resolve(ctx,
		pgSpecificProfiles,
		ndb_api.SoftwareOOBProfileResolverForSingleInstance)

	assert.Nil(t, err)
	assert.Equal(t, resolvedSoftwareProfile.Name, "Software_Profile_1")
}

func TestResolveSoftwareProfile_ByNameMismatch_throwsError(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()
	pgSpecificProfiles := util.Filter(allProfiles, func(p ndb_api.ProfileResponse) bool {
		return p.EngineType == common.DATABASE_ENGINE_TYPE_POSTGRES
	})

	inputProfile := GetProfileResolvers(v1alpha1.Database{
		Spec: v1alpha1.DatabaseSpec{
			Instance: v1alpha1.Instance{
				Profiles: v1alpha1.Profiles{
					Software: v1alpha1.Profile{
						Name: "Software_Profile_#1", // profile with this name does not exist
					},
				},
			},
		},
	})[common.PROFILE_TYPE_SOFTWARE]

	resolvedSoftwareProfile, err := inputProfile.Resolve(ctx,
		pgSpecificProfiles,
		ndb_api.SoftwareOOBProfileResolverForSingleInstance)

	assert.NotNil(t, err)
	// should return an error and an empty profile
	assert.Equal(t, resolvedSoftwareProfile, (ndb_api.ProfileResponse{}))

}

func TestResolveComputeProfileByName_resolvesOk(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()

	inputProfile := GetProfileResolvers(v1alpha1.Database{
		Spec: v1alpha1.DatabaseSpec{
			Instance: v1alpha1.Instance{
				Profiles: v1alpha1.Profiles{
					Compute: v1alpha1.Profile{
						Name: "Compute_Profile_1",
					},
				},
			},
		},
	})[common.PROFILE_TYPE_COMPUTE]

	resolvedComputeProfile, err := inputProfile.Resolve(ctx,
		allProfiles,
		ndb_api.ComputeOOBProfileResolver)

	assert.Nil(t, err)
	assert.Equal(t, resolvedComputeProfile.Name, "Compute_Profile_1")
}

// case mismatch is not supported, profile name is case-sensitive
func TestResolveComputeProfileByNameCaseMismatch_throwsError(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()

	inputProfile := GetProfileResolvers(v1alpha1.Database{
		Spec: v1alpha1.DatabaseSpec{
			Instance: v1alpha1.Instance{
				Profiles: v1alpha1.Profiles{
					Compute: v1alpha1.Profile{
						Name: "compute_Profile_1",
					},
				},
			},
		},
	})[common.PROFILE_TYPE_COMPUTE]

	resolvedComputeProfile, err := inputProfile.Resolve(ctx,
		allProfiles,
		ndb_api.ComputeOOBProfileResolver)

	assert.NotNil(t, err)
	assert.Equal(t, resolvedComputeProfile, ndb_api.ProfileResponse{})
}

func TestResolveComputeProfileById_resolvesOk(t *testing.T) {
	ctx := context.Background()
	allProfiles := profilesListGenerator()

	inputProfile := GetProfileResolvers(v1alpha1.Database{
		Spec: v1alpha1.DatabaseSpec{
			Instance: v1alpha1.Instance{
				Profiles: v1alpha1.Profiles{
					Compute: v1alpha1.Profile{
						Id: "cp-id-2",
					},
				},
			},
		},
	})[common.PROFILE_TYPE_COMPUTE]

	resolvedComputeProfile, err := inputProfile.Resolve(ctx,
		allProfiles,
		ndb_api.ComputeOOBProfileResolver)

	assert.Nil(t, err)
	assert.Equal(t, resolvedComputeProfile.Id, "cp-id-2")
}
