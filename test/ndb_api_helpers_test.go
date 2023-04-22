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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
)

func TestGetNoneTimeMachineSLA(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	sla, err := v1alpha1.GetNoneTimeMachineSLA(context.Background(), ndbclient)

	//Assert
	if err != nil {
		t.Errorf("Could not get NONE TM, error: %s", err)
	}
	if sla.Name != v1alpha1.SLA_NAME_NONE {
		t.Error("Could not fetch mock slas")
	}
}

func TestGetNoneTimeMachineSLAReturnsErrorWhenNoneTimeMachineNotFound(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			resp, _ := json.Marshal([]v1alpha1.SLAResponse{
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
			})
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	sla, err := v1alpha1.GetNoneTimeMachineSLA(context.Background(), ndbclient)
	//Assert
	if err == nil {
		t.Errorf("GetNoneTimeMachineSLA should return an error when NONE time machine does not exists")
	}
	if sla != (v1alpha1.SLAResponse{}) {
		t.Error("GetNoneTimeMachineSLA should respond with an empty SLA when NONE time machine is not found")
	}
}

func TestGetOOBProfiles(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		profileMap, _ := v1alpha1.GetOOBProfiles(context.Background(), ndbclient, dbType)

		//Assert
		profileTypes := []string{
			v1alpha1.PROFILE_TYPE_COMPUTE,
			v1alpha1.PROFILE_TYPE_STORAGE,
			v1alpha1.PROFILE_TYPE_SOFTWARE,
			v1alpha1.PROFILE_TYPE_NETWORK,
			v1alpha1.PROFILE_TYPE_DATABASE_PARAMETER,
		}
		for _, profileType := range profileTypes {
			profile := profileMap[profileType]
			//Assert that no profileType is empty
			if profile == (v1alpha1.ProfileResponse{}) {
				t.Errorf("Empty profile type %s for dbType %s", profileType, dbType)
			}
			//Assert that profile EngineType matches the database engine or the generic type
			if profile.EngineType != v1alpha1.GetDatabaseEngineName(dbType) && profile.EngineType != v1alpha1.DATABASE_ENGINE_TYPE_GENERIC {
				t.Errorf("Profile engine type %s for dbType %s does not match", profile.EngineType, dbType)
			}
		}
	}
}

func TestGetOOBProfilesOnlyGetsTheSmallOOBComputeProfile(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		profileMap, _ := v1alpha1.GetOOBProfiles(context.Background(), ndbclient, dbType)

		//Assert
		computeProfile := profileMap[v1alpha1.PROFILE_TYPE_COMPUTE]
		if !strings.Contains(strings.ToLower(computeProfile.Name), "small") {
			t.Errorf("Expected small oob compute profile, but got: %s", computeProfile.Name)
		}
	}
}

func TestGetOOBProfilesReturnsErrorWhenSomeProfileIsNotFound(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			resp, _ := json.Marshal([]v1alpha1.ProfileResponse{
				{
					Id:              "1",
					Name:            "a",
					Type:            "test type",
					EngineType:      "test engine",
					LatestVersionId: "v-id-1",
					Topology:        "test topology",
				},
			})
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		_, err := v1alpha1.GetOOBProfiles(context.Background(), ndbclient, dbType)
		// None of the profile criteria should match the mocked response
		// t.Log(err)
		if err == nil {
			t.Errorf("GetOOBProdiles should have retuned an error when none of the profiles matc the criteria.")
		}

	}
}

func TestGenerateProvisioningRequestReturnsErrorIfNoneTMNotFound(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response interface{}
			if r.URL.Path == "/profiles" {
				response = []v1alpha1.ProfileResponse{
					{
						Id:              "1",
						Name:            "a",
						Type:            "test type",
						EngineType:      "test engine",
						LatestVersionId: "v-id-1",
						Topology:        "test topology",
					},
				}
			} else if r.URL.Path == "/slas" {
				response = []v1alpha1.SLAResponse{
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
				}
			}
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		dbSpec := v1alpha1.DatabaseSpec{
			NDB: v1alpha1.NDB{
				Server:           "abc.def.ghi.jkl/v99/api",
				ClusterId:        "test-cluster-id",
				CredentialSecret: "qwertyuiop",
			},
			Instance: v1alpha1.Instance{
				DatabaseNames:        []string{"a", "b", "c", "d"},
				Type:                 dbType,
				DatabaseInstanceName: dbType + "-instance-test",
				TimeZone:             "UTC",
			},
		}

		reqData := map[string]interface{}{
			v1alpha1.NDB_PARAM_PASSWORD:       "qwerty",
			v1alpha1.NDB_PARAM_SSH_PUBLIC_KEY: "qwertyuiop",
		}

		_, err := v1alpha1.GenerateProvisioningRequest(context.Background(), ndbclient, dbSpec, reqData)
		t.Log(err)
		if err == nil {
			t.Errorf("GenerateProvisioningRequest should return an error when NONE time machine is not found")
		}
	}
}

