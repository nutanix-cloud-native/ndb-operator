package api

var DefaultDatabaseNames = []string{"database_one", "database_two", "database_three"}

var AllowedDatabaseTypes = map[string]bool{
	"mysql":    true,
	"postgres": true,
	"mongodb":  true,
	"mssql":    true,
}

var ClosedSourceDatabaseTypes = map[string]bool{
	"mssql": true,
}

var AllowedLogCatchupFrequencyInMinutes = map[int]bool{
	15:  true,
	30:  true,
	45:  true,
	60:  true,
	90:  true,
	120: true,
}

var AllowedWeeklySnapshotDays = map[string]bool{
	"MONDAY":    true,
	"TUESDAY":   true,
	"WEDNESDAY": true,
	"THURSDAY":  true,
	"FRIDAY":    true,
	"SATURDAY":  true,
	"SUNDAY":    true,
}

var AllowedQuarterlySnapshotMonths = map[string]bool{
	"Jan": true,
	"Feb": true,
	"Mar": true,
}

var AllowedMySqlTypeDetails = map[string]bool{
	"server_collation":           true,
	"database_collation":         true,
	"vm_win_license_key":         true,
	"vm_dbserver_admin_password": true,
	"authentication_mode":        true,
	"sql_user_name":              true,
	"sql_user_password":          true,
	"windows_domain_profile_id":  true,
	"vm_db_server_user":          true,
}

var AllowedPostGresTypeDetails = map[string]bool{
	"listener_port": true,
}

var AllowedMongoDBTypeDetails = map[string]bool{
	"listener_port": true,
	"log_size":      true,
	"journal_size":  true,
}

var AllowedMySQLTypeDetails = map[string]bool{
	"listener_port": true,
}

var AllowedMariaDBTypeDetails = map[string]bool{
	"listener_port": true,
}
