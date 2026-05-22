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
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func GenerateLinkedDatabaseProvisionRequest(databaseName string) (*LinkedDatabaseProvisionRequest, error) {
	databaseName = strings.TrimSpace(databaseName)
	if databaseName == "" {
		return nil, fmt.Errorf("database name is empty")
	}
	return &LinkedDatabaseProvisionRequest{
		Databases: []LinkedDatabaseRequest{{DatabaseName: databaseName}},
	}, nil
}

func ProvisionLinkedDatabase(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, sourceDatabaseId string, req *LinkedDatabaseProvisionRequest) (task *TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	sourceDatabaseId = strings.TrimSpace(sourceDatabaseId)
	if sourceDatabaseId == "" {
		err = fmt.Errorf("source database id is empty")
		log.Error(err, "no source database id provided")
		return
	}
	path := fmt.Sprintf("databases/%s/linked-databases", sourceDatabaseId)
	if _, err = sendRequest(ctx, ndbClient, http.MethodPost, path, req, &task); err != nil {
		log.Error(err, "Error in ProvisionLinkedDatabase")
		return
	}
	return
}

func GetLinkedDatabaseByName(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, sourceDatabaseId, databaseName string) (*LinkedDatabaseResponse, error) {
	sourceDatabaseId = strings.TrimSpace(sourceDatabaseId)
	if sourceDatabaseId == "" {
		return nil, fmt.Errorf("source database id is empty")
	}
	databaseName = strings.TrimSpace(databaseName)
	if databaseName == "" {
		return nil, fmt.Errorf("database name is empty")
	}

	sourceDatabase, err := GetDatabaseById(ctx, ndbClient, sourceDatabaseId)
	if err != nil {
		return nil, err
	}
	if sourceDatabase == nil {
		return nil, nil
	}
	for _, linkedDatabase := range sourceDatabase.LinkedDatabases {
		if linkedDatabase.DatabaseName == databaseName {
			return &linkedDatabase, nil
		}
	}
	return nil, nil
}
