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

func TestGetAllDatabases(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockNDBClient.On("NewRequest", http.MethodGet, "databases?detailed=true", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(
			`[{"id":"1", "name":"database-1"},{"id":"2", "name":"database-2"}]`,
		)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "databases?detailed=true", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)
	tests := []struct {
		name          string
		args          args
		wantDatabases []DatabaseResponse
		wantErr       bool
	}{
		{
			name: "Test 1: GetAllDatabases returns an error when sendRequest returns an error",
			args: args{
				context.TODO(),
				mockNDBClient,
			},
			wantDatabases: nil,
			wantErr:       true,
		},
		{
			name: "Test 2: GetAllDatabases returns a slice of databases when sendRequest does not return an error",
			args: args{
				context.TODO(),
				mockNDBClient,
			},
			wantDatabases: []DatabaseResponse{
				{
					Id:   "1",
					Name: "database-1",
				},
				{
					Id:   "2",
					Name: "database-2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatabases, err := GetAllDatabases(tt.args.ctx, tt.args.ndbClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllDatabases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDatabases, tt.wantDatabases) {
				t.Errorf("GetAllDatabases() = %v, want %v", gotDatabases, tt.wantDatabases)
			}
		})
	}
}

func TestGetDatabaseById(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		id        string
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockNDBClient.On("NewRequest", http.MethodGet, "databases/databaseid?detailed=true", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"databaseid", "name":"database-1"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "databases/databaseid?detailed=true", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)
	tests := []struct {
		name         string
		args         args
		wantDatabase *DatabaseResponse
		wantErr      bool
	}{
		{
			name: "Test 1: GetDatabaseById returns an error when a request with empty id is passed to it",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "",
			},
			wantDatabase: nil,
			wantErr:      true,
		},
		{
			name: "Test 2: GetDatabaseById returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "databaseid",
			},
			wantDatabase: nil,
			wantErr:      true,
		},
		{
			name: "Test 3: GetDatabaseById returns a DatabaseResponse when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "databaseid",
			},
			wantDatabase: &DatabaseResponse{Id: "databaseid", Name: "database-1"},
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDatabase, err := GetDatabaseById(tt.args.ctx, tt.args.ndbClient, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabaseById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDatabase, tt.wantDatabase) {
				t.Errorf("GetDatabaseById() = %v, want %v", gotDatabase, tt.wantDatabase)
			}
		})
	}
}

func TestProvisionDatabase(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		req       *DatabaseProvisionRequest
	}
	tests := []struct {
		name     string
		args     args
		wantTask *TaskInfoSummaryResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTask, err := ProvisionDatabase(tt.args.ctx, tt.args.ndbClient, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisionDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("ProvisionDatabase() = %v, want %v", gotTask, tt.wantTask)
			}
		})
	}
}
