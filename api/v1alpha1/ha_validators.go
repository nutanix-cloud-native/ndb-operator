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
}

// getHAValidator returns the registered HAParamsValidator for the given database type,
// or (nil, false) when the type does not support HA.
func getHAValidator(dbType string) (HAParamsValidator, bool) {
	v, ok := haValidators[dbType]
	return v, ok
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
