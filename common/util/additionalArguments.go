package util

import (
	"fmt"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

// Returns a tuple that consists of the following:
//  1. A map where the keys are the allowed additional arguments for the database type, and the corresponding values indicates whether the key is an action argument (where true=yes and false=no).
//     Currently, all additional arguments are action arguments but this might not always be the case, thus this distinction is made so actual action arguments are appended to the appropriate provisioning body property.
//  2. An error if there is no allowed additional arguments for the corresponding type, in other words, if the dbType is not MSSQL, MongoDB, PostGres, or MYSQL. Else nil.
func GetAllowedAdditionalArguments(isClone bool, dbType string) (map[string]bool, error) {
	if isClone {
		return GetAllowedAdditionalArgumentsForClone(dbType)
	} else {
		return GetAllowedAdditionalArgumentsForDatabase(dbType)
	}
}

func GetAllowedAdditionalArgumentsForClone(dbType string) (map[string]bool, error) {
	switch dbType {
	case common.DATABASE_TYPE_MSSQL:
		return map[string]bool{
			/* Has a default */
			"vm_name":                    true,
			"database_name":              true,
			"vm_dbserver_admin_password": true,
			"dbserver_description":       true,
			"sql_user_name":              true,
			"authentication_mode":        true,
			"instance_name":              true,
			/* No default */
			"windows_domain_profile_id":   true,
			"era_worker_service_user":     true,
			"sql_service_startup_account": true,
			"vm_win_license_key":          true,
			"target_mountpoints_location": true,
			"expireInDays":                false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"expiryDateTimezone":          false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"deleteDatabase":              false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"refreshInDays":               false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshTime":                 false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshDateTimezone":         false, // In lcmConfig.refreshDetails.refreshDetails
		}, nil
	case common.DATABASE_TYPE_MONGODB:
		return map[string]bool{
			/* No default */
			"expireInDays":        false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"expiryDateTimezone":  false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"deleteDatabase":      false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"refreshInDays":       false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshTime":         false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshDateTimezone": false, // In lcmConfig.refreshDetails.refreshDetails
		}, nil

	case common.DATABASE_TYPE_POSTGRES:
		return map[string]bool{
			/* No default */
			"expireInDays":        false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"expiryDateTimezone":  false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"deleteDatabase":      false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"refreshInDays":       false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshTime":         false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshDateTimezone": false, // In lcmConfig.refreshDetails.refreshDetails
		}, nil
	case common.DATABASE_TYPE_MYSQL:
		return map[string]bool{
			/* No default */
			"expireInDays":        false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"expiryDateTimezone":  false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"deleteDatabase":      false, // In lcmConfig.databaseLCMConfig.expiryDetails
			"refreshInDays":       false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshTime":         false, // In lcmConfig.refreshDetails.refreshDetails
			"refreshDateTimezone": false, // In lcmConfig.refreshDetails.refreshDetails
		}, nil
	default:
		return map[string]bool{}, fmt.Errorf("could not find allowed additional arguments for clone of type: %s. Please ensure database type is one of the following: %s ", dbType, common.DATABASE_TYPES)
	}
}

func GetAllowedAdditionalArgumentsForDatabase(dbType string) (map[string]bool, error) {
	switch dbType {
	case common.DATABASE_TYPE_MSSQL:
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
		}, nil
	case common.DATABASE_TYPE_MONGODB:
		return map[string]bool{
			/* Has a default */
			"listener_port": true,
			"log_size":      true,
			"journal_size":  true,
		}, nil
	case common.DATABASE_TYPE_POSTGRES:
		return map[string]bool{
			/* Has a default */
			"listener_port":           true,
			"proxy_read_port":         true,
			"proxy_write_port":        true,
			"enable_synchronous_mode": true,
			"auto_tune_staging_drive": true,
			"backup_policy":           true,
			"db_password":             true,
			"database_names":          true,
			"provision_virtual_ip":    true,
			"deploy_haproxy":          true,
			"failover_mode":           true,
			"node_type":               true,
			"allocate_pg_hugepage":    true,
			"cluster_database":        true,
			"archive_wal_expire_days": true,
			"enable_peer_auth":        true,
			"cluster_name":            true,
			"patroni_cluster_name":    true,
		}, nil
	case common.DATABASE_TYPE_MYSQL:
		return map[string]bool{
			"listener_port": true,
		}, nil
	default:
		return map[string]bool{}, fmt.Errorf("could not find allowed additional arguments for database of type: %s. Please ensure database type is one of the following: %s ", dbType, common.DATABASE_TYPES)
	}
}
