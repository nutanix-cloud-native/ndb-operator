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

package test

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
)

const mock_username = "username"
const mock_password = "password"

const NONE_SLA_ID = "NONE_SLA_ID"

var MockResponsesMap = map[string]interface{}{
	"GET /slas": []v1alpha1.SLAResponse{
		{
			Id:                 "sla-1-id",
			Name:               "SLA 1",
			UniqueName:         "SLA 1 Unique Name",
			Description:        "SLA 1 Description",
			DailyRetention:     1,
			WeeklyRetention:    2,
			MonthlyRetention:   3,
			QuarterlyRetention: 4,
			YearlyRetention:    5,
		},
		{
			Id:                 "sla-2-id",
			Name:               "SLA 2",
			UniqueName:         "SLA 2 Unique Name",
			Description:        "SLA 2 Description",
			DailyRetention:     1,
			WeeklyRetention:    2,
			MonthlyRetention:   3,
			QuarterlyRetention: 4,
			YearlyRetention:    5,
		},
		{
			Id:                 NONE_SLA_ID,
			Name:               v1alpha1.SLA_NAME_NONE,
			UniqueName:         "SLA 3 Unique Name",
			Description:        "SLA 3 Description",
			DailyRetention:     1,
			WeeklyRetention:    2,
			MonthlyRetention:   3,
			QuarterlyRetention: 4,
			YearlyRetention:    5,
		},
	},

	"GET /profiles": []v1alpha1.ProfileResponse{
		{
			Id:              "1",
			Name:            "a",
			Type:            v1alpha1.PROFILE_TYPE_COMPUTE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_GENERIC,
			LatestVersionId: "v-id-1",
			Topology:        v1alpha1.TOPOLOGY_ALL,
		},
		{
			Id:              "1.1",
			Name:            "DEFAULT_OOB_SMALL_COMPUTE",
			Type:            v1alpha1.PROFILE_TYPE_COMPUTE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_GENERIC,
			LatestVersionId: "v-id-1",
			Topology:        v1alpha1.TOPOLOGY_ALL,
		},
		{
			Id:              "2",
			Name:            "b",
			Type:            v1alpha1.PROFILE_TYPE_STORAGE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_GENERIC,
			LatestVersionId: "v-id-2",
			Topology:        v1alpha1.TOPOLOGY_ALL,
		},
		{
			Id:              "3",
			Name:            "c",
			Type:            v1alpha1.PROFILE_TYPE_SOFTWARE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-3",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "4",
			Name:            "d",
			Type:            v1alpha1.PROFILE_TYPE_SOFTWARE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-4",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "5",
			Name:            "e",
			Type:            v1alpha1.PROFILE_TYPE_SOFTWARE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-5",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "6",
			Name:            "f",
			Type:            v1alpha1.PROFILE_TYPE_NETWORK,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-6",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "7",
			Name:            "g",
			Type:            v1alpha1.PROFILE_TYPE_NETWORK,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-7",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "8",
			Name:            "h",
			Type:            v1alpha1.PROFILE_TYPE_NETWORK,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-8",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "9",
			Name:            "i",
			Type:            v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-9",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "10",
			Name:            "j",
			Type:            v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-10",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "11",
			Name:            "k",
			Type:            v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-11",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "12",
			Name:            "custom postgres software profile",
			Type:            v1alpha1.PROFILE_TYPE_SOFTWARE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-12",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "13",
			Name:            "custom mysql software profile",
			Type:            v1alpha1.PROFILE_TYPE_SOFTWARE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-13",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "14",
			Name:            "custom mongodb software profile",
			Type:            v1alpha1.PROFILE_TYPE_SOFTWARE,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-14",
			Topology:        v1alpha1.TOPOLOGY_SINGLE,
		},
		{
			Id:              "15",
			Name:            "custom network profile for postgres",
			Type:            v1alpha1.PROFILE_TYPE_NETWORK,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-15",
			Topology:        v1alpha1.TOPOLOGY_ALL,
		},
		{
			Id:              "16",
			Name:            "custom network profile for mysql",
			Type:            v1alpha1.PROFILE_TYPE_NETWORK,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-16",
			Topology:        v1alpha1.TOPOLOGY_ALL,
		},
		{
			Id:              "17",
			Name:            "custom network profile for mongodb",
			Type:            v1alpha1.PROFILE_TYPE_NETWORK,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-17",
			Topology:        v1alpha1.TOPOLOGY_ALL,
		},
		{
			Id:              "18",
			Name:            "custom database param profile for postgres",
			Type:            v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-18",
			Topology:        v1alpha1.TOPOLOGY_INSTANCE,
		},
		{
			Id:              "19",
			Name:            "custom database param profile for mysql",
			Type:            v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-19",
			Topology:        v1alpha1.TOPOLOGY_INSTANCE,
		},
		{
			Id:              "20",
			Name:            "custom database param profile for mongodb",
			Type:            v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      v1alpha1.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-20",
			Topology:        v1alpha1.TOPOLOGY_INSTANCE,
		},
	},
}

