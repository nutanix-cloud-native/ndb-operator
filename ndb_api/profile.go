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

// Fetches and returns all the available profiles as a profile slice
func GetAllProfiles(ctx context.Context, ndbClient *ndb_client.NDBClient) (profiles []ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetAllProfiles")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	res, err := ndbClient.Get("profiles")
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /profiles responded with %d", res.StatusCode)
			} else {
				err = fmt.Errorf("GET /profiles responded with nil response")
			}
		}
		log.Error(err, "Error occurred while fetching profiles")
		return
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in GetAllProfiles")
		return
	}
	err = json.Unmarshal(body, &profiles)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetAllProfiles")
	return
}
