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

type ProfileResponse struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	EngineType      string `json:"engineType"`
	LatestVersionId string `json:"latestVersionId"`
	Topology        string `json:"topology"`
	SystemProfile   bool   `json:"systemProfile"`
	Status          string `json:"status"`
}
