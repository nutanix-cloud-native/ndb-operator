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
	Properties       []Property     `json:"properties"`
}

type DatabaseServer struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	IPAddresses []string `json:"ipAddresses"`
	NxClusterId string   `json:"nxClusterId"`
}

// DBServerDetails represents a DB server entry returned by GET /dbservers or GET /dbservers/{id}.
// DbserverClusterId is populated for HA instances (all nodes in a cluster share the same value).
type DBServerDetails struct {
	Id                string   `json:"id"`
	Name              string   `json:"name"`
	IPAddresses       []string `json:"ipAddresses"`
	DbserverClusterId string   `json:"dbserverClusterId"`
}

// PrimarySlaDetails holds the SLA and cluster scope for multi-cluster HA Time Machines.
type PrimarySlaDetails struct {
	SlaId        string   `json:"slaId"`
	NxClusterIds []string `json:"nxClusterIds"`
}

// SlaDetails wraps PrimarySlaDetails for the timeMachineInfo payload (used in HA provisioning).
type SlaDetails struct {
	PrimarySla PrimarySlaDetails `json:"primarySla"`
}

type TimeMachineInfo struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	SlaId            string      `json:"slaId,omitempty"`
	SlaDetails       *SlaDetails `json:"slaDetails,omitempty"`
	Schedule         Schedule    `json:"schedule"`
	Tags             []string    `json:"tags"`
	AutoTuneLogDrive bool        `json:"autoTuneLogDrive"`
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

// NodeProperty is used in outgoing provisioning request node entries.
// It is intentionally lightweight (no Description) to match the NDB API shape.
type NodeProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Node struct {
	VmName              string         `json:"vmName"`
	ComputeProfileId    string         `json:"computeProfileId,omitempty"`
	NetworkProfileId    string         `json:"networkProfileId,omitempty"`
	NewDbServerTimeZone string         `json:"newDbServerTimeZone,omitempty"`
	NxClusterId         string         `json:"nxClusterId,omitempty"`
	Properties          []NodeProperty `json:"properties"`
}

// Property is used in NDB API responses (includes Description).
type Property struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}
