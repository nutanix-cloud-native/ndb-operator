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
	"strconv"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// FetchConfigMapDefaults fetches a ConfigMap and returns its data as a map
func FetchConfigMapDefaults(ctx context.Context, k8sClient client.Client, namespace, configMapName string) (map[string]string, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Fetching defaults ConfigMap", "namespace", namespace, "configMapName", configMapName)

	configMap := &corev1.ConfigMap{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      configMapName,
		Namespace: namespace,
	}, configMap)

	if err != nil {
		log.Error(err, "Failed to fetch defaults ConfigMap", "configMapName", configMapName)
		return nil, fmt.Errorf("failed to fetch ConfigMap '%s': %w", configMapName, err)
	}

	if configMap.Data == nil || len(configMap.Data) == 0 {
		log.Info("ConfigMap is empty", "configMapName", configMapName)
		return make(map[string]string), nil
	}

	log.Info("Successfully fetched defaults ConfigMap", "configMapName", configMapName, "keysCount", len(configMap.Data))
	return configMap.Data, nil
}

// GetDefaultValue returns a value from the defaults map, trying type-specific key first, then generic key
// Example: For dbType="postgres" and key="size", tries "postgres.size" first, then "size"
func GetDefaultValue(defaults map[string]string, key string, dbType string) string {
	// Try type-specific key first (e.g., "postgres.clusterName")
	if dbType != "" {
		typeSpecificKey := dbType + "." + key
		if val, exists := defaults[typeSpecificKey]; exists && val != "" {
			return val
		}
	}

	// Fall back to generic key (e.g., "clusterName")
	if val, exists := defaults[key]; exists {
		return val
	}

	return ""
}

// GetDefaultValueInt returns an integer value from defaults, trying type-specific key first
func GetDefaultValueInt(defaults map[string]string, key string, dbType string) (int, bool) {
	strVal := GetDefaultValue(defaults, key, dbType)
	if strVal == "" {
		return 0, false
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, false
	}

	return intVal, true
}

// ApplyDefaultsToDatabase applies defaults from ConfigMap to Database resource
func ApplyDefaultsToDatabase(ctx context.Context, database *ndbv1alpha1.Database, defaults map[string]string) {
	log := ctrllog.FromContext(ctx)
	log.Info("Applying defaults to Database resource", "databaseName", database.Name)

	// Determine database type
	dbType := getDatabaseType(database)

	if database.Spec.IsClone {
		applyCloneDefaults(ctx, database.Spec.Clone, defaults, dbType)
	} else {
		applyInstanceDefaults(ctx, database.Spec.Instance, defaults, dbType)
	}

	log.Info("Defaults applied successfully", "databaseName", database.Name, "dbType", dbType)
}

// getDatabaseType extracts the database type from the Database resource
func getDatabaseType(database *ndbv1alpha1.Database) string {
	if database.Spec.IsClone && database.Spec.Clone != nil {
		return database.Spec.Clone.Type
	}
	if database.Spec.Instance != nil {
		return database.Spec.Instance.Type
	}
	return ""
}

// applyInstanceDefaults applies defaults to a database instance
func applyInstanceDefaults(ctx context.Context, instance *ndbv1alpha1.Instance, defaults map[string]string, dbType string) {
	if instance == nil {
		return
	}

	log := ctrllog.FromContext(ctx)

	// Apply cluster defaults
	if instance.ClusterId == "" && instance.ClusterName == "" {
		if val := GetDefaultValue(defaults, "clusterName", dbType); val != "" {
			instance.ClusterName = val
			log.Info("Applied default clusterName", "value", val)
		}
	}

	// Apply timezone default
	if instance.TimeZone == "" {
		if val := GetDefaultValue(defaults, "timezone", dbType); val != "" {
			instance.TimeZone = val
			log.Info("Applied default timezone", "value", val)
		}
	}

	// Apply size default
	if instance.Size == 0 {
		if val, ok := GetDefaultValueInt(defaults, "size", dbType); ok {
			instance.Size = val
			log.Info("Applied default size", "value", val)
		}
	}

	// Apply profile defaults
	if instance.Profiles == nil {
		instance.Profiles = &ndbv1alpha1.Profiles{}
	}
	applyProfileDefaults(ctx, instance.Profiles, defaults, dbType)

	// Apply Time Machine defaults
	if instance.TMInfo == nil {
		instance.TMInfo = &ndbv1alpha1.DBTimeMachineInfo{}
	}
	applyTimeMachineDefaults(ctx, instance.TMInfo, defaults, dbType)
}

