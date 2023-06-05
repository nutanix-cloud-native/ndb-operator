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

type DatabaseProvisionRequest struct {
	DatabaseType             string           `json:"databaseType"`
	Name                     string           `json:"name"`
	DatabaseDescription      string           `json:"databaseDescription"`
	SoftwareProfileId        string           `json:"softwareProfileId" validate:"uuid"`
	SoftwareProfileVersionId string           `json:"softwareProfileVersionId" validate:"uuid"`
	ComputeProfileId         string           `json:"computeProfileId" validate:"uuid"`
	NetworkProfileId         string           `json:"networkProfileId" validate:"uuid"`
	DbParameterProfileId     string           `json:"dbParameterProfileId" validate:"uuid"`
	NewDbServerTimeZone      string           `json:"newDbServerTimeZone"`
	CreateDbServer           bool             `json:"createDbserver"`
	NodeCount                int              `json:"nodeCount"`
	NxClusterId              string           `json:"nxClusterId"`
	SSHPublicKey             string           `json:"sshPublicKey"`
	Clustered                bool             `json:"clustered"`
	AutoTuneStagingDrive     bool             `json:"autoTuneStagingDrive"`
	TimeMachineInfo          TimeMachineInfo  `json:"timeMachineInfo"`
	ActionArguments          []ActionArgument `json:"actionArguments"`
	Nodes                    []Node           `json:"nodes"`
}

type DatabaseDeprovisionRequest struct {
	Delete               bool `json:"delete"`
	Remove               bool `json:"remove"`
	SoftRemove           bool `json:"softRemove"`
	Forced               bool `json:"forced"`
	DeleteTimeMachine    bool `json:"deleteTimeMachine"`
	DeleteLogicalCluster bool `json:"deleteLogicalCluster"`
}

type DatabaseRestoreRequest struct {
	SnapshotId string `json:"snapshotId" validate:"required,uuid"`
}
