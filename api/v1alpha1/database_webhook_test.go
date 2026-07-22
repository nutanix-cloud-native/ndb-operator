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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDatabaseCustomDefaulter_Default(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ns := "default"

	t.Run("provision applies ConfigMap defaults then standard defaulter", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "ndb-def"},
			Data: map[string]string{
				"clusterName": "cm-cluster",
				"size":        "25",
				"timezone":    "America/Chicago",
			},
		}
		c := fake.NewClientBuilder().WithObjects(cm).Build()
		d := &DatabaseCustomDefaulter{Client: c}

		db := provisionDB(func(i *Instance) {
			i.ClusterId, i.ClusterName = "", ""
			i.Size = 0
			i.TimeZone = ""
		})
		db.Namespace = ns
		db.Spec.DefaultsConfigMapRef = "ndb-def"

		require.NoError(t, d.Default(ctx, db))
		assert.Equal(t, "cm-cluster", db.Spec.Instance.ClusterName)
		assert.Equal(t, 25, db.Spec.Instance.Size)
		assert.Equal(t, "America/Chicago", db.Spec.Instance.TimeZone)
		assert.Contains(t, db.Spec.Instance.Description, "Database provisioned by ndb-operator")
	})

	t.Run("clone applies ConfigMap defaults for cluster and timezone", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "ndb-clone-def"},
			Data: map[string]string{
				"clone.clusterName": "clone-fleet",
				"clone.timezone":    "Europe/London",
			},
		}
		c := fake.NewClientBuilder().WithObjects(cm).Build()
		d := &DatabaseCustomDefaulter{Client: c}

		db := cloneDB("clone-webhook-ut", func(cl *Clone) {
			cl.ClusterId, cl.ClusterName = "", ""
			cl.TimeZone = ""
		})
		db.Namespace = ns
		db.Spec.DefaultsConfigMapRef = "ndb-clone-def"

		require.NoError(t, d.Default(ctx, db))
		assert.Equal(t, "clone-fleet", db.Spec.Clone.ClusterName)
		assert.Equal(t, "Europe/London", db.Spec.Clone.TimeZone)
		assert.Contains(t, db.Spec.Clone.Description, "Clone created by ndb-operator")
	})

	t.Run("missing ConfigMap does not fail Default; standard defaulter still runs", func(t *testing.T) {
		c := fake.NewClientBuilder().Build()
		d := &DatabaseCustomDefaulter{Client: c}

		db := provisionDB()
		db.Namespace = ns
		db.Spec.DefaultsConfigMapRef = "does-not-exist"

		require.NoError(t, d.Default(ctx, db))
		assert.Contains(t, db.Spec.Instance.Description, "Database provisioned by ndb-operator")
	})

	t.Run("empty ConfigMap Data skips ApplyDefaultsFromConfigMap", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "empty-cm"},
			Data:       map[string]string{},
		}
		c := fake.NewClientBuilder().WithObjects(cm).Build()
		d := &DatabaseCustomDefaulter{Client: c}

		db := provisionDB(func(i *Instance) {
			i.ClusterId, i.ClusterName = "", ""
		})
		db.Namespace = ns
		db.Spec.DefaultsConfigMapRef = "empty-cm"

		require.NoError(t, d.Default(ctx, db))
		assert.Empty(t, db.Spec.Instance.ClusterName)
		assert.Empty(t, db.Spec.Instance.ClusterId)
	})
}
