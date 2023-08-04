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
