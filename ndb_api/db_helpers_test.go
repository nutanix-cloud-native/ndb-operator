package ndb_api

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// Test constants
const (
	TEST_PASSWORD      = "testPassword"
	TEST_SSHKEY        = "testSSHKey"
	TEST_DB_NAMES      = "testDB"
	TEST_INSTANCE_TYPE = "testInstance"
	TEST_TIMEZONE      = "test-timezone"
	TEST_CLUSTER_ID    = "test-cluster-id"
	TEST_INSTANCE_SIZE = 100
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

// Tests the GetRequestAppenderByType() function for different database types
func TestGetRequestAppenderByType(t *testing.T) {

	// test data map
	tests := []struct {
		databaseType       string
		isHighAvailability bool
		expected           interface{}
	}{
		{databaseType: common.DATABASE_TYPE_POSTGRES,
			isHighAvailability: false,
			expected:           &PostgresRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_POSTGRES,
			isHighAvailability: true,
			expected:           &PostgresHARequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MYSQL,
			isHighAvailability: false,
			expected:           &MySqlRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MSSQL,
			isHighAvailability: false,
			expected:           &MSSQLRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MONGODB,
			isHighAvailability: false,
			expected:           &MongoDbRequestAppender{},
		},
		{databaseType: "test",
			isHighAvailability: false,
			expected:           nil,
		},
	}

	for _, tc := range tests {
		got, _ := GetRequestAppender(tc.databaseType, tc.isHighAvailability)
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
	}
}

// Tests PostgresProvisionRequestAppender(), without additional arguments, positive workflow
func TestPostgresProvisionRequestAppender_withoutAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "listener_port",
			Value: "5432",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "false",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: TEST_DB_NAMES,
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)
	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Checks if expected and retrieved action arguments are equal
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}

	// Checks requestAppender.appendProvisioningRequest return type has no error and resultRequest.ActionArguments correctly configured
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests PostgresProvisionRequestAppender(), with additional arguments, positive workflow
func TestPostgresProvisionRequestAppender_withAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"listener_port": "0000",
	})
	mockDatabase.On("IsClone").Return(false)

	expectedActionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "0000",
		},
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "false",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: TEST_DB_NAMES,
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests PostgresProvisionRequestAppender(), with additional arguments, negative workflow
func TestPostgresProvisionRequestAppender_withAdditionalArguments_negativeWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"invalid-key": "invalid-value",
	})
	mockDatabase.On("IsClone").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Checks if error was returned
	if err == nil {
		t.Errorf("Should have errored. Expected: Setting configured action arguments failed! invalid-key is not an allowed additional argument, Got: %v", err)
	}
	// Checks if resultRequestIsNil
	if resultRequest != nil {
		t.Errorf("Should have errored. Expected: resultRequest to be nil, Got: %v", resultRequest)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests PostgresHAProvisionRequestAppender(), without additional arguments, positive workflow
func TestPostgresHAProvisionRequestAppender_withoutAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetName").Return("TestPostgresHADB")
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("GetClusterId").Return(TEST_CLUSTER_ID)
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "listener_port",
			Value: "5432",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "true",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: TEST_DB_NAMES,
		},
		{
			Name:  "provision_virtual_ip",
			Value: "true",
		},
		{
			Name:  "deploy_haproxy",
			Value: "true",
		},
		{
			Name:  "failover_mode",
			Value: "Automatic",
		},
		{
			Name:  "node_type",
			Value: "database",
		},
		{
			Name:  "allocate_pg_hugepage",
			Value: "false",
		},
		{
			Name:  "cluster_database",
			Value: "false",
		},
		{
			Name:  "archive_wal_expire_days",
			Value: "-1",
		},
		{
			Name:  "enable_peer_auth",
			Value: "false",
		},
		{
			Name:  "cluster_name",
			Value: "psqlcluster",
		},
		{
			Name:  "patroni_cluster_name",
			Value: "patroni",
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES, true)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)
	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Checks if expected and retrieved action arguments are equal
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}

	// Checks requestAppender.appendProvisioningRequest return type has no error and resultRequest.ActionArguments correctly configured
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Test PostgresHAProvisionRequestAppender(), with additional arguments, positive workflow
func TestPostgresHAProvisionRequestAppender_withAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetName").Return("TestPostgresHADB")
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"listener_port": "0000",
	})
	mockDatabase.On("GetClusterId").Return(TEST_CLUSTER_ID)
	mockDatabase.On("IsClone").Return(false)

	expectedActionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "0000",
		},
		{
			Name:  "proxy_read_port",
			Value: "5001",
		},
		{
			Name:  "proxy_write_port",
			Value: "5000",
		},
		{
			Name:  "enable_synchronous_mode",
			Value: "true",
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: TEST_DB_NAMES,
		},
		{
			Name:  "provision_virtual_ip",
			Value: "true",
		},
		{
			Name:  "deploy_haproxy",
			Value: "true",
		},
		{
			Name:  "failover_mode",
			Value: "Automatic",
		},
		{
			Name:  "node_type",
			Value: "database",
		},
		{
			Name:  "allocate_pg_hugepage",
			Value: "false",
		},
		{
			Name:  "cluster_database",
			Value: "false",
		},
		{
			Name:  "archive_wal_expire_days",
			Value: "-1",
		},
		{
			Name:  "enable_peer_auth",
			Value: "false",
		},
		{
			Name:  "cluster_name",
			Value: "psqlcluster",
		},
		{
			Name:  "patroni_cluster_name",
			Value: "patroni",
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES, true)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Test PostgresHAProvisionRequestAppender(), with additional arguments, negative workflow
func TestPostgresHAProvisionRequestAppender_withoutAdditionalArguments_negativeWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetName").Return("TestPostgresHADB")
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"invalid-key": "invalid-value",
	})
	mockDatabase.On("GetClusterId").Return(TEST_CLUSTER_ID)
	mockDatabase.On("IsClone").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES, true)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Checks if error was returned
	if err == nil {
		t.Errorf("Should have errored. Expected: Setting configured action arguments failed! invalid-key is not an allowed additional argument, Got: %v", err)
	}
	// Checks if resultRequestIsNil
	if resultRequest != nil {
		t.Errorf("Should have errored. Expected: resultRequest to be nil, Got: %v", resultRequest)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MSSQLProvisionRequestAppender(), without additional arguments, positive workflow
func TestMSSQLProvisionRequestAppender_withoutAdditionalArguments_positiveWorklow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	profileResponse := ProfileResponse{
		Id:              "123",
		Name:            "Test Profile",
		Type:            "Test Type",
		EngineType:      "Sample Engine",
		LatestVersionId: "456",
		Topology:        "Test Topology",
		SystemProfile:   true,
		Status:          "Active",
	}
	profileMap := map[string]ProfileResponse{
		common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: profileResponse,
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD: TEST_PASSWORD,
		common.PROFILE_MAP_PARAM:  profileMap}
	adminPassword := reqData[common.NDB_PARAM_PASSWORD].(string)

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetName").Return("testInstance")
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MSSQL)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "working_dir",
			Value: "C:\\temp",
		},
		{
			Name:  "sql_user_name",
			Value: "sa",
		},
		{
			Name:  "authentication_mode",
			Value: "windows",
		},
		{
			Name:  "delete_vm_on_failure",
			Value: "false",
		},
		{
			Name:  "is_gmsa_sql_service_account",
			Value: "false",
		},
		{
			Name:  "provision_from_backup",
			Value: "false",
		},
		{
			Name:  "distribute_database_data",
			Value: "true",
		},
		{
			Name:  "retain_database_in_restoring_mode",
			Value: "false",
		},
		{
			Name:  "dbserver_name",
			Value: mockDatabase.GetName(),
		},
		{
			Name:  "server_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "database_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "dbParameterProfileIdInstance",
			Value: profileResponse.Id,
		},
		{
			Name:  "vm_dbserver_admin_password",
			Value: adminPassword,
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MSSQL, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.DatabaseName != mockDatabase.GetInstanceDatabaseNames() {
		t.Errorf("Unexpected Database Name. Expected: %s, Got: %s", mockDatabase.GetInstanceDatabaseNames(), resultRequest.DatabaseName)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MSSQLProvisionRequestAppender(), with additional arguments, positive workflow
func TestMSSQLProvisionRequestAppender_withAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	profileResponse := ProfileResponse{
		Id:              "123",
		Name:            "Test Profile",
		Type:            "Test Type",
		EngineType:      "Sample Engine",
		LatestVersionId: "456",
		Topology:        "Test Topology",
		SystemProfile:   true,
		Status:          "Active",
	}
	profileMap := map[string]ProfileResponse{
		common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: profileResponse,
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD: TEST_PASSWORD,
		common.PROFILE_MAP_PARAM:  profileMap}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetName").Return("testInstance")
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MSSQL)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"sql_user_name":             "admin",
		"sql_user_password":         TEST_PASSWORD,
		"authentication_mode":       "mixed",
		"windows_domain_profile_id": "<windows-domain-profile-id>",
		"vm_db_server_user":         "<vm-db-server-user>",
	})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "sql_user_name",
			Value: "admin",
		},
		{
			Name:  "sql_user_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "authentication_mode",
			Value: "mixed",
		},
		{
			Name:  "windows_domain_profile_id",
			Value: "<windows-domain-profile-id>",
		},
		{
			Name:  "vm_db_server_user",
			Value: "<vm-db-server-user>",
		},
		{
			Name:  "working_dir",
			Value: "C:\\temp",
		},
		{
			Name:  "delete_vm_on_failure",
			Value: "false",
		},
		{
			Name:  "is_gmsa_sql_service_account",
			Value: "false",
		},
		{
			Name:  "provision_from_backup",
			Value: "false",
		},
		{
			Name:  "distribute_database_data",
			Value: "true",
		},
		{
			Name:  "retain_database_in_restoring_mode",
			Value: "false",
		},
		{
			Name:  "dbserver_name",
			Value: mockDatabase.GetName(),
		},
		{
			Name:  "server_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "database_collation",
			Value: "SQL_Latin1_General_CP1_CI_AS",
		},
		{
			Name:  "dbParameterProfileIdInstance",
			Value: profileResponse.Id,
		},
		{
			Name:  "vm_dbserver_admin_password",
			Value: TEST_PASSWORD,
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MSSQL, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.DatabaseName != mockDatabase.GetInstanceDatabaseNames() {
		t.Errorf("Unexpected Database Name. Expected: %s, Got: %s", mockDatabase.GetInstanceDatabaseNames(), resultRequest.DatabaseName)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")

}

// Tests MSSQLProvisionRequestAppender(), with additionalArguments, negative workflow
func TestMSSQLProvisionRequestAppender_withAdditionalArguments_negativeWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	profileResponse := ProfileResponse{
		Id:              "123",
		Name:            "Test Profile",
		Type:            "Test Type",
		EngineType:      "Sample Engine",
		LatestVersionId: "456",
		Topology:        "Test Topology",
		SystemProfile:   true,
		Status:          "Active",
	}
	profileMap := map[string]ProfileResponse{
		common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: profileResponse,
	}

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD: TEST_PASSWORD,
		common.PROFILE_MAP_PARAM:  profileMap}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetName").Return("testInstance")
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MSSQL)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"invalid-key":  "invalid-value",
		"invalid-key2": "invalid-value",
	})
	mockDatabase.On("IsClone").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MSSQL, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Checks if error was returned
	if err == nil {
		t.Errorf("Should have errored. Expected: Setting configured action arguments failed! invalid-key is not an allowed additional argument, Got: %v", err)
	}
	// Checks if resultRequestIsNil
	if resultRequest != nil {
		t.Errorf("Should have errored. Expected: resultRequest to be nil, Got: %v", resultRequest)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")

}

