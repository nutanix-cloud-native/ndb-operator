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

	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ResolveClusterNameToId resolves a cluster name to its UUID by fetching clusters from NDB API.
// Note: This function requires the clusters API endpoint to be available in NDB.
// If the clusters API is not available, cluster name resolution will fail.
// Clusters must have unique names in NDB for this to work correctly.
func ResolveClusterNameToId(ctx context.Context, ndbClient *ndb_client.NDBClient, clusterName string) (clusterId string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Resolving cluster name to ID", "clusterName", clusterName)

	// Try to fetch clusters directly from NDB API
	clusters, err := GetAllClusters(ctx, ndbClient)
	if err != nil {
		return "", fmt.Errorf("clusters API endpoint not available or error fetching clusters: %w. Cluster name resolution requires the clusters API endpoint", err)
	}

	// Search for cluster with matching name
	cluster, err := util.FindFirst(clusters, func(c ClusterResponse) bool {
		return c.Name == clusterName
	})
	if err != nil {
		return "", fmt.Errorf("cluster with name '%s' not found", clusterName)
	}

	log.Info("Resolved cluster name to ID", "clusterName", clusterName, "clusterId", cluster.Id)
	return cluster.Id, nil
}

// GetAllClusters fetches all clusters from NDB
// Returns empty slice if clusters API endpoint doesn't exist
func GetAllClusters(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface) (clusters []ClusterResponse, err error) {
	log := ctrllog.FromContext(ctx)
	// Try to fetch clusters - this endpoint may not exist in all NDB versions
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "clusters", nil, &clusters); err != nil {
		log.V(1).Info("Clusters endpoint not available or error fetching clusters", "error", err)
		return []ClusterResponse{}, fmt.Errorf("clusters endpoint not available: %w", err)
	}
	return
}

// ClusterResponse represents a cluster in NDB
type ClusterResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
