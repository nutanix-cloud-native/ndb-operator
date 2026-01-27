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

package controller_adapters

import (
	"context"
	"fmt"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ResolveNamesToUUIDs resolves name fields to UUID fields in the database spec
// This function should be called before making NDB API calls that require UUIDs
func ResolveNamesToUUIDs(ctx context.Context, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) error {
	log := ctrllog.FromContext(ctx)
	log.Info("Resolving names to UUIDs for database", "name", database.Name)

	if database.Spec.IsClone {
		return resolveCloneNamesToUUIDs(ctx, ndbClient, database)
	}
	return resolveInstanceNamesToUUIDs(ctx, ndbClient, database)
}

// resolveInstanceNamesToUUIDs resolves names to UUIDs for provisioning instances
func resolveInstanceNamesToUUIDs(ctx context.Context, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) error {
	instance := database.Spec.Instance

	// Resolve cluster name to ID if needed
	if err := resolveClusterNameToId(ctx, ndbClient, instance.ClusterId, instance.ClusterName, &instance.ClusterId); err != nil {
		return err
	}

	return nil
}

// resolveCloneNamesToUUIDs resolves names to UUIDs for cloning
func resolveCloneNamesToUUIDs(ctx context.Context, ndbClient *ndb_client.NDBClient, database *ndbv1alpha1.Database) error {
	clone := database.Spec.Clone

	// Resolve cluster name to ID if needed
	if err := resolveClusterNameToId(ctx, ndbClient, clone.ClusterId, clone.ClusterName, &clone.ClusterId); err != nil {
		return err
	}

	// Resolve source database name to ID if needed
	if clone.SourceDatabaseId == "" && clone.SourceDatabaseName != "" {
		log := ctrllog.FromContext(ctx)
		log.Info("Resolving source database name to ID", "sourceDatabaseName", clone.SourceDatabaseName)
		databaseId, err := ndb_api.ResolveSourceDatabaseNameToId(ctx, ndbClient, clone.SourceDatabaseName)
		if err != nil {
			return fmt.Errorf("failed to resolve source database name '%s': %w", clone.SourceDatabaseName, err)
		}
		clone.SourceDatabaseId = databaseId
		log.Info("Resolved source database name to ID", "sourceDatabaseName", clone.SourceDatabaseName, "databaseId", databaseId)
	}

	// Resolve snapshot name to ID if needed (requires source database ID to be resolved first)
	if clone.SnapshotId == "" && clone.SnapshotName != "" {
		if clone.SourceDatabaseId == "" {
			return fmt.Errorf("sourceDatabaseId or sourceDatabaseName must be provided to resolve snapshot name")
		}
		if err := resolveSnapshotNameToId(ctx, ndbClient, clone.SourceDatabaseId, clone.SnapshotName, clone); err != nil {
			return err
		}
	}

	return nil
}

// resolveClusterNameToId resolves a cluster name to its UUID if needed
func resolveClusterNameToId(ctx context.Context, ndbClient *ndb_client.NDBClient, clusterId, clusterName string, clusterIdPtr *string) error {
	if clusterId != "" || clusterName == "" {
		return nil // Already has ID or no name provided
	}

	log := ctrllog.FromContext(ctx)
	log.Info("Resolving cluster name to ID", "clusterName", clusterName)
	resolvedClusterId, err := ndb_api.ResolveClusterNameToId(ctx, ndbClient, clusterName)
	if err != nil {
		return fmt.Errorf("failed to resolve cluster name '%s': %w", clusterName, err)
	}
	*clusterIdPtr = resolvedClusterId
	log.Info("Resolved cluster name to ID", "clusterName", clusterName, "clusterId", resolvedClusterId)
	return nil
}

// resolveSnapshotNameToId resolves a snapshot name to its UUID
// It fetches the source database to get the time machine ID, then resolves the snapshot name
func resolveSnapshotNameToId(ctx context.Context, ndbClient *ndb_client.NDBClient, sourceDatabaseId string, snapshotName string, clone *ndbv1alpha1.Clone) error {
	log := ctrllog.FromContext(ctx)
	log.Info("Fetching source database to get time machine ID", "sourceDatabaseId", sourceDatabaseId)
	sourceDatabase, err := ndb_api.GetDatabaseById(ctx, ndbClient, sourceDatabaseId)
	if err != nil {
		return fmt.Errorf("failed to fetch source database '%s': %w", sourceDatabaseId, err)
	}
	if sourceDatabase.TimeMachineId == "" {
		return fmt.Errorf("source database '%s' does not have a time machine", sourceDatabaseId)
	}

	log.Info("Resolving snapshot name to ID", "snapshotName", snapshotName, "timeMachineId", sourceDatabase.TimeMachineId)
	snapshotId, err := ndb_api.ResolveSnapshotNameToId(ctx, ndbClient, snapshotName, sourceDatabase.TimeMachineId)
	if err != nil {
		return fmt.Errorf("failed to resolve snapshot name '%s': %w", snapshotName, err)
	}
	clone.SnapshotId = snapshotId
	log.Info("Resolved snapshot name to ID", "snapshotName", snapshotName, "snapshotId", snapshotId)
	return nil
}