// Tests MongoDbProvisionRequestAppender(), without additionalArguments, positive workflow
func TestMongoDbProvisionRequestAppender_withoutAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MONGODB)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "27017",
		},
		{
			Name:  "log_size",
			Value: "100",
		},
		{
			Name:  "journal_size",
			Value: "100",
		},
		{
			Name:  "restart_mongod",
			Value: "true",
		},
		{
			Name:  "working_dir",
			Value: "/tmp",
		},
		{
			Name:  "db_user",
			Value: "admin",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: mockDatabase.GetInstanceDatabaseNames(),
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MONGODB, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MongoDbProvisionRequestAppender(), with additionalArguments, positive workflow
func TestMongoDbProvisionRequestAppender_withAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MONGODB)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"listener_port": "1111",
		"log_size":      "1",
		"journal_size":  "1",
	})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "1111",
		},
		{
			Name:  "log_size",
			Value: "1",
		},
		{
			Name:  "journal_size",
			Value: "1",
		},
		{
			Name:  "restart_mongod",
			Value: "true",
		},
		{
			Name:  "working_dir",
			Value: "/tmp",
		},
		{
			Name:  "db_user",
			Value: "admin",
		},
		{
			Name:  "backup_policy",
			Value: "primary_only",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: mockDatabase.GetInstanceDatabaseNames(),
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MONGODB, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MongoDbProvisionRequestAppender(), with additionalArguments, negative workflow
func TestMongoDbProvisionRequestAppender_withAdditionalArguments_negativeWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MONGODB)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"invalid-key": "invalid-value",
	})
	mockDatabase.On("IsClone").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MONGODB, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Checks if error was returned
	if err == nil {
		t.Errorf("Should have errored. Expected: Setting configured action arguments failed! invalid-key is not an allowed additional argument, Got: %v", err)
	}
	// Checks if resultRequestIsNil
	if resultRequest != nil {
		t.Errorf("Should have errored. Expected: resultRequest to be nil, Got: %v", resultRequest)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MySqlProvisionRequestAppender(), without additional arguments, positive workflow
func TestMySqlProvisionRequestAppender_withoutAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MYSQL)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "3306",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: mockDatabase.GetInstanceDatabaseNames(),
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MYSQL, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MySqlProvisionRequestAppender(), with additional arguments, positive workflow
func TestMySqlProvisionRequestAppender_withAdditionalArguments_positiveWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MYSQL)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"listener_port": "1111",
	})
	mockDatabase.On("IsClone").Return(false)
	expectedActionArgs := []ActionArgument{
		{
			Name:  "listener_port",
			Value: "1111",
		},
		{
			Name:  "db_password",
			Value: TEST_PASSWORD,
		},
		{
			Name:  "database_names",
			Value: mockDatabase.GetInstanceDatabaseNames(),
		},
		{
			Name:  "auto_tune_staging_drive",
			Value: "true",
		},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MYSQL, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort expected and retrieved action arguments
	sortWantAndGotActionArgsByName(expectedActionArgs, resultRequest.ActionArguments)

	// Checks if no error was returned
	if err != nil {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", nil, err)
	}
	// Check if the lengths of expected and retrieved action arguments are equal
	if !reflect.DeepEqual(expectedActionArgs, resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments. Expected: %v, Got: %v", expectedActionArgs, resultRequest.ActionArguments)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")
}

