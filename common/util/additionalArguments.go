package util

import (
	"fmt"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

// Returns a tuple that consists of the following:
//  1. A map where the keys are the allowed additional arguments for the database type, and the corresponding values indicates whether the key is an action argument (where true=yes and false=no).
//     Currently, all additional arguments are action arguments but this might not always be the case, thus this distinction is made so actual action arguments are appended to the appropriate provisioning body property.
//  2. An error if there is no allowed additional arguments for the corresponding type, in other words, if the dbType is not MSSQL, MongoDB, PostGres, MySQL, or Oracle. Else nil.
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
	case common.DATABASE_TYPE_ORACLE:
		return map[string]bool{
			/* Core - Has defaults (added programmatically in clone_helpers.go) */
			"vm_name":                   true, // VM name for the clone
			"dbserver_description":      true, // DB server description
			"db_password":               true, // Password for SYS user (replaces sys_password in provisioning)
			"new_db_sid":                true, // New Oracle SID for clone (max 8 chars, our successful manual test used this)
			"oracle_sid":                true, // Some NDB versions may use this instead of new_db_sid
			"listener_port":             true, // Listener port
			"enable_ha":                 true, // High Availability (RAC)
			"scan_port":                 true, // SCAN port for RAC
			"delete_logs_post_recovery": true, // Delete archive logs after recovery
			"asm_driver":                true, // ASM driver type
			/* Additional Oracle clone-specific parameters */
			"sys_asm_password":     false, // ASM password (required if using ASM/Grid)
			"client_public_key":    false, // SSH key for oracle/grid users
			"guest_os":             false, // Guest OS identifier
			"working_dir":          false, // Temporary working directory
			"delete_vm_on_failure": false, // Delete VM if clone fails
			"pre_clone_cmd":        false, // Pre-clone script hook
			"post_clone_cmd":       false, // Post-clone script hook
			/* LCM configs */
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
			"listener_port": true,
		}, nil
	case common.DATABASE_TYPE_MYSQL:
		return map[string]bool{
			"listener_port": true,
		}, nil
	case common.DATABASE_TYPE_ORACLE:
		return map[string]bool{
			/* Core - Has defaults, users can override */
			"listener_port":           true,
			"dbserver_name":           true,
			"oracle_sid":              true,
			"global_database_name":    true,
			"db_unique_name":          true,
			"database_size":           true,
			"sys_password":            true,
			"system_password":         true,
			"db_character_set":        true,
			"national_character_set":  true,
			"database_fra_size":       true,
			"enable_cdb":              true,
			"enable_tde":              true,
			"enable_ha":               true,
			"auto_tune_staging_drive": true,
			"working_dir":             true,
			"delete_logs_older_than":  true,
			"dbserver_description":    true,
			/* Additional Oracle-specific parameters */
			"sys_asm_password":            false, // ASM password (required if using ASM/Grid)
			"tde_encryption_passphrase":   false, // TDE passphrase (required if enable_tde=true)
			"client_public_key":           false, // SSH key for oracle/grid users
			"dbserver_timezone":           false,
			"cluster_name":                false, // RAC cluster name
			"scan_name":                   false, // RAC SCAN name
			"nodes":                       false, // Number of RAC nodes
			"asm_driver":                  false, // ASM driver type (None, asmlib, afd)
			"provision_type":              false, // "pdb" for pluggable database
			"pdb_name":                    false, // Pluggable database name
			"application_id":              false, // Parent CDB UUID for PDB
			"redo_log_size":               false,
			"no_of_redo_log_groups":       false,
			"pre_create_script":           true, // Pre-create script (action argument with default)
			"post_create_script":          true, // Post-create script (action argument with default)
			"pre_rollback_command":        false,
			"ensure_vm_host_distribution": false,
			/* Data Guard specific */
			"mount_mode":      false,
			"protection_mode": false,
			"DelayMins":       false,
			/* Oracle parameter overrides */
			"pga_aggregate_target": false,
			"pga_aggregate_limit":  false,
			"sga_target":           false,
			"sga_min_size":         false,
			"shared_servers":       false,
			"undo_tablespace":      false,
			"cpu_count":            false,
		}, nil
	default:
		return map[string]bool{}, fmt.Errorf("could not find allowed additional arguments for database of type: %s. Please ensure database type is one of the following: %s ", dbType, common.DATABASE_TYPES)
	}
}
