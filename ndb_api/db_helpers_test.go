package ndb_api

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		databaseType string
		expected     interface{}
	}{
		{databaseType: common.DATABASE_TYPE_POSTGRES,
			expected: &PostgresRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MYSQL,
			expected: &MySqlRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MSSQL,
			expected: &MSSQLRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MONGODB,
			expected: &MongoDbRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_ORACLE,
			expected: &OracleRequestAppender{},
		},
		{databaseType: "test",
			expected: nil,
		},
	}

	for _, tc := range tests {
		got, _ := GetRequestAppender(tc.databaseType)
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
	mockDatabase.On("IsPostgresHA").Return(false)
	mockDatabase.On("GetInstanceHAConfig").Return((*HAConfig)(nil))
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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES)

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
	mockDatabase.On("IsPostgresHA").Return(false)
	mockDatabase.On("GetInstanceHAConfig").Return((*HAConfig)(nil))

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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES)

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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES)

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

// Tests PostgresProvisionRequestAppender for HA (multi-cluster) positive workflow.
// Verifies that HA-specific action arguments are appended when IsPostgresHA returns true.
func TestPostgresProvisionRequestAppender_HA_positiveWorkflow(t *testing.T) {
	baseRequest := &DatabaseProvisionRequest{}
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	haConfig := &HAConfig{
		PatroniClusterName:    "test-patroni-cluster",
		ClusterName:           "test-ha-cluster",
		EnableSynchronousMode: true,
		ProvisionVirtualIP:    true,
	}

	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("IsClone").Return(false)
	mockDatabase.On("IsPostgresHA").Return(true)
	mockDatabase.On("GetInstanceHAConfig").Return(haConfig)

	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES)
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Build a lookup map for the resulting action arguments
	argMap := make(map[string]string)
	for _, arg := range resultRequest.ActionArguments {
		argMap[arg.Name] = arg.Value
	}

	// Verify HA-specific arguments are present
	haExpected := map[string]string{
		"provision_virtual_ip":    "true",
		"deploy_haproxy":          "true",
		"failover_mode":           "Automatic",
		"enable_synchronous_mode": "true",
		"patroni_cluster_name":    "test-patroni-cluster",
		"cluster_name":            "test-ha-cluster",
		"cluster_database":        "false",
		"db_user":                 "postgres",
	}
	for k, v := range haExpected {
		if got, ok := argMap[k]; !ok {
			t.Errorf("Missing expected HA action argument: %s", k)
		} else if got != v {
			t.Errorf("Action argument %s: expected %s, got %s", k, v, got)
		}
	}

	// Verify SSH key was set
	if resultRequest.SSHPublicKey != TEST_SSHKEY {
		t.Errorf("Unexpected SSHPublicKey. Expected: %s, Got: %s", TEST_SSHKEY, resultRequest.SSHPublicKey)
	}
}

// Verifies that provision_virtual_ip is "false" when ProvisionVirtualIP is not set (cross-cluster HAProxy).
func TestPostgresProvisionRequestAppender_HA_noVirtualIP(t *testing.T) {
	baseRequest := &DatabaseProvisionRequest{}
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	haConfig := &HAConfig{
		PatroniClusterName:    "test-patroni-cluster",
		ClusterName:           "test-ha-cluster",
		EnableSynchronousMode: false,
		ProvisionVirtualIP:    false, // HAProxy nodes on different PE clusters
	}

	mockDatabase.On("GetInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetInstanceType").Return(common.DATABASE_TYPE_POSTGRES)
	mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
	mockDatabase.On("IsClone").Return(false)
	mockDatabase.On("IsPostgresHA").Return(true)
	mockDatabase.On("GetInstanceHAConfig").Return(haConfig)

	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_POSTGRES)
	resultRequest, err := requestAppender.appendProvisioningRequest(baseRequest, mockDatabase, reqData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	argMap := make(map[string]string)
	for _, arg := range resultRequest.ActionArguments {
		argMap[arg.Name] = arg.Value
	}

	if got := argMap["provision_virtual_ip"]; got != "false" {
		t.Errorf("provision_virtual_ip: expected false, got %s", got)
	}
	if got := argMap["deploy_haproxy"]; got != "true" {
		t.Errorf("deploy_haproxy: expected true, got %s", got)
	}
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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MSSQL)

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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MSSQL)

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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MSSQL)

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
	mockDatabase.On("IsMongoHA").Return(false)
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
			Value: "mongod",
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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MONGODB)

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
	mockDatabase.On("IsMongoHA").Return(false)
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
			Value: "mongod",
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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MONGODB)

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
	mockDatabase.On("IsMongoHA").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MONGODB)

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
	mockDatabase.On("IsMysqlHA").Return(false)
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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MYSQL)

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
	mockDatabase.On("IsMysqlHA").Return(false)
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
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MYSQL)

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
	mockDatabase.On("IsMysqlHA").Return(false)
	// Get specific implementation of RequestAppender
	requestAppender, _ := GetRequestAppender(common.DATABASE_TYPE_MYSQL)

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
		mockDatabase.On("IsMysqlHA").Return(false)

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
		mockDatabase.On("IsMysqlHA").Return(false)

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
			expectedError: errors.New("invalid database type: supported values: mssql, mysql, postgres, mongodb, oracle"),
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
		mockDatabase.On("IsClone").Return(false)
		mockDatabase.On("IsPostgresHA").Return(false)
		mockDatabase.On("IsMysqlHA").Return(false)
		mockDatabase.On("IsMongoHA").Return(false)
		mockDatabase.On("GetInstanceHAConfig").Return((*HAConfig)(nil))

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

