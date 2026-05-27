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

// mysqlRouterNode returns a minimal valid MySQL Router node for use in test fixtures.
func mysqlRouterNode(name string) InstanceHANode {
	return InstanceHANode{VmName: name, NodeType: common.HA_NODE_TYPE_MYSQLROUTER, ClusterName: "cluster-a"}
}

// masterNode returns a minimal valid Master database node for use in test fixtures.
func masterNode(name string) InstanceHANode {
	return InstanceHANode{VmName: name, NodeType: common.HA_NODE_TYPE_DATABASE, Role: common.HA_NODE_ROLE_MASTER, ClusterName: "cluster-a"}
}

// replicaNode returns a minimal valid Replica database node for use in test fixtures.
func replicaNode(name string) InstanceHANode {
	return InstanceHANode{VmName: name, NodeType: common.HA_NODE_TYPE_DATABASE, Role: common.HA_NODE_ROLE_REPLICA, ClusterName: "cluster-b"}
}

// validMySQLConfig returns a minimal valid MySQLHAConfig for use in test fixtures.
func validMySQLConfig() *MySQLHAConfig {
	return &MySQLHAConfig{InnoDBClusterName: "innodb-cluster"}
}

func TestGetHAValidator(t *testing.T) {
	t.Run("returns validator for postgres", func(t *testing.T) {
		v, ok := getHAValidator(common.DATABASE_TYPE_POSTGRES)
		assert.True(t, ok)
		assert.NotNil(t, v)
		assert.IsType(t, &PostgresHAParamsValidator{}, v)
	})

	t.Run("returns validator for mysql", func(t *testing.T) {
		v, ok := getHAValidator(common.DATABASE_TYPE_MYSQL)
		assert.True(t, ok)
		assert.NotNil(t, v)
		assert.IsType(t, &MysqlHAParamsValidator{}, v)
	})

	t.Run("returns false for unsupported engine types", func(t *testing.T) {
		for _, unsupported := range []string{"mongodb", "mssql", "oracle", ""} {
			v, ok := getHAValidator(unsupported)
			assert.False(t, ok, "expected no validator for type %q", unsupported)
			assert.Nil(t, v)
		}
	})
}

func TestMysqlHAParamsValidator_Validate(t *testing.T) {
	validator := &MysqlHAParamsValidator{}
	haPath := field.NewPath("haConfig")

	t.Run("valid config with router enabled produces no errors", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: &MySQLHAConfig{InnoDBClusterName: "innodb-cluster", DeployMySQLRouter: true},
			Nodes: []InstanceHANode{mysqlRouterNode("router1"), mysqlRouterNode("router2"), masterNode("db1"), replicaNode("db2"), replicaNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Empty(t, *errors)
	})

	t.Run("valid config with router disabled produces no errors", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: validMySQLConfig(),
			Nodes: []InstanceHANode{masterNode("db1"), replicaNode("db2"), replicaNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Empty(t, *errors)
	})

	t.Run("deployMySQLRouter true but no router nodes returns invalid error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: &MySQLHAConfig{InnoDBClusterName: "innodb-cluster", DeployMySQLRouter: true},
			Nodes: []InstanceHANode{masterNode("db1"), replicaNode("db2"), replicaNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.mysql.deployMySQLRouter", (*errors)[0].Field)
		assert.Contains(t, (*errors)[0].Detail, "no mysqlrouter nodes")
	})

	t.Run("router nodes present but deployMySQLRouter false returns invalid error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: validMySQLConfig(), // DeployMySQLRouter defaults to false
			Nodes: []InstanceHANode{mysqlRouterNode("router1"), masterNode("db1"), replicaNode("db2")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes", (*errors)[0].Field)
		assert.Contains(t, (*errors)[0].Detail, "deployMySQLRouter is false")
	})

	t.Run("nil mysql returns required error and stops further node validation", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: nil,
			Nodes: []InstanceHANode{masterNode("db1"), replicaNode("db2")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeRequired, (*errors)[0].Type)
		assert.Equal(t, "haConfig.mysql", (*errors)[0].Field)
	})

	t.Run("empty innoDBClusterName returns invalid error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: &MySQLHAConfig{InnoDBClusterName: ""},
			Nodes: []InstanceHANode{masterNode("db1"), replicaNode("db2")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.mysql.innoDBClusterName", (*errors)[0].Field)
	})

	t.Run("invalid nodeType returns error for that node index", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: validMySQLConfig(),
			Nodes: []InstanceHANode{
				{VmName: "bad", NodeType: "haproxy", ClusterName: "cluster-a"},
				masterNode("db1"),
			},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes[0].nodeType", (*errors)[0].Field)
	})

	t.Run("mysqlrouter node with role set returns error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: &MySQLHAConfig{InnoDBClusterName: "innodb-cluster", DeployMySQLRouter: true},
			Nodes: []InstanceHANode{
				{VmName: "router1", NodeType: common.HA_NODE_TYPE_MYSQLROUTER, Role: "Master", ClusterName: "cluster-a"},
				masterNode("db1"),
			},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes[0].role", (*errors)[0].Field)
	})

	t.Run("zero master database nodes returns invalid nodes error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: validMySQLConfig(),
			Nodes: []InstanceHANode{replicaNode("db1"), replicaNode("db2"), replicaNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes", (*errors)[0].Field)
		assert.Contains(t, (*errors)[0].Detail, "exactly one")
	})

	t.Run("multiple master database nodes returns invalid nodes error", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: validMySQLConfig(),
			Nodes: []InstanceHANode{masterNode("db1"), masterNode("db2"), replicaNode("db3")},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Len(t, *errors, 1)
		assert.Equal(t, field.ErrorTypeInvalid, (*errors)[0].Type)
		assert.Equal(t, "haConfig.nodes", (*errors)[0].Field)
		assert.Contains(t, (*errors)[0].Detail, "exactly one")
	})

	t.Run("empty nodes list skips master count check", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: validMySQLConfig(),
			Nodes: []InstanceHANode{},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		assert.Empty(t, *errors)
	})

	t.Run("multiple violations accumulate all errors", func(t *testing.T) {
		haConfig := &InstanceHAConfig{
			MySQL: &MySQLHAConfig{InnoDBClusterName: ""},
			Nodes: []InstanceHANode{
				{VmName: "bad", NodeType: "haproxy", ClusterName: "cluster-a"},
				replicaNode("db1"),
				// no master node
			},
		}
		errors := &field.ErrorList{}
		validator.Validate(haConfig, haPath, errors)
		// 3 errors: empty innoDBClusterName + invalid nodeType for node[0] + no master
		assert.Len(t, *errors, 3)
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
