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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestAppendLCMConfigDetailsToRequest(t *testing.T) {
	t.Run("nil additionalArguments is treated like empty", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		require.NoError(t, appendLCMConfigDetailsToRequest(req, nil))
		require.NotNil(t, req.LcmConfig)
		assert.Equal(t, ExpiryDetails{}, req.LcmConfig.DatabaseLCMConfig.ExpiryDetails)
		assert.Equal(t, RefreshDetails{}, req.LcmConfig.DatabaseLCMConfig.RefreshDetails)
	})

	t.Run("empty map only initializes LcmConfig", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		require.NoError(t, appendLCMConfigDetailsToRequest(req, map[string]string{}))
		require.NotNil(t, req.LcmConfig)
	})

	t.Run("all three expiry fields set ExpiryDetails", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		args := map[string]string{
			"expireInDays":       "7",
			"expiryDateTimezone": "UTC",
			"deleteDatabase":     "true",
		}
		require.NoError(t, appendLCMConfigDetailsToRequest(req, args))
		assert.Equal(t, "7", req.LcmConfig.DatabaseLCMConfig.ExpiryDetails.ExpireInDays)
		assert.Equal(t, "UTC", req.LcmConfig.DatabaseLCMConfig.ExpiryDetails.ExpiryDateTimezone)
		assert.Equal(t, "true", req.LcmConfig.DatabaseLCMConfig.ExpiryDetails.DeleteDatabase)
	})

	t.Run("partial expiry fields returns error", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		args := map[string]string{
			"expireInDays": "7",
		}
		err := appendLCMConfigDetailsToRequest(req, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expireInDays, expiryDateTimezone, and deleteDatabase")
		assert.Contains(t, err.Error(), "1/3")
	})

	t.Run("all three refresh fields set RefreshDetails", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		args := map[string]string{
			"refreshInDays":       "3",
			"refreshTime":         "02:00:00",
			"refreshDateTimezone": "UTC",
		}
		require.NoError(t, appendLCMConfigDetailsToRequest(req, args))
		assert.Equal(t, "3", req.LcmConfig.DatabaseLCMConfig.RefreshDetails.RefreshInDays)
		assert.Equal(t, "02:00:00", req.LcmConfig.DatabaseLCMConfig.RefreshDetails.RefreshTime)
		assert.Equal(t, "UTC", req.LcmConfig.DatabaseLCMConfig.RefreshDetails.RefreshDateTimezone)
	})

	t.Run("partial refresh fields returns error", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		args := map[string]string{
			"refreshInDays": "3",
			"refreshTime":   "02:00:00",
		}
		err := appendLCMConfigDetailsToRequest(req, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "refreshInDays, refreshTime, refreshDateTimezone")
		assert.Contains(t, err.Error(), "2/3")
	})

	t.Run("full expiry and full refresh both applied", func(t *testing.T) {
		req := &DatabaseCloneRequest{}
		args := map[string]string{
			"expireInDays":        "1",
			"expiryDateTimezone":  "UTC",
			"deleteDatabase":      "false",
			"refreshInDays":       "2",
			"refreshTime":         "01:00:00",
			"refreshDateTimezone": "America/New_York",
		}
		require.NoError(t, appendLCMConfigDetailsToRequest(req, args))
		assert.Equal(t, "1", req.LcmConfig.DatabaseLCMConfig.ExpiryDetails.ExpireInDays)
		assert.Equal(t, "2", req.LcmConfig.DatabaseLCMConfig.RefreshDetails.RefreshInDays)
		assert.Equal(t, "America/New_York", req.LcmConfig.DatabaseLCMConfig.RefreshDetails.RefreshDateTimezone)
	})
}

func TestOracleRequestAppender_appendCloningRequest(t *testing.T) {
	appender := &OracleRequestAppender{}
	
	t.Run("Oracle clone request has correct default action arguments", func(t *testing.T) {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("oraclone1")
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
		
		req := &DatabaseCloneRequest{}
		reqData := map[string]interface{}{
			"password":       "TestPassword123",
			"ssh_public_key": "ssh-rsa test-key",
		}
		
		result, err := appender.appendCloningRequest(req, mockDatabase, reqData)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "ssh-rsa test-key", result.SSHPublicKey)
		
		// Check that required Oracle clone action arguments are present
		actionArgsMap := make(map[string]string)
		for _, arg := range result.ActionArguments {
			actionArgsMap[arg.Name] = arg.Value
		}
		
		// Verify Oracle-specific clone parameters
		assert.Equal(t, "oraclone1", actionArgsMap["vm_name"])
		assert.Equal(t, "DB Server VM for oraclone1", actionArgsMap["dbserver_description"])
		assert.Equal(t, "TestPassword123", actionArgsMap["db_password"])
		assert.Equal(t, "oraclone1", actionArgsMap["new_db_sid"]) // Oracle clones use new_db_sid, not oracle_sid
		assert.Equal(t, "1521", actionArgsMap["listener_port"])
		assert.Equal(t, "false", actionArgsMap["enable_ha"])
		assert.Equal(t, "1521", actionArgsMap["scan_port"])
		assert.Equal(t, "false", actionArgsMap["delete_logs_post_recovery"])
		assert.Equal(t, "None", actionArgsMap["asm_driver"])
	})
	
	t.Run("Oracle clone request allows overriding defaults via additionalArguments", func(t *testing.T) {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("oraclone2")
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
			"listener_port": "1522",
			"enable_ha":     "true",
		})
		
		req := &DatabaseCloneRequest{}
		reqData := map[string]interface{}{
			"password":       "Password456",
			"ssh_public_key": "ssh-rsa another-key",
		}
		
		result, err := appender.appendCloningRequest(req, mockDatabase, reqData)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		
		actionArgsMap := make(map[string]string)
		for _, arg := range result.ActionArguments {
			actionArgsMap[arg.Name] = arg.Value
		}
		
		// Verify overridden values
		assert.Equal(t, "1522", actionArgsMap["listener_port"])
		assert.Equal(t, "true", actionArgsMap["enable_ha"])
		// Verify defaults remain
		assert.Equal(t, "oraclone2", actionArgsMap["new_db_sid"])
	})
	
	t.Run("Oracle clone with LCM config", func(t *testing.T) {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("oraclone3")
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
			"expireInDays":       "7",
			"expiryDateTimezone": "UTC",
			"deleteDatabase":     "true",
		})
		
		req := &DatabaseCloneRequest{}
		reqData := map[string]interface{}{
			"password":       "Pass789",
			"ssh_public_key": "ssh-rsa lcm-key",
		}
		
		result, err := appender.appendCloningRequest(req, mockDatabase, reqData)
		
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.LcmConfig)
		assert.Equal(t, "7", result.LcmConfig.DatabaseLCMConfig.ExpiryDetails.ExpireInDays)
		assert.Equal(t, "UTC", result.LcmConfig.DatabaseLCMConfig.ExpiryDetails.ExpiryDateTimezone)
		assert.Equal(t, "true", result.LcmConfig.DatabaseLCMConfig.ExpiryDetails.DeleteDatabase)
	})
}

