/*
Copyright 2021-2022 Nutanix, Inc.

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

const (
	DATABASE_CR_STATUS_DELETING     = "DELETING"
	DATABASE_CR_STATUS_EMPTY        = ""
	DATABASE_CR_STATUS_READY        = "READY"
	DATABASE_CR_STATUS_PROVISIONING = "PROVISIONING"

	DATABASE_DEFAULT_PORT_MONGODB  = 27017
	DATABASE_DEFAULT_PORT_MYSQL    = 3306
	DATABASE_DEFAULT_PORT_POSTGRES = 5432

	DATABASE_ENGINE_TYPE_GENERIC  = "Generic"
	DATABASE_ENGINE_TYPE_POSTGRES = "postgres_database"
	DATABASE_ENGINE_TYPE_MYSQL    = "mysql_database"
	DATABASE_ENGINE_TYPE_MONGODB  = "mongodb_database"

	FINALIZER_DATABASE_INSTANCE = "ndb.nutanix.com/finalizerdatabaseinstance"
	FINALIZER_DATABASE_SERVER   = "ndb.nutanix.com/finalizerdatabaseserver"

	PROFILE_TYPE_COMPUTE            = "Compute"
	PROFILE_TYPE_DATABASE_PARAMETER = "Database_Parameter"
	PROFILE_TYPE_NETWORK            = "Network"
	PROFILE_TYPE_SOFTWARE           = "Software"
	PROFILE_TYPE_STORAGE            = "Storage"

	PROPERTY_NAME_VM_IP = "vm_ip"

	SLA_NAME_NONE = "NONE"

	TOPOLOGY_ALL      = "ALL"
	TOPOLOGY_CLUSTER  = "cluster"
	TOPOLOGY_DATABASE = "database"
	TOPOLOGY_INSTANCE = "instance"
	TOPOLOGY_SINGLE   = "single"
)
