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

func TestAuthValidate(t *testing.T) {
	type args struct {
		ctx       context.Context
		ndbClient ndb_client.NDBClientHTTPInterface
	}
	// Mocks of the NDB Client interface
	mockNDBClient := &MockNDBClientHTTPInterface{}
	mockNDBClient.On("NewRequest", http.MethodGet, "auth/validate", nil).Once().Return(nil, errors.New("mock-error-new-request"))

	req := &http.Request{Method: http.MethodGet}
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"status":"TEST-STATUS", "message":"TEST-MESSAGE"}`)),
	}
	mockNDBClient.On("NewRequest", http.MethodGet, "auth/validate", nil).Once().Return(req, nil)
	mockNDBClient.On("Do", req).Once().Return(res, nil)

	tests := []struct {
		name                     string
		args                     args
		wantAuthValidateResponse AuthValidateResponse
		wantErr                  bool
	}{
		{
			name: "Test 1: AuthValidate returns an error when sendRequest",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
			},
			wantAuthValidateResponse: AuthValidateResponse{},
			wantErr:                  true,
		},
		{
			name: "Test 2: AuthValidate returns an auth response when sendRequest does not return an error",
			args: args{
				ctx:       context.TODO(),
				ndbClient: mockNDBClient,
			},
			wantAuthValidateResponse: AuthValidateResponse{
				Status:  "TEST-STATUS",
				Message: "TEST-MESSAGE",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuthValidateResponse, err := AuthValidate(tt.args.ctx, tt.args.ndbClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthValidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAuthValidateResponse, tt.wantAuthValidateResponse) {
				t.Errorf("AuthValidate() = %v, want %v", gotAuthValidateResponse, tt.wantAuthValidateResponse)
			}
		})
	}
}
