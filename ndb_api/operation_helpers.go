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

const OPERATION_STATUS_FAILED = "FAILED"
const OPERATION_STATUS_PASSED = "PASSED"

// Returns an operation status string
func GetOperationStatus(o OperationResponse) string {
	status := ""
	// Statuses on NDB
	// 2: STOPPED
	// 3: SUSPENDED
	// 4: FAILED
	// 5: PASSED
	switch o.Status {
	case "2", "3", "4":
		status = OPERATION_STATUS_FAILED
	case "5":
		status = OPERATION_STATUS_PASSED
	default:
		status = ""
	}
	return status
}
