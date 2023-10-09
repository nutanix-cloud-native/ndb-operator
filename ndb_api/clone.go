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

// Fetches all the clones on the NDB instance and retutns a slice of the databases
func GetAllClones(ctx context.Context, ndbClient *ndb_client.NDBClient) (clones []DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetAllClones")
	if ndbClient == nil {
		err = errors.New("nil reference: received nil reference for ndbClient")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	res, err := ndbClient.Get("clones?detailed=true")
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /clones responded with %d", res.StatusCode)
			} else {
				err = fmt.Errorf("GET /clones responded with a nil response")
			}
		}
		log.Error(err, "Error occurred fetching all clones")
		return
	}
	log.Info("GET /clones", "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in GetAllClones")
		return
	}
	err = json.Unmarshal(body, &clones)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetAllClones")
	return
}

// Deprovisions a clone instance given a clone id
// Returns the task info summary response for the operation
func DeprovisionClone(ctx context.Context, ndbClient *ndb_client.NDBClient, id string, req CloneDeprovisionRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.DeprovisionClone", "cloneId", id)
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no clone id provided")
		return
	}
	res, err := ndbClient.Delete("clones/"+id, req)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("DELETE /clones/%s responded with %d", id, res.StatusCode)
			} else {
				err = fmt.Errorf("DELETE /clones/%s responded with nil response", id)
			}
		}
		log.Error(err, "Error occurred deprovisioning clones")
		return
	}
	log.Info("DELETE /clones/"+id, "HTTP status code", res.StatusCode)
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
	log.Info("Returning from ndb_api.DeprovisionClone")
	return
}
