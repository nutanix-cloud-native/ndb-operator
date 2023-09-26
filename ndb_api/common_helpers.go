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

package ndb_api

import (
	"sort"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

func GetDatabaseEngineName(dbType string) string {
	switch dbType {
	case common.DATABASE_TYPE_POSTGRES:
		return common.DATABASE_ENGINE_TYPE_POSTGRES
	case common.DATABASE_TYPE_MYSQL:
		return common.DATABASE_ENGINE_TYPE_MYSQL
	case common.DATABASE_TYPE_MONGODB:
		return common.DATABASE_ENGINE_TYPE_MONGODB
	case common.DATABASE_TYPE_MSSQL:
		return common.DATABASE_ENGINE_TYPE_MSSQL
	default:
		return ""
	}
}

func GetDatabasePortByType(dbType string) int32 {
	switch dbType {
	case common.DATABASE_TYPE_POSTGRES:
		return common.DATABASE_DEFAULT_PORT_POSTGRES
	case common.DATABASE_TYPE_MONGODB:
		return common.DATABASE_DEFAULT_PORT_MONGODB
	case common.DATABASE_TYPE_MYSQL:
		return common.DATABASE_DEFAULT_PORT_MYSQL
	case common.DATABASE_TYPE_MSSQL:
		return common.DATABASE_DEFAULT_PORT_MSSQL
	default:
		return -1
	}
}

// Sorts action arguments in place by name
func sortActionArgsByName(actionArgs []ActionArgument) {
	sort.Slice(actionArgs, func(i, j int) bool {
		return actionArgs[i].Name < actionArgs[j].Name
	})
}

// Gets MSSQL overridable action arguments.
func getMsSQLOverridableActionArgs(dbParameterProfileIdInstance string, vmDbServerAdminPassword string) []ActionArgument {
	return []ActionArgument{
		{Name: "sql_user_name", Value: "sa"},
		{Name: "authentication_mode", Value: "windows"},
		{Name: "server_collation", Value: "SQL_Latin1_General_CP1_CI_AS"},
		{Name: "database_collation", Value: "SQL_Latin1_General_CP1_CI_AS"},
		{Name: "dbParameterProfileIdInstance", Value: dbParameterProfileIdInstance},
		{Name: "vm_dbserver_admin_password", Value: vmDbServerAdminPassword},
	}
}

// Gets MSSQL non-overridable action arguments.
func getMsSQLNonOverridableActionArgs(dbServerName string) []ActionArgument {
	return []ActionArgument{
		{Name: "working_dir", Value: "C:\\temp"},
		{Name: "delete_vm_on_failure", Value: "false"},
		{Name: "is_gmsa_sql_service_account", Value: "false"},
		{Name: "provision_from_backup", Value: "false"},
		{Name: "distribute_database_data", Value: "true"},
		{Name: "retain_database_in_restoring_mode", Value: "false"},
		{Name: "dbserver_name", Value: dbServerName},
	}
}

// Gets MongoDB overridable action arguments.
func getMongoDbOverridableActionArgs() []ActionArgument {
	return []ActionArgument{
		{Name: "listener_port", Value: "27017"},
		{Name: "log_size", Value: "100"},
		{Name: "journal_size", Value: "100"},
	}
}

// Gets MongoDB non-overridable action arguments.
func getMongoDbNonOverridableActionArgs(dbPassword string, databaseNames string) []ActionArgument {
	return []ActionArgument{
		{Name: "restart_mongod", Value: "true"},
		{Name: "working_dir", Value: "/tmp"},
		{Name: "db_user", Value: "admin"},
		{Name: "backup_policy", Value: "primary_only"},
		{Name: "db_password", Value: dbPassword},
		{Name: "database_names", Value: databaseNames},
	}
}

// Gets Postgres overridable action arguments.
func getPostgresOverridableActionArgs() []ActionArgument {
	return []ActionArgument{
		{Name: "listener_port", Value: "5432"},
	}
}

// Gets Postgres non-overridable action arguments.
func getPostgresNonOverridableActionArgs(dbPassword string, databaseNames string) []ActionArgument {
	return []ActionArgument{
		{Name: "proxy_read_port", Value: "5001"},
		{Name: "proxy_write_port", Value: "5000"},
		{Name: "enable_synchronous_mode", Value: "false"},
		{Name: "auto_tune_staging_drive", Value: "true"},
		{Name: "backup_policy", Value: "primary_only"},
		{Name: "db_password", Value: dbPassword},
		{Name: "database_names", Value: databaseNames},
	}
}

// Gets MYSQL overridable action arguments.
func getMySQLOverridableActionArgs() []ActionArgument {
	return []ActionArgument{
		{Name: "listener_port", Value: "3306"},
	}
}

// Gets MYSQL non-overridable action arguments.
func getMySQLNonOverridableActionArgs(dbPassword string, databaseNames string) []ActionArgument {
	return []ActionArgument{
		{Name: "db_password", Value: dbPassword},
		{Name: "database_names", Value: databaseNames},
	}
}
