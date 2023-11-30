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

// Fetches all snapshots on the NDB instance and returns a slice of the snapshots
func GetAllSnapshots(ctx context.Context, ndbClient *ndb_client.NDBClient) (snapshots []SnapshotResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetAllSnapshots")
	if ndbClient == nil {
		err = errors.New("nil reference: received nil reference for ndbClient")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	res, err := ndbClient.Get("snapshots")
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /snapshots responded with %d", res.StatusCode)
			} else {
				err = fmt.Errorf("GET /snapshots responded with a nil response")
			}
		}
		log.Error(err, "Error occurred fetching all snapshots")
		return
	}
	log.Info("GET /snapshots", "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in GetAllSnapshots")
		return
	}
	err = json.Unmarshal(body, &snapshots)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetAllSnapshots")
	return
}

// Fetches and returns a snapshot by an Id
func GetSnapshotById(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (snapshot SnapshotResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetSnapshotById", "snapshotId", id)
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary otherwise the request becomes a call to get all snapshots (/snapshots)
	if id == "" {
		err = fmt.Errorf("snapshot id is empty")
		log.Error(err, "no snapshot id provided")
		return
	}
	getSnapshotPath := fmt.Sprintf("snapshots/%s", id)
	res, err := ndbClient.Get(getSnapshotPath)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET %s responded with %d", getSnapshotPath, res.StatusCode)
			} else {
				err = fmt.Errorf("GET %s responded with a nil response", getSnapshotPath)
			}
		}
		log.Error(err, "Error occurred fetching snapshot")
		return
	}
	log.Info(getSnapshotPath, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in Get Snapshot by ID")
		return
	}
	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetSnapshotById")
	return
}

// Takes a snapshot of the database upon request
// Returns the task info summary response for the operation
func TakeSnapshot(ctx context.Context, ndbClient *ndb_client.NDBClient, req *SnapshotRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.TakeSnapshot")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	if req.TimeMachineId == "" {
		err = errors.New("empty timeMachineId")
		log.Error(err, "Received empty timeMachineId in request")
		return
	}
	snapshotEndPoint := "tms/" + req.TimeMachineId + "/snapshots"
	res, err := ndbClient.Post(snapshotEndPoint, req)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("POST %s responded with %d", snapshotEndPoint, res.StatusCode)
			} else {
				err = fmt.Errorf("POST %s responded with nil response", snapshotEndPoint)
			}
		}
		log.Error(err, "Error taking snapshot of database")
		return
	}
	log.Info("POST "+snapshotEndPoint, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in TakeSnapshot")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.TakeSnapshot")
	return
}

// Deletes a snapshot given a snapshot id
// Returns the task info summary response for the operation
func DeleteSnapshot(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.DeleteSnapshot", "SnapshotId", id)
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no snapshot id provided")
		return
	}
	res, err := ndbClient.Delete("snapshots/"+id, nil)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("DELETE /snapshots/%s responded with %d", id, res.StatusCode)
			} else {
				err = fmt.Errorf("DELETE /snapshots/%s responded with nil response", id)
			}
		}
		log.Error(err, "Error occurred deleting snapshots")
		return
	}
	log.Info("DELETE /snapshots/"+id, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.DeleteSnapshot")
	return
}
