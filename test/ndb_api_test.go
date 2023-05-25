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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// func TestGetAllDatabases(t *testing.T) {

// 	server := util.GetServerTestHelper(t, "/databases")
// 	defer server.Close()
// 	ndb_client := ndb_client.NewNDBClient("username", "passwdord", server.URL)
// 	value := GetAllDatabases(ndb_client)
// 	t.Log(value)
// }

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
