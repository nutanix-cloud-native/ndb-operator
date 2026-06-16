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
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// HAParamsValidator validates engine-specific constraints within an InstanceHAConfig.
// Each database engine that supports HA provides its own implementation.
// +kubebuilder:object:generate=false
type HAParamsValidator interface {
	// Validate checks engine-specific fields (e.g. patroniClusterName for Postgres)
	// and engine-specific node constraints (e.g. valid nodeType values, Primary count).
	Validate(haConfig *InstanceHAConfig, haPath *field.Path, errors *field.ErrorList)
}

// haValidators maps a database engine type to its HAParamsValidator implementation.
// To add HA support for a new engine, register its validator here.
var haValidators = map[string]HAParamsValidator{
	common.DATABASE_TYPE_POSTGRES: &PostgresHAParamsValidator{},
	common.DATABASE_TYPE_MYSQL:    &MysqlHAParamsValidator{},
	common.DATABASE_TYPE_MONGODB:  &MongoHAParamsValidator{},
}

// getHAValidator returns the registered HAParamsValidator for the given database type,
// or (nil, false) when the type does not support HA.
func getHAValidator(dbType string) (HAParamsValidator, bool) {
	v, ok := haValidators[dbType]
	return v, ok
}

// MysqlHAParamsValidator validates InnoDB Cluster and MySQL Router specific fields for MySQL HA.
// +kubebuilder:object:generate=false
type MysqlHAParamsValidator struct{}

// Validate checks MySQL-specific HA constraints. Fields validated:
//   - haConfig.mysql                    — must be present (required)
//   - haConfig.mysql.innoDBClusterName  — must be non-empty
//   - haConfig.nodes[*].nodeType        — must be "database" or "mysqlrouter"
//   - haConfig.nodes                    — exactly one database node must have role "Master"
//   - haConfig.nodes                    — mysqlrouter nodes must not have a role set
//   - haConfig.mysql.deployMySQLRouter  — if true, at least one mysqlrouter node must be present;
//     if false, no mysqlrouter nodes must be present
func (v *MysqlHAParamsValidator) Validate(haConfig *InstanceHAConfig, haPath *field.Path, errors *field.ErrorList) {
	myPath := haPath.Child("mysql")

	if haConfig.MySQL == nil {
		*errors = append(*errors, field.Required(myPath,
			"mysql config must be specified in haConfig when database type is mysql"))
		return
	}

	my := haConfig.MySQL
	if my.InnoDBClusterName == "" {
		*errors = append(*errors, field.Invalid(myPath.Child("innoDBClusterName"),
			my.InnoDBClusterName, "innoDBClusterName must be specified"))
	}

	// Validate node types, enforce exactly one Master database node, and
	// count mysqlrouter nodes for the deployMySQLRouter consistency check below.
	masterCount := 0
	routerCount := 0
	for i, node := range haConfig.Nodes {
		nodePath := haPath.Child("nodes").Index(i)
		if node.NodeType != common.HA_NODE_TYPE_DATABASE && node.NodeType != common.HA_NODE_TYPE_MYSQLROUTER {
			*errors = append(*errors, field.Invalid(nodePath.Child("nodeType"), node.NodeType,
				"nodeType must be either '"+common.HA_NODE_TYPE_DATABASE+"' or '"+common.HA_NODE_TYPE_MYSQLROUTER+"' for mysql HA"))
		}
		if node.NodeType == common.HA_NODE_TYPE_MYSQLROUTER && node.Role != "" {
			*errors = append(*errors, field.Invalid(nodePath.Child("role"), node.Role,
				"role must not be set for mysqlrouter nodes"))
		}
		if node.NodeType == common.HA_NODE_TYPE_DATABASE && node.Role == common.HA_NODE_ROLE_MASTER {
			masterCount++
		}
		if node.NodeType == common.HA_NODE_TYPE_MYSQLROUTER {
			routerCount++
		}
	}

	if len(haConfig.Nodes) > 0 && masterCount != 1 {
		*errors = append(*errors, field.Invalid(haPath.Child("nodes"), haConfig.Nodes,
			"exactly one database node must have role 'Master'"))
	}

	// Enforce consistency between deployMySQLRouter and the nodes list.
	if my.DeployMySQLRouter && routerCount == 0 {
		*errors = append(*errors, field.Invalid(myPath.Child("deployMySQLRouter"), my.DeployMySQLRouter,
			"deployMySQLRouter is true but no mysqlrouter nodes are present in haConfig.nodes"))
	}
	if !my.DeployMySQLRouter && routerCount > 0 {
		*errors = append(*errors, field.Invalid(haPath.Child("nodes"), haConfig.Nodes,
			"mysqlrouter nodes are present but deployMySQLRouter is false"))
	}
}

// MongoHAParamsValidator validates Replica Set specific fields for MongoDB HA.
// +kubebuilder:object:generate=false
type MongoHAParamsValidator struct{}

