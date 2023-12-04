/*
Copyright 2022-2023 Nutanix, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller_adapters

import (
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/stretchr/testify/assert"
)

// Tests that GetName() retrieves Name correctly
func TestDatabase_GetName(t *testing.T) {

	tests := []struct {
		name             string
		database         Database
		wantInstanceName string
	}{
		{
			name: "Contains Instance Name",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Name: "test-instance-name",
						},
					},
				},
			},
			wantInstanceName: "test-instance-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotInstanceName := tt.database.GetName()
			if gotInstanceName != tt.wantInstanceName {
				t.Errorf("Database.GetName() gotInstanceName = %v, want %v", gotInstanceName, tt.wantInstanceName)
			}
		})
	}
}

// Tests the GetDescription() function against the following:
// 1. Description is NOT empty
// 2. Description IS empty, in this case, a description is created for the user based on instance name
func TestDatabase_GetDescription(t *testing.T) {

	tests := []struct {
		name            string
		database        Database
		wantDescription string
	}{
		{
			name: "Description is NOT empty",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Description: "test-description",
						},
					},
				},
			},
			wantDescription: "test-description",
		},
		{
			name: "Description IS empty",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Name:        "test-instance-name",
							Description: "",
						},
					},
				},
			},
			wantDescription: "Created by ndb-operator: test-instance-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotDescription := tt.database.GetDescription()
			if gotDescription != tt.wantDescription {
				t.Errorf("Database.GetDescription() gotDescription = %v, want %v", gotDescription, tt.wantDescription)
			}
		})
	}
}

// Tests the GetInstanceType() retrieves Type correctly:
func TestDatabase_GetInstanceType(t *testing.T) {

	name := "Contains Type"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					Type: "test-type",
				},
			},
		},
	}
	wantType := "test-type"

	t.Run(name, func(t *testing.T) {

		gotType := database.GetInstanceType()
		if gotType != wantType {
			t.Errorf("Database.GetInstanceType() gotType = %v, want %v", gotType, wantType)
		}
	})
}

// Tests the GetAdditionalArguments() retrieves AdditionalArguments correctly:
func TestDatabase_GetAdditionalArguments(t *testing.T) {

	name := "Contains Additional Arguments"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					AdditionalArguments: map[string]string{
						"valid_key": "valid_value",
					},
				},
			},
		},
	}
	wantAdditionalArguments := map[string]string{
		"valid_key": "valid_value",
	}

	t.Run(name, func(t *testing.T) {

		gotAdditionalArguments := database.GetAdditionalArguments()
		if !reflect.DeepEqual(wantAdditionalArguments, gotAdditionalArguments) {
			t.Errorf("Database.GetInstanceTypeDetails gotTypeDetails = %v, want %v", gotAdditionalArguments, wantAdditionalArguments)
		}
	})
}

// Tests the GetInstanceDatabaseNames() retrieves DatabaseNames correctly:
func TestDatabase_GetInstanceDatabaseNames(t *testing.T) {

	name := "Contains DatabaseNames"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					DatabaseNames: []string{"database_one", "database_two", "database_three"},
				},
			},
		},
	}
	wantDatabaseNames := "database_one,database_two,database_three"

	t.Run(name, func(t *testing.T) {

		gotDatabaseNames := database.GetInstanceDatabaseNames()
		if gotDatabaseNames != wantDatabaseNames {
			t.Errorf("Database.GetInstanceDatabaseNames() gotDatabaseNames = %v, want %v", gotDatabaseNames, wantDatabaseNames)
		}
	})
}

// Tests the GetTimeZone() function retrieves TimeZone correctly:
func TestDatabase_GetTimeZone(t *testing.T) {

	name := "Contains TimeZone"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					TimeZone: "UTC",
				},
			},
		},
	}
	wantTimeZone := "UTC"

	t.Run(name, func(t *testing.T) {

		gotTimeZone := database.GetTimeZone()
		if gotTimeZone != wantTimeZone {
			t.Errorf("Database.GetInstanceTimeZone() gotTimeZone = %v, want %v", gotTimeZone, wantTimeZone)
		}
	})
}

// Tests the GetInstanceSize() function retrieves Size correctly:
func TestDatabase_GetInstanceSize(t *testing.T) {

	name := "Contains Size"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					Size: 10,
				},
			},
		},
	}
	wantSize := 10

	t.Run(name, func(t *testing.T) {

		gotSize := database.GetInstanceSize()
		if gotSize != wantSize {
			t.Errorf("Database.GetInstanceSize() gotSize= %v, want %v", gotSize, wantSize)
		}
	})
}

// Tests the GetInstanceIsHighAvailability() function retrieves Size correctly:
func TestDatabase_GetInstanceIsHighAvailability(t *testing.T) {

	name := "Contains IsHighAvailability"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					IsHighAvailability: true,
				},
			},
		},
	}
	wantIsHighAvailability := true

	t.Run(name, func(t *testing.T) {

		gotIsHighAvailability := database.GetInstanceIsHighAvailability()
		if gotIsHighAvailability != wantIsHighAvailability {
			t.Errorf("Database.GetInstanceIsHighAvailability() gotIsHighAvailability= %v, want %v", gotIsHighAvailability, wantIsHighAvailability)
		}
	})
}

// Tests the GetClusterId() function retrieves ClusterId correctly:
func TestDatabase_GetClusterId(t *testing.T) {

	name := "Contains ClusterId"
	database := Database{
		Database: v1alpha1.Database{
			Spec: v1alpha1.DatabaseSpec{
				Instance: &v1alpha1.Instance{
					ClusterId: "test-cluster-id",
				},
			},
		},
	}
	wantClusterId := "test-cluster-id"

	t.Run(name, func(t *testing.T) {
		gotClusterId := database.GetClusterId()
		if gotClusterId != wantClusterId {
			t.Errorf("Database.GetClusterId() gotClusterId= %v, want %v", gotClusterId, wantClusterId)
		}
	})
}

// Tests the GetTMScheduleForInstance() function against the following:
// 1. All inputs are valid, no error is returned
// 2. DailySnapshotTime has incorrect values for hour, returns an error
// 3. DailySnapshotTime has incorrect values for minutes, returns an error
// 4. DailySnapshotTime has incorrect values for seconds, returns an error
// 5. DailySnapshotTime has incorrect values (all), returns an error
// 6. DailySnapshotTime has incorrect format, returns an error
func TestDatabase_GetTMScheduleForInstance(t *testing.T) {

	tests := []struct {
		name         string
		database     Database
		wantSchedule ndb_api.Schedule
		wantErr      bool
	}{
		{
			name: "All inputs are valid, no error is returned",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "tm-name", Description: "tm-description", SLAName: "sla-name", DailySnapshotTime: "12:34:56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantSchedule: ndb_api.Schedule{
				SnapshotTimeOfDay:  ndb_api.SnapshotTimeOfDay{Hours: 12, Minutes: 34, Seconds: 56},
				ContinuousSchedule: ndb_api.ContinuousSchedule{Enabled: true, LogBackupInterval: 30, SnapshotsPerDay: 1},
				WeeklySchedule:     ndb_api.WeeklySchedule{Enabled: true, DayOfWeek: "FRIDAY"},
				MonthlySchedule:    ndb_api.MonthlySchedule{Enabled: true, DayOfMonth: 15},
				QuarterlySchedule:  ndb_api.QuarterlySchedule{Enabled: true, StartMonth: "JANUARY", DayOfMonth: 15},
				YearlySchedule:     ndb_api.YearlySchedule{Enabled: false, DayOfMonth: 31, Month: "DECEMBER"},
			},
			wantErr: false,
		},
		{
			name: "DailySnapshotTime has incorrect values for hour, returns an error",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "tm-name", Description: "tm-description", SLAName: "sla-name", DailySnapshotTime: "xy-34-56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantSchedule: ndb_api.Schedule{},
			wantErr:      true,
		},
		{
			name: "DailySnapshotTime has incorrect values for minutes, returns an error",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "tm-name", Description: "tm-description", SLAName: "sla-name", DailySnapshotTime: "12:xy:56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantSchedule: ndb_api.Schedule{},
			wantErr:      true,
		},
		{
			name: "DailySnapshotTime has incorrect values for seconds, returns an error",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "tm-name", Description: "tm-description", SLAName: "sla-name", DailySnapshotTime: "12:34:xy", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantSchedule: ndb_api.Schedule{},
			wantErr:      true,
		},
		{
			name: "DailySnapshotTime has incorrect values (all), returns an error",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "tm-name", Description: "tm-description", SLAName: "sla-name", DailySnapshotTime: "a:b:c", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantSchedule: ndb_api.Schedule{},
			wantErr:      true,
		},
		{
			name: "DailySnapshotTime has incorrect format, returns an error",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "tm-name", Description: "tm-description", SLAName: "sla-name", DailySnapshotTime: "1:2", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantSchedule: ndb_api.Schedule{},
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotSchedule, err := tt.database.GetTMScheduleForInstance()

			if tt.wantErr {
				assert.Error(t, err)
			}
			if !reflect.DeepEqual(gotSchedule, tt.wantSchedule) {
				t.Errorf("Database.GetTMScheduleForInstance() = %v, want %v", gotSchedule, tt.wantSchedule)
			}
		})
	}
}

// Tests the GetInstanceTMDetails() function against the following test cases:
// 1. TM name, description and sla name are empty, returns default values
// 2. TM name is non empty, returns default values for other empty fields
// 3. TM description is non empty, returns default values for other empty fields
// 4. SLA name is non empty, returns default values for other empty fields
func TestDatabase_GetInstanceTMDetails(t *testing.T) {

	tests := []struct {
		name              string
		database          Database
		wantTmName        string
		wantTmDescription string
		wantSlaName       string
	}{
		{
			name: "TM name, description and sla name are empty, returns default values",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Name:   "test-database",
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "", Description: "", SLAName: "", DailySnapshotTime: "12:34:56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantTmName:        "test-database_TM",
			wantTmDescription: "Time Machine for test-database",
			wantSlaName:       "NONE",
		},
		{
			name: "TM name is non empty, returns default values for other empty fields",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Name:   "test-database",
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "test-name", Description: "", SLAName: "", DailySnapshotTime: "12:34:56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantTmName:        "test-name",
			wantTmDescription: "Time Machine for test-database",
			wantSlaName:       "NONE",
		},
		{
			name: "TM description is non empty, returns default values for other empty fields",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Name:   "test-database",
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "", Description: "test-description", SLAName: "", DailySnapshotTime: "12:34:56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantTmName:        "test-database_TM",
			wantTmDescription: "test-description",
			wantSlaName:       "NONE",
		},
		{
			name: "SLA name is non empty, returns default values for other empty fields",
			database: Database{
				Database: v1alpha1.Database{
					Spec: v1alpha1.DatabaseSpec{
						Instance: &v1alpha1.Instance{
							Name:   "test-database",
							TMInfo: &v1alpha1.DBTimeMachineInfo{Name: "", Description: "", SLAName: "test-sla", DailySnapshotTime: "12:34:56", SnapshotsPerDay: 1, LogCatchUpFrequency: 30, WeeklySnapshotDay: "FRIDAY", MonthlySnapshotDay: 15, QuarterlySnapshotMonth: "Jan"},
						},
					},
				},
			},
			wantTmName:        "test-database_TM",
			wantTmDescription: "Time Machine for test-database",
			wantSlaName:       "test-sla",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotTmName, gotTmDescription, gotSlaName := tt.database.GetInstanceTMDetails()
			if gotTmName != tt.wantTmName {
				t.Errorf("Database.GetInstanceTMDetails() gotTmName = %v, want %v", gotTmName, tt.wantTmName)
			}
			if gotTmDescription != tt.wantTmDescription {
				t.Errorf("Database.GetInstanceTMDetails() gotTmDescription = %v, want %v", gotTmDescription, tt.wantTmDescription)
			}
			if gotSlaName != tt.wantSlaName {
				t.Errorf("Database.GetInstanceTMDetails() gotSlaName = %v, want %v", gotSlaName, tt.wantSlaName)
			}
		})
	}
}