// applyCloneDefaults applies defaults to a clone
func applyCloneDefaults(ctx context.Context, clone *ndbv1alpha1.Clone, defaults map[string]string, dbType string) {
	if clone == nil {
		return
	}

	log := ctrllog.FromContext(ctx)

	// Apply cluster defaults (try clone-specific first)
	if clone.ClusterId == "" && clone.ClusterName == "" {
		// Try clone-specific default first
		if val := GetDefaultValue(defaults, "clone.clusterName", ""); val != "" {
			clone.ClusterName = val
			log.Info("Applied default clone.clusterName", "value", val)
		} else if val := GetDefaultValue(defaults, "clusterName", dbType); val != "" {
			clone.ClusterName = val
			log.Info("Applied default clusterName", "value", val)
		}
	}

	// Apply timezone default
	if clone.TimeZone == "" {
		// Try clone-specific default first
		if val := GetDefaultValue(defaults, "clone.timezone", ""); val != "" {
			clone.TimeZone = val
			log.Info("Applied default clone.timezone", "value", val)
		} else if val := GetDefaultValue(defaults, "timezone", dbType); val != "" {
			clone.TimeZone = val
			log.Info("Applied default timezone", "value", val)
		}
	}

	// Apply profile defaults
	if clone.Profiles == nil {
		clone.Profiles = &ndbv1alpha1.Profiles{}
	}
	// Try clone-specific profile defaults first, then fall back to generic
	applyCloneProfileDefaults(ctx, clone.Profiles, defaults, dbType)
}

// applyProfileDefaults applies profile defaults
func applyProfileDefaults(ctx context.Context, profiles *ndbv1alpha1.Profiles, defaults map[string]string, dbType string) {
	log := ctrllog.FromContext(ctx)

	// Software profile
	if profiles.Software.Id == "" && profiles.Software.Name == "" {
		if val := GetDefaultValue(defaults, "profiles.software.name", dbType); val != "" {
			profiles.Software.Name = val
			log.Info("Applied default profiles.software.name", "value", val)
		}
	}

	// Compute profile
	if profiles.Compute.Id == "" && profiles.Compute.Name == "" {
		if val := GetDefaultValue(defaults, "profiles.compute.name", dbType); val != "" {
			profiles.Compute.Name = val
			log.Info("Applied default profiles.compute.name", "value", val)
		}
	}

	// Network profile
	if profiles.Network.Id == "" && profiles.Network.Name == "" {
		if val := GetDefaultValue(defaults, "profiles.network.name", dbType); val != "" {
			profiles.Network.Name = val
			log.Info("Applied default profiles.network.name", "value", val)
		}
	}

	// DbParam profile
	if profiles.DbParam.Id == "" && profiles.DbParam.Name == "" {
		if val := GetDefaultValue(defaults, "profiles.dbParam.name", dbType); val != "" {
			profiles.DbParam.Name = val
			log.Info("Applied default profiles.dbParam.name", "value", val)
		}
	}

	// DbParamInstance profile
	if profiles.DbParamInstance.Id == "" && profiles.DbParamInstance.Name == "" {
		if val := GetDefaultValue(defaults, "profiles.dbParamInstance.name", dbType); val != "" {
			profiles.DbParamInstance.Name = val
			log.Info("Applied default profiles.dbParamInstance.name", "value", val)
		}
	}
}

