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

func TestDeprovisionDatabaseServer(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		id        string
		req       *DatabaseServerDeprovisionRequest
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodDelete, "dbservers/dbserverid", GenerateDeprovisionDatabaseServerRequest()).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"name":"test-name", "entityId":"test-id"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodDelete, "dbservers/dbserverid", GenerateDeprovisionDatabaseServerRequest()).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)
	tests := []struct {
		name     string
		args     args
		wantTask *TaskInfoSummaryResponse
		wantErr  bool
	}{
		{
			name: "Test 1: DeprovisionDatabaseServer returns an error when a request with empty id is passed to it",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "",
				req:       nil,
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 2: DeprovisionDatabaseServer returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "dbserverid",
				req:       GenerateDeprovisionDatabaseServerRequest(),
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 3: DeprovisionDatabaseServer returns a TaskInfoSummary response when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "dbserverid",
				req:       GenerateDeprovisionDatabaseServerRequest(),
			},
			wantTask: &TaskInfoSummaryResponse{
				Name:     "test-name",
				EntityId: "test-id",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, err := DeprovisionDatabaseServer(tt.args.ctx, tt.args.ndbClient, tt.args.id, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeprovisionDatabaseServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("DeprovisionDatabaseServer() = %v, want %v", gotTask, tt.wantTask)
			}
		})
	}
}
