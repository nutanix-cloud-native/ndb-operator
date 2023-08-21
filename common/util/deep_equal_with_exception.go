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

package util

import "reflect"

// Returns if two structs/objects are equal with the exception of
// the field under the exceptionKey
func DeepEqualWithException(a, b interface{}, exceptionKey string) bool {
	v1 := reflect.ValueOf(a)
	v2 := reflect.ValueOf(b)
	// Check if a and b are of the same type
	if v1.Type() != v2.Type() {
		return false
	}
	// Iterate over struct fields
	for i := 0; i < v1.NumField(); i++ {
		fieldName := v1.Type().Field(i).Name
		// Exclude the specified exception key
		if fieldName == exceptionKey {
			continue
		}
		// Use reflect.DeepEqual to compare field values
		if !reflect.DeepEqual(v1.Field(i).Interface(), v2.Field(i).Interface()) {
			return false
		}
	}
	return true
}
