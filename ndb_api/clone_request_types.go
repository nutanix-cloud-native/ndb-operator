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

type CloneDeprovisionRequest struct {
	SoftRemove           bool `json:"softRemove"`
	Remove               bool `json:"remove"`
	Delete               bool `json:"delete"`
	Forced               bool `json:"forced"`
	DeleteDataDrives     bool `json:"deleteDataDrives"`
	DeleteLogicalCluster bool `json:"deleteLogicalCluster"`
	RemoveLogicalCluster bool `json:"removeLogicalCluster"`
	DeleteTimeMachine    bool `json:"deleteTimeMachine"`
}
