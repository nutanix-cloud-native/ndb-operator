package api

var DefaultDatabaseInstanceName = "database_instance_name"

var DefaultDatabaseNames = []string{"database_one", "database_two", "database_three"}

var DatabaseInstanceTypes = []string{"database_one", "database_two", "database_three"}

var AllowedDatabaseTypes = map[string]bool{
	"mysql":    true,
	"postgres": true,
	"mongodb":  true,
	"mssql":    true,
	"oracle":   true,
}

var ClosedSourceDatabaseTypes = map[string]bool{
	"oracle": true,
	"mssql":  true,
}
