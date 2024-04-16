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
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	"github.com/stretchr/testify/assert"
)

func Test_sendRequest(t *testing.T) {
	type args struct {
		ctx          context.Context
		ndbClient    ndb_client.NDBClientHTTPInterface
		method       string
		endpoint     string
		payload      interface{}
		responseBody interface{}
	}

	// Mock NDB Client interface
	// Required to mock the http responses from NDB.
	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockNDBClient.On("NewRequest", "TEST-METHOD-3", "TEST-ENDPOINT-3", "TEST-PAYLOAD-3").Return(nil, errors.New("mock-error-new-request"))

	request4 := &http.Request{}
	mockNDBClient.On("NewRequest", "TEST-METHOD-4", "TEST-ENDPOINT-4", "TEST-PAYLOAD-4").Return(request4, nil)
	mockNDBClient.On("Do", request4).Return(nil, errors.New("mock-error-do"))

	request5 := &http.Request{Method: "TEST-METHOD-5"}
	mockNDBClient.On("NewRequest", "TEST-METHOD-5", "TEST-ENDPOINT-5", "TEST-PAYLOAD-5").Return(request5, nil)
	mockNDBClient.On("Do", request5).Return(nil, nil)

	request6 := &http.Request{Method: "TEST-METHOD-6"}
	response6 := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data": "TEST-RESPONSE-DATA"}`)),
	}
	mockNDBClient.On("NewRequest", "TEST-METHOD-6", "TEST-ENDPOINT-6", "TEST-PAYLOAD-6").Return(request6, nil)
	mockNDBClient.On("Do", request6).Return(response6, nil)

	request7 := &http.Request{Method: "TEST-METHOD-7"}
	response7 := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data": "ERROR-DATA"}`)),
	}
	mockNDBClient.On("NewRequest", "TEST-METHOD-7", "TEST-ENDPOINT-7", "TEST-PAYLOAD-7").Return(request7, nil)
	mockNDBClient.On("Do", request7).Return(response7, nil)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test 1: sendRequest returns an error when ndbClient is nil",
			args: args{
				ctx:          context.Background(),
				ndbClient:    nil,
				method:       "TEST-METHOD-1",
				endpoint:     "TEST-ENDPOINT-1",
				payload:      "TEST-PAYLOAD-1",
				responseBody: &struct{}{},
			},
			wantErr: true,
		},
		{
			name: "Test 2: sendRequest returns an error when responseBody is nil",
			args: args{
				ctx:          context.Background(),
				ndbClient:    mockNDBClient,
				method:       "TEST-METHOD-2",
				endpoint:     "TEST-ENDPOINT-2",
				payload:      "TEST-PAYLOAD-2",
				responseBody: nil,
			},
			wantErr: true,
		},
		{
			name: "Test 3: sendRequest returns an error when ndbClient.NewRequest returns an error",
			args: args{
				ctx:          context.Background(),
				ndbClient:    mockNDBClient,
				method:       "TEST-METHOD-3",
				endpoint:     "TEST-ENDPOINT-3",
				payload:      "TEST-PAYLOAD-3",
				responseBody: &struct{}{},
			},
			wantErr: true,
		},
		{
			name: "Test 4: sendRequest returns an error when ndbClient.Do returns an error",
			args: args{
				ctx:          context.Background(),
				ndbClient:    mockNDBClient,
				method:       "TEST-METHOD-4",
				endpoint:     "TEST-ENDPOINT-4",
				payload:      "TEST-PAYLOAD-4",
				responseBody: &struct{}{},
			},
			wantErr: true,
		},
		{
			name: "Test 5: sendRequest returns no error when ndbClient.Do returns nil response and no error",
			args: args{
				ctx:          context.Background(),
				ndbClient:    mockNDBClient,
				method:       "TEST-METHOD-5",
				endpoint:     "TEST-ENDPOINT-5",
				payload:      "TEST-PAYLOAD-5",
				responseBody: &struct{}{},
			},
			wantErr: false,
		},
		{
			name: "Test 6: sendRequest returns no error when ndbClient.Do returns non nil 200 response and no error",
			args: args{
				ctx:       context.Background(),
				ndbClient: mockNDBClient,
				method:    "TEST-METHOD-6",
				endpoint:  "TEST-ENDPOINT-6",
				payload:   "TEST-PAYLOAD-6",
				responseBody: &struct {
					Data string `json:"data"`
				}{},
			},
			wantErr: false,
		},
		{
			name: "Test 7: sendRequest returns an error when ndbClient.Do returns non nil 400 response and no error",
			args: args{
				ctx:       context.Background(),
				ndbClient: mockNDBClient,
				method:    "TEST-METHOD-7",
				endpoint:  "TEST-ENDPOINT-7",
				payload:   "TEST-PAYLOAD-7",
				responseBody: &struct {
					Data string `json:"data"`
				}{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := sendRequest(tt.args.ctx, tt.args.ndbClient, tt.args.method, tt.args.endpoint, tt.args.payload, tt.args.responseBody); (err != nil) != tt.wantErr {
				t.Errorf("sendRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDatabaseEngineName(t *testing.T) {
	// Test cases for GetDatabaseEngineName
	testCases := []struct {
		dbType         string
		expectedEngine string
	}{
		{common.DATABASE_TYPE_POSTGRES, common.DATABASE_ENGINE_TYPE_POSTGRES},
		{common.DATABASE_TYPE_MYSQL, common.DATABASE_ENGINE_TYPE_MYSQL},
		{common.DATABASE_TYPE_MONGODB, common.DATABASE_ENGINE_TYPE_MONGODB},
		{common.DATABASE_TYPE_MSSQL, common.DATABASE_ENGINE_TYPE_MSSQL},
		{"invalidType", ""},
	}

	for _, tc := range testCases {
		result := GetDatabaseEngineName(tc.dbType)
		assert.Equal(t, tc.expectedEngine, result)
	}
}

func TestGetDatabaseTypeFromEngine(t *testing.T) {
	// Test cases for GetDatabaseTypeFromEngine
	testCases := []struct {
		engine         string
		expectedDbType string
	}{
		{common.DATABASE_ENGINE_TYPE_POSTGRES, common.DATABASE_TYPE_POSTGRES},
		{common.DATABASE_ENGINE_TYPE_MYSQL, common.DATABASE_TYPE_MYSQL},
		{common.DATABASE_ENGINE_TYPE_MONGODB, common.DATABASE_TYPE_MONGODB},
		{common.DATABASE_ENGINE_TYPE_MSSQL, common.DATABASE_TYPE_MSSQL},
		{"invalidEngine", ""},
	}

	for _, tc := range testCases {
		result := GetDatabaseTypeFromEngine(tc.engine)
		assert.Equal(t, tc.expectedDbType, result)
	}
}

func TestGetDatabasePortByType(t *testing.T) {
	// Test cases for GetDatabasePortByType
	testCases := []struct {
		dbType       string
		expectedPort int32
	}{
		{common.DATABASE_TYPE_POSTGRES, common.DATABASE_DEFAULT_PORT_POSTGRES},
		{common.DATABASE_TYPE_MYSQL, common.DATABASE_DEFAULT_PORT_MYSQL},
		{common.DATABASE_TYPE_MONGODB, common.DATABASE_DEFAULT_PORT_MONGODB},
		{common.DATABASE_TYPE_MSSQL, common.DATABASE_DEFAULT_PORT_MSSQL},
		{"invalidType", -1},
	}

	for _, tc := range testCases {
		result := GetDatabasePortByType(tc.dbType)
		assert.Equal(t, tc.expectedPort, result)
	}
}

func TestGetRequestAppender(t *testing.T) {
	// Test cases for GetRequestAppender
	testCases := []struct {
		databaseType   string
		expectedResult bool
	}{
		{common.DATABASE_TYPE_POSTGRES, true},
		{common.DATABASE_TYPE_MYSQL, true},
		{common.DATABASE_TYPE_MONGODB, true},
		{common.DATABASE_TYPE_MSSQL, true},
		{"invalidType", false},
	}

	for _, tc := range testCases {
		result, err := GetRequestAppender(tc.databaseType, false)
		if tc.expectedResult {
			assert.NotNil(t, result)
			assert.NoError(t, err)
		} else {
			assert.Nil(t, result)
			assert.Error(t, err)
		}
	}
}
