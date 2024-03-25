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

type DatabaseNode struct {
	Id               string         `json:"id"`
	Name             string         `json:"name"`
	DatabaseServerId string         `json:"dbServerId"`
	DbServer         DatabaseServer `json:"dbserver"`
}

type DatabaseServer struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	IPAddresses []string `json:"ipAddresses"`
	NxClusterId string   `json:"nxClusterId"`
}

type TimeMachineInfo struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	SlaId            string   `json:"slaId"`
	Schedule         Schedule `json:"schedule"`
	Tags             []string `json:"tags"`
	AutoTuneLogDrive bool     `json:"autoTuneLogDrive"`
}

type Schedule struct {
	SnapshotTimeOfDay  SnapshotTimeOfDay  `json:"snapshotTimeOfDay"`
	ContinuousSchedule ContinuousSchedule `json:"continuousSchedule"`
	WeeklySchedule     WeeklySchedule     `json:"weeklySchedule"`
	MonthlySchedule    MonthlySchedule    `json:"monthlySchedule"`
	QuarterlySchedule  QuarterlySchedule  `json:"quartelySchedule"`
	YearlySchedule     YearlySchedule     `json:"yearlySchedule"`
}

type SnapshotTimeOfDay struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type ContinuousSchedule struct {
	Enabled           bool `json:"enabled"`
	LogBackupInterval int  `json:"logBackupInterval"`
	SnapshotsPerDay   int  `json:"snapshotsPerDay"`
}

type WeeklySchedule struct {
	Enabled   bool   `json:"enabled"`
	DayOfWeek string `json:"dayOfWeek"`
}

type MonthlySchedule struct {
	Enabled    bool `json:"enabled"`
	DayOfMonth int  `json:"dayOfMonth"`
}

type QuarterlySchedule struct {
	Enabled    bool   `json:"enabled"`
	StartMonth string `json:"startMonth"`
	DayOfMonth int    `json:"dayOfMonth"`
}

type YearlySchedule struct {
	Enabled    bool   `json:"enabled"`
	DayOfMonth int    `json:"dayOfMonth"`
	Month      string `json:"month"`
}

type ActionArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Node struct {
	VmName              string              `json:"vmName"`
	ComputeProfileId    string              `json:"computeProfileId,omitempty"`
	NetworkProfileId    string              `json:"networkProfileId,omitempty"`
	NewDbServerTimeZone string              `json:"newDbServerTimeZone,omitempty"`
	NxClusterId         string              `json:"nxClusterId,omitempty"`
	Properties          []map[string]string `json:"properties"`
}

type Property struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}
