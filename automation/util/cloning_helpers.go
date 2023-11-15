package util

import (
	"context"
	"fmt"
	"os"
	"strconv"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/automation"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	corev1 "k8s.io/api/core/v1"
)

// Checking that TM info that was specified in yaml is returned in TM response
func CheckTmInfo(ctx context.Context, database *ndbv1alpha1.Database, tmResponse *ndb_api.TimeMachineResponse) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CheckTmInfo() starting...")

	tmInfo := database.Spec.Instance.TMInfo
	invalidProperties := make([]string, 0, 10)

	if tmInfo.Name != tmResponse.Name {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'name', expected: %s, got: %s", tmInfo.Name, tmResponse.Name))
	}

	if tmInfo.Description != tmResponse.Description {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'description', expected: %s, got: %s", tmInfo.Description, tmResponse.Description))
	}

	if tmInfo.SLAName != tmResponse.Sla.Name {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'slaName', expected: %s, got: %s", tmInfo.Name, tmResponse.Sla.Name))
	}

	gotSnapHour := tmResponse.Schedule.SnapshotTimeOfDay.Hours
	gotSnapMinute := tmResponse.Schedule.SnapshotTimeOfDay.Minutes
	gotSnapSecond := tmResponse.Schedule.SnapshotTimeOfDay.Seconds
	expectedSnapHour, expectedSnapMinute, expectedSnapSecond, err := extractDailySnapshotTime(ctx, tmInfo.DailySnapshotTime)
	if err != nil {
		invalidProperties = append(invalidProperties, fmt.Sprintf("failed to convert dailySnapshotTime: %s", err))
	} else {
		if gotSnapHour != expectedSnapHour || gotSnapMinute != expectedSnapMinute || gotSnapSecond != expectedSnapSecond {
			invalidProperties = append(invalidProperties, fmt.Sprintf("for 'dailySnapshotTime', expected %d:%d:%d, got: %d:%d:%d", expectedSnapHour, expectedSnapMinute, expectedSnapSecond, gotSnapHour, gotSnapMinute, gotSnapSecond))
		}
	}

	if tmInfo.SnapshotsPerDay != tmResponse.Schedule.ContinuousSchedule.SnapshotsPerDay {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'snapshotsPerDay', expected: %d, got: %d", tmInfo.SnapshotsPerDay, tmResponse.Schedule.ContinuousSchedule.SnapshotsPerDay))
	}

	if tmInfo.LogCatchUpFrequency != tmResponse.Schedule.ContinuousSchedule.LogBackupInterval {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'logCatchUpFrequency', expected: %d, got: %d", tmInfo.LogCatchUpFrequency, tmResponse.Schedule.ContinuousSchedule.LogBackupInterval))
	}

	if tmInfo.WeeklySnapshotDay != tmResponse.Schedule.WeeklySchedule.DayOfWeek {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'weeklySnapshotDay', expected: %s, got: %s", tmInfo.WeeklySnapshotDay, tmResponse.Schedule.WeeklySchedule.DayOfWeek))
	}

	if tmInfo.MonthlySnapshotDay != tmResponse.Schedule.MonthlySchedule.DayOfMonth {
		invalidProperties = append(invalidProperties, fmt.Sprintf("for 'monthlySnapshotDay', expected: %d, got: %d", tmInfo.MonthlySnapshotDay, tmResponse.Schedule.MonthlySchedule.DayOfMonth))
	}

	// if tmInfo.QuarterlySnapshotMonth != tmResponse.Schedule.QuarterlySchedule.StartMonth {
	// 	invalidProperties = append(invalidProperties, fmt.Sprintf("for 'quarterlySnapshotMonth', expected: %s, got: %s", tmInfo.QuarterlySnapshotMonth, tmResponse.Schedule.QuarterlySchedule.StartMonth))
	// }

	logger.Println("CheckTmInfo() ended!")

	if len(invalidProperties) == 0 {
		return nil
	} else {
		return fmt.Errorf("CheckTmInfo() failed! Found invalid properties: %v", invalidProperties)
	}
}

// Extras hour, minute, and second from e.g: 12:12:12
func extractDailySnapshotTime(ctx context.Context, dailySnapshotTime string) (hour int, minute int, second int, err error) {
	logger := GetLogger(ctx)
	logger.Println("extractDailySnapshotTime() starting...")

	hour, err = strconv.Atoi(dailySnapshotTime[0:2]) // 12:12:12
	if err != nil {
		fmt.Println("Conversion error for hour:", err)
		return -1, -1, -1, fmt.Errorf("ExtractDailySnapshotTime() failed! Conversion error for hour: %v", err)
	} else {
		logger.Println(fmt.Sprintf("Extracted hour: %d", hour))
	}

	minute, err = strconv.Atoi(dailySnapshotTime[3:5])
	if err != nil {
		fmt.Println("Conversion error for minute:", err)
		return -1, -1, -1, fmt.Errorf("ExtractDailySnapshotTime() failed! Conversion error for minute: %v", err)
	} else {
		logger.Println(fmt.Sprintf("Extracted minute: %d", minute))
	}

	second, err = strconv.Atoi(dailySnapshotTime[6:8])
	if err != nil {
		fmt.Println("Conversion error for second:", err)
		return -1, -1, -1, fmt.Errorf("ExtractDailySnapshotTime() failed! Conversion error for second: %v", err)
	} else {
		logger.Println(fmt.Sprintf("Extracted second: %d", second))
	}

	logger.Println("extractDailySnapshotTime() ended!...")
	return
}

// Checks if name is in environment. If not, defaults
func getDatabaseNameFromEnvElseDefault(envName string, defaultName string) (string, bool) {
	name, ok := os.LookupEnv(envName)
	if ok {
		return name, true
	} else {
		return defaultName, false
	}
}

