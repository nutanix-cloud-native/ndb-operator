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