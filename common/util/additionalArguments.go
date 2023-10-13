package util

import (
	"errors"
	"fmt"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

// Returns a tuple that consists of the following:
//  1. A map where the keys are the allowed additional arguments for the database type, and the corresponding values indicates whether the key is an action argument (where true=yes and false=no).
//     Currently, all additional arguments are action arguments but this might not always be the case, thus this distinction is made so actual action arguments are appended to the appropriate provisioning body property.
//  2. An error if there is no allowed additional arguments for the corresponding type, in other words, if the dbType is not MSSQL, MngoDB, PostGres, or MYSQL. Else nil.
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
