/*
Copyright 2021-2022 Nutanix, Inc.

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

package v1alpha1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nutanix-cloud-native/ndb-operator/ndbclient"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func GetAllDatabases(ctx context.Context, ndbClient *ndbclient.NDBClient) (databases []DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetAllDatabases")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	res, err := ndbClient.Get("databases")
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /databases responded with %d", res.StatusCode)
			} else {
				err = fmt.Errorf("GET /databases responded with a nil response")
			}
		}
		log.Error(err, "Error occured fetching all databases")
		return
	}
	log.Info("GET /databases", "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body in GetAllDatabases")
		return
	}
	err = json.Unmarshal(body, &databases)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetAllDatabases")
	return
}

func GetDatabaseById(ctx context.Context, ndbClient *ndbclient.NDBClient, id string) (database DatabaseResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetDatabaseById", "databaseId", id)
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	if id == "" {
		err = fmt.Errorf("database id is empty")
		log.Error(err, "no database id provided")
		return
	}
	res, err := ndbClient.Get("databases/" + id)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /databases/%s responded with %d", id, res.StatusCode)
			} else {
				err = fmt.Errorf("GET /databases/%s responded with a nil response", id)
			}
		}
		log.Error(err, "Error occured fetching database")
		return
	}
	log.Info("GET /databases/"+id, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body in GetAllDatabases")
		return
	}
	err = json.Unmarshal(body, &database)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetDatabaseById")
	return
}

func ProvisionDatabase(ctx context.Context, ndbClient *ndbclient.NDBClient, req *DatabaseProvisionRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.ProvisionDatabase")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	res, err := ndbClient.Post("databases/provision", req)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("POST databases/provision responded with %d", res.StatusCode)
			} else {
				err = fmt.Errorf("POST databases/provision responded with nil response")
			}
		}
		log.Error(err, "Error occured provisioning database")
		return
	}
	log.Info("POST databases/provision", "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body in ProvisionDatabase")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.ProvisionDatabase")
	return
}

func GetAllProfiles(ctx context.Context, ndbClient *ndbclient.NDBClient) (profiles []ProfileResponse, err error) {
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
		log.Error(err, "Error occured while fetching profiles")
		return
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body in GetAllProfiles")
		return
	}
	err = json.Unmarshal(body, &profiles)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetAllProfiles")
	return
}

func GetAllSLAs(ctx context.Context, ndbClient *ndbclient.NDBClient) (slas []SLAResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetAllSLAs")
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	res, err := ndbClient.Get("slas")
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /slas responded with %d", res.StatusCode)
			} else {
				err = fmt.Errorf("GET /slas responded with nil response")
			}
		}
		log.Error(err, "Error occured while fetching slas")
		return
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body in GetAllSLAs")
		return
	}
	err = json.Unmarshal(body, &slas)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetAllSLAs")
	return
}

func DeprovisionDatabase(ctx context.Context, ndbClient *ndbclient.NDBClient, id string, req DatabaseDeprovisionRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.DeprovisionDatabase", "databaseId", id)
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no database id provided")
		return
	}
	res, err := ndbClient.Delete("databases/"+id, req)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("DELETE /databases/%s responded with %d", id, res.StatusCode)
			} else {
				err = fmt.Errorf("DELETE /databases/%s responded with nil response", id)
			}
		}
		log.Error(err, "Error occured deprovisioning database")
		return
	}
	log.Info("DELETE /databases/"+id, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.DeprovisionDatabase")
	return
}

func DeprovisionDatabaseServer(ctx context.Context, ndbClient *ndbclient.NDBClient, id string, req DatabaseServerDeprovisionRequest) (task TaskInfoSummaryResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.DeprovisionDatabaseServer", "dbServerId", id)
	if ndbClient == nil {
		err = errors.New("nil reference")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	if id == "" {
		err = fmt.Errorf("id is empty")
		log.Error(err, "no database server id provided")
		return
	}
	res, err := ndbClient.Delete("dbservers/"+id, req)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("DELETE /dbservers/%s responded with %d", id, res.StatusCode)
			} else {
				err = fmt.Errorf("DELETE /dbservers/%s responded with nil response", id)
			}
		}
		log.Error(err, "Error occured deprovisioning database server")
		return
	}
	log.Info("DELETE /dbservers/"+id, "HTTP status code", res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occured reading response.Body in DeprovisionDatabaseServer")
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		log.Error(err, "Error occured trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.DeprovisionDatabaseServer")
	return
}
