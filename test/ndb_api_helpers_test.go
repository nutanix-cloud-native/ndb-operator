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
	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
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

func TestMatchAndGetProfilesWhenProfilesMatch(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}

	for _, dbType := range dbTypes {

		// get custom profile based upon the database type
		customProfile := GetCustomProfileForDBType(dbType)

		profileMap, _ := v1alpha1.GetProfiles(context.Background(), ndbclient, dbType, customProfile)

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
			if reflect.DeepEqual(profile, v1alpha1.ProfileResponse{}) {
				t.Errorf("Empty profile type %s for dbType %s", profileType, dbType)
			}
			//Assert that profile EngineType matches the database engine or the generic type
			if profile.EngineType != v1alpha1.GetDatabaseEngineName(dbType) && profile.EngineType != v1alpha1.DATABASE_ENGINE_TYPE_GENERIC {
				t.Errorf("Profile engine type %s for dbType %s does not match", profile.EngineType, dbType)
			}
			obtainedProfile := v1alpha1.GetProfileByType(profileType, customProfile)
			// Ignoring Storage Profile Type as the Profile struct currently only supports compute, software, network and dbParam
			if profileType != v1alpha1.PROFILE_TYPE_STORAGE && profile.Id != obtainedProfile.Id && profile.LatestVersionId == obtainedProfile.VersionId {
				t.Errorf("Custom Profile Enrichment failed for profileType = %s and dbType = %s", profileType, dbType)
			}
		}
	}
}

func TestMatchAndGetProfilesWhenNonMatchingProfilesProvided(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres_invalid_profiles", "mysql_invalid_profiles", "mongodb_invalid_profiles"}

	for _, dbType := range dbTypes {

		// get custom profile based upon the database type
		customProfile := GetCustomProfileForDBType(dbType)

		profileMap, _ := v1alpha1.GetProfiles(context.Background(), ndbclient, dbType, customProfile)

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
			//Assert that profile EngineType matches the database engine or the generic type
			if profile.EngineType != v1alpha1.GetDatabaseEngineName(dbType) && profile.EngineType != v1alpha1.DATABASE_ENGINE_TYPE_GENERIC {
				t.Errorf("Profile engine type %s for dbType %s does not match", profile.EngineType, dbType)
			}
			/*
				Since custom profile is passed it should not default to OOB, and err should be raised stating the custom profile passed does
				not exist, and thus database provisioning does not occur
			*/
			if !reflect.DeepEqual(profile, v1alpha1.ProfileResponse{}) {
				t.Errorf("Incorrect Profile Match found for profile type = %s and dbType = %s", profileType, dbType)
			}
		}
	}
}

func TestMatchAndGetProfilesForDefaultProfiles(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		profileMap, _ := v1alpha1.GetProfiles(context.Background(), ndbclient, dbType, v1alpha1.Profiles{})

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
			if reflect.DeepEqual(profile, v1alpha1.ProfileResponse{}) {
				t.Errorf("Empty profile type %s for dbType %s", profileType, dbType)
			}
			//Assert that profile EngineType matches the database engine or the generic type
			if profile.EngineType != v1alpha1.GetDatabaseEngineName(dbType) && profile.EngineType != v1alpha1.DATABASE_ENGINE_TYPE_GENERIC {
				t.Errorf("Profile engine type %s for dbType %s does not match", profile.EngineType, dbType)
			}
		}
	}
}

func TestMatchAndGetProfilesOnlyGetsTheSmallOOBComputeProfile(t *testing.T) {

	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndbclient := ndbclient.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	dbTypes := []string{"postgres", "mysql", "mongodb"}
	for _, dbType := range dbTypes {
		profileMap, _ := v1alpha1.GetProfiles(context.Background(), ndbclient, dbType, v1alpha1.Profiles{})
		//Assert
		computeProfile := profileMap[v1alpha1.PROFILE_TYPE_COMPUTE]
		if !strings.Contains(strings.ToLower(computeProfile.Name), "small") {
			t.Errorf("Expected small oob compute profile, but got: %s", computeProfile.Name)
		}
	}
}

