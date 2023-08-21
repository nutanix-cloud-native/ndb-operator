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
	"reflect"
	"testing"
)

// Tests the following scenarios:
// 1. Creates a map with given field as key
// 2. Returns error when key is empty
// 3. Returns error when key does not exist
func TestCreateMapForKey(t *testing.T) {
	type randomType struct {
		Foo int
		Bar string
		Baz float32
	}
	type args struct {
		objects []randomType
		key     string
	}
	tests := []struct {
		name    string
		args    args
		wantM   map[string]randomType
		wantErr bool
	}{
		{
			name: "Creates a map with given field as key",
			args: args{
				key: "Bar",
				objects: []randomType{
					{1, "a", 1.1},
					{2, "b", 2.2},
				},
			},
			wantM: map[string]randomType{
				"a": {1, "a", 1.1},
				"b": {2, "b", 2.2},
			},
			wantErr: false,
		},
		{
			name: "Returns error when key is empty",
			args: args{
				key: "",
				objects: []randomType{
					{1, "a", 1.1},
					{2, "b", 2.2},
				},
			},
			wantM:   map[string]randomType{},
			wantErr: true,
		},
		{
			name: "Returns error when key does not exist",
			args: args{
				key: "qwertyuiop",
				objects: []randomType{
					{1, "a", 1.1},
					{2, "b", 2.2},
				},
			},
			wantM:   map[string]randomType{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotM, err := CreateMapForKey(tt.args.objects, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMapForKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("CreateMapForKey() = %v, want %v", gotM, tt.wantM)
			}
		})
	}
}