func TestGenerateProvisioningRequestReturnsErrorIfProfilesNotFound(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response interface{}
			if r.URL.Path == "/profiles" {
				response = []v1alpha1.ProfileResponse{
					{
						Id:              "1",
						Name:            "a",
						Type:            "test type",
						EngineType:      "test engine",
						LatestVersionId: "v-id-1",
						Topology:        "test topology",
					},
				}
			} else if r.URL.Path == "/slas" {
				response = []v1alpha1.SLAResponse{
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
				}
			}
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		dbSpec := v1alpha1.DatabaseSpec{
			NDB: v1alpha1.NDB{
				Server:           "abc.def.ghi.jkl/v99/api",
				ClusterId:        "test-cluster-id",
				CredentialSecret: "test-credential-secret-name",
			},
			Instance: v1alpha1.Instance{
				DatabaseNames:        []string{"a", "b", "c", "d"},
				Type:                 dbType,
				DatabaseInstanceName: dbType + "-instance-test",
				TimeZone:             "UTC",
			},
		}

		reqData := map[string]interface{}{
			v1alpha1.NDB_PARAM_PASSWORD:       "qwerty",
			v1alpha1.NDB_PARAM_SSH_PUBLIC_KEY: "qwertyuiop",
		}

		_, err := v1alpha1.GenerateProvisioningRequest(context.Background(), ndbclient, dbSpec, reqData)
		t.Log(err)
		if err == nil {
			t.Errorf("GenerateProvisioningRequest should return an error when profiles are not found")
		}
	}
}
func TestGenerateProvisioningRequest(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		dbSpec := v1alpha1.DatabaseSpec{
			NDB: v1alpha1.NDB{
				Server:           "abc.def.ghi.jkl/v99/api",
				ClusterId:        "test-cluster-id",
				CredentialSecret: "test-credential-secret-name",
			},
			Instance: v1alpha1.Instance{
				DatabaseNames:        []string{"a", "b", "c", "d"},
				Type:                 dbType,
				DatabaseInstanceName: dbType + "-instance-test",
				TimeZone:             "UTC",
			},
		}

		reqData := map[string]interface{}{
			v1alpha1.NDB_PARAM_PASSWORD:       "qwerty",
			v1alpha1.NDB_PARAM_SSH_PUBLIC_KEY: "qwertyuiop",
		}

		request, _ := v1alpha1.GenerateProvisioningRequest(context.Background(), ndbclient, dbSpec, reqData)

		//Assert
		if request.DatabaseType != v1alpha1.GetDatabaseEngineName(dbType) {
			t.Errorf("Database Engine type mismatch. Want: %s, got: %s", v1alpha1.GetDatabaseEngineName(dbType), request.DatabaseType)
		}

		if request.SoftwareProfileId == "" || request.SoftwareProfileVersionId == "" {
			t.Logf("SoftwareProfileId or SoftwareProfileVersionId is empty")
		}
		if request.ComputeProfileId == "" {
			t.Logf("ComputeProfileId is empty")
		}
		if request.NetworkProfileId == "" {
			t.Logf("NetworkProfileId is empty")
		}
		if request.DbParameterProfileId == "" {
			t.Logf("DbParameterProfileId is empty")
		}
		if request.TimeMachineInfo.SlaId != NONE_SLA_ID {
			t.Logf("NONE time machine sla not selected")
		}
	}
}

