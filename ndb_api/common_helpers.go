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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Makes the request via the ndb http client to ndb.
// Used by functions in the ndb_api package.
func sendRequest(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, method, endpoint string, payload interface{}, responseBody interface{}) (resp *http.Response, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.sendRequest", "Method", method, "Endpoint", endpoint)

	if ndbClient == nil {
		err = errors.New("nil reference: received nil reference for ndbClient")
		return
	}

	if responseBody == nil {
		err = errors.New("nil reference: received nil reference for responseBody, expected an initialized/non-nil variable")
		return
	}

	req, err := ndbClient.NewRequest(method, endpoint, payload)
	if err != nil {
		log.Error(err, "An error occurred while creating the HTTP request")
		return
	}

	res, err := ndbClient.Do(req)
	if err != nil {
		log.Error(err, "An error occurred while calling the HTTP endpoint")
		return
	}

	// Read the NDB API response.
	// Perform error checks.
	// Unmarshal into the response body passed by the caller.
	if res != nil {
		body, readErr := io.ReadAll(res.Body)
		defer res.Body.Close()
		if readErr != nil {
			log.Error(readErr, "Error occurred reading response.Body")
			return
		}
		// Considering any status >= 400 to be an error.
		if res.StatusCode >= http.StatusBadRequest {
			err = fmt.Errorf("%s %s error", method, endpoint)
			if body != nil {
				err = errors.Join(err, fmt.Errorf("ndb api error response status: %d, response body: %s", res.StatusCode, body))
			}
			log.Error(err, "NDB API error")
			return
		}
		err = json.Unmarshal(body, responseBody)
		if err != nil {
			log.Error(err, "Error occurred while unmarshalling the response body")
			return
		}
	} else {
		log.Info("Received an empty/nil response")
	}
	log.Info("Returning from ndb_api.sendRequest")
	return
}

func GetDatabaseEngineName(dbType string) string {
	switch dbType {
	case common.DATABASE_TYPE_POSTGRES:
		return common.DATABASE_ENGINE_TYPE_POSTGRES
	case common.DATABASE_TYPE_MYSQL:
		return common.DATABASE_ENGINE_TYPE_MYSQL
	case common.DATABASE_TYPE_MONGODB:
		return common.DATABASE_ENGINE_TYPE_MONGODB
	case common.DATABASE_TYPE_MSSQL:
		return common.DATABASE_ENGINE_TYPE_MSSQL
	default:
		return ""
	}
}

func GetDatabaseTypeFromEngine(engine string) string {
	switch engine {
	case common.DATABASE_ENGINE_TYPE_POSTGRES:
		return common.DATABASE_TYPE_POSTGRES
	case common.DATABASE_ENGINE_TYPE_MYSQL:
		return common.DATABASE_TYPE_MYSQL
	case common.DATABASE_ENGINE_TYPE_MONGODB:
		return common.DATABASE_TYPE_MONGODB
	case common.DATABASE_ENGINE_TYPE_MSSQL:
		return common.DATABASE_TYPE_MSSQL
	default:
		return ""
	}
}

func GetDatabasePortByType(dbType string) int32 {
	switch dbType {
	case common.DATABASE_TYPE_POSTGRES:
		return common.DATABASE_DEFAULT_PORT_POSTGRES
	case common.DATABASE_TYPE_MONGODB:
		return common.DATABASE_DEFAULT_PORT_MONGODB
	case common.DATABASE_TYPE_MYSQL:
		return common.DATABASE_DEFAULT_PORT_MYSQL
	case common.DATABASE_TYPE_MSSQL:
		return common.DATABASE_DEFAULT_PORT_MSSQL
	default:
		return -1
	}
}

// Get specific implementation of the DBProvisionRequestAppender interface based on the provided databaseType
func GetRequestAppender(databaseType string, isHighAvailability bool) (requestAppender RequestAppender, err error) {
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		requestAppender = &MySqlRequestAppender{}
	case common.DATABASE_TYPE_POSTGRES:
		if isHighAvailability {
			requestAppender = &PostgresHARequestAppender{}
		} else {
			requestAppender = &PostgresRequestAppender{}
		}
	case common.DATABASE_TYPE_MONGODB:
		requestAppender = &MongoDbRequestAppender{}
	case common.DATABASE_TYPE_MSSQL:
		requestAppender = &MSSQLRequestAppender{}
	default:
		return nil, errors.New("invalid database type: supported values: mssql, mysql, postgres, mongodb")
	}
	return
}
