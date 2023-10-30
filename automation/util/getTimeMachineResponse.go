package util

import (
	"context"
	"fmt"
	"strconv"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Wrapper function called in all TestSuite TestProvisioningSuccess methods. Returns a DatabaseResponse which indicates if provison was succesful
func GetTimemachineResponseByDatabaseId(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (timemachineResponse ndb_api.TimeMachineResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetTimemachineResponse() starting...")
	errBaseMsg := "Error: GetTimemachineResponse() ended"

	// Get NDBServer CR
	ndbServer, err := v1alpha1ClientSet.NDBServers(st.NdbServer.Namespace).Get(st.NdbServer.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Could not fetch ndbServer '%s' CR! %s\n", errBaseMsg, ndbServer.Name, err)
	} else {
		logger.Printf("Retrieved ndbServer '%s' CR from v1alpha1ClientSet", ndbServer.Name)
	}

	// Get Database CR
	database, err := v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Could not fetch database '%s' CR! %s\n", errBaseMsg, database.Name, err)
	} else {
		logger.Printf("Retrieved database '%s' CR from v1alpha1ClientSet", database.Name)
	}

	// Get NDB username and password from NDB CredentialSecret
	ndb_secret_name := ndbServer.Spec.CredentialSecret
	secret, err := clientset.CoreV1().Secrets(database.Namespace).Get(context.TODO(), ndb_secret_name, metav1.GetOptions{})
	username, password := string(secret.Data[common.SECRET_DATA_KEY_USERNAME]), string(secret.Data[common.SECRET_DATA_KEY_PASSWORD])
	if err != nil || username == "" || password == "" {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Could not fetch data from secret! %s\n", errBaseMsg, err)
	}

	// Create ndbClient and getting databaseResponse so we can get timemachine id
	ndbClient := ndb_client.NewNDBClient(username, password, ndbServer.Spec.Server, "", true)
	databaseResponse, err := ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Database response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	// Get timemachine
	timemachineResponse, err = ndb_api.GetTimeMachineById(context.TODO(), ndbClient, databaseResponse.TimeMachineId)
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! time machine response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	logger.Printf("Timemachine response.status: %s.\n", timemachineResponse.Status)
	logger.Println("GetTimemachineResponse() ended!")

	return timemachineResponse, nil
}

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
