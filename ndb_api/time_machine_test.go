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
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

func TestCreateSnapshotForTM(t *testing.T) {
	type args struct {
		ctx                context.Context
		ndbClient          ndb_client.NDBClientHTTPInterface
		tmId               string
		snapshotName       string
		expiryDateTimezone string
		expireInDays       string
	}

	tmId := "1"
	snapshotName := "mySnapshot"
	expiryDateTimezone := "UTC"
	expireInDays := "7"
	takeSnapshotPath := fmt.Sprintf("tms/%s/snapshots", tmId)

	snapshotRequest := GenerateSnapshotRequest(snapshotName, expiryDateTimezone, expireInDays)

	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodPost, takeSnapshotPath, snapshotRequest).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"name":"test-name", "entityId":"test-id"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodPost, takeSnapshotPath, snapshotRequest).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name     string
		args     args
		wantTask *TaskInfoSummaryResponse
		wantErr  bool
	}{
		{
			name: "Test 1: TestCreateSnapshotForTM returns an error when a request with empty id is passed to it",
			args: args{
				ctx:                context.TODO(),
				ndbClient:          mockNDBClient,
				tmId:               "",
				snapshotName:       snapshotName,
				expiryDateTimezone: expiryDateTimezone,
				expireInDays:       expireInDays,
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 2: TestCreateSnapshotForTM returns an error when sendRequest returns an error",
			args: args{
				ctx:                context.TODO(),
				ndbClient:          mockNDBClient,
				tmId:               tmId,
				snapshotName:       snapshotName,
				expiryDateTimezone: expiryDateTimezone,
				expireInDays:       expireInDays,
			},
			wantTask: nil,
			wantErr:  true,
		},
		{
			name: "Test 3: TestCreateSnapshotForTM returns a TaskInfoSummary response when sendRequest returns a response without error",
			args: args{
				ctx:                context.TODO(),
				ndbClient:          mockNDBClient,
				tmId:               tmId,
				snapshotName:       snapshotName,
				expiryDateTimezone: expiryDateTimezone,
				expireInDays:       expireInDays,
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
			gotTaskInfoSummaryResponse, err := CreateSnapshotForTM(tt.args.ctx, tt.args.ndbClient, tt.args.tmId, tt.args.snapshotName, tt.args.expiryDateTimezone, tt.args.expireInDays)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestCreateSnapshotForTM) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTaskInfoSummaryResponse, tt.wantTask) {
				t.Errorf("TestCreateSnapshotForTM() = %v, want %v", gotTaskInfoSummaryResponse, tt.wantTask)
			}
		})
	}
}

func TestGetSnapshotsForTM(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		tmId      string
	}

	tmId := "1"
	getSnapshotsPath := fmt.Sprintf("tms/%s/snapshots", tmId);

	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodGet, getSnapshotsPath, nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"snapshotsPerNxCluster":null}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, getSnapshotsPath, nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name                   string
		args                   args
		wantTMSnapshotResponse *TimeMachineGetSnapshotsResponse
		wantErr                bool
	}{
		{
			name: "Test 1: TestGetSnapshotsForTM returns an error when a request with empty id is passed to it",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				tmId:      "",
			},
			wantTMSnapshotResponse: nil,
			wantErr:                true,
		},
		{
			name: "Test 2: TestGetSnapshotsForTM returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				tmId:      tmId,
			},
			wantTMSnapshotResponse: nil,
			wantErr:                true,
		},
		{
			name: "Test 3: TestGetSnapshotsForTM returns a TimeMachineGetSnapshotsResponse response when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				tmId:      tmId,
			},
			wantTMSnapshotResponse: &TimeMachineGetSnapshotsResponse{
				SnapshotsPerNxCluster: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTMSnapshotResponse, err := GetSnapshotsForTM(tt.args.ctx, tt.args.ndbClient, tt.args.tmId)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetSnapshotsForTM) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTMSnapshotResponse, tt.wantTMSnapshotResponse) {
				t.Errorf("TestGetSnapshotsForTM() = %v, want %v", gotTMSnapshotResponse, tt.wantTMSnapshotResponse)
			}
		})
	}
}

func TestGetTimeMachineById(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
		tmId      string
	}

	tmId := "1"
	getTmDetailedPath := fmt.Sprintf("tms/%s", tmId);

	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}

	mockNDBClient.On("NewRequest", http.MethodGet, getTmDetailedPath, nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"test-id", "name":"test-name"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, getTmDetailedPath, nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name                   string
		args                   args
		wantTMResponse *TimeMachineResponse
		wantErr                bool
	}{
		{
			name: "Test 1: TestGetTimeMachineById returns an error when a request with empty id is passed to it",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				tmId:      "",
			},
			wantTMResponse: nil,
			wantErr:                true,
		},
		{
			name: "Test 2: TestGetTimeMachineById returns an error when sendRequest returns an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				tmId:      tmId,
			},
			wantTMResponse: nil,
			wantErr:                true,
		},
		{
			name: "Test 3: TestGetTimeMachineById returns a TimeMachineResponse response when sendRequest returns a response without error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
				tmId:      tmId,
			},
			wantTMResponse: &TimeMachineResponse{
				Id: "test-id",
				Name: "test-name",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTMResponse, err := GetTimeMachineById(tt.args.ctx, tt.args.ndbClient, tt.args.tmId)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetTimeMachineById) error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTMResponse, tt.wantTMResponse) {
				t.Errorf("TestGetTimeMachineById() = %v, want %v", gotTMResponse, tt.wantTMResponse)
			}
		})
	}
}