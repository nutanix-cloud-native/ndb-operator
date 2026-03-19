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

package v1alpha1

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// FetchConfigMapDefaults fetches a ConfigMap and returns its data as a map.
// Used by the defaulter webhook to inject values before validation.
func FetchConfigMapDefaults(ctx context.Context, k8sClient client.Client, namespace, configMapName string) (map[string]string, error) {
	log := logf.FromContext(ctx)
	log.Info("Fetching defaults ConfigMap", "namespace", namespace, "configMapName", configMapName)

	configMap := &corev1.ConfigMap{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      configMapName,
		Namespace: namespace,
	}, configMap)

	if err != nil {
		log.Info("ConfigMap not found or error fetching", "configMapName", configMapName, "error", err)
		return nil, fmt.Errorf("failed to fetch ConfigMap '%s': %w", configMapName, err)
	}

	if configMap.Data == nil || len(configMap.Data) == 0 {
		log.Info("ConfigMap is empty", "configMapName", configMapName)
		return make(map[string]string), nil
	}

	log.Info("Successfully fetched defaults ConfigMap", "configMapName", configMapName, "keysCount", len(configMap.Data))
	return configMap.Data, nil
}

// ApplyDefaultsFromConfigMap applies defaults from ConfigMap to Database resource.
// Called by the defaulter webhook before validation so the CR is fully populated.
func ApplyDefaultsFromConfigMap(ctx context.Context, database *Database, defaults map[string]string) {
	log := logf.FromContext(ctx)
	log.Info("Applying defaults from ConfigMap", "databaseName", database.Name)

	dbType := getDatabaseTypeForDefaults(database)

	if database.Spec.IsClone {
		applyCloneDefaultsFromConfigMap(ctx, database.Spec.Clone, defaults, dbType)
	} else {
		applyInstanceDefaultsFromConfigMap(ctx, database.Spec.Instance, defaults, dbType)
	}

	log.Info("Defaults applied successfully", "databaseName", database.Name, "dbType", dbType)
}

func getDatabaseTypeForDefaults(database *Database) string {
	if database.Spec.IsClone && database.Spec.Clone != nil {
		return database.Spec.Clone.Type
	}
	if database.Spec.Instance != nil {
		return database.Spec.Instance.Type
	}
	return ""
}

func getDefaultValue(defaults map[string]string, key string, dbType string) string {
	if dbType != "" {
		typeSpecificKey := dbType + "." + key
		if val, exists := defaults[typeSpecificKey]; exists && val != "" {
			return val
		}
	}
	if val, exists := defaults[key]; exists {
		return val
	}
	return ""
}

func getDefaultValueInt(defaults map[string]string, key string, dbType string) (int, bool) {
	strVal := getDefaultValue(defaults, key, dbType)
	if strVal == "" {
		return 0, false
	}
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, false
	}
	return intVal, true
}

func applyInstanceDefaultsFromConfigMap(ctx context.Context, instance *Instance, defaults map[string]string, dbType string) {
	if instance == nil {
		return
	}
	log := logf.FromContext(ctx)

	if instance.ClusterId == "" && instance.ClusterName == "" {
		if val := getDefaultValue(defaults, "clusterName", dbType); val != "" {
			instance.ClusterName = val
			log.Info("Applied default clusterName", "value", val)
		}
	}
	if instance.TimeZone == "" {
		if val := getDefaultValue(defaults, "timezone", dbType); val != "" {
			instance.TimeZone = val
			log.Info("Applied default timezone", "value", val)
		}
	}
	if instance.Size == 0 {
		if val, ok := getDefaultValueInt(defaults, "size", dbType); ok {
			instance.Size = val
			log.Info("Applied default size", "value", val)
		}
	}
	if instance.Profiles == nil {
		instance.Profiles = &Profiles{}
	}
	applyProfileDefaultsFromConfigMap(ctx, instance.Profiles, defaults, dbType)
	if instance.TMInfo == nil {
		instance.TMInfo = &DBTimeMachineInfo{}
	}
	applyTimeMachineDefaultsFromConfigMap(ctx, instance.TMInfo, defaults, dbType)
}

