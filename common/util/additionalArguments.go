package util

import (
	"errors"
	"fmt"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

// Gets allowed additional argument types and indicates whether it is an action argument. Also returns if there is an error or not.
func GetAllowedAdditionalArgumentsForType(dbType string) (map[string]bool, error) {
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
	}
	// Return error
	return map[string]bool{}, errors.New(fmt.Sprintf("Could not find allowed additional arguments for database type: %s. Please ensure database type is one of the following: %s ", dbType, common.DATABASE_TYPES))
}