func checkAuthTestHelper(r *http.Request) bool {
	username, password, ok := r.BasicAuth()

	if ok {
		usernameHash := sha256.Sum256([]byte(username))
		passwordHash := sha256.Sum256([]byte(password))
		expectedUsernameHash := sha256.Sum256([]byte(mock_username))
		expectedPasswordHash := sha256.Sum256([]byte(mock_password))

		usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
		passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

		if usernameMatch && passwordMatch {
			return true
		}
	}
	return false
}

func GetServerTestHelper(t *testing.T) *httptest.Server {

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response = MockResponsesMap[r.Method+" "+r.URL.Path]
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
}

func GetCustomProfileForDBType(dbType string) (profiles v1alpha1.Profiles) {
	switch dbType {
	case v1alpha1.DATABASE_TYPE_POSTGRES:
		profiles = v1alpha1.Profiles{
			// Custom Software Profile Name = "custom postgres software profile"
			Software: v1alpha1.Profile{
				Id:        "12",
				VersionId: "v-id-12",
			},
			// Custom ompute Name = "a"
			Compute: v1alpha1.Profile{
				Id:        "1",
				VersionId: "v-id-1",
			},
			Network: v1alpha1.Profile{
				Id:        "15",
				VersionId: "v-id-15",
			},
			DbParam: v1alpha1.Profile{
				Id:        "18",
				VersionId: "v-id-18",
			},
		}
		return profiles
	case v1alpha1.DATABASE_TYPE_MYSQL:
		profiles = v1alpha1.Profiles{
			// Custom Software Profile Name = "custom mysql software profile"
			Software: v1alpha1.Profile{
				Id:        "13",
				VersionId: "v-id-13",
			},
			// Custom Compute Name = "a"
			Compute: v1alpha1.Profile{
				Id:        "1",
				VersionId: "v-id-1",
			},
			Network: v1alpha1.Profile{
				Id:        "16",
				VersionId: "v-id-16",
			},
			DbParam: v1alpha1.Profile{
				Id:        "19",
				VersionId: "v-id-19",
			},
		}
		return profiles
	case v1alpha1.DATABASE_TYPE_MONGODB:
		profiles = v1alpha1.Profiles{
			// Custom Software Profile Name = "custom mongodb software profile"
			Software: v1alpha1.Profile{
				Id:        "14",
				VersionId: "v-id-14",
			},
			// Custom Compute Name = "a"
			Compute: v1alpha1.Profile{
				Id:        "1",
				VersionId: "v-id-1",
			},
			Network: v1alpha1.Profile{
				Id:        "17",
				VersionId: "v-id-17",
			},
			DbParam: v1alpha1.Profile{
				Id:        "20",
				VersionId: "v-id-20",
			},
		}
		return profiles
	case v1alpha1.DATABASE_TYPE_MONGODB_INVALID_PROFILE, v1alpha1.DATABASE_TYPE_MYSQL_INVALID_PROFILE, v1alpha1.DATABASE_TYPE_POSTGRES_INVALID_PROFILE:
		// below custom profiles do not exist and will be used for the negative scenario
		profiles = v1alpha1.Profiles{
			Software: v1alpha1.Profile{
				Id:        "140",
				VersionId: "v-id-140",
			},
			Compute: v1alpha1.Profile{
				Id:        "100",
				VersionId: "v-id-100",
			},
			Network: v1alpha1.Profile{
				Id:        "170",
				VersionId: "v-id-170",
			},
			DbParam: v1alpha1.Profile{
				Id:        "200",
				VersionId: "v-id-200",
			},
		}
		return profiles
	default:
		return v1alpha1.Profiles{}
	}
}