// Tests MySqlProvisionRequestAppender(), with additional arguments, negative workflow
func TestMySqlProvisionRequestAppender_withAdditionalArguments_negativeWorkflow(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_MYSQL)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
		"invalid-key": "invalid-value",
	})
	mockDatabase.On("IsClone").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MYSQL, false)

	// Call function being tested
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	// Checks if error was returned
	if err == nil {
		t.Errorf("Should have errored. Expected: Setting configured action arguments failed! invalid-key is not an allowed additional argument, Got: %v", err)
	}
	// Checks if resultRequestIsNil
	if resultRequest != nil {
		t.Errorf("Should have errored. Expected: resultRequest to be nil, Got: %v", resultRequest)
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetInstanceDatabaseNames")

}

// Test the error scenarios in GenerateProvisioningRequest function with different TM details
// 1. SLA is found, but error while getting/generating the TM schedule
// 2. SLA not found, no error in getting the TM schedule
// 3. SLA not found and error in getting the TM schedule
func TestGenerateProvisioningRequest_WithoutValidTMDetails_ReturnsError(t *testing.T) {

	// Set
	tests := []struct {
		slaName       string
		tmSchedule    Schedule
		tmScheduleErr error
		expectedError error
	}{
		// SLA is found, but error while getting/generating the TM schedule
		{
			slaName:       "SLA 1",
			tmSchedule:    Schedule{},
			tmScheduleErr: errors.New("err_xyz"),
			expectedError: errors.New("err_xyz"),
		},
		// SLA not found, no error in getting the TM schedule.
		{
			slaName:       "SLA-NOT-FOUND",
			tmSchedule:    Schedule{},
			tmScheduleErr: nil,
			expectedError: errors.New("SLA SLA-NOT-FOUND not found"),
		},
		// SLA not found and error in getting the TM schedule
		{
			slaName:       "SLA-NOT-FOUND",
			tmSchedule:    Schedule{},
			tmScheduleErr: errors.New("err_xyz"),
			expectedError: errors.New("SLA SLA-NOT-FOUND not found"),
		},
	}

	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)
	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
	}

	for _, tc := range tests {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("db_instance_name")
		mockDatabase.On("GetInstanceType").Return("db_instance_type")
		mockDatabase.On("GetInstanceTMDetails").Return("tm_name", "rm_description", tc.slaName)
		mockDatabase.On("GetTMScheduleForInstance").Return(tc.tmSchedule, tc.tmScheduleErr)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})

		// Test
		_, err := GenerateProvisioningRequest(context.Background(), ndb_client, mockDatabase, reqData)

		// Assert
		if err != tc.expectedError && err.Error() != tc.expectedError.Error() {
			t.Fatalf("expected: %v, got: %v", tc.expectedError, err)
		}
	}
}

