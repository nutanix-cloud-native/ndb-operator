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

// ResolveSnapshotNameToId resolves a snapshot name to its UUID
// This requires the time machine ID and searches through snapshots for the matching name
// If multiple snapshots have the same name, returns the most recent one
func ResolveSnapshotNameToId(ctx context.Context, ndbClient *ndb_client.NDBClient, snapshotName string, timeMachineId string) (snapshotId string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Resolving snapshot name to ID", "snapshotName", snapshotName, "timeMachineId", timeMachineId)

	// Get all snapshots for the time machine
	snapshotsResponse, err := GetSnapshotsForTM(ctx, ndbClient, timeMachineId)
	if err != nil {
		log.Error(err, "Error fetching snapshots for time machine")
		return "", fmt.Errorf("failed to fetch snapshots: %w", err)
	}

	// Collect all snapshots with matching name
	type snapshotWithTimestamp struct {
		id        string
		timestamp int64
	}
	var matchingSnapshots []snapshotWithTimestamp

	// Search through all snapshots across all clusters
	for _, snapshotsPerCluster := range snapshotsResponse.SnapshotsPerNxCluster {
		for _, snapshotParent := range snapshotsPerCluster {
			for _, snapshot := range snapshotParent.Snapshots {
				// Fetch detailed snapshot info to get the name
				detailedSnapshot, err := GetSnapshotById(ctx, ndbClient, snapshot.Id)
				if err != nil {
					log.V(1).Info("Could not fetch detailed snapshot info", "snapshotId", snapshot.Id, "error", err)
					continue
				}
				if detailedSnapshot != nil && detailedSnapshot.Name == snapshotName {
					// Found a matching snapshot - add to list with timestamp
					matchingSnapshots = append(matchingSnapshots, snapshotWithTimestamp{
						id:        snapshot.Id,
						timestamp: detailedSnapshot.SnapshotTimeStampDate,
					})
				}
			}
		}
	}

	// If no snapshots found, return error
	if len(matchingSnapshots) == 0 {
		return "", fmt.Errorf("snapshot with name '%s' not found in time machine '%s'", snapshotName, timeMachineId)
	}

	// Find the most recent snapshot (highest timestamp)
	mostRecentSnapshot := matchingSnapshots[0]
	for _, s := range matchingSnapshots {
		if s.timestamp > mostRecentSnapshot.timestamp {
			mostRecentSnapshot = s
		}
	}

	log.Info("Resolved snapshot name to ID (most recent)", "snapshotName", snapshotName, "snapshotId", mostRecentSnapshot.id, "matchingCount", len(matchingSnapshots))
	return mostRecentSnapshot.id, nil
}

// GetSnapshotById fetches detailed snapshot information by ID
func GetSnapshotById(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface, snapshotId string) (snapshot *SnapshotResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if snapshotId == "" {
		err = fmt.Errorf("snapshot id is empty")
		log.Error(err, "no snapshot id provided")
		return
	}
	snapshotPath := fmt.Sprintf("snapshots/%s", snapshotId)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, snapshotPath, nil, &snapshot); err != nil {
		log.Error(err, "Error in GetSnapshotById")
		return
	}
	return
}
