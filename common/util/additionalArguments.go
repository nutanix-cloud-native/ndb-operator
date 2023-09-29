package util

import "github.com/nutanix-cloud-native/ndb-operator/common"

// Gets allowed additional argument types and indicates whether it is an action argument
func GetAllowedAdditionalArgumentsForType(typ string) map[string]bool {
	switch typ {
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
		}
	case common.DATABASE_TYPE_MONGODB:
		return map[string]bool{
			/* Has a default */
			"listener_port": true,
			"log_size":      true,
			"journal_size":  true,
		}
	case common.DATABASE_TYPE_POSTGRES:
		return map[string]bool{
			/* Has a default */
			"listener_port": true,
		}
	case common.DATABASE_TYPE_MYSQL:
		return map[string]bool{
			"listener_port": true,
		}
	}
	// Should never happen
	return map[string]bool{}
}