func TestGenerateProvisioningRequestReturnsErrorIfDBPasswordIsEmpty(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response interface{}
			if r.URL.Path == "/profiles" {
				response = []v1alpha1.ProfileResponse{
					{
						Id:              "1",
						Name:            "a",
						Type:            "test type",
						EngineType:      "test engine",
						LatestVersionId: "v-id-1",
						Topology:        "test topology",
					},
				}
			} else if r.URL.Path == "/slas" {
				response = []v1alpha1.SLAResponse{
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
				}
			}
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		dbSpec := v1alpha1.DatabaseSpec{
			NDB: v1alpha1.NDB{
				Server:           "abc.def.ghi.jkl/v99/api",
				ClusterId:        "test-cluster-id",
				CredentialSecret: "test-credential-secret-name",
			},
			Instance: v1alpha1.Instance{
				DatabaseNames:        []string{"a", "b", "c", "d"},
				Type:                 dbType,
				DatabaseInstanceName: dbType + "-instance-test",
				TimeZone:             "UTC",
			},
		}

		reqData := map[string]interface{}{
			v1alpha1.NDB_PARAM_PASSWORD:       "",
			v1alpha1.NDB_PARAM_SSH_PUBLIC_KEY: "qwertyuiop",
		}

		_, err := v1alpha1.GenerateProvisioningRequest(context.Background(), ndbclient, dbSpec, reqData)
		t.Log(err)
		if err == nil {
			t.Errorf("GenerateProvisioningRequest should return an error when db password is empty")
		}
	}
}

func TestGenerateProvisioningRequestReturnsErrorIfSSHKeyIsEmpty(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response interface{}
			if r.URL.Path == "/profiles" {
				response = []v1alpha1.ProfileResponse{
					{
						Id:              "1",
						Name:            "a",
						Type:            "test type",
						EngineType:      "test engine",
						LatestVersionId: "v-id-1",
						Topology:        "test topology",
					},
				}
			} else if r.URL.Path == "/slas" {
				response = []v1alpha1.SLAResponse{
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
				}
			}
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		dbSpec := v1alpha1.DatabaseSpec{
			NDB: v1alpha1.NDB{
				Server:           "abc.def.ghi.jkl/v99/api",
				ClusterId:        "test-cluster-id",
				CredentialSecret: "test-credential-secret-name",
			},
			Instance: v1alpha1.Instance{
				DatabaseNames:        []string{"a", "b", "c", "d"},
				Type:                 dbType,
				DatabaseInstanceName: dbType + "-instance-test",
				TimeZone:             "UTC",
			},
		}

		reqData := map[string]interface{}{
			v1alpha1.NDB_PARAM_PASSWORD:       "qwertyuiop",
			v1alpha1.NDB_PARAM_SSH_PUBLIC_KEY: "",
		}

		_, err := v1alpha1.GenerateProvisioningRequest(context.Background(), ndbclient, dbSpec, reqData)
		t.Log(err)
		if err == nil {
			t.Errorf("GenerateProvisioningRequest should return an error when ssh key is empty")
		}
	}
}

func TestGetActionArgumentsByDatabaseType(t *testing.T) {
	cases := []struct {
		dbType       string
		expectedArgs []v1alpha1.ActionArgument
	}{
		{
			dbType: "mysql",
			expectedArgs: []v1alpha1.ActionArgument{
				{
					Name:  "listener_port",
					Value: "3306",
				},
			},
		},
		{
			dbType: "postgres",
			expectedArgs: []v1alpha1.ActionArgument{
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
					Name:  "backup_policy",
					Value: "primary_only",
				},
			},
		},
		{
			dbType:       "unsupported_database_type",
			expectedArgs: nil,
		},
	}

	for _, c := range cases {
		args := v1alpha1.GetActionArgumentsByDatabaseType(c.dbType)
		if args == nil && c.expectedArgs != nil {
			t.Errorf("Unexpected nil value for database type '%s'", c.dbType)
			continue
		}
		if args != nil && c.expectedArgs == nil {
			t.Errorf("Expected nil value for database type '%s', but got %v", c.dbType, args)
			continue
		}
		if args == nil && c.expectedArgs == nil {
			continue
		}

		actionArgs := args.GetActionArguments()
		if len(actionArgs) != len(c.expectedArgs) {
			t.Errorf("Expected %d action arguments for database type '%s', but got %d", len(c.expectedArgs), c.dbType, len(actionArgs))
			continue
		}

		for i, expectedArg := range c.expectedArgs {
			if expectedArg.Name != actionArgs[i].Name {
				t.Errorf("Expected action argument name '%s' for database type '%s', but got '%s'", expectedArg.Name, c.dbType, actionArgs[i].Name)
			}
			if expectedArg.Value != actionArgs[i].Value {
				t.Errorf("Expected action argument value '%s' for database type '%s' and name '%s', but got '%s'", expectedArg.Value, c.dbType, expectedArg.Name, actionArgs[i].Value)
			}
		}
	}
}