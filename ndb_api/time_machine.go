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

// Creates a snapshot for a time machine
// Returns the task info summary response for the operation TODO
func CreateSnapshotForTM(
	ctx context.Context,
	ndbClient *ndb_client.NDBClient,
	tmName string,
	snapshotName string,
	expiryDateTimezone string,
	ExpireInDays int) (task TaskInfoSummaryResponse, err error) {

	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.CreateSnapshotForTM")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary to create a snapshot for the timemachine (/tms/{timemachine_id}/snapshots)
	if tmName == "" {
		err = fmt.Errorf("snapshot id is empty")
		log.Error(err, "no snapshot id provided")
		return
	}
	takeSnapshotPath := fmt.Sprintf("tms/%s/snapshots", tmName)
	requestBody := GenerateSnapshotRequest(snapshotName, expiryDateTimezone, ExpireInDays)

	res, err := ndbClient.Post(takeSnapshotPath, requestBody)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("POST %s responded with %d", takeSnapshotPath, res.StatusCode)
			} else {
				err = fmt.Errorf("POST %s responded with nil response", takeSnapshotPath)
			}
		}

		log.Error(err, "Error occurred taking TM snapshot")
		return
	}

	log.Info(fmt.Sprintf("POST %s", takeSnapshotPath), "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in CreateSnapshotForTM")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.CreateSnapshotForTM")
	return
}

// Gets snapshots for a time machine
// Returns the task info summary response for the operation TODO
func GetSnapshotsForTM(ctx context.Context, ndbClient *ndb_client.NDBClient, tmId string) (response TimeMachineGetSnapshotsResponse, err error) {

	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetSnapshotsForTM")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary to create a snapshot for the timemachine (/tms/{timemachine_id}/snapshots)
	if tmId == "" {
		err = fmt.Errorf("snapshot id is empty")
		log.Error(err, "no snapshot id provided")
		return
	}
	getTmSnapshotsPath := fmt.Sprintf("tms/%s/snapshots", tmId)
	res, err := ndbClient.Get(getTmSnapshotsPath)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET %s responded with %d", getTmSnapshotsPath, res.StatusCode)
			} else {
				err = fmt.Errorf("GET %s responded with nil response", getTmSnapshotsPath)
			}
		}

		log.Error(err, "Error occurred taking TM snapshot")
		return
	}

	log.Info(fmt.Sprintf("GET %s", getTmSnapshotsPath), "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in GetSnapshotsForTM")
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetSnapshotsForTM")
	return
}

// Gets TimeMachine by id
func GetTimeMachineById(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (timeMachine TimeMachineResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetTimeMachineById", "timeMachineId", id)

	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary otherwise the request becomes a call to get all databases (/databases)
	if id == "" {
		err = fmt.Errorf("timemachine id is empty")
		log.Error(err, "no timemachine id provided")
		return
	}

	getTmDetailedPath := fmt.Sprintf("tms/%s", id)
	res, err := ndbClient.Get(getTmDetailedPath)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET %s responded with %d", getTmDetailedPath, res.StatusCode)
			} else {
				err = fmt.Errorf("GET %s responded with a nil response", getTmDetailedPath)
			}
		}
		log.Error(err, "Error occurred fetching Time Machine")
		return
	}
	log.Info(getTmDetailedPath, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in Get TM by ID")
		return
	}
	err = json.Unmarshal(body, &timeMachine)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetTimeMachineById")
	return
}
