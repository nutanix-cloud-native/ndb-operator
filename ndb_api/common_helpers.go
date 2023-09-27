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

// Converts a map to an action arguments list
func convertMapToActionArguments(myMap map[string]string) []ActionArgument {
	actionArgs := []ActionArgument{}
	for name, value := range myMap {
		actionArgs = append(actionArgs, ActionArgument{Name: name, Value: value})
	}
	return actionArgs
}

// Gets MSSQL allowed additional arguments.
// If key is an action argument value is true, else false.
func GetMsSQLAllowedAdditionalArguments() map[string]bool {
	return map[string]bool{
		/* Has a default */
		"sql_user_name":                true,
		"authentication_mode":          true,
		"server_collation":             true,
		"database_collation":           true,
		"dbParameterProfileIdInstance": true,
		"vm_dbserver_admin_password":   true,
		/* No default */
		"sql_user_password":         true,
		"vm_win_license_key":        true,
		"windows_domain_profile_id": true,
		"vm_db_server_user":         true,
	}
}

// Gets MSSQL default action arguments.
func getMsSQLDefaultActionArguments(dbServerName string, dbParamInstanceProfileId string, adminPassword string) map[string]string {
	return map[string]string{
		"working_dir":                       "C:\\temp",
		"sql_user_name":                     "sa",
		"authentication_mode":               "windows",
		"delete_vm_on_failure":              "false",
		"is_gmsa_sql_service_account":       "false",
		"provision_from_backup":             "false",
		"distribute_database_data":          "true",
		"retain_database_in_restoring_mode": "false",
		"dbserver_name":                     dbServerName,
		"server_collation":                  "SQL_Latin1_General_CP1_CI_AS",
		"database_collation":                "SQL_Latin1_General_CP1_CI_AS",
		"dbParameterProfileIdInstance":      dbParamInstanceProfileId,
		"vm_dbserver_admin_password":        adminPassword,
	}
}

// Gets MongoDb allowed additional arguments.
// If key is an action argument value is true, else false.
func GetMongoDbAllowedAdditionalArguments() map[string]bool {
	return map[string]bool{
		/* Has a default */
		"listener_port": true,
		"log_size":      true,
		"journal_size":  true,
	}
}

// Gets MongoDB default action arguments.
func getMongoDbDefaultActionArguments(dbPassword string, databaseNames string) map[string]string {
	return map[string]string{
		"listener_port":  "27017",
		"log_size":       "100",
		"journal_size":   "100",
		"restart_mongod": "true",
		"working_dir":    "/tmp",
		"db_user":        "admin",
		"backup_policy":  "primary_only",
		"db_password":    dbPassword,
		"database_names": databaseNames,
	}
}

// Gets Postgres allowed additional arguments.
// If key is an action argument value is true, else false.
func GetPostgresAllowedAdditionalArguments() map[string]bool {
	return map[string]bool{
		/* Has a default */
		"listener_port": true,
	}
}

// Gets Postgres default action arguments.
func getPostgresDefaultActionArguments(dbPassword string, databaseNames string) map[string]string {
	return map[string]string{
		"proxy_read_port":         "5001",
		"listener_port":           "5432",
		"proxy_write_port":        "5000",
		"enable_synchronous_mode": "false",
		"auto_tune_staging_drive": "true",
		"backup_policy":           "primary_only",
		"db_password":             "dbPassword",
		"database_names":          "databaseNames",
	}
}

// Gets MYSQL allowed additional arguments.
// If key is an action argument value is true, else false.
func GetMySQLAllowedAdditionalArguments() map[string]bool {
	/* Has a default */
	return map[string]bool{
		"listener_port": true,
	}
}

// Gets MYSQL default action arguments.
// If key is an action argument value is true, else false.
func getMySQLDefaultActionArguments(dbPassword string, databaseNames string) map[string]string {
	return map[string]string{
		"listener_port":           "3306",
		"db_password":             dbPassword,
		"database_names":          databaseNames,
		"auto_tune_staging_drive": "true",
	}
}
