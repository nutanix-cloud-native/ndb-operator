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

// Provisions a clone based on the clone provisioning request
// Returns the task info summary response for the operation
// Although, a clone is provisioned via the TimeMachine API,
// it is deprovisioned via clone APIs.
func ProvisionClone(ctx context.Context, ndbClient *ndb_client.NDBClient, req *DatabaseCloneRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.ProvisionClone")
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
	cloneEndpoint := "tms/" + req.TimeMachineId + "/clones"
	res, err := ndbClient.Post(cloneEndpoint, req)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("POST %s responded with %d", cloneEndpoint, res.StatusCode)
			} else {
				err = fmt.Errorf("POST %s responded with nil response", cloneEndpoint)
			}
		}
		log.Error(err, "Error occurred while cloning database")
		return
	}
	log.Info("POST "+cloneEndpoint, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in ProvisionClone")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.ProvisionClone")
	return
}

// Takes a snapshot for a time machine
// Returns the task info summary response for the operation TODO
func TakeSnapshotForTM(ctx context.Context, ndbClient *ndb_client.NDBClient, tmName string) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.TakeSnapshotForTM")
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
	requestBody := &SnapshotRequest{
		Name: "snapshot2",
		SnapshotLcmConfig: SnapshotLcmConfig{
			SnapshotLCMConfigDetailed: SnapshotLcmConfigDetailed{
				ExpiryDetails: ExpiryDetails{
					ExpiryDateTimezone: "Asia/Calcutta",
					ExpireInDays:       "10",
				},
			},
		},
	}

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
		log.Error(err, "Error occurred reading response.Body in TakeSnapshotForTM")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.TakeSnapshotForTM")
	return
}