// Test the error scenarios in GenerateProvisioningRequest function with different ProfileResolver errors
// 1. Software Profile returns an error
// 2. Compute Profile returns an error
// 3. Network Profile returns an error
// 4. DBParam Profile returns an error
// 5. DBParamInstance Profile returns an error
// Test cases are self explanatory.
func TestGenerateProvisioningRequest(t *testing.T) {

	// Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	reqData := map[string]interface{}{
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
	}

	getResolver := func(p ProfileResponse, e error) *MockProfileResolverInterface {
		profileResolver := MockProfileResolverInterface{}
		profileResolver.On("GetId").Return(p.Id)
		profileResolver.On("GetName").Return(p.Name)
		profileResolver.On("Resolve").Return(p, e)
		return &profileResolver
	}
	softwareError := errors.New("test-error-software")
	computeError := errors.New("test-error-compute")
	networkError := errors.New("test-error-network")
	dbParamError := errors.New("test-error-dbParam")
	dbParamInstanceError := errors.New("test-error-dbParamInstance")

	tests := []struct {
		databaseType         string
		softwareError        error
		computeError         error
		networkError         error
		dbParamError         error
		dbParamInstanceError error
		expectedError        error
	}{
		{
			softwareError: softwareError,
			expectedError: softwareError,
		},
		{
			computeError:  computeError,
			expectedError: computeError,
		},
		{
			networkError:  networkError,
			expectedError: networkError,
		},
		{
			dbParamError:  dbParamError,
			expectedError: dbParamError,
		},
		{
			databaseType:         common.DATABASE_TYPE_MSSQL,
			dbParamInstanceError: dbParamInstanceError,
			expectedError:        dbParamInstanceError,
		},
	}

	for _, tc := range tests {

		software := getResolver(ProfileResponse{}, tc.softwareError)
		compute := getResolver(ProfileResponse{}, tc.computeError)
		network := getResolver(ProfileResponse{}, tc.networkError)
		dbParam := getResolver(ProfileResponse{}, tc.dbParamError)
		dbParamInstance := getResolver(ProfileResponse{}, tc.dbParamInstanceError)

		instanceType := tc.databaseType
		// We're explicitly setting values on software profile (mock) because
		// MSSQL (and other closed source engines) have a special check in
		// ResolveProfiles function that looks for empty id and name in software profile.
		if instanceType == common.DATABASE_TYPE_MSSQL {
			software = &MockProfileResolverInterface{}
			software.On("GetName").Return("test-mssql-software-profile-name")
			software.On("GetId").Return("test-mssql-software-profile-id")
			software.On("Resolve").Return(ProfileResponse{
				Id:              "test-mssql-software-profile-id",
				Name:            "test-mssql-software-profile-name",
				Type:            common.PROFILE_TYPE_SOFTWARE,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-mssql",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   false,
			}, nil)
		}

		profileResolvers := ProfileResolvers{
			common.PROFILE_TYPE_SOFTWARE:                    software,
			common.PROFILE_TYPE_COMPUTE:                     compute,
			common.PROFILE_TYPE_NETWORK:                     network,
			common.PROFILE_TYPE_DATABASE_PARAMETER:          dbParam,
			common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: dbParamInstance,
		}

		mockDatabase := MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("db_instance_name")
		mockDatabase.On("GetInstanceType").Return(instanceType)
		mockDatabase.On("GetInstanceTMDetails").Return("tm_name", "rm_description", "SLA 1")
		mockDatabase.On("GetTMScheduleForInstance").Return(Schedule{}, nil)
		mockDatabase.On("GetProfileResolvers").Return(profileResolvers)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})

		// Test
		_, err := GenerateProvisioningRequest(context.Background(), ndb_client, &mockDatabase, reqData)

		// Assert
		if err != tc.expectedError && err.Error() != tc.expectedError.Error() {
			t.Fatalf("expected: %v, got: %v", tc.expectedError, err)
		}
	}
}