// Validate checks MongoDB-specific HA constraints. Fields validated:
//   - haConfig.mongodb                   — must be present (required)
//   - haConfig.mongodb.replicaSetName    — must be non-empty
//   - haConfig.nodes[*].nodeType         — must be "database" or "arbiter"
//   - haConfig.nodes                     — arbiter nodes must not have a role set
//   - haConfig.nodes                     — exactly one database node must have role "primary"
//   - haConfig.mongodb.deployArbiter     — if true, exactly one "arbiter" node must be present;
//     if false, no "arbiter" nodes may be present
func (v *MongoHAParamsValidator) Validate(haConfig *InstanceHAConfig, haPath *field.Path, errors *field.ErrorList) {
	mgPath := haPath.Child("mongodb")

	if haConfig.MongoDB == nil {
		*errors = append(*errors, field.Required(mgPath,
			"mongodb config must be specified in haConfig when database type is mongodb"))
		return
	}

	mg := haConfig.MongoDB
	if mg.ReplicaSetName == "" {
		*errors = append(*errors, field.Invalid(mgPath.Child("replicaSetName"),
			mg.ReplicaSetName, "replicaSetName must be specified"))
	}

	primaryCount := 0
	arbiterCount := 0
	for i, node := range haConfig.Nodes {
		nodePath := haPath.Child("nodes").Index(i)
		if node.NodeType != common.HA_NODE_TYPE_DATABASE && node.NodeType != common.HA_NODE_TYPE_ARBITER {
			*errors = append(*errors, field.Invalid(nodePath.Child("nodeType"), node.NodeType,
				"nodeType must be either '"+common.HA_NODE_TYPE_DATABASE+"' or '"+common.HA_NODE_TYPE_ARBITER+"' for mongodb HA"))
		}
		if node.NodeType == common.HA_NODE_TYPE_ARBITER && node.Role != "" {
			*errors = append(*errors, field.Invalid(nodePath.Child("role"), node.Role,
				"role must not be set for arbiter nodes; role is implied by nodeType"))
		}
		if node.NodeType == common.HA_NODE_TYPE_DATABASE && node.Role == common.HA_NODE_ROLE_MONGO_PRIMARY {
			primaryCount++
		}
		if node.NodeType == common.HA_NODE_TYPE_ARBITER {
			arbiterCount++
		}
	}

	if len(haConfig.Nodes) > 0 && primaryCount != 1 {
		*errors = append(*errors, field.Invalid(haPath.Child("nodes"), haConfig.Nodes,
			"exactly one database node must have role '"+common.HA_NODE_ROLE_MONGO_PRIMARY+"'"))
	}

	// Enforce consistency between deployArbiter and the nodes list.
	if mg.DeployArbiter && arbiterCount != 1 {
		*errors = append(*errors, field.Invalid(mgPath.Child("deployArbiter"), mg.DeployArbiter,
			"deployArbiter is true but exactly one arbiter node must be present in haConfig.nodes"))
	}
	if !mg.DeployArbiter && arbiterCount > 0 {
		*errors = append(*errors, field.Invalid(haPath.Child("nodes"), haConfig.Nodes,
			"arbiter nodes are present but deployArbiter is false"))
	}
}

// PostgresHAParamsValidator validates Patroni and HAProxy specific fields for Postgres HA.
// +kubebuilder:object:generate=false
type PostgresHAParamsValidator struct{}

// Validate checks Postgres-specific HA constraints. Fields validated:
//   - haConfig.postgres         — must be present (required)
//   - haConfig.postgres.patroniClusterName — must be non-empty
//   - haConfig.nodes[*].nodeType — must be "haproxy" or "database"
//   - haConfig.nodes            — exactly one database node must have role "Primary"
func (v *PostgresHAParamsValidator) Validate(haConfig *InstanceHAConfig, haPath *field.Path, errors *field.ErrorList) {
	pgPath := haPath.Child("postgres")

	if haConfig.Postgres == nil {
		*errors = append(*errors, field.Required(pgPath,
			"postgres config must be specified in haConfig when database type is postgres"))
		return
	}

	pg := haConfig.Postgres
	if pg.PatroniClusterName == "" {
		*errors = append(*errors, field.Invalid(pgPath.Child("patroniClusterName"),
			pg.PatroniClusterName, "patroniClusterName must be specified"))
	}

	// Validate node types (only "haproxy" and "database" are valid for Postgres HA)
	// and enforce exactly one Primary database node (Patroni constraint).
	primaryCount := 0
	for i, node := range haConfig.Nodes {
		nodePath := haPath.Child("nodes").Index(i)
		if node.NodeType != common.HA_NODE_TYPE_HAPROXY && node.NodeType != common.HA_NODE_TYPE_DATABASE {
			*errors = append(*errors, field.Invalid(nodePath.Child("nodeType"), node.NodeType,
				"nodeType must be either '"+common.HA_NODE_TYPE_HAPROXY+"' or '"+common.HA_NODE_TYPE_DATABASE+"' for postgres HA"))
		}
		if node.NodeType == common.HA_NODE_TYPE_DATABASE && node.Role == common.HA_NODE_ROLE_PRIMARY {
			primaryCount++
		}
	}

	if len(haConfig.Nodes) > 0 && primaryCount != 1 {
		*errors = append(*errors, field.Invalid(haPath.Child("nodes"), haConfig.Nodes,
			"exactly one database node must have role 'Primary'"))
	}
}
