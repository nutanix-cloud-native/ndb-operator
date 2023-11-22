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

// Creates a snapshot for a time machine
// Returns the task info summary response for the operation TODO
func CreateSnapshotForTM(
	ctx context.Context,
	ndbClient *ndb_client.NDBClient,
	tmId string,
	snapshotName string,
	expiryDateTimezone string,
	ExpireInDays string) (task *TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	// Checking if id is empty, this is necessary to create a snapshot for the timemachine (/tms/{timemachine_id}/snapshots)
	if tmId == "" {
		err = fmt.Errorf("timemachine id is empty")
		log.Error(err, "no timemachine id provided")
		return
	}
	takeSnapshotPath := fmt.Sprintf("tms/%s/snapshots", tmId)
	req := GenerateSnapshotRequest(snapshotName, expiryDateTimezone, ExpireInDays)
	if _, err = sendRequest(ctx, ndbClient, http.MethodPost, takeSnapshotPath, req, &task); err != nil {
		log.Error(err, "Error in CreateSnapshotForTM")
		return
	}
	return
}

// Gets snapshots for a time machine
// Returns the task info summary response for the operation
func GetSnapshotsForTM(ctx context.Context, ndbClient *ndb_client.NDBClient, tmId string) (response *TimeMachineGetSnapshotsResponse, err error) {
	log := ctrllog.FromContext(ctx)
	// Checking if id is empty, this is necessary to get a snapshot for the timemachine (/tms/{timemachine_id}/snapshots)
	if tmId == "" {
		err = fmt.Errorf("timemachine id is empty")
		log.Error(err, "no timemachine id provided")
		return
	}
	getTmSnapshotsPath := fmt.Sprintf("tms/%s/snapshots", tmId)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, getTmSnapshotsPath, nil, &response); err != nil {
		log.Error(err, "Error in GetSnapshotsForTM")
		return
	}
	return
}

// Gets TimeMachine by id
func GetTimeMachineById(ctx context.Context, ndbClient *ndb_client.NDBClient, tmId string) (timeMachine *TimeMachineResponse, err error) {
	log := ctrllog.FromContext(ctx)
	// Checking if id is empty, this is necessary to get a timemachine (/tms/{timemachine_id})
	if tmId == "" {
		err = fmt.Errorf("timemachine id is empty")
		log.Error(err, "no timemachine id provided")
		return
	}
	getTmDetailedPath := fmt.Sprintf("tms/%s", tmId)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, getTmDetailedPath, nil, &timeMachine); err != nil {
		log.Error(err, "Error in GetTimeMachineById")
		return
	}
	return
}