// Test the error scenarios in GenerateProvisioningRequest function for different parameters:
// 1. ReqData with empty db password for any database
// 2. ReqData with with empty ssh key for Non-MSSQL database
// 3. ReqData with with empty ssh key MSSQL database
// 4. Invalid instance type
func TestGenerateProvisioningRequest_AgainstDifferentReqData(t *testing.T) {

	// Set
	server := GetServerTestHelper(t)
	defer server.Close()
	ndb_client := ndb_client.NewNDBClient("username", "password", server.URL, "", true)

	getResolver := func(p ProfileResponse, e error) *MockProfileResolverInterface {
		profileResolver := MockProfileResolverInterface{}
		profileResolver.On("GetId").Return(p.Id)
		profileResolver.On("GetName").Return(p.Name)
		profileResolver.On("Resolve").Return(p, e)
		return &profileResolver
	}

	tests := []struct {
		databaseType  string
		reqData       map[string]interface{}
		expectedError error
	}{
		{
			// Database with empty db password
			databaseType: common.DATABASE_TYPE_POSTGRES,
			reqData: map[string]interface{}{
				common.NDB_PARAM_PASSWORD:       "",
				common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
			},
			expectedError: errors.New("invalid database password"),
		},
		{
			//  Non-MSSQL database with empty ssh key
			databaseType: common.DATABASE_TYPE_MYSQL,
			reqData: map[string]interface{}{
				common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
				common.NDB_PARAM_SSH_PUBLIC_KEY: "",
			},
			expectedError: errors.New("invalid ssh public key"),
		},
		{ // MSSQL database with empty ssh key
			databaseType: common.DATABASE_TYPE_MSSQL,
			reqData: map[string]interface{}{
				common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
				common.NDB_PARAM_SSH_PUBLIC_KEY: "",
			},
			expectedError: nil,
		},
		{ // Invalid database type
			databaseType: TEST_INSTANCE_TYPE,
			reqData: map[string]interface{}{
				common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
				common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
			},
			expectedError: errors.New("invalid database type: supported values: mssql, mysql, postgres, mongodb"),
		},
	}

	for _, tc := range tests {

		software := getResolver(ProfileResponse{}, nil)
		compute := getResolver(ProfileResponse{}, nil)
		network := getResolver(ProfileResponse{}, nil)
		dbParam := getResolver(ProfileResponse{}, nil)
		dbParamInstance := getResolver(ProfileResponse{}, nil)

		instanceType := tc.databaseType
		if instanceType == common.DATABASE_TYPE_MSSQL {
			software = &MockProfileResolverInterface{}
			software.On("GetName").Return("test-mssql-software-profile-name")
			software.On("GetId").Return("test-mssql-software-profile-id")
			software.On("Resolve").Return(ProfileResponse{
				Id:              "test-mssql-software-profile-id",
				Name:            "test-mssql-software-profile-name",
				Type:            common.PROFILE_TYPE_SOFTWARE,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-mssql",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   false,
			}, nil)

			dbParamInstance = &MockProfileResolverInterface{}
			dbParamInstance.On("GetName").Return("test-mssql-dbParamInstance-profile-name")
			dbParamInstance.On("GetId").Return("test-mssql-dbParamInstance-profile-id")
			dbParamInstance.On("Resolve").Return(ProfileResponse{
				Id:              "test-mssql-dbParamInstance-profile-id",
				Name:            "test-mssql-dbParamInstance-profile-name",
				Type:            common.PROFILE_TYPE_DATABASE_PARAMETER,
				EngineType:      common.DATABASE_ENGINE_TYPE_MSSQL,
				LatestVersionId: "v-id-mssql",
				Topology:        common.TOPOLOGY_SINGLE,
				Status:          "READY",
				SystemProfile:   false,
			}, nil)
		}

		profileResolvers := ProfileResolvers{
			common.PROFILE_TYPE_SOFTWARE:                    software,
			common.PROFILE_TYPE_COMPUTE:                     compute,
			common.PROFILE_TYPE_NETWORK:                     network,
			common.PROFILE_TYPE_DATABASE_PARAMETER:          dbParam,
			common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE: dbParamInstance,
		}

		mockDatabase := MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("db_instance_name")
		mockDatabase.On("GetDescription").Return("db_instance_description")
		mockDatabase.On("GetInstanceType").Return(instanceType)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
		mockDatabase.On("GetInstanceTMDetails").Return("tm_name", "rm_description", "SLA 1")
		mockDatabase.On("GetTMScheduleForInstance").Return(Schedule{}, nil)
		mockDatabase.On("GetProfileResolvers").Return(profileResolvers)
		mockDatabase.On("GetTimeZone").Return(TEST_TIMEZONE)
		mockDatabase.On("GetClusterId").Return(TEST_CLUSTER_ID)
		mockDatabase.On("GetInstanceSize").Return(TEST_INSTANCE_SIZE)
		mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
		mockDatabase.On("GetInstanceIsHighAvailability").Return(false)
		mockDatabase.On("IsClone").Return(false)

		// Test
		_, err := GenerateProvisioningRequest(context.Background(), ndb_client, &mockDatabase, tc.reqData)

		// Assert
		if err != tc.expectedError && err.Error() != tc.expectedError.Error() {
			t.Fatalf("expected: %v, got: %v", tc.expectedError, err)
		}
	}
}

// Sorts want and got action args by name
func sortWantAndGotActionArgsByName(wantActionArgs, gotActionArgs []ActionArgument) {
	sort.Slice(wantActionArgs, func(i, j int) bool {
		return wantActionArgs[i].Name < wantActionArgs[j].Name
	})
	sort.Slice(gotActionArgs, func(i, j int) bool {
		return gotActionArgs[i].Name < gotActionArgs[j].Name
	})
}