// applyCloneProfileDefaults applies clone-specific profile defaults, falling back to generic
func applyCloneProfileDefaults(ctx context.Context, profiles *ndbv1alpha1.Profiles, defaults map[string]string, dbType string) {
	log := ctrllog.FromContext(ctx)

	// Software profile (try clone-specific first)
	if profiles.Software.Id == "" && profiles.Software.Name == "" {
		if val := GetDefaultValue(defaults, "clone.profiles.software.name", ""); val != "" {
			profiles.Software.Name = val
			log.Info("Applied default clone.profiles.software.name", "value", val)
		} else if val := GetDefaultValue(defaults, "profiles.software.name", dbType); val != "" {
			profiles.Software.Name = val
			log.Info("Applied default profiles.software.name", "value", val)
		}
	}

	// Compute profile
	if profiles.Compute.Id == "" && profiles.Compute.Name == "" {
		if val := GetDefaultValue(defaults, "clone.profiles.compute.name", ""); val != "" {
			profiles.Compute.Name = val
			log.Info("Applied default clone.profiles.compute.name", "value", val)
		} else if val := GetDefaultValue(defaults, "profiles.compute.name", dbType); val != "" {
			profiles.Compute.Name = val
			log.Info("Applied default profiles.compute.name", "value", val)
		}
	}

	// Network profile
	if profiles.Network.Id == "" && profiles.Network.Name == "" {
		if val := GetDefaultValue(defaults, "clone.profiles.network.name", ""); val != "" {
			profiles.Network.Name = val
			log.Info("Applied default clone.profiles.network.name", "value", val)
		} else if val := GetDefaultValue(defaults, "profiles.network.name", dbType); val != "" {
			profiles.Network.Name = val
			log.Info("Applied default profiles.network.name", "value", val)
		}
	}

	// DbParam profile
	if profiles.DbParam.Id == "" && profiles.DbParam.Name == "" {
		if val := GetDefaultValue(defaults, "clone.profiles.dbParam.name", ""); val != "" {
			profiles.DbParam.Name = val
			log.Info("Applied default clone.profiles.dbParam.name", "value", val)
		} else if val := GetDefaultValue(defaults, "profiles.dbParam.name", dbType); val != "" {
			profiles.DbParam.Name = val
			log.Info("Applied default profiles.dbParam.name", "value", val)
		}
	}

	// DbParamInstance profile
	if profiles.DbParamInstance.Id == "" && profiles.DbParamInstance.Name == "" {
		if val := GetDefaultValue(defaults, "clone.profiles.dbParamInstance.name", ""); val != "" {
			profiles.DbParamInstance.Name = val
			log.Info("Applied default clone.profiles.dbParamInstance.name", "value", val)
		} else if val := GetDefaultValue(defaults, "profiles.dbParamInstance.name", dbType); val != "" {
			profiles.DbParamInstance.Name = val
			log.Info("Applied default profiles.dbParamInstance.name", "value", val)
		}
	}
}

// applyTimeMachineDefaults applies Time Machine defaults
func applyTimeMachineDefaults(ctx context.Context, tmInfo *ndbv1alpha1.DBTimeMachineInfo, defaults map[string]string, dbType string) {
	log := ctrllog.FromContext(ctx)

	// SLA
	if tmInfo.SLAName == "" {
		if val := GetDefaultValue(defaults, "timeMachine.sla", dbType); val != "" {
			tmInfo.SLAName = val
			log.Info("Applied default timeMachine.sla", "value", val)
		}
	}

	// Daily snapshot time
	if tmInfo.DailySnapshotTime == "" {
		if val := GetDefaultValue(defaults, "timeMachine.dailySnapshotTime", dbType); val != "" {
			tmInfo.DailySnapshotTime = val
			log.Info("Applied default timeMachine.dailySnapshotTime", "value", val)
		}
	}

	// Snapshots per day
	if tmInfo.SnapshotsPerDay == 0 {
		if val, ok := GetDefaultValueInt(defaults, "timeMachine.snapshotsPerDay", dbType); ok {
			tmInfo.SnapshotsPerDay = val
			log.Info("Applied default timeMachine.snapshotsPerDay", "value", val)
		}
	}

	// Log catch up frequency
	if tmInfo.LogCatchUpFrequency == 0 {
		if val, ok := GetDefaultValueInt(defaults, "timeMachine.logCatchUpFrequency", dbType); ok {
			tmInfo.LogCatchUpFrequency = val
			log.Info("Applied default timeMachine.logCatchUpFrequency", "value", val)
		}
	}

	// Weekly snapshot day
	if tmInfo.WeeklySnapshotDay == "" {
		if val := GetDefaultValue(defaults, "timeMachine.weeklySnapshotDay", dbType); val != "" {
			tmInfo.WeeklySnapshotDay = val
			log.Info("Applied default timeMachine.weeklySnapshotDay", "value", val)
		}
	}

	// Monthly snapshot day
	if tmInfo.MonthlySnapshotDay == 0 {
		if val, ok := GetDefaultValueInt(defaults, "timeMachine.monthlySnapshotDay", dbType); ok {
			tmInfo.MonthlySnapshotDay = val
			log.Info("Applied default timeMachine.monthlySnapshotDay", "value", val)
		}
	}

	// Quarterly snapshot month
	if tmInfo.QuarterlySnapshotMonth == "" {
		if val := GetDefaultValue(defaults, "timeMachine.quarterlySnapshotMonth", dbType); val != "" {
			tmInfo.QuarterlySnapshotMonth = val
			log.Info("Applied default timeMachine.quarterlySnapshotMonth", "value", val)
		}
	}
}
