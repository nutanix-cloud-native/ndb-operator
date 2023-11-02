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

// Fetches all the snapshots on the NDB instance and returns a slice of the snapshots
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

// Fetches and returns a snapshot by id
func GetSnapshotById(ctx context.Context, ndbClient *ndb_client.NDBClient, id string) (snapshots SnapshotResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered ndb_api.GetSnapshotById")
	if ndbClient == nil {
		err = errors.New("nil reference: received nil reference for ndbClient")
		log.Error(err, "Received nil ndbClient reference")
		return
	}
	// Checking if id is empty, this is necessary otherwise the request becomes a call to get all databases (/databases)
	if id == "" {
		err = fmt.Errorf("snapshot id is empty")
		log.Error(err, "no snapshot id provided")
		return
	}
	getSnapshotIdPath := fmt.Sprintf("snapshots/%s", id)
	res, err := ndbClient.Get(getSnapshotIdPath)
	if err != nil || res == nil || res.StatusCode != http.StatusOK {
		if err == nil {
			if res != nil {
				err = fmt.Errorf("GET /%s responded with %d", getSnapshotIdPath, res.StatusCode)
			} else {
				err = fmt.Errorf("GET /%s responded with a nil response", getSnapshotIdPath)
			}
		}
		log.Error(err, "Error occurred fetching all snapshots")
		return
	}
	log.Info("GET /%s", "HTTP status code", getSnapshotIdPath, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error(err, "Error occurred reading response.Body in GetSnapshotById")
		return
	}
	err = json.Unmarshal(body, &snapshots)
	if err != nil {
		log.Error(err, "Error occurred trying to unmarshal.")
		return
	}
	log.Info("Returning from ndb_api.GetSnapshotById")
	return
}
