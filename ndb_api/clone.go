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
	"errors"
	"fmt"
	"net/http"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Fetches all the clones on the NDB instance and retutns a slice of the databases
func GetAllClones(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface) (clones []DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "clones?detailed=true", nil, &clones); err != nil {
		log.Error(err, "Error in GetAllClones")
		return
	}
	return
}

// Provisions a clone based on the clone provisioning request
// Returns the task info summary response for the operation
func ProvisionClone(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, req *DatabaseCloneRequest) (task *TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if req.TimeMachineId == "" {
		err = errors.New("empty timeMachineId")
		log.Error(err, "Received empty timeMachineId in request")
		return
	}
	cloneEndpoint := "tms/" + req.TimeMachineId + "/clones"
	if _, err = sendRequest(ctx, ndbClient, http.MethodPost, cloneEndpoint, req, &task); err != nil {
		log.Error(err, "Error in ProvisionClone")
		return
	}
	return
}

// Fetches clone by id
func GetCloneById(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (clone DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetCloneById")
	if ndbClient == nil {
		err = errors.New("nil reference: received nil reference for ndbClient")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary otherwise the request becomes a call to get all databases (/databases)
	if id == "" {
		err = fmt.Errorf("clone id is empty")
		log.Error(err, "no clone id provided")
		return
	}
	getCloneIdPath := fmt.Sprintf("clones/%s", id)
	res, err := ndbClient.Get(getCloneIdPath)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /%s responded with %d", getCloneIdPath, res.StatusCode)
			} else {
				err = fmt.Errorf("GET /%s responded with a nil response", getCloneIdPath)
			}
		}
		log.Error(err, "Error occurred fetching all snapshots")
		return
	}
	log.Info("GET /%s", "HTTP status code", getCloneIdPath, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in GetCloneById")
		return
	}
	err = json.Unmarshal(body, &clone)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetCloneById")
	return
}

// Deprovisions a clone instance given a clone id
// Returns the task info summary response for the operation
func DeprovisionClone(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, id string, req *CloneDeprovisionRequest) (task *TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no clone id provided")
		return
	}
	if _, err = sendRequest(ctx, ndbClient, http.MethodDelete, "clones/"+id, req, &task); err != nil {
		log.Error(err, "Error in DeprovisionClone")
		return
	}
	return
}
