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

import "testing"

func TestToString(t *testing.T) {

	tests := []struct {
		name string
		args interface{}
		want string
	}{
		{
			name: "Test 1: Convert int to string",
			args: 42,
			want: "42",
		},
		{
			name: "Test 2: Convert string to string",
			args: "hello",
			want: "\"hello\"",
		},
		{
			name: "Test 3: Convert struct to string",
			args: struct{ Name string }{Name: "TONY STARK"},
			want: `{"Name":"TONY STARK"}`,
		},
		{
			name: "Test 4: Convert slice to string",
			args: []int{1, 2, 3},
			want: "[1,2,3]",
		},
		{
			name: "Test 5: Convert map to string",
			args: map[string]interface{}{"key": "value", "number": 8},
			want: `{"key":"value","number":8}`,
		},
		{
			name: "Test 6: Convert nil to string",
			args: nil,
			want: "null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.args); got != tt.want {
				t.Errorf("ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
