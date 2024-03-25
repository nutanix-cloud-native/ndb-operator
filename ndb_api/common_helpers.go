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
	"errors"

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

func GetDatabaseTypeFromEngine(engine string) string {
	switch engine {
	case common.DATABASE_ENGINE_TYPE_POSTGRES:
		return common.DATABASE_TYPE_POSTGRES
	case common.DATABASE_ENGINE_TYPE_MYSQL:
		return common.DATABASE_TYPE_MYSQL
	case common.DATABASE_ENGINE_TYPE_MONGODB:
		return common.DATABASE_TYPE_MONGODB
	case common.DATABASE_ENGINE_TYPE_MSSQL:
		return common.DATABASE_TYPE_MSSQL
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

// Get specific implementation of the DBProvisionRequestAppender interface based on the provided databaseType
func GetRequestAppender(databaseType string, isHighAvailability bool) (requestAppender RequestAppender, err error) {
	switch databaseType {
	case common.DATABASE_TYPE_MYSQL:
		requestAppender = &MySqlRequestAppender{}
	case common.DATABASE_TYPE_POSTGRES:
		if isHighAvailability {
			requestAppender = &PostgresHARequestAppender{}
		} else {
			requestAppender = &PostgresRequestAppender{}
		}
	case common.DATABASE_TYPE_MONGODB:
		requestAppender = &MongoDbRequestAppender{}
	case common.DATABASE_TYPE_MSSQL:
		requestAppender = &MSSQLRequestAppender{}
	default:
		return nil, errors.New("invalid database type: supported values: mssql, mysql, postgres, mongodb")
	}
	return
}
