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

type TaskInfoSummaryResponse struct {
	Name                 string                    `json:"name"`
	WorkId               string                    `json:"workId"`
	OperationId          string                    `json:"operationId"`
	DbServerId           string                    `json:"dbserverId"`
	Message              string                    `json:"messgae"`
	EntityId             string                    `json:"entityId"`
	EntityName           string                    `json:"entityName"`
	EntityType           string                    `json:"entityType"`
	Status               string                    `json:"status"`
	AssociatedOperations []TaskInfoSummaryResponse `json:"associatedOperations"`
	DependencyReport     interface{}               `json:"dependencyReport"`
}
