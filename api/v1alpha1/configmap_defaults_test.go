/*
Copyright 2022-2026 Nutanix, Inc.

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
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testSourceDBUUID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

// provisionDB builds a Database with a non-nil Instance; optional fns tweak the Instance in order.
func provisionDB(modify ...func(*Instance)) *Database {
	inst := &Instance{
		Name:             "test-db",
		ClusterName:      "cluster",
		TimeZone:         "UTC",
		Size:             10,
		Type:             common.DATABASE_TYPE_POSTGRES,
		CredentialSecret: "secret",
		Profiles:         &Profiles{},
		TMInfo:           &DBTimeMachineInfo{},
	}
	for _, f := range modify {
		f(inst)
	}
	return &Database{
		ObjectMeta: metav1.ObjectMeta{Name: inst.Name, Namespace: "default"},
		Spec:       DatabaseSpec{IsClone: false, Instance: inst},
	}
}

func cloneDB(name string, modify ...func(*Clone)) *Database {
	c := &Clone{
		Name:             name,
		Type:             common.DATABASE_TYPE_POSTGRES,
		ClusterId:        "cid",
		ClusterName:      "c",
		TimeZone:         "UTC",
		CredentialSecret: "secret",
		SourceDatabaseId: testSourceDBUUID,
		SnapshotId:       "snap",
		Profiles:         &Profiles{},
	}
	for _, f := range modify {
		f(c)
	}
	return &Database{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec:       DatabaseSpec{IsClone: true, Clone: c},
	}
}

func TestFetchConfigMapDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ns, name := "test-ns", "ndb-db-defaults"

	t.Run("returns data when ConfigMap exists", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Data:       map[string]string{"clusterName": "fleet-cluster", "timezone": "UTC"},
		}
		c := fake.NewClientBuilder().WithObjects(cm).Build()
		got, err := FetchConfigMapDefaults(ctx, c, ns, name)
		require.NoError(t, err)
		assert.Equal(t, "fleet-cluster", got["clusterName"])
		assert.Equal(t, "UTC", got["timezone"])
	})

	t.Run("returns empty map when Data is nil or empty", func(t *testing.T) {
		for _, label := range []struct {
			name string
			data map[string]string
		}{
			{"nil", nil},
			{"empty", map[string]string{}},
		} {
			t.Run(label.name, func(t *testing.T) {
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
					Data:       label.data,
				}
				c := fake.NewClientBuilder().WithObjects(cm).Build()
				got, err := FetchConfigMapDefaults(ctx, c, ns, name)
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Empty(t, got)
			})
		}
	})

	t.Run("error when ConfigMap missing", func(t *testing.T) {
		c := fake.NewClientBuilder().Build()
		_, err := FetchConfigMapDefaults(ctx, c, ns, "does-not-exist")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch ConfigMap")
	})
}

func TestApplyDefaultsFromConfigMap_Provision(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("cluster timezone size from defaults", func(t *testing.T) {
		db := provisionDB(func(i *Instance) {
			i.ClusterId, i.ClusterName = "", ""
			i.TimeZone, i.Size = "", 0
		})
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"clusterName": "my-cluster", "timezone": "America/New_York", "size": "20",
		})
		assert.Equal(t, "my-cluster", db.Spec.Instance.ClusterName)
		assert.Equal(t, "America/New_York", db.Spec.Instance.TimeZone)
		assert.Equal(t, 20, db.Spec.Instance.Size)
	})

	t.Run("type-specific size overrides generic", func(t *testing.T) {
		db := provisionDB(func(i *Instance) { i.Size = 0 })
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"size": "10", "postgres.size": "25", "mysql.size": "15",
		})
		assert.Equal(t, 25, db.Spec.Instance.Size)
	})

	t.Run("CR fields already set are not overwritten", func(t *testing.T) {
		db := provisionDB(func(i *Instance) {
			i.ClusterName = "cr-cluster"
			i.TimeZone = "Europe/London"
			i.Size = 50
		})
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"clusterName": "configmap-cluster", "timezone": "UTC", "size": "20",
		})
		assert.Equal(t, "cr-cluster", db.Spec.Instance.ClusterName)
		assert.Equal(t, "Europe/London", db.Spec.Instance.TimeZone)
		assert.Equal(t, 50, db.Spec.Instance.Size)
	})

	t.Run("type-prefixed timezone", func(t *testing.T) {
		db := provisionDB(func(i *Instance) {
			i.ClusterName, i.ClusterId = "", ""
			i.TimeZone = ""
			i.Size = 0
		})
		def := map[string]string{"timezone": "UTC", "postgres.timezone": "Asia/Tokyo", "mysql.timezone": "America/Denver"}
		ApplyDefaultsFromConfigMap(ctx, db, def)
		assert.Equal(t, "Asia/Tokyo", db.Spec.Instance.TimeZone)

		mysqlInst := provisionDB(func(i *Instance) {
			i.Name = "mysql-db"
			i.Type = common.DATABASE_TYPE_MYSQL
			i.TimeZone = ""
		})
		ApplyDefaultsFromConfigMap(ctx, mysqlInst, def)
		assert.Equal(t, "America/Denver", mysqlInst.Spec.Instance.TimeZone)
	})

	// Single case covering applyProfileDefaultsFromConfigMap + applyTimeMachineDefaultsFromConfigMap
	t.Run("profiles timeMachine and TM int fields", func(t *testing.T) {
		db := provisionDB()
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"profiles.software.name":             "POSTGRES_15_OOB",
			"profiles.compute.name":              "DEFAULT_COMPUTE",
			"profiles.network.name":              "DEFAULT_NETWORK",
			"profiles.dbParam.name":              "DEFAULT_PARAMS",
			"profiles.dbParamInstance.name":      "PARAM_INST_OOB",
			"timeMachine.sla":                    "BRASS_SLA",
			"timeMachine.dailySnapshotTime":      "12:00:00",
			"timeMachine.snapshotsPerDay":        "4",
			"timeMachine.logCatchUpFrequency":    "60",
			"timeMachine.weeklySnapshotDay":      "MONDAY",
			"timeMachine.monthlySnapshotDay":     "7",
			"timeMachine.quarterlySnapshotMonth": "Mar",
		})
		p, tm := db.Spec.Instance.Profiles, db.Spec.Instance.TMInfo
		assert.Equal(t, "POSTGRES_15_OOB", p.Software.Name)
		assert.Equal(t, "DEFAULT_COMPUTE", p.Compute.Name)
		assert.Equal(t, "DEFAULT_NETWORK", p.Network.Name)
		assert.Equal(t, "DEFAULT_PARAMS", p.DbParam.Name)
		assert.Equal(t, "PARAM_INST_OOB", p.DbParamInstance.Name)
		assert.Equal(t, "BRASS_SLA", tm.SLAName)
		assert.Equal(t, "12:00:00", tm.DailySnapshotTime)
		assert.Equal(t, 4, tm.SnapshotsPerDay)
		assert.Equal(t, 60, tm.LogCatchUpFrequency)
		assert.Equal(t, "MONDAY", tm.WeeklySnapshotDay)
		assert.Equal(t, 7, tm.MonthlySnapshotDay)
		assert.Equal(t, "Mar", tm.QuarterlySnapshotMonth)
	})
}

func TestApplyDefaultsFromConfigMap_Clone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("clone-prefixed cluster and timezone beat generic", func(t *testing.T) {
		db := cloneDB("c1", func(c *Clone) {
			c.ClusterId, c.ClusterName = "", ""
			c.TimeZone = ""
		})
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"clusterName": "generic-cluster", "clone.clusterName": "clone-specific-cluster",
			"timezone": "UTC", "clone.timezone": "America/Los_Angeles",
		})
		assert.Equal(t, "clone-specific-cluster", db.Spec.Clone.ClusterName)
		assert.Equal(t, "America/Los_Angeles", db.Spec.Clone.TimeZone)
	})

	t.Run("falls back to generic clone keys", func(t *testing.T) {
		db := cloneDB("c2", func(c *Clone) {
			c.ClusterId, c.ClusterName = "", ""
			c.TimeZone = ""
		})
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"clusterName": "generic-cluster", "timezone": "Europe/Paris",
		})
		assert.Equal(t, "generic-cluster", db.Spec.Clone.ClusterName)
		assert.Equal(t, "Europe/Paris", db.Spec.Clone.TimeZone)
	})

	t.Run("clone.profiles.* and typed fallback for software", func(t *testing.T) {
		db := cloneDB("c3")
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"profiles.software.name": "generic-software", "clone.profiles.software.name": "clone-software",
		})
		assert.Equal(t, "clone-software", db.Spec.Clone.Profiles.Software.Name)
	})

	t.Run("all clone.profiles.* keys", func(t *testing.T) {
		db := cloneDB("c4")
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"clone.profiles.software.name": "C_SW", "clone.profiles.compute.name": "C_COMP",
			"clone.profiles.network.name": "C_NET", "clone.profiles.dbParam.name": "C_DBP",
			"clone.profiles.dbParamInstance.name": "C_DBPI",
		})
		p := db.Spec.Clone.Profiles
		assert.Equal(t, "C_SW", p.Software.Name)
		assert.Equal(t, "C_COMP", p.Compute.Name)
		assert.Equal(t, "C_NET", p.Network.Name)
		assert.Equal(t, "C_DBP", p.DbParam.Name)
		assert.Equal(t, "C_DBPI", p.DbParamInstance.Name)
	})

	t.Run("clone software falls back to postgres.profiles.software.name", func(t *testing.T) {
		db := cloneDB("c5")
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{
			"profiles.software.name": "GEN_SW", "postgres.profiles.software.name": "PG_SW",
		})
		assert.Equal(t, "PG_SW", db.Spec.Clone.Profiles.Software.Name)
	})
}

func TestApplyDefaultsFromConfigMap_EdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("nil instance or clone does not panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			ApplyDefaultsFromConfigMap(ctx, &Database{ObjectMeta: metav1.ObjectMeta{Name: "x"}, Spec: DatabaseSpec{Instance: nil}}, map[string]string{"clusterName": "c"})
		})
		require.NotPanics(t, func() {
			ApplyDefaultsFromConfigMap(ctx, &Database{ObjectMeta: metav1.ObjectMeta{Name: "x"}, Spec: DatabaseSpec{IsClone: true, Clone: nil}}, map[string]string{"clusterName": "c"})
		})
	})

	t.Run("empty defaults", func(t *testing.T) {
		db := provisionDB(func(i *Instance) {
			i.ClusterName, i.TimeZone = "", ""
			i.Size = 0
		})
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{})
		assert.Empty(t, db.Spec.Instance.ClusterName)
		assert.Empty(t, db.Spec.Instance.TimeZone)
		assert.Zero(t, db.Spec.Instance.Size)
	})

	t.Run("profile Id set blocks name default", func(t *testing.T) {
		db := provisionDB(func(i *Instance) {
			i.Profiles = &Profiles{Software: Profile{Id: "existing-id", Name: ""}}
		})
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{"profiles.software.name": "should-not-apply"})
		assert.Empty(t, db.Spec.Instance.Profiles.Software.Name)
		assert.Equal(t, "existing-id", db.Spec.Instance.Profiles.Software.Id)
	})

	t.Run("invalid numeric strings ignored for size and TM ints", func(t *testing.T) {
		db := provisionDB(func(i *Instance) { i.Size = 0 })
		ApplyDefaultsFromConfigMap(ctx, db, map[string]string{"size": "not-a-number"})
		assert.Zero(t, db.Spec.Instance.Size)

		db2 := provisionDB()
		ApplyDefaultsFromConfigMap(ctx, db2, map[string]string{
			"timeMachine.snapshotsPerDay": "bogus", "timeMachine.logCatchUpFrequency": "x", "timeMachine.monthlySnapshotDay": "xx",
		})
		tm := db2.Spec.Instance.TMInfo
		assert.Zero(t, tm.SnapshotsPerDay)
		assert.Zero(t, tm.LogCatchUpFrequency)
		assert.Zero(t, tm.MonthlySnapshotDay)
	})
}
