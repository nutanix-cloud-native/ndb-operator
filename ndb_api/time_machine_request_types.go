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

type DatabaseCloneRequest struct {
	Name                       string           `json:"name"`
	Description                string           `json:"description"`
	CreateDbServer             bool             `json:"createDbserver"`
	Clustered                  bool             `json:"clustered"`
	NxClusterId                string           `json:"nxClusterId"`
	SSHPublicKey               string           `json:"sshPublicKey,omitempty"`
	DbServerId                 string           `json:"dbserverId,omitempty"`
	DbServerClusterId          string           `json:"dbserverClusterId,omitempty"`
	DbserverLogicalClusterId   string           `json:"dbserverLogicalClusterId,omitempty"`
	TimeMachineId              string           `json:"timeMachineId"`
	SnapshotId                 string           `json:"snapshotId,omitempty"`
	UserPitrTimestamp          string           `json:"userPitrTimestamp,omitempty"`
	TimeZone                   string           `json:"timeZone"`
	LatestSnapshot             bool             `json:"latestSnapshot"`
	NodeCount                  int              `json:"nodeCount"`
	Nodes                      []Node           `json:"nodes"`
	ActionArguments            []ActionArgument `json:"actionArguments"`
	Tags                       interface{}      `json:"tags"`
	LcmConfig                  *LcmConfig       `json:"lcmConfig,omitempty"`
	VmPassword                 string           `json:"vmPassword"`
	ComputeProfileId           string           `json:"computeProfileId"`
	NetworkProfileId           string           `json:"networkProfileId"`
	DatabaseParameterProfileId string           `json:"databaseParameterProfileId"`
}

type LcmConfig struct {
	DatabaseLCMConfig DatabaseLCMConfig `json:"databaseLCMConfig,omitempty"`
}

type DatabaseLCMConfig struct {
	ExpiryDetails  ExpiryDetails  `json:"expiryDetails,omitempty"`
	RefreshDetails RefreshDetails `json:"refreshDetails,omitempty"`
}

type ExpiryDetails struct {
	ExpireInDays       string `json:"expireInDays"`
	ExpiryDateTimezone string `json:"expiryDateTimezone"`
	DeleteDatabase     string `json:"deleteDatabase,omitempty"`
}

type RefreshDetails struct {
	RefreshInDays       string `json:"refreshInDays"`
	RefreshTime         string `json:"refreshTime"`
	RefreshDateTimezone string `json:"refreshDateTimezone"`
}

type SnapshotRequest struct {
	Name              string            `json:"name"`
	SnapshotLcmConfig SnapshotLcmConfig `json:"lcmConfig"`
}

type SnapshotLcmConfig struct {
	SnapshotLCMConfigDetailed SnapshotLcmConfigDetailed `json:"snapshotLCMConfig"`
}

type SnapshotLcmConfigDetailed struct {
	ExpiryDetails ExpiryDetails `json:"expiryDetails"`
}