// Gets database name for clone test
func getDatabaseName(ctx context.Context, database *ndbv1alpha1.Database) (name string, err error) {
	logger := GetLogger(ctx)
	logger.Println("getDatabaseName() starting...")

	var databaseType string
	var fromEnv bool

	switch database.Spec.Clone.Type {
	case common.DATABASE_TYPE_MONGODB:
		databaseType = common.DATABASE_TYPE_MONGODB
		name, fromEnv = getDatabaseNameFromEnvElseDefault(automation.MONGO_SI_CLONING_NAME_ENV, automation.MONGO_SI_CLONING_NAME_DEFAULT)
	case common.DATABASE_TYPE_MSSQL:
		databaseType = common.DATABASE_TYPE_MSSQL
		name, fromEnv = getDatabaseNameFromEnvElseDefault(automation.MSSQL_SI_CLONING_NAME_ENV, automation.MSSQL_SI_CLONING_NAME_DEFAULT)
	case common.DATABASE_TYPE_MYSQL:
		databaseType = common.DATABASE_TYPE_MYSQL
		name, fromEnv = getDatabaseNameFromEnvElseDefault(automation.MYSQL_SI_CLONING_NAME_ENV, automation.MYSQL_SI_CLONING_NAME_DEFAULT)
	case common.DATABASE_TYPE_POSTGRES:
		databaseType = common.DATABASE_TYPE_POSTGRES
		name, fromEnv = getDatabaseNameFromEnvElseDefault(automation.POSTGRES_SI_CLONING_NAME_ENV, automation.POSTGRES_SI_CLONING_NAME_DEFAULT)
	default:
		err = fmt.Errorf("Invalid database type: %s. Valid database types are: %s", database.Spec.Clone.Type, common.DATABASE_TYPES)
	}

	if err == nil {
		if fromEnv {
			logger.Printf("Database named: %s of type: %s was retrieved from the environment.", name, databaseType)
		} else {
			logger.Printf("Database named: %s of type: %s could not be found in the environment and was defaulted.", name, databaseType)
		}
	}
	logger.Println("getDatabaseName() exited!")

	return
}

// Gets snapshot id from TimeMachineGetSnapshotsResponse
func getSnapshotId(ctx context.Context, response ndb_api.TimeMachineGetSnapshotsResponse, nxClusterId string) (string, error) {
	logger := GetLogger(ctx)
	logger.Println("getSnapshotId() starting...")

	snapshotsPerNxCluster, ok := response.SnapshotsPerNxCluster[nxClusterId]
	if !ok || len(snapshotsPerNxCluster) == 0 {
		return "", fmt.Errorf("No snapshots for cluster id: %s found!", nxClusterId)
	}

	// Return a snapshot id (index 0 is DAILY, index 1 is CONTINUOUS, and index 2 is MANUAL)
	for i := 0; i < len(snapshotsPerNxCluster); i++ {
		snapshots := snapshotsPerNxCluster[i].Snapshots
		// Return the first available snapshot
		for j := 0; j < len(snapshots); j++ {
			logger.Println("getSnapshotId() ended!")
			return snapshots[i].Id, nil
		}
	}

	return "", fmt.Errorf("No snapshots for cluster id: %s found!", nxClusterId)
}

// Updates clones sourceDatabaseId, nxClusterId, and snapshotId
func updateClone(ctx context.Context, database *ndbv1alpha1.Database, ndbServer *ndbv1alpha1.NDBServer, ndbSecret *corev1.Secret) (err error) {
	logger := GetLogger(ctx)
	logger.Println("updateClone() starting...")

	// Get database name
	databaseName, err := getDatabaseName(ctx, database)
	if err != nil {
		return fmt.Errorf("updateClone() failed! Error: %s", err)
	}

	// Create ndb client
	ndbClient := ndb_client.NewNDBClient(
		ndbSecret.StringData[common.SECRET_DATA_KEY_USERNAME],
		ndbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD],
		ndbServer.Spec.Server, "", true)

	// Fetch database by name and extract sourceDatabaseId, tmId, nxClusterId
	fetchedDatabase, err := ndb_api.GetDatabaseByName(ctx, ndbClient, databaseName)
	if err != nil {
		return fmt.Errorf("updateClone() failed! Error: %s", err)
	} else if fetchedDatabase == nil {
		return fmt.Errorf("updateClone() failed! database is null, there is no database of name: %s ", databaseName)
	}
	sourceDatabaseId := fetchedDatabase.Id
	tmId := fetchedDatabase.TimeMachineId
	nxClusterId := fetchedDatabase.DatabaseNodes[0].DbServer.NxClusterId
	logger.Printf("Fetched database: %s with sourceDatabaseId: %s, tmId: %s, nxClusterId: %s.", databaseName, sourceDatabaseId, tmId, nxClusterId)

	// Get snapshots response from time machine
	response, err := ndb_api.GetSnapshotsForTM(ctx, ndbClient, tmId)
	if err != nil {
		return fmt.Errorf("UpdateClone() failed! Error: %s", err)
	} else {
		logger.Printf("Called GetSnapshotsForTM and retrieved the following snapshots for cluster ids: %s", response)
	}

	// Get a snapshot
	snapshotId, err := getSnapshotId(ctx, response, nxClusterId)
	if err != nil {
		return fmt.Errorf("UpdateClone() failed! %s", err)
	} else {
		logger.Printf("Retrieved snapshot: %s for clusterId: %s", snapshotId, nxClusterId)
	}

	// Update sourceDatabaseId, nxClusterId, and snapshotId
	database.Spec.Clone.SourceDatabaseId = sourceDatabaseId
	database.Spec.Clone.ClusterId = nxClusterId
	database.Spec.Clone.SnapshotId = snapshotId

	return
}
