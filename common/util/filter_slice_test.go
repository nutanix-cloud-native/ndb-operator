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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	type args struct {
		items []int
		fn    func(item int) bool
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "Test 1: Filter returns a result slice based on the provided function",
			args: args{
				items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				fn:    func(item int) bool { return item%2 == 0 },
			},
			want: []int{2, 4, 6, 8, 10},
		},
		{
			name: "Test 2: Filter returns an empty slice if no elements filter through the provided function",
			args: args{
				items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				fn:    func(item int) bool { return item > 9999 },
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.args.items, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindFirstInt(t *testing.T) {
	tests := []struct {
		name           string
		input          []int
		filterFunction func(item int) bool
		expectedOutput int
		expectedError  error
	}{
		{
			name:           "Test 1: Find the first even number",
			input:          []int{1, 3, 5, 2, 4, 6},
			filterFunction: func(item int) bool { return item%2 == 0 },
			expectedOutput: 2,
			expectedError:  nil,
		},
		{
			name:           "Test 2: Find the first number greater than 10",
			input:          []int{5, 8, 12, 15},
			filterFunction: func(item int) bool { return item > 10 },
			expectedOutput: 12,
			expectedError:  nil,
		},
		{
			name:           "Test 3: No match in the input slice",
			input:          []int{2, 4, 6, 8, 10},
			filterFunction: func(item int) bool { return item > 100 },
			expectedOutput: 0,
			expectedError:  errors.New("no element found matching the provided criteria"),
		},
		{
			name:           "Test 4: Empty input slice",
			input:          []int{},
			filterFunction: func(item int) bool { return item > 5 },
			expectedOutput: 0,
			expectedError:  errors.New("no element found matching the provided criteria"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FindFirst(tt.input, tt.filterFunction)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error(), "Unexpected error")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tt.expectedOutput, result, "FindFirst result is not as expected")
			}
		})
	}
}