func TestMatchAndGetProfilesReturnsErrorWhenSomeProfileIsNotFound(t *testing.T) {

	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			resp, _ := json.Marshal([]v1alpha1.ProfileResponse{
				{
					Id:              "112",
					Name:            "a",
					Type:            "test type",
					EngineType:      "test engine",
					LatestVersionId: "v-id-1",
					Topology:        "test topology",
					Versions: []v1alpha1.Version{
						{
							Id:          "version-id",
							Name:        "version-name",
							Description: "version-description",
						},
					},
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
		_, err := v1alpha1.GetProfiles(context.Background(), ndbclient, dbType, v1alpha1.Profiles{})
		// None of the profile criteria should match the mocked response
		// t.Log(err)
		if err == nil {
			t.Errorf("GetProfiles should have returned an error when none of the profiles match the criteria.")
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
	// Test with MySQL database type
	MySQLExpectedArgs, err := v1alpha1.GetActionArgumentsByDatabaseType(v1alpha1.DATABASE_TYPE_MYSQL)

	if err != nil {
		t.Error("Error while fetching mysql args", "err", err)
	}

	expectedMySqlArgs := []v1alpha1.ActionArgument{
		{
			Name:  "listener_port",
			Value: "3306",
		},
	}

	if !reflect.DeepEqual(MySQLExpectedArgs.GetActionArguments(v1alpha1.DatabaseSpec{Instance: v1alpha1.Instance{DatabaseInstanceName: "test"}}), expectedMySqlArgs) {
		t.Errorf("Expected %v, but got %v", expectedMySqlArgs, MySQLExpectedArgs.GetActionArguments(v1alpha1.DatabaseSpec{Instance: v1alpha1.Instance{DatabaseInstanceName: "test"}}))
	}

	// Test with Postgres database type
	postgresArgs, err := v1alpha1.GetActionArgumentsByDatabaseType(v1alpha1.DATABASE_TYPE_POSTGRES)
	if err != nil {
		t.Error("Error while fetching postgres args", "err", err)
	}
	expectedPostgresArgs := []v1alpha1.ActionArgument{
		{Name: "proxy_read_port", Value: "5001"},
		{Name: "listener_port", Value: "5432"},
		{Name: "proxy_write_port", Value: "5000"},
		{Name: "enable_synchronous_mode", Value: "false"},
		{Name: "auto_tune_staging_drive", Value: "true"},
		{Name: "backup_policy", Value: "primary_only"},
	}
	if !reflect.DeepEqual(postgresArgs.GetActionArguments(v1alpha1.DatabaseSpec{Instance: v1alpha1.Instance{DatabaseInstanceName: "test"}}), expectedPostgresArgs) {
		t.Errorf("Expected %v, but got %v", expectedPostgresArgs, postgresArgs.GetActionArguments(v1alpha1.DatabaseSpec{Instance: v1alpha1.Instance{DatabaseInstanceName: "test"}}))
	}

	// Test with MongoDB database type
	mongodbArgs, err := v1alpha1.GetActionArgumentsByDatabaseType(v1alpha1.DATABASE_TYPE_MONGODB)
	if err != nil {
		t.Error("Error while fetching mongodbArgs", "err", err)
	}
	expectedMongodbArgs := []v1alpha1.ActionArgument{
		{Name: "listener_port", Value: "27017"},
		{Name: "log_size", Value: "100"},
		{Name: "journal_size", Value: "100"},
		{Name: "restart_mongod", Value: "true"},
		{Name: "working_dir", Value: "/tmp"},
		{Name: "db_user", Value: "test"},
		{Name: "backup_policy", Value: "primary_only"},
	}
	if !reflect.DeepEqual(mongodbArgs.GetActionArguments(v1alpha1.DatabaseSpec{Instance: v1alpha1.Instance{DatabaseInstanceName: "test"}}), expectedMongodbArgs) {
		t.Errorf("Expected %v, but got %v", expectedMongodbArgs, mongodbArgs.GetActionArguments(v1alpha1.DatabaseSpec{Instance: v1alpha1.Instance{DatabaseInstanceName: "test"}}))
	}

	// Test with unknown database type
	unknownArgs, err := v1alpha1.GetActionArgumentsByDatabaseType("unknown")
	if err == nil {
		t.Errorf("Expected error for unknown database type, but got %v", unknownArgs)
	}
}
