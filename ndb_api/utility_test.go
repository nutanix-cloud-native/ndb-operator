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
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const mock_username = "username"
const mock_password = "password"

const NONE_SLA_ID = "NONE_SLA_ID"

func getMockedResponseMap() map[string]interface{} {
	return map[string]interface{}{
		"GET /slas":     getMockSLAResponses(),
		"GET /profiles": getMockProfileResponses(),
	}
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
	mockResponsesMap := getMockedResponseMap()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response = mockResponsesMap[r.Method+" "+r.URL.Path]
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
}

// responseMap holds the responses that will be returned by this server.
// The key is in the format [Method endpoint], ex - GET /slas, GET /profiles.
func GetServerTestHelperWithResponseMap(t *testing.T, responseMap map[string]interface{}) *httptest.Server {

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuthTestHelper(r) {
			t.Errorf("Invalid Authentication Credentials")
		} else {
			var response = responseMap[r.Method+" "+r.URL.Path]
			resp, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}))
}
