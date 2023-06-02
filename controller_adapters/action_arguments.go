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
	"errors"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
)

type DatabaseActionArgs interface {
	Get(dbSpec v1alpha1.DatabaseSpec) []ndb_api.ActionArgument
}

// MysqlActionArgs implements the DatabaseActionArgs interface
type MysqlActionArgs struct{}

// PostgresActionArgs implements the DatabaseActionArgs interface
type PostgresActionArgs struct{}

// MongodbActionArgs implements the DatabaseActionArgs interface
type MongodbActionArgs struct{}

// DatabaseActionArgs implementation for the type MysqlActionArgs
func (m *MysqlActionArgs) Get(dbSpec v1alpha1.DatabaseSpec) []ndb_api.ActionArgument {
	return []ndb_api.ActionArgument{
		{
			Name:  "listener_port",
			Value: "3306",
		},
	}
}

// DatabaseActionArgs implementation for the type PostgresActionArgs
func (p *PostgresActionArgs) Get(dbSpec v1alpha1.DatabaseSpec) []ndb_api.ActionArgument {
	return []ndb_api.ActionArgument{
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "listener_port",
			Value: "5432",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "false",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
	}
}

// DatabaseActionArgs implementation for the type MongodbActionArgs
func (m *MongodbActionArgs) Get(dbSpec v1alpha1.DatabaseSpec) []ndb_api.ActionArgument {
	return []ndb_api.ActionArgument{
		{
			Name:  "listener_port",
			Value: "27017",
		},
		{
			Name:  "log_size",
			Value: "100",
		},
		{
			Name:  "journal_size",
			Value: "100",
		},
		{
			Name:  "restart_mongod",
			Value: "true",
		},
		{
			Name:  "working_dir",
			Value: "/tmp",
		},
		{
			Name:  "db_user",
			Value: "admin",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
	}
}

// Returns the concrete implementation type of action arguments based on the type of database
func GetActionArgumentsByDatabaseType(databaseType string) (DatabaseActionArgs, error) {
	var dbTypeActionArgs DatabaseActionArgs
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		dbTypeActionArgs = &MysqlActionArgs{}
	case common.DATABASE_TYPE_POSTGRES:
		dbTypeActionArgs = &PostgresActionArgs{}
	case common.DATABASE_TYPE_MONGODB:
		dbTypeActionArgs = &MongodbActionArgs{}
	default:
		return nil, errors.New("invalid database type: supported values: mysql, postgres, mongodb")
	}
	return dbTypeActionArgs, nil
}