func applyCloneDefaultsFromConfigMap(ctx context.Context, clone *Clone, defaults map[string]string, dbType string) {
	if clone == nil {
		return
	}
	log := logf.FromContext(ctx)

	if clone.ClusterId == "" && clone.ClusterName == "" {
		if val := getDefaultValue(defaults, "clone.clusterName", ""); val != "" {
			clone.ClusterName = val
			log.Info("Applied default clone.clusterName", "value", val)
		} else if val := getDefaultValue(defaults, "clusterName", dbType); val != "" {
			clone.ClusterName = val
			log.Info("Applied default clusterName", "value", val)
		}
	}
	if clone.TimeZone == "" {
		if val := getDefaultValue(defaults, "clone.timezone", ""); val != "" {
			clone.TimeZone = val
			log.Info("Applied default clone.timezone", "value", val)
		} else if val := getDefaultValue(defaults, "timezone", dbType); val != "" {
			clone.TimeZone = val
			log.Info("Applied default timezone", "value", val)
		}
	}
	if clone.Profiles == nil {
		clone.Profiles = &Profiles{}
	}
	applyCloneProfileDefaultsFromConfigMap(ctx, clone.Profiles, defaults, dbType)
}

func applyProfileDefaultsFromConfigMap(ctx context.Context, profiles *Profiles, defaults map[string]string, dbType string) {
	log := logf.FromContext(ctx)
	if profiles.Software.Id == "" && profiles.Software.Name == "" {
		if val := getDefaultValue(defaults, "profiles.software.name", dbType); val != "" {
			profiles.Software.Name = val
			log.Info("Applied default profiles.software.name", "value", val)
		}
	}
	if profiles.Compute.Id == "" && profiles.Compute.Name == "" {
		if val := getDefaultValue(defaults, "profiles.compute.name", dbType); val != "" {
			profiles.Compute.Name = val
			log.Info("Applied default profiles.compute.name", "value", val)
		}
	}
	if profiles.Network.Id == "" && profiles.Network.Name == "" {
		if val := getDefaultValue(defaults, "profiles.network.name", dbType); val != "" {
			profiles.Network.Name = val
			log.Info("Applied default profiles.network.name", "value", val)
		}
	}
	if profiles.DbParam.Id == "" && profiles.DbParam.Name == "" {
		if val := getDefaultValue(defaults, "profiles.dbParam.name", dbType); val != "" {
			profiles.DbParam.Name = val
			log.Info("Applied default profiles.dbParam.name", "value", val)
		}
	}
	if profiles.DbParamInstance.Id == "" && profiles.DbParamInstance.Name == "" {
		if val := getDefaultValue(defaults, "profiles.dbParamInstance.name", dbType); val != "" {
			profiles.DbParamInstance.Name = val
			log.Info("Applied default profiles.dbParamInstance.name", "value", val)
		}
	}
}

