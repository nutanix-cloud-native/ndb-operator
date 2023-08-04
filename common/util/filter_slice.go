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
	"errors"
)

// A utility function to filter the items based on the conditions
// defined in the predicate 'fn'
func Filter[T any](items []T, fn func(item T) bool) []T {
	filteredItems := []T{}
	for _, value := range items {
		if fn(value) {
			filteredItems = append(filteredItems, value)
		}
	}
	return filteredItems
}

// A utility function to find the first match for the given filter function
// in the case of no match, returns an empty struct instance and an error
func FindFirst[T any](items []T, filter func(item T) bool) (T, error) {
	for _, value := range items {
		if filter(value) {
			return value, nil
		}
	}

	// returning an empty instance of T in the case of no match
	var empty T
	return empty, errors.New("no element found matching the provided criteria")
}
