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
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// haproxyNode returns a minimal valid HAProxy node for use in test fixtures.
func haproxyNode(name string) InstanceHANode {
	return InstanceHANode{VmName: name, NodeType: common.HA_NODE_TYPE_HAPROXY, ClusterName: "cluster-a"}
}

// primaryNode returns a minimal valid Primary database node for use in test fixtures.
func primaryNode(name string) InstanceHANode {
	return InstanceHANode{VmName: name, NodeType: common.HA_NODE_TYPE_DATABASE, Role: common.HA_NODE_ROLE_PRIMARY, ClusterName: "cluster-a"}
}

// secondaryNode returns a minimal valid Secondary database node for use in test fixtures.
func secondaryNode(name string) InstanceHANode {
	return InstanceHANode{VmName: name, NodeType: common.HA_NODE_TYPE_DATABASE, Role: common.HA_NODE_ROLE_SECONDARY, ClusterName: "cluster-b"}
}

// validPGConfig returns a minimal valid PostgresHAConfig for use in test fixtures.
func validPGConfig() *PostgresHAConfig {
	return &PostgresHAConfig{PatroniClusterName: "patroni-cluster"}
}

func TestGetHAValidator(t *testing.T) {
	t.Run("returns validator for postgres", func(t *testing.T) {
		v, ok := getHAValidator(common.DATABASE_TYPE_POSTGRES)
		assert.True(t, ok)
		assert.NotNil(t, v)
		assert.IsType(t, &PostgresHAParamsValidator{}, v)
	})

	t.Run("returns false for unsupported engine types", func(t *testing.T) {
		for _, unsupported := range []string{"mysql", "mongodb", "mssql", "oracle", ""} {
			v, ok := getHAValidator(unsupported)
			assert.False(t, ok, "expected no validator for type %q", unsupported)
			assert.Nil(t, v)
		}
	})
}

func TestPostgresHAParamsValidator_Validate(t *testing.T) {
	validator := &PostgresHAParamsValidator{}
	haPath := field.NewPath("haConfig")

	t.Run("valid config produces no errors", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: validPGConfig(),
			Nodes:    []InstanceHANode{haproxyNode("proxy1"), haproxyNode("proxy2"), primaryNode("db1"), secondaryNode("db2"), secondaryNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Empty(t, *errors)
	})

	t.Run("nil postgres returns required error and stops further node validation", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: nil,
			Nodes:    []InstanceHANode{haproxyNode("proxy1"), primaryNode("db1")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeRequired, (*errors)[0].Type)
		assert.Equal(t, "haConfig.postgres", (*errors)[0].Field)
	})

	t.Run("empty patroniClusterName returns invalid error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: &PostgresHAConfig{PatroniClusterName: ""},
			Nodes:    []InstanceHANode{haproxyNode("proxy1"), primaryNode("db1")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.postgres.patroniClusterName", (*errors)[0].Field)
	})

	t.Run("invalid nodeType returns error for that node index", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: validPGConfig(),
			Nodes: []InstanceHANode{
				{VmName: "bad", NodeType: "unknown-type", ClusterName: "cluster-a"},
				primaryNode("db1"),
			},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes[0].nodeType", (*errors)[0].Field)
	})

	t.Run("zero primary database nodes returns invalid nodes error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: validPGConfig(),
			Nodes:    []InstanceHANode{haproxyNode("proxy1"), secondaryNode("db1"), secondaryNode("db2")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes", (*errors)[0].Field)
		assert.Contains(t, (*errors)[0].Detail, "exactly one")
	})

	t.Run("multiple primary database nodes returns invalid nodes error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: validPGConfig(),
			Nodes:    []InstanceHANode{haproxyNode("proxy1"), primaryNode("db1"), primaryNode("db2"), secondaryNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes", (*errors)[0].Field)
		assert.Contains(t, (*errors)[0].Detail, "exactly one")
	})

	t.Run("empty nodes list skips primary count check", func(t *testing.T) {
		// Generic node-emptiness check is enforced by webhook_helpers; this validator
		// skips the primary-count check when Nodes is empty to avoid a double error.
		haConfig := &InstanceHAConfig{
			Postgres: validPGConfig(),
			Nodes:    []InstanceHANode{},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Empty(t, *errors)
	})

	t.Run("multiple violations accumulate all errors", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			Postgres: &PostgresHAConfig{PatroniClusterName: ""},
			Nodes: []InstanceHANode{
				{VmName: "bad", NodeType: "unknown-type", ClusterName: "cluster-a"},
				secondaryNode("db1"),
				// no primary node
			},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		// 3 errors: empty patroniClusterName + invalid nodeType for node[0] + no primary
		assert.Len(t, *errors, 3)
	})

	t.Run("haproxy node with database role is treated as haproxy not as primary", func(t *testing.T) {
		// Ensures the primary count check only looks at database-typed nodes.
		haConfig := &InstanceHAConfig{
			Postgres: validPGConfig(),
			Nodes: []InstanceHANode{
				{VmName: "proxy1", NodeType: common.HA_NODE_TYPE_HAPROXY, Role: common.HA_NODE_ROLE_PRIMARY, ClusterName: "cluster-a"},
				secondaryNode("db1"),
				// no actual database primary node
			},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		// The haproxy node with role=Primary should not count toward primaryCount.
		assert.Len(t, *errors, 1)
		assert.Equal(t, "haConfig.nodes", (*errors)[0].Field)
	})
}
