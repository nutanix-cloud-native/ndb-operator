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
	Id                        string                `json:"id"`
	Name                      string                `json:"name"`
	Description               string                `json:"description"`
	Clustered                 bool                  `json:"clustered"`
	Clone                     bool                  `json:"clone"`
	DatabaseId                string                `json:"databaseId"`
	LogDriveId                string                `json:"logDriveId"`
	Type                      string                `json:"type"`
	Status                    string                `json:"Status"`
	EAStatus                  string                `json:"eaStatus"`
	Scope                     string                `json:"scope"`
	SlaId                     string                `json:"slaId"`
	ScheduleId                string                `json:"scheduleId"`
	OwnerId                   string                `json:"ownerId"`
	DateCreated               string                `json:"dateCreated"`
	DateModified              string                `json:"dateModified"`
	Info                      interface{}           `json:"Info"`
	Metadata                  interface{}           `json:"metadata"`
	Properties                []TimeMachineProperty `json:"properties"`
	Tags                      []TimeMachineTags     `json:"tags"`
	LogDrive                  TimeMachineLogDrive   `json:"logDrive"`
	Sla                       TimeMachineSLA        `json:"sla"`
	Schedule                  TimeMachineSchedule   `json:"schedule"`
	Database                  TimeMachineDatabase   `json:"database"`
	Clones                    []interface{}         `json:"clones"`
	ZeroSla                   bool                  `json:"zeroSla"`
	SlaSet                    bool                  `json:"slaSet"`
	ContinuousRecoveryEnabled bool                  `json:"continuousRecoveryEnabled"`
	SnapshotableState         bool                  `json:"snapshotableState"`
}

