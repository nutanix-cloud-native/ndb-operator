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

package ndb_api

type TimeMachineResponse struct {
	Id          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	DatabaseId  string              `json:"databaseId"`
	Status      string              `json:"Status"`
	Sla         TimeMachineSLA      `json:"sla"`
	Schedule    TimeMachineSchedule `json:"schedule"`
}

type TimeMachineSLA struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type TimeMachineSchedule struct {
	Id                 string                        `json:"id"`
	Name               string                        `json:"name"`
	SnapshotTimeOfDay  TimeMachineSnapshotTimeOfDay  `json:"snapshotTimeOfDay"`
	ContinuousSchedule TimeMachineContinuousSchedule `json:"continuousSchedule"`
	WeeklySchedule     TimeMachineWeeklySchedule     `json:"weeklySchedule"`
	MonthlySchedule    TimeMachineMonthlySchedule    `json:"monthlySchedule"`
	QuarterlySchedule  TimeMachineQuarterlySchedule  `json:"quarterlySchedule"`
}

type TimeMachineSnapshotTimeOfDay struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type TimeMachineContinuousSchedule struct {
	LogBackupInterval int `json:"logBackupInterval"`
	SnapshotsPerDay   int `json:"snapshotsPerDay"`
}

type TimeMachineWeeklySchedule struct {
	DayOfWeek string `json:"dayOfWeek"`
}

type TimeMachineMonthlySchedule struct {
	DayOfMonth int `json:"dayOfMonth"`
}

type TimeMachineQuarterlySchedule struct {
	StartMonth string `json:"startMonth"`
}
