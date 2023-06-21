package ndb_api

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

// Test constants
const (
	TEST_PASSWORD = "testPassword"
	TEST_SSHKEY   = "testSSHKey"
	TEST_DB_NAMES = "testDB"
)

// Tests the validateReqData() function with different values of password and sshkey
func TestValidateReqData(t *testing.T) {
	//Set
	server := GetServerTestHelper(t)
	defer server.Close()

	type reqData map[string]interface{}
	errorInvalidPassword := errors.New("invalid database password")
	errorInvalidSSHKey := errors.New("invalid ssh public key")

	// test data map
	tests := []struct {
		databaseType string
		reqData      reqData
		expected     interface{}
	}{
		// No error
		{databaseType: common.DATABASE_TYPE_POSTGRES,
			reqData: reqData{
				common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
				common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
			},
			expected: nil,
		},
		{ //Throw error when password is empty
			databaseType: common.DATABASE_TYPE_POSTGRES,
			reqData: reqData{
				common.NDB_PARAM_PASSWORD:       "",
				common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
			},
			expected: errorInvalidPassword,
		},
		{ // Throw error if database is not MSSQL and SSHKey is empty
			databaseType: common.DATABASE_TYPE_POSTGRES,
			reqData: reqData{
				common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
				common.NDB_PARAM_SSH_PUBLIC_KEY: "",
			},
			expected: errorInvalidSSHKey,
		},
		{ // No error if datbase is MSSQL and SSHKey is empty
			databaseType: common.DATABASE_TYPE_MSSQL,
			reqData: reqData{
				common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
				common.NDB_PARAM_SSH_PUBLIC_KEY: "",
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		got := validateReqData(context.Background(), tc.databaseType, tc.reqData)
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
	}
}
