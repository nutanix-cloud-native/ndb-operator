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
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
)

const (
	MSSQL_TEST_SW_PROFILE_NAME  = "mssqlSW"
	MSSQL_TEST_SW_PROFILE_ID    = "id-mssql-sw-1"
	MSSQL_TEST_DBI_PROFILE_NAME = "mssqlDBI"
	MSSQL_TEST_DBI_PROFILE_ID   = "id-mssql-dbi-1"
	mock_username               = "username"
	mock_password               = "password"
	NONE_SLA_ID                 = "NONE_SLA_ID"
)

var MockResponsesMap = map[string]interface{}{
	"GET /slas": []ndb_api.SLAResponse{
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
			Name:               common.SLA_NAME_NONE,
			UniqueName:         "SLA 3 Unique Name",
			Description:        "SLA 3 Description",
			DailyRetention:     1,
			WeeklyRetention:    2,
			MonthlyRetention:   3,
			QuarterlyRetention: 4,
			YearlyRetention:    5,
		},
	},

	"GET /profiles": []ndb_api.ProfileResponse{
		{
			Id:              "1",
			Name:            "a",
			Type:            common.PROFILE_TYPE_COMPUTE,
			EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
			LatestVersionId: "v-id-1",
			Topology:        common.TOPOLOGY_ALL,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "1.1",
			Name:            "DEFAULT_OOB_SMALL_COMPUTE",
			Type:            common.PROFILE_TYPE_COMPUTE,
			EngineType:      common.DATABASE_ENGINE_TYPE_GENERIC,
			LatestVersionId: "v-id-1",
			Topology:        common.TOPOLOGY_ALL,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "3",
			Name:            "c",
			Type:            common.PROFILE_TYPE_SOFTWARE,
			EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-3",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "3NotReady",
			Name:            "Software_Profile_Not_Ready",
			Type:            common.PROFILE_TYPE_SOFTWARE,
			EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-3",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "NOT_YET_CREATED",
			SystemProfile:   true,
		},
		{
			Id:              "4",
			Name:            "d",
			Type:            common.PROFILE_TYPE_SOFTWARE,
			EngineType:      common.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-4",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "5",
			Name:            "e",
			Type:            common.PROFILE_TYPE_SOFTWARE,
			EngineType:      common.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-5",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "6",
			Name:            "f",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-6",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "7",
			Name:            "g",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-7",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "8",
			Name:            "h",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-8",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "9",
			Name:            "i",
			Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-9",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "10",
			Name:            "j",
			Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      common.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-10",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "11",
			Name:            "k",
			Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      common.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-11",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "id-pg-nw-1",
			Name:            "DEFAULT_OOB_POSTGRESQL_NETWORK",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_POSTGRES,
			LatestVersionId: "v-id-6",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "id-mongo-nw-1",
			Name:            "DEFAULT_OOB_MONGODB_NETWORK",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_MONGODB,
			LatestVersionId: "v-id-6",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "id-mysql-nw-1",
			Name:            "DEFAULT_OOB_MYSQL_NETWORK",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_MYSQL,
			LatestVersionId: "v-id-6",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              MSSQL_TEST_SW_PROFILE_ID,
			Name:            MSSQL_TEST_SW_PROFILE_NAME,
			Type:            common.PROFILE_TYPE_SOFTWARE,
			EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
			LatestVersionId: "mssql-id-1",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "id-mssql-nw-1",
			Name:            "mssqlNW",
			Type:            common.PROFILE_TYPE_NETWORK,
			EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
			LatestVersionId: "mssql-id-1",
			Topology:        common.TOPOLOGY_SINGLE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              "id-mssql-db-1",
			Name:            "mssqlDB",
			Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
			LatestVersionId: "v-id-10",
			Topology:        common.TOPOLOGY_DATABASE,
			Status:          "READY",
			SystemProfile:   true,
		},
		{
			Id:              MSSQL_TEST_DBI_PROFILE_ID,
			Name:            MSSQL_TEST_DBI_PROFILE_NAME,
			Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
			EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
			LatestVersionId: "v-id-10",
			Topology:        common.TOPOLOGY_INSTANCE,
			Status:          "READY",
			SystemProfile:   true,
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

		usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
		passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

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