func applyCloneProfileDefaultsFromConfigMap(ctx context.Context, profiles *Profiles, defaults map[string]string, dbType string) {
	log := logf.FromContext(ctx)
	if profiles.Software.Id == "" && profiles.Software.Name == "" {
		if val := getDefaultValue(defaults, "clone.profiles.software.name", ""); val != "" {
			profiles.Software.Name = val
			log.Info("Applied default clone.profiles.software.name", "value", val)
		} else if val := getDefaultValue(defaults, "profiles.software.name", dbType); val != "" {
			profiles.Software.Name = val
			log.Info("Applied default profiles.software.name", "value", val)
		}
	}
	if profiles.Compute.Id == "" && profiles.Compute.Name == "" {
		if val := getDefaultValue(defaults, "clone.profiles.compute.name", ""); val != "" {
			profiles.Compute.Name = val
			log.Info("Applied default clone.profiles.compute.name", "value", val)
		} else if val := getDefaultValue(defaults, "profiles.compute.name", dbType); val != "" {
			profiles.Compute.Name = val
			log.Info("Applied default profiles.compute.name", "value", val)
		}
	}
	if profiles.Network.Id == "" && profiles.Network.Name == "" {
		if val := getDefaultValue(defaults, "clone.profiles.network.name", ""); val != "" {
			profiles.Network.Name = val
			log.Info("Applied default clone.profiles.network.name", "value", val)
		} else if val := getDefaultValue(defaults, "profiles.network.name", dbType); val != "" {
			profiles.Network.Name = val
			log.Info("Applied default profiles.network.name", "value", val)
		}
	}
	if profiles.DbParam.Id == "" && profiles.DbParam.Name == "" {
		if val := getDefaultValue(defaults, "clone.profiles.dbParam.name", ""); val != "" {
			profiles.DbParam.Name = val
			log.Info("Applied default clone.profiles.dbParam.name", "value", val)
		} else if val := getDefaultValue(defaults, "profiles.dbParam.name", dbType); val != "" {
			profiles.DbParam.Name = val
			log.Info("Applied default profiles.dbParam.name", "value", val)
		}
	}
	if profiles.DbParamInstance.Id == "" && profiles.DbParamInstance.Name == "" {
		if val := getDefaultValue(defaults, "clone.profiles.dbParamInstance.name", ""); val != "" {
			profiles.DbParamInstance.Name = val
			log.Info("Applied default clone.profiles.dbParamInstance.name", "value", val)
		} else if val := getDefaultValue(defaults, "profiles.dbParamInstance.name", dbType); val != "" {
			profiles.DbParamInstance.Name = val
			log.Info("Applied default profiles.dbParamInstance.name", "value", val)
		}
	}
}

func applyTimeMachineDefaultsFromConfigMap(ctx context.Context, tmInfo *DBTimeMachineInfo, defaults map[string]string, dbType string) {
	log := logf.FromContext(ctx)
	if tmInfo.SLAName == "" {
		if val := getDefaultValue(defaults, "timeMachine.sla", dbType); val != "" {
			tmInfo.SLAName = val
			log.Info("Applied default timeMachine.sla", "value", val)
		}
	}
	if tmInfo.DailySnapshotTime == "" {
		if val := getDefaultValue(defaults, "timeMachine.dailySnapshotTime", dbType); val != "" {
			tmInfo.DailySnapshotTime = val
			log.Info("Applied default timeMachine.dailySnapshotTime", "value", val)
		}
	}
	if tmInfo.SnapshotsPerDay == 0 {
		if val, ok := getDefaultValueInt(defaults, "timeMachine.snapshotsPerDay", dbType); ok {
			tmInfo.SnapshotsPerDay = val
			log.Info("Applied default timeMachine.snapshotsPerDay", "value", val)
		}
	}
	if tmInfo.LogCatchUpFrequency == 0 {
		if val, ok := getDefaultValueInt(defaults, "timeMachine.logCatchUpFrequency", dbType); ok {
			tmInfo.LogCatchUpFrequency = val
			log.Info("Applied default timeMachine.logCatchUpFrequency", "value", val)
		}
	}
	if tmInfo.WeeklySnapshotDay == "" {
		if val := getDefaultValue(defaults, "timeMachine.weeklySnapshotDay", dbType); val != "" {
			tmInfo.WeeklySnapshotDay = val
			log.Info("Applied default timeMachine.weeklySnapshotDay", "value", val)
		}
	}
	if tmInfo.MonthlySnapshotDay == 0 {
		if val, ok := getDefaultValueInt(defaults, "timeMachine.monthlySnapshotDay", dbType); ok {
			tmInfo.MonthlySnapshotDay = val
			log.Info("Applied default timeMachine.monthlySnapshotDay", "value", val)
		}
	}
	if tmInfo.QuarterlySnapshotMonth == "" {
		if val := getDefaultValue(defaults, "timeMachine.quarterlySnapshotMonth", dbType); val != "" {
			tmInfo.QuarterlySnapshotMonth = val
			log.Info("Applied default timeMachine.quarterlySnapshotMonth", "value", val)
		}
	}
}
