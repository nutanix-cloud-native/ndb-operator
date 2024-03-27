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

package controller_adapters

import (
	"context"
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
)

func TestProfile_Resolve(t *testing.T) {
	type fields struct {
		Profile     v1alpha1.Profile
		ProfileType string
	}
	type args struct {
		ctx         context.Context
		allProfiles []ndb_api.ProfileResponse
		filter      func(p ndb_api.ProfileResponse) bool
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantProfile ndb_api.ProfileResponse
		wantErr     bool
	}{
		{
			name: "Test 1: Resolve profile by name",
			fields: fields{
				Profile:     v1alpha1.Profile{Name: "profile-name"},
				ProfileType: "test_type",
			},
			args: args{
				ctx: context.Background(),
				allProfiles: []ndb_api.ProfileResponse{
					{Type: "test_type", Name: "oob_profile"},
					{Type: "test_type", Name: "profile-name", Id: "profile-id"},
				},
				filter: func(p ndb_api.ProfileResponse) bool { return p.Name == "oob_profile" },
			},
			wantProfile: ndb_api.ProfileResponse{Type: "test_type", Name: "profile-name", Id: "profile-id"},
			wantErr:     false,
		},
		{
			name: "Test 2: Resolve profile by id",
			fields: fields{
				Profile:     v1alpha1.Profile{Id: "profile-id"},
				ProfileType: "test_type",
			},
			args: args{
				ctx: context.Background(),
				allProfiles: []ndb_api.ProfileResponse{
					{Type: "test_type", Name: "oob_profile"},
					{Type: "test_type", Name: "profile-name", Id: "profile-id"},
				},
				filter: func(p ndb_api.ProfileResponse) bool { return p.Name == "oob_profile" },
			},
			wantProfile: ndb_api.ProfileResponse{Type: "test_type", Name: "profile-name", Id: "profile-id"},
			wantErr:     false,
		},
		{
			name: "Test 3: Resolve profile by name and id both",
			fields: fields{
				Profile:     v1alpha1.Profile{Name: "profile-name", Id: "profile-id"},
				ProfileType: "test_type",
			},
			args: args{
				ctx: context.Background(),
				allProfiles: []ndb_api.ProfileResponse{
					{Type: "test_type", Name: "oob_profile"},
					{Type: "test_type", Name: "profile-name", Id: "profile-id"},
				},
				filter: func(p ndb_api.ProfileResponse) bool { return p.Name == "oob_profile" },
			},
			wantProfile: ndb_api.ProfileResponse{Type: "test_type", Name: "profile-name", Id: "profile-id"},
			wantErr:     false,
		},
		{
			name: "Test 4: Resolve oob profile",
			fields: fields{
				Profile:     v1alpha1.Profile{},
				ProfileType: "test_type",
			},
			args: args{
				ctx: context.Background(),
				allProfiles: []ndb_api.ProfileResponse{
					{Type: "test_type", Name: "oob_profile"},
					{Type: "test_type", Name: "profile-name", Id: "profile-id"},
				},
				filter: func(p ndb_api.ProfileResponse) bool { return p.Name == "oob_profile" },
			},
			wantProfile: ndb_api.ProfileResponse{Type: "test_type", Name: "oob_profile"},
			wantErr:     false,
		},
		{
			name: "Test 5: Return error if different profiles are resolved by name and id",
			fields: fields{
				Profile:     v1alpha1.Profile{Name: "profile-name-1", Id: "profile-id-2"},
				ProfileType: "test_type",
			},
			args: args{
				ctx: context.Background(),
				allProfiles: []ndb_api.ProfileResponse{
					{Type: "test_type", Name: "oob_profile"},
					{Type: "test_type", Name: "profile-name-1", Id: "profile-id-1"},
					{Type: "test_type", Name: "profile-name-2", Id: "profile-id-2"},
				},
				filter: func(p ndb_api.ProfileResponse) bool { return p.Name == "oob_profile" },
			},
			wantProfile: ndb_api.ProfileResponse{},
			wantErr:     true,
		},
		{
			name: "Test 6: Return error if no profile is resolved",
			fields: fields{
				Profile:     v1alpha1.Profile{Name: "profile-name-1", Id: "profile-id-2"},
				ProfileType: "test_type",
			},
			args: args{
				ctx: context.Background(),
				allProfiles: []ndb_api.ProfileResponse{
					{Type: "test_type", Name: "qwertyuiop"},
				},
				filter: func(p ndb_api.ProfileResponse) bool { return p.Name == "oob_profile" },
			},
			wantProfile: ndb_api.ProfileResponse{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputProfile := &Profile{
				Profile:     tt.fields.Profile,
				ProfileType: tt.fields.ProfileType,
			}
			gotProfile, err := inputProfile.Resolve(tt.args.ctx, tt.args.allProfiles, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Profile.Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProfile, tt.wantProfile) {
				t.Errorf("Profile.Resolve() = %v, want %v", gotProfile, tt.wantProfile)
			}
		})
	}
}
