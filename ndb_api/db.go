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
	"fmt"
	"net/http"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Fetches all the databases on the NDB instance and retutns a slice of the databases
func GetAllDatabases(ctx context.Context, ndbClient *ndb_client.NDBClient) (databases []DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "databases?detailed=true", nil, &databases); err != nil {
		log.Error(err, "Error in GetAllDatabases")
		return
	}
	return
}

// Fetches and returns a database by an Id
func GetDatabaseById(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (database DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	// Checking if id is empty, this is necessary otherwise the request becomes a call to get all databases (/databases)
	if id == "" {
		err = fmt.Errorf("database id is empty")
		log.Error(err, "no database id provided")
		return
	}
	getDbDetailedPath := fmt.Sprintf("databases/%s?detailed=true", id)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, getDbDetailedPath, nil, &database); err != nil {
		log.Error(err, "Error in GetDatabaseById")
		return
	}
	return
}

// Provisions a database instance based on the database provisioning request
// Returns the task info summary response for the operation
func ProvisionDatabase(ctx context.Context, ndbClient *ndb_client.NDBClient, req *DatabaseProvisionRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if _, err = sendRequest(ctx, ndbClient, http.MethodPost, "databases/provision", req, &task); err != nil {
		log.Error(err, "Error in ProvisionDatabase")
		return
	}
	return
}

// Deprovisions a database instance given a database id
// Returns the task info summary response for the operation
func DeprovisionDatabase(ctx context.Context, ndbClient *ndb_client.NDBClient, id string, req *DatabaseDeprovisionRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no database id provided")
		return
	}
	if _, err = sendRequest(ctx, ndbClient, http.MethodDelete, "databases/"+id, req, &task); err != nil {
		log.Error(err, "Error in DeprovisionDatabase")
		return
	}
	return
}
