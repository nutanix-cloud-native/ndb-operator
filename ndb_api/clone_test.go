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

func TestGetAllClones(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockNDBClient.On("NewRequest", http.MethodGet, "clones?detailed=true", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(bytes.NewBufferString(
			`[{"id":"1", "name":"clone-1"},{"id":"2", "name":"clone-2"}]`,
		)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "clones?detailed=true", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)
	tests := []struct {
		name       string
		args       args
		wantClones []DatabaseResponse
		wantErr    bool
	}{
		{
			name: "Test 1: GetAllClones returns an error when sendRequest returns an error",
			args: args{
				context.TODO(),
				mockNDBClient,
			},
			wantClones: nil,
			wantErr:    true,
		},
		{
			name: "Test 2: GetAllClones returns a slice of clones when sendRequest does not return an error",
			args: args{
				context.TODO(),
				mockNDBClient,
			},
			wantClones: []DatabaseResponse{
				{
					Id:   "1",
					Name: "clone-1",
				},
				{
					Id:   "2",
					Name: "clone-2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClones, err := GetAllClones(tt.args.ctx, tt.args.ndbClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllClones() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotClones, tt.wantClones) {
				t.Errorf("GetAllClones() = %v, want %v", gotClones, tt.wantClones)
			}
		})
	}
}

func TestProvisionClone(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		req       *DatabaseCloneRequest
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodPost, "tms/tmid/clones", &DatabaseCloneRequest{TimeMachineId: "tmid"}).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"name":"test-name", "entityId":"test-id"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodPost, "tms/tmid/clones", &DatabaseCloneRequest{TimeMachineId: "tmid"}).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)
	tests := []struct {
		name     string
		args     args
		wantTask *TaskInfoSummaryResponse
		wantErr  bool
	}{
		{
			name: "Test 1: ProvisionClone returns an error when a request with empty timeMachineId is passed to it",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				req: &DatabaseCloneRequest{
					TimeMachineId: "",
				},
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 2: ProvisionClone returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				req:       &DatabaseCloneRequest{TimeMachineId: "tmid"},
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 3: ProvisionClone returns a TaskInfoSummary response when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				req: &DatabaseCloneRequest{
					TimeMachineId: "tmid",
				},
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
			gotTask, err := ProvisionClone(tt.args.ctx, tt.args.ndbClient, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProvisionClone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("ProvisionClone() = %v, want %v", gotTask, tt.wantTask)
			}
		})
	}
}

func TestDeprovisionClone(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		id        string
		req       *CloneDeprovisionRequest
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodDelete, "clones/cloneid", GenerateDeprovisionCloneRequest()).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"name":"test-name", "entityId":"test-id"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodDelete, "clones/cloneid", GenerateDeprovisionCloneRequest()).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)
	tests := []struct {
		name     string
		args     args
		wantTask *TaskInfoSummaryResponse
		wantErr  bool
	}{
		{
			name: "Test 1: DeprovisionClone returns an error when a request with empty id is passed to it",
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
			name: "Test 2: DeprovisionClone returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "cloneid",
				req:       GenerateDeprovisionCloneRequest(),
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 3: DeprovisionClone returns a TaskInfoSummary response when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				id:        "cloneid",
				req:       GenerateDeprovisionCloneRequest(),
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
			gotTask, err := DeprovisionClone(tt.args.ctx, tt.args.ndbClient, tt.args.id, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeprovisionClone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTask, tt.wantTask) {
				t.Errorf("DeprovisionClone() = %v, want %v", gotTask, tt.wantTask)
			}
		})
	}
}
