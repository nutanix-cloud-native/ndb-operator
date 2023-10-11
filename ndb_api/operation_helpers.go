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

// Returns a boolean indicating if an operation has reached a terminal state
func HasOperationFailed(o OperationResponse) bool {
	status := o.Status
	return status == "2" || status == "3" || status == "4"
}

// Returns a boolean indicating if an operation was successful
func HasOperationPassed(o OperationResponse) bool {
	return o.Status == "5"
}
