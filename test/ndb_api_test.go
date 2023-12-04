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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

func TestGetAllSLAs(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	value, _ := ndb_api.GetAllSLAs(context.Background(), ndb_client)
	if len(value) == 0 {
		t.Error("Could not fetch mock slas")
	}
}

func TestGetAllSLAsThrowsErrorWhenClientReturnsNon200(t *testing.T) {
	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	_, err := ndb_api.GetAllSLAs(context.Background(), ndb_client)
	if err == nil {
		t.Error("GetAllSLAs should return an error when client responds with non 200 status.")
	}
}

func TestGetAllProfiles(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	value, _ := ndb_api.GetAllProfiles(context.Background(), ndb_client)
	t.Log(len(value))
	if len(value) == 0 {
		t.Error("Could not fetch mock profiles")
	}
}

func TestGetAllProfileThrowsErrorWhenClientReturnsNon200(t *testing.T) {
	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	_, err := ndb_api.GetAllProfiles(context.Background(), ndb_client)
	if err == nil {
		t.Error("TestGetAllProfiles should return an error when client responds with non 200 status.")
	}
}

func TestGetAllSnapshots(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	value, _ := ndb_api.GetAllSnapshots(context.Background(), ndb_client)
	t.Log(len(value))
	if len(value) == 0 {
		t.Error("Could not fetch Snapshot profiles")
	}
}

func TestGetAllSnapshotsThrowsErrorWhenClientReturnsNon200(t *testing.T) {
	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	_, err := ndb_api.GetAllSnapshots(context.Background(), ndb_client)
	if err == nil {
		t.Error("GetAllSnapshots should return an error when client responds with non 200 status.")
	}
}

func TestGetSnapshotById(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	value, _ := ndb_api.GetSnapshotById(context.Background(), ndb_client, "id")
	t.Log(value)
	if value.Id != "id" {
		t.Error("Could not fetch Snapshot profiles")
	}
}

func TestGetSnapshotByIdThrowsErrorWhenClientReturnsNon200(t *testing.T) {
	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	_, err := ndb_api.GetSnapshotById(context.Background(), ndb_client, "id")
	if err == nil {
		t.Error("GetAllSnapshots should return an error when client responds with non 200 status.")
	}
}

func TestTakeSnapshot(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)
	exp := ndb_api.SnapshotExpiryDetails("timezone", 1)
	detailedConfig := ndb_api.SnapshotLcmConfigDetailed(exp)
	config := ndb_api.SnapshotLcmConfig(detailedConfig)
	request := ndb_api.SnapshotRequest("name", config)

	//Test
	value, _ := ndb_api.TakeSnapshot(context.Background(), ndb_client, request)
	t.Log(value)
	if value.Name != "name" {
		t.Error("Could not create Snapshot profiles")
	}
}

func TestTakeSnapshotThrowsErrorWhenClientReturnsNon200(t *testing.T) {
	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)
	exp := ndb_api.SnapshotExpiryDetails("timezone", 1)
	detailedConfig := ndb_api.SnapshotLcmConfigDetailed(exp)
	config := ndb_api.SnapshotLcmConfig(detailedConfig)
	request := ndb_api.SnapshotRequest("name", config)

	//Test
	_, err := ndb_api.TakeSnapshot(context.Background(), ndb_client, request)
	if err == nil {
		t.Error("GetAllSnapshots should return an error when client responds with non 200 status.")
	}
}

func TestDeleteSnapshot(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	value, _ := ndb_api.DeleteSnapshot(context.Background(), ndb_client, "id")
	t.Log(value)
	if value.EntityId != "id" {
		t.Error("Could not delete Snapshot profiles")
	}
}

func TestDeleteSnapshotThrowsErrorWhenClientReturnsNon200(t *testing.T) {
	//Set
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	//Test
	_, err := ndb_api.DeleteSnapshot(context.Background(), ndb_client, "id")
	if err == nil {
		t.Error("DeleteSnapshot should return an error when client responds with non 200 status.")
	}
}
