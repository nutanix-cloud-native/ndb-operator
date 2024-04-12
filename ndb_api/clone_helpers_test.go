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

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

func TestGenerateCloningRequest(t *testing.T) {
	type args struct {
		ctx        context.Context
		ndb_client ndb_client.NDBClientHTTPInterface
		database   DatabaseInterface
		reqData    map[string]interface{}
	}
	// Mocks of the interfaces
	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockDatabase := &MockDatabaseInterface{}

	// Common stubs for all the test cases
	mockDatabase.On("GetName").Return("test-clone-name")
	mockDatabase.On("IsClone").Return(true)
	// Stubs for Test 1
	mockDatabase.On("GetCloneSourceDBId").Return("").Once()

	// Stubs for Test 2
	mockDatabase.On("GetCloneSourceDBId").Return("test-sourcedb-id")
	reqGetDatabaseById := &http.Request{}
	resGetDatabaseById := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"test-sourcedb-id"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "databases/test-sourcedb-id?detailed=true", nil).Return(reqGetDatabaseById, nil)
	mockNDBClient.On("Do", reqGetDatabaseById).Return(resGetDatabaseById, nil)
	mockDatabase.On("GetProfileResolvers").Once().Return(ProfileResolvers{})
	mockNDBClient.On("NewRequest", http.MethodGet, "profiles", nil).Return(nil, errors.New("profiles-error")).Once()

	tests := []struct {
		name            string
		args            args
		wantRequestBody *DatabaseCloneRequest
		wantErr         bool
	}{
		{
			name: "Test 1: GenerateCloningRequest returns an error if source database is not found",
			args: args{
				ctx:        context.TODO(),
				ndb_client: mockNDBClient,
				database:   mockDatabase,
				reqData:    make(map[string]interface{}),
			},
			wantRequestBody: nil,
			wantErr:         true,
		},
		{
			name: "Test 2: GenerateCloningRequest returns an error when ResolveProfiles returns an error",
			args: args{
				ctx:        context.TODO(),
				ndb_client: mockNDBClient,
				database:   mockDatabase,
				reqData:    make(map[string]interface{}),
			},
			wantRequestBody: nil,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRequestBody, err := GenerateCloningRequest(tt.args.ctx, tt.args.ndb_client, tt.args.database, tt.args.reqData)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCloningRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRequestBody, tt.wantRequestBody) {
				t.Errorf("GenerateCloningRequest() = %v, want %v", gotRequestBody, tt.wantRequestBody)
			}
		})
	}
}
