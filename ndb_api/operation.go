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

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Fetches and returns a operation by an Id
func GetOperationById(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (operation OperationResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if ndbClient == nil {
		err = errors.New("nil reference for ndbClient")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary otherwise the request becomes a call to get all operations (/operations)
	if id == "" {
		err = fmt.Errorf("operation id is empty")
		log.Error(err, "no operation id provided")
		return
	}
	// ?display=true is added to fetch the operations even when it
	// is not yet created in the operation table on NDB. It causes
	// the operation details to be fetched from the WORK table instead.
	getOperationPath := fmt.Sprintf("operations/%s?display=true", id)
	res, err := ndbClient.Get(getOperationPath)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET %s responded with %d", getOperationPath, res.StatusCode)
			} else {
				err = fmt.Errorf("GET %s responded with a nil response", getOperationPath)
			}
		}
		log.Error(err, "Error occurred fetching operation", "operationId", id)
		return
	}
	log.Info(getOperationPath, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in Get Operation by ID", "operation id", id)
		return
	}
	err = json.Unmarshal(body, &operation)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	return
}