type TimeMachineProperty struct {
	RefID       string `json:"ref_id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Secure      bool   `json:"secure"`
	Description string `json:"description"`
}

type TimeMachineTags struct {
	TagId      string `json:"tagId"`
	EntityId   string `json:"entityId"`
	EntityType string `json:"entityType"`
	Value      string `json:"value"`
	TagName    string `json:"tagName"`
}

type TimeMachineLogDrive struct {
	Id                 string              `json:"id"`
	Path               string              `json:"path"`
	VgUuid             string              `json:"vgUuid"`
	TimeMachineId      string              `json:"timeMachineId"`
	ClusterId          string              `json:"clusterId"`
	ProtectionDomainId string              `json:"protectionDomainId"`
	Status             string              `json:"status"`
	TotalSize          string              `json:"totalSize"`
	UsedSize           string              `json:"usedSize"`
	Info               interface{}         `json:"info"`
	DateCreated        string              `json:"dateCreated"`
	DateModified       string              `json:"dateModified"`
	OwnerId            string              `json:"ownerId"`
	Metadata           TimeMachineMetadata `json:"metadata"`
	LogDisks           TimeMachineLogDisks `json:"logDisks"`
	Message            string              `json:"message"`
}

type TimeMachineMetadata struct {
	SecureInfo     interface{}               `json:"secureInfo"`
	Info           interface{}               `json:"info"`
	DeregisterInfo TimeMachineDeregisterInfo `json:"deregisterInfo"`
}

type TimeMachineDeregisterInfo struct {
	Message    string   `json:"message"`
	Operations []string `json:"operations"`
}

type TimeMachineLogDisks struct {
	Id            string      `json:"id"`
	VdiskUuid     string      `json:"vdiskUuid"`
	EraLogDriveId string      `json:"eraLogDriveId"`
	Status        string      `json:"status"`
	TotalSize     int         `json:"totalSize"`
	UsedSize      int         `json:"usedSize"`
	Info          interface{} `json:"info"`
	DateCreated   string      `json:"dateCreated"`
	DateModified  string      `json:"dateModified"`
	OwnerId       string      `json:"ownerId"`
	Message       string      `json:"message"`
}

type TimeMachineSLA struct {
	Id                      string `json:"id"`
	Name                    string `json:"name"`
	UniqueName              string `json:"uniqueName"`
	Description             string `json:"description"`
	OwnerId                 string `json:"ownerId"`
	SystemSla               bool   `json:"systemSla"`
	DateCreated             string `json:"dateCreated"`
	DateModified            string `json:"dateModified"`
	ContinuousRetention     int    `json:"continuousRetention"`
	DailyRetention          int    `json:"dailyRetention"`
	WeeklyRetention         int    `json:"weeklyRetention"`
	MonthlyRetention        int    `json:"monthlyRetention"`
	QuarterlyRetention      int    `json:"quarterlyRetention"`
	YearlyRetention         int    `json:"yearlyRetention"`
	ReferenceCountRetention int    `json:"referenceCountRetention"`
}

type TimeMachineSchedule struct {
	Id                        string                        `json:"id"`
	Name                      string                        `json:"name"`
	UniqueName                string                        `json:"uniqueName"`
	Description               string                        `json:"description"`
	OwnerId                   string                        `json:"ownerId"`
	SystemPolicy              bool                          `json:"systemPolicy"`
	GlobalPolicy              bool                          `json:"globalPolicy"`
	DateCreated               string                        `json:"dateCreated"`
	DateModified              string                        `json:"dateModified"`
	SnapshotTimeOfDay         TimeMachineSnapshotTimeOfDay  `json:"snapshotTimeOfDay"`
	ContinuousSchedule        TimeMachineContinuousSchedule `json:"continuousSchedule"`
	DailySchedule             TimeMachineDailySchedule      `json:"dailySchedule"`
	WeeklySchedule            TimeMachineWeeklySchedule     `json:"weeklySchedule"`
	MonthlySchedule           TimeMachineMonthlySchedule    `json:"monthlySchedule"`
	YearlySchedule            TimeMachineYearlySchedule     `json:"yearlySchedule"`
	ReferenceCount            int                           `json:"referenceCount"`
	StartTime                 string                        `json:"startTime"`
	ContinuousScheduleEnabled bool                          `json:"continuousScheduleEnabled"`
	QuarterlySchedule         TimeMachineQuarterlySchedule  `json:"quarterlySchedule"`
}

type TimeMachineSnapshotTimeOfDay struct {
	Hours                   int  `json:"hours"`
	Minutes                 int  `json:"minutes"`
	Seconds                 int  `json:"seconds"`
	Extra                   bool `json:"extra"`
	ValidScheduleTime       bool `json:"validScheduleTime"`
	RepeatIntervalInMinutes bool `json:"repeatIntervalInMinutes"`
}

type TimeMachineContinuousSchedule struct {
	Enabled                          bool `json:"enabled"`
	LogBackupInterval                int  `json:"logBackupInterval"`
	SnapshotsPerDay                  int  `json:"snapshotsPerDay"`
	LogBackupRepeatIntervalInMinutes int  `json:"logBackupRepeatIntervalInMinutes"`
	SnapshotRepeatIntervalInMinutes  int  `json:"snapshotRepeatIntervalInMinutes"`
}

type TimeMachineDailySchedule struct {
	Enabled bool `json:"enabled"`
}

type TimeMachineWeeklySchedule struct {
	Enabled        bool   `json:"enabled"`
	DayOfWeek      string `json:"dayOfWeek"`
	DayOfWeekValue string `json:"dayOfWeekValue"`
}

type TimeMachineMonthlySchedule struct {
	Enabled    bool `json:"enabled"`
	DayOfMonth int  `json:"dayOfMonth"`
}

type TimeMachineYearlySchedule struct {
	Enabled    bool   `json:"enabled"`
	Month      string `json:"Month"`
	MonthValue string `json:"MonthValue"`
	DayOfMonth int    `json:"DayOfMonth"`
}

type TimeMachineQuarterlySchedule struct {
	Enabled         bool   `json:"enabled"`
	StartMonth      string `json:"startMonth"`
	StartMonthValue string `json:"startMonthValue"`
	DayOfMonth      string `json:"DayOfMonth"`
}

type TimeMachineDatabase struct {
	Id                       string                `json:"id"`
	Name                     string                `json:"name"`
	Description              string                `json:"description"`
	OwnerId                  string                `json:"ownerId"`
	DateCreated              string                `json:"dateCreated"`
	DateModified             string                `json:"dateModified"`
	Properties               []TimeMachineProperty `json:"properties"`
	Tags                     []TimeMachineTags     `json:"tags"`
	Clustered                bool                  `json:"clustered"`
	Clone                    bool                  `json:"clone"`
	EraCreated               bool                  `json:"eraCreated"`
	Placeholder              bool                  `json:"placeHolder"`
	DatabaseName             string                `json:"databaseName"`
	Type                     string                `json:"type"`
	DatabaseClusterType      string                `json:"databaseClusterType"`
	Status                   string                `json:"status"`
	DatabaseStatus           string                `json:"databaseStatus"`
	DbserverLogicalClusterId string                `json:"dbserverLogicalClusterId "`
	TimeMachineId            string                `json:"timeMachineId"`
	ParentTimeMachineId      string                `json:"ParentTimeMachineId"`
	TimeZone                 string                `json:"timeZone"`
	Info                     TimeMachineInfo       `json:"info"`
	Metadata                 TimeMachineMetadata   `json:"metadata"`
	LcmConfig                LcmConfig             `json:"lcmConfig"`
	DbserverlogicalCluster   interface{}           `json:"dbserverlogicalCluster"`
	DatabaseNodes            interface{}           `json:"databaseNodes"`
}
