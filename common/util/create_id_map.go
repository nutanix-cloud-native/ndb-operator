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

import (
	"fmt"
	"reflect"
)

// Returns a map from a slice of objects of type 'T'
// Uses the key provided in the arguments
// Returns an error if the desired key is not a field in any object of the slice
func CreateMapForKey[T any](objects []T, key string) (m map[string]T, err error) {
	m = make(map[string]T)
	for i, obj := range objects {
		keyField := reflect.ValueOf(obj).FieldByName(key)
		if !keyField.IsValid() {
			err = fmt.Errorf("%s field not found in object at index %d", key, i)
			return
		}

		keyValue := fmt.Sprintf("%v", keyField.Interface())
		m[keyValue] = obj
	}

	return
}