func TestOracleRequestAppender_appendProvisioningRequest(t *testing.T) {
	appender := &OracleRequestAppender{}

	t.Run("Oracle provision request has correct default action arguments", func(t *testing.T) {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("oradb1")
		mockDatabase.On("GetInstanceDatabaseNames").Return("oradb1")
		mockDatabase.On("GetInstanceSize").Return(50)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
		mockDatabase.On("IsClone").Return(false)
		mockDatabase.On("GetInstanceType").Return("oracle")

		req := &DatabaseProvisionRequest{}
		reqData := map[string]interface{}{
			"password":       "OraclePassword123",
			"ssh_public_key": "ssh-rsa oracle-key",
		}

		result, err := appender.appendProvisioningRequest(req, mockDatabase, reqData)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "oradb1", result.DatabaseName)
		assert.Equal(t, "ssh-rsa oracle-key", result.SSHPublicKey)

		// Check that required Oracle provisioning action arguments are present
		actionArgsMap := make(map[string]string)
		for _, arg := range result.ActionArguments {
			actionArgsMap[arg.Name] = arg.Value
		}

		// Verify Oracle-specific provisioning parameters
		assert.Equal(t, "1521", actionArgsMap["listener_port"])
		assert.Equal(t, "true", actionArgsMap["auto_tune_staging_drive"])
		assert.Equal(t, "/tmp", actionArgsMap["working_dir"])
		assert.Equal(t, "oradb1_VM", actionArgsMap["dbserver_name"])
		assert.Equal(t, "oradb1", actionArgsMap["oracle_sid"])
		assert.Equal(t, "oradb1", actionArgsMap["global_database_name"])
		assert.Equal(t, "oradb1", actionArgsMap["db_unique_name"])
		assert.Equal(t, "OraclePassword123", actionArgsMap["sys_password"])
		assert.Equal(t, "OraclePassword123", actionArgsMap["system_password"])
		assert.Equal(t, "AL32UTF8", actionArgsMap["db_character_set"])
		assert.Equal(t, "AL16UTF16", actionArgsMap["national_character_set"])
		assert.Equal(t, "50", actionArgsMap["database_fra_size"])
		assert.Equal(t, "false", actionArgsMap["enable_cdb"])
		assert.Equal(t, "false", actionArgsMap["enable_tde"])
		assert.Equal(t, "false", actionArgsMap["enable_ha"])
	})

	t.Run("Oracle provision request allows overriding defaults via additionalArguments", func(t *testing.T) {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("oradb2")
		mockDatabase.On("GetInstanceDatabaseNames").Return("oradb2")
		mockDatabase.On("GetInstanceSize").Return(100)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{
			"listener_port":      "1522",
			"enable_tde":         "true",
			"db_character_set":   "UTF8",
			"pre_create_script":  "echo 'Pre-create'",
			"post_create_script": "echo 'Post-create'",
		})
		mockDatabase.On("IsClone").Return(false)
		mockDatabase.On("GetInstanceType").Return("oracle")

		req := &DatabaseProvisionRequest{}
		reqData := map[string]interface{}{
			"password":       "CustomPass456",
			"ssh_public_key": "ssh-rsa custom-key",
		}

		result, err := appender.appendProvisioningRequest(req, mockDatabase, reqData)

		require.NoError(t, err)
		require.NotNil(t, result)

		actionArgsMap := make(map[string]string)
		for _, arg := range result.ActionArguments {
			actionArgsMap[arg.Name] = arg.Value
		}

		// Verify overridden values
		assert.Equal(t, "1522", actionArgsMap["listener_port"])
		assert.Equal(t, "true", actionArgsMap["enable_tde"])
		assert.Equal(t, "UTF8", actionArgsMap["db_character_set"])
		assert.Equal(t, "echo 'Pre-create'", actionArgsMap["pre_create_script"])
		assert.Equal(t, "echo 'Post-create'", actionArgsMap["post_create_script"])
		// Verify defaults remain for non-overridden values
		assert.Equal(t, "false", actionArgsMap["enable_cdb"])
		assert.Equal(t, "oradb2", actionArgsMap["oracle_sid"])
	})

	t.Run("Oracle provision request correctly sets database names", func(t *testing.T) {
		mockDatabase := &MockDatabaseInterface{}
		mockDatabase.On("GetName").Return("oradb3")
		mockDatabase.On("GetInstanceDatabaseNames").Return("CustomOraDB")
		mockDatabase.On("GetInstanceSize").Return(75)
		mockDatabase.On("GetAdditionalArguments").Return(map[string]string{})
		mockDatabase.On("IsClone").Return(false)
		mockDatabase.On("GetInstanceType").Return("oracle")

		req := &DatabaseProvisionRequest{}
		reqData := map[string]interface{}{
			"password":       "Pass789",
			"ssh_public_key": "ssh-rsa test-key",
		}

		result, err := appender.appendProvisioningRequest(req, mockDatabase, reqData)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify top-level DatabaseName field
		assert.Equal(t, "CustomOraDB", result.DatabaseName)

		actionArgsMap := make(map[string]string)
		for _, arg := range result.ActionArguments {
			actionArgsMap[arg.Name] = arg.Value
		}

		// Verify all three Oracle identifiers use the custom database name
		assert.Equal(t, "CustomOraDB", actionArgsMap["oracle_sid"])
		assert.Equal(t, "CustomOraDB", actionArgsMap["global_database_name"])
		assert.Equal(t, "CustomOraDB", actionArgsMap["db_unique_name"])
	})
}
