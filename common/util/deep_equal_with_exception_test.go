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

// Tests the following :
// 1. Returns true when both objects are same
// 2. Returns true when both objects are same with different exceptionKey field
// 3. Returns false when both objects are differet
// 4. Returns false when both objects are of different type
func TestDeepEqualWithException(t *testing.T) {
	type randomType struct {
		Foo int
		Bar string
		Baz []float32
	}
	type anotherRandomType struct {
		Foo int
		Bar string
		Baz []float32
	}
	type args struct {
		a            interface{}
		b            interface{}
		exceptionKey string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Returns true when both objects are same",
			args: args{
				a:            randomType{1, "xyz", []float32{1.1, 2.2, 3.3}},
				b:            randomType{1, "xyz", []float32{1.1, 2.2, 3.3}},
				exceptionKey: "Foo",
			},
			want: true,
		},
		{
			name: "Returns true when both objects are same with different exceptionKey field",
			args: args{
				a:            randomType{1, "xyz", []float32{1.1, 2.2, 3.3}},
				b:            randomType{2, "xyz", []float32{1.1, 2.2, 3.3}},
				exceptionKey: "Foo",
			},
			want: true,
		},
		{
			name: "Returns false when both objects are differet",
			args: args{
				a:            randomType{1, "abc", nil},
				b:            randomType{1, "xyz", []float32{1.1, 2.2, 3.3}},
				exceptionKey: "Foo",
			},
			want: false,
		},
		{
			name: "Returns false when both objects are of different type",
			args: args{
				a:            randomType{1, "abc", []float32{1.1, 2.2, 3.3}},
				b:            anotherRandomType{1, "abc", []float32{1.1, 2.2, 3.3}},
				exceptionKey: "Foo",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeepEqualWithException(tt.args.a, tt.args.b, tt.args.exceptionKey); got != tt.want {
				t.Errorf("DeepEqualWithException() = %v, want %v", got, tt.want)
			}
		})
	}
}
