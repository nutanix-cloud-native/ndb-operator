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
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

func TestResolveProfiles(t *testing.T) {
	// Set
	server := GetServerTestHelper(t)
	defer server.Close()
	client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	getResolver := func(p ProfileResponse, e error) *MockProfileResolverInterface {
		profileResolver := MockProfileResolverInterface{}
		profileResolver.On("GetId").Return(p.Id)
		profileResolver.On("GetName").Return(p.Name)
		profileResolver.On("Resolve").Return(p, e)
		return &profileResolver
	}

	tests := []struct {
		name                 string
		ctx                  context.Context
		ndbClient            *ndb_client.NDBClient
		databaseType         string
		compute              ProfileResponse
		software             ProfileResponse
		network              ProfileResponse
		dbParam              ProfileResponse
		dbParamInstance      ProfileResponse
		softwareError        error
		computeError         error
		networkError         error
		dbParamError         error
		dbParamInstanceError error
		wantProfilesMap      map[string]ProfileResponse
		wantErr              bool
	}{
		{
			name:         "Non MSSQL databases have empty db param instance profile",
			ctx:          context.TODO(),
			ndbClient:    client,
			databaseType: common.DATABASE_TYPE_POSTGRES,
			compute: ProfileResponse{
				Id:              "1.1",
				Name:            "DEFAULT_OOB_SMALL_COMPUTE",
				Type:            common.PROFILE_TYPE_COMPUTE,
				EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
				LatestVersionId: "v-id-1",
				Topology:        common.TOPOLOGY_ALL,
				Status:          "READY",
				SystemProfile:   true,
			},
			software: ProfileResponse{
				Id:              "3",
				Name:            "c",
				Type:            common.PROFILE_TYPE_SOFTWARE,
				EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
				LatestVersionId: "v-id-3",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			network: ProfileResponse{
				Id:              "6",
				Name:            "f",
				Type:            common.PROFILE_TYPE_NETWORK,
				EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
				LatestVersionId: "v-id-6",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			dbParam: ProfileResponse{
				Id:              "9",
				Name:            "i",
				Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
				EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
				LatestVersionId: "v-id-9",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			wantProfilesMap: map[string]ProfileResponse{
				common.PROFILE_TYPE_COMPUTE: {
					Id:              "1.1",
					Name:            "DEFAULT_OOB_SMALL_COMPUTE",
					Type:            common.PROFILE_TYPE_COMPUTE,
					EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
					LatestVersionId: "v-id-1",
					Topology:        common.TOPOLOGY_ALL,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_SOFTWARE: {
					Id:              "3",
					Name:            "c",
					Type:            common.PROFILE_TYPE_SOFTWARE,
					EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
					LatestVersionId: "v-id-3",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_NETWORK: {
					Id:              "6",
					Name:            "f",
					Type:            common.PROFILE_TYPE_NETWORK,
					EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
					LatestVersionId: "v-id-6",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_DATABASE_PARAMETER: {
					Id:              "9",
					Name:            "i",
					Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
					EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
					LatestVersionId: "v-id-9",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: {},
			},
			wantErr: false,
		},
		{
			name:         "MSSQL databases when db param instance profile is resolved",
			ctx:          context.TODO(),
			ndbClient:    client,
			databaseType: common.DATABASE_TYPE_MSSQL,
			compute: ProfileResponse{
				Id:              "1.1",
				Name:            "DEFAULT_OOB_SMALL_COMPUTE",
				Type:            common.PROFILE_TYPE_COMPUTE,
				EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
				LatestVersionId: "v-id-1",
				Topology:        common.TOPOLOGY_ALL,
				Status:          "READY",
				SystemProfile:   true,
			},
			software: ProfileResponse{
				Id:              "15",
				Name:            "MSSQL SW Profile",
				Type:            common.PROFILE_TYPE_SOFTWARE,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-4",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			network: ProfileResponse{
				Id:              "14",
				Name:            "MSSQL NW Profile",
				Type:            common.PROFILE_TYPE_NETWORK,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-8",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			dbParam: ProfileResponse{
				Id:              "12",
				Name:            "MSSQL DB PARAM",
				Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-db-p",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			dbParamInstance: ProfileResponse{
				Id:              "13",
				Name:            "MSSQL DB PARAM INSTANCE",
				Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-db-p-i",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			wantProfilesMap: map[string]ProfileResponse{
				common.PROFILE_TYPE_COMPUTE: {
					Id:              "1.1",
					Name:            "DEFAULT_OOB_SMALL_COMPUTE",
					Type:            common.PROFILE_TYPE_COMPUTE,
					EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
					LatestVersionId: "v-id-1",
					Topology:        common.TOPOLOGY_ALL,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_SOFTWARE: {
					Id:              "15",
					Name:            "MSSQL SW Profile",
					Type:            common.PROFILE_TYPE_SOFTWARE,
					EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
					LatestVersionId: "v-id-4",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_NETWORK: {
					Id:              "14",
					Name:            "MSSQL NW Profile",
					Type:            common.PROFILE_TYPE_NETWORK,
					EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
					LatestVersionId: "v-id-8",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_DATABASE_PARAMETER: {
					Id:              "12",
					Name:            "MSSQL DB PARAM",
					Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
					EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
					LatestVersionId: "v-id-db-p",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
				common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: {
					Id:              "13",
					Name:            "MSSQL DB PARAM INSTANCE",
					Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
					EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
					LatestVersionId: "v-id-db-p-i",
					Topology:        common.TOPOLOGY_SINGLE,
					Status:          "READY",
					SystemProfile:   true,
				},
			},
			wantErr: false,
		},
		{
			name:         "MSSQL databases when db param instance profile is not resolved",
			ctx:          context.TODO(),
			ndbClient:    client,
			databaseType: common.DATABASE_TYPE_MSSQL,
			compute: ProfileResponse{
				Id:              "1.1",
				Name:            "DEFAULT_OOB_SMALL_COMPUTE",
				Type:            common.PROFILE_TYPE_COMPUTE,
				EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
				LatestVersionId: "v-id-1",
				Topology:        common.TOPOLOGY_ALL,
				Status:          "READY",
				SystemProfile:   true,
			},
			software: ProfileResponse{
				Id:              "15",
				Name:            "MSSQL SW Profile",
				Type:            common.PROFILE_TYPE_SOFTWARE,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-4",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			network: ProfileResponse{
				Id:              "14",
				Name:            "MSSQL NW Profile",
				Type:            common.PROFILE_TYPE_NETWORK,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-8",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			dbParam: ProfileResponse{
				Id:              "12",
				Name:            "MSSQL DB PARAM",
				Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-db-p",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			wantProfilesMap: nil,
			wantErr:         true,
		},
		{
			name:         "MSSQL databases when software profile info is not provided",
			ctx:          context.TODO(),
			ndbClient:    client,
			databaseType: common.DATABASE_TYPE_MSSQL,
			compute: ProfileResponse{
				Id:              "1.1",
				Name:            "DEFAULT_OOB_SMALL_COMPUTE",
				Type:            common.PROFILE_TYPE_COMPUTE,
				EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
				LatestVersionId: "v-id-1",
				Topology:        common.TOPOLOGY_ALL,
				Status:          "READY",
				SystemProfile:   true,
			},
			software: ProfileResponse{},
			network: ProfileResponse{
				Id:              "14",
				Name:            "MSSQL NW Profile",
				Type:            common.PROFILE_TYPE_NETWORK,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-8",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			dbParam: ProfileResponse{
				Id:              "12",
				Name:            "MSSQL DB PARAM",
				Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-db-p",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   true,
			},
			wantProfilesMap: nil,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			software := getResolver(tt.software, tt.softwareError)
			compute := getResolver(tt.compute, tt.computeError)
			network := getResolver(tt.network, tt.networkError)
			dbParam := getResolver(tt.dbParam, tt.dbParamError)
			dbParamInstance := getResolver(tt.dbParamInstance, tt.dbParamInstanceError)

			profileResolvers := map[string]ProfileResolver{
				common.PROFILE_TYPE_SOFTWARE:                    software,
				common.PROFILE_TYPE_COMPUTE:                     compute,
				common.PROFILE_TYPE_NETWORK:                     network,
				common.PROFILE_TYPE_DATABASE_PARAMETER:          dbParam,
				common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: dbParamInstance,
			}

			gotProfilesMap, err := ResolveProfiles(tt.ctx, tt.ndbClient, tt.databaseType, profileResolvers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveProfiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProfilesMap, tt.wantProfilesMap) {
				t.Errorf("ResolveProfiles() = %v, want %v", gotProfilesMap, tt.wantProfilesMap)
			}
		})
	}
}
