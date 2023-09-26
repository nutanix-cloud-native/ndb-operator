package ndb_api

import (
	"context"
	"errors"
	"reflect"
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
		databaseType string
		expected     interface{}
	}{
		{databaseType: common.DATABASE_TYPE_POSTGRES,
			expected: &PostgresProvisionRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MYSQL,
			expected: &MySqlProvisionRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MSSQL,
			expected: &MSSQLProvisionRequestAppender{},
		},
		{databaseType: common.DATABASE_TYPE_MONGODB,
			expected: &MongoDbProvisionRequestAppender{},
		},
		{databaseType: "test",
			expected: nil,
		},
	}

	for _, tc := range tests {
		got, _ := GetDbProvRequestAppender(tc.databaseType)
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
	}
}

// Tests if PostgresProvisionRequestAppender() function appends requests correctly without type details
func TestPostgresProvisionRequestAppenderWithoutTypeDetails(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

	var expectedActionArgs []ActionArgument
	expectedActionArgs = append(
		expectedActionArgs,
		postgresReplacableActionArgs()...,
	)
	expectedActionArgs = append(
		expectedActionArgs,
		postgresDefaultActionArgs(TEST_PASSWORD, mockDatabase.GetDBInstanceDatabaseNames())...,
	)

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_POSTGRES)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")
}

// Tests if PostgresProvisionRequestAppender() function appends requests correctly with type details
func TestPostgresProvisionRequestAppenderWithTypeDetails(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

	baseRequest.ActionArguments = append(baseRequest.ActionArguments, []ActionArgument{
		{Name: "listener_port", Value: "5432"},
	}...)

	expectedActionArgs := []ActionArgument{
		{Name: "listener_port", Value: "5432"},
		{Name: "proxy_read_port", Value: "5001"},
		{Name: "proxy_write_port", Value: "5000"},
		{Name: "enable_synchronous_mode", Value: "false"},
		{Name: "auto_tune_staging_drive", Value: "true"},
		{Name: "backup_policy", Value: "primary_only"},
		{Name: "db_password", Value: TEST_PASSWORD},
		{Name: "database_names", Value: mockDatabase.GetDBInstanceDatabaseNames()},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_POSTGRES)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")
}

// Tests if MSSQLProvisionRequestAppender() function appends requests correctly with no typeDetails specified
func TestMSSQLProvisionRequestAppenderWithoutTypeDetails(t *testing.T) {

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
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetDBInstanceName").Return("testInstance")

	var expectedActionArgs []ActionArgument
	expectedActionArgs = append(
		expectedActionArgs,
		mssqlReplacableActionArgs(
			profileMap[common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE].Id,
			adminPassword,
		)...,
	)
	expectedActionArgs = append(
		expectedActionArgs,
		mssqlDefaultActionArgs(mockDatabase.GetDBInstanceName())...,
	)

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_MSSQL)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.DatabaseName != mockDatabase.GetDBInstanceDatabaseNames() {
		t.Errorf("Unexpected Database Name. Expected: %s, Got: %s", mockDatabase.GetDBInstanceDatabaseNames(), resultRequest.DatabaseName)
	}

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")
}

// Tests if MSSQLProvisionRequestAppender() function appends requests correctly with typeDetails specified
func TestMSSQLProvisionRequestAppenderWithTypeDetails(t *testing.T) {

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
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)
	mockDatabase.On("GetDBInstanceName").Return("testInstance")

	baseRequest.ActionArguments = append(baseRequest.ActionArguments, []ActionArgument{
		{Name: "server_collation", Value: "SQL_Latin1_General_CPI_CI_AS"},
		{Name: "database_collation", Value: "SQL_Latin1_General_CPI_CI_AS"},
		{Name: "vm_win_license_key", Value: "XXXX-XXXXX-XXXXX-XXXXX-XXXXX"},
		{Name: "vm_dbserver_admin_password", Value: "<password>"},
		{Name: "authentication_mode", Value: "mixed"},
		{Name: "sql_user_name", Value: "sa"},
		{Name: "sql_user_password", Value: "<password>"},
		{Name: "windows_domain_profile_id", Value: "<XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"},
		{Name: "vm_db_server_user", Value: "<prod.cdm.com\\<user>"},
	}...)

	expectedActionArgs := []ActionArgument{
		{Name: "server_collation", Value: "SQL_Latin1_General_CPI_CI_AS"},
		{Name: "database_collation", Value: "SQL_Latin1_General_CPI_CI_AS"},
		{Name: "vm_win_license_key", Value: "XXXX-XXXXX-XXXXX-XXXXX-XXXXX"},
		{Name: "vm_dbserver_admin_password", Value: "<password>"},
		{Name: "authentication_mode", Value: "mixed"},
		{Name: "sql_user_name", Value: "sa"},
		{Name: "sql_user_password", Value: "<password>"},
		{Name: "windows_domain_profile_id", Value: "<XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"},
		{Name: "vm_db_server_user", Value: "<prod.cdm.com\\<user>"},
		{Name: "working_dir", Value: "C:\\temp"},
		{Name: "delete_vm_on_failure", Value: "false"},
		{Name: "is_gmsa_sql_service_account", Value: "false"},
		{Name: "provision_from_backup", Value: "false"},
		{Name: "distribute_database_data", Value: "true"},
		{Name: "retain_database_in_restoring_mode", Value: "false"},
		{Name: "dbserver_name", Value: mockDatabase.GetDBInstanceName()},
		{Name: "dbParameterProfileIdInstance", Value: profileResponse.Id},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_MSSQL)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.DatabaseName != mockDatabase.GetDBInstanceDatabaseNames() {
		t.Errorf("Unexpected Database Name. Expected: %s, Got: %s", mockDatabase.GetDBInstanceDatabaseNames(), resultRequest.DatabaseName)
	}

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")

}

// Tests if MongoDbProvisionRequestAppender() function appends requests correctly without type details.
func TestMongoDbProvisionRequestAppenderWithoutTypeDetails(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

	var expectedActionArgs []ActionArgument
	expectedActionArgs = append(
		expectedActionArgs,
		mongoDbReplacableActionArgs()...,
	)
	expectedActionArgs = append(
		expectedActionArgs,
		mongoDbDefaultActionArgs(TEST_PASSWORD, mockDatabase.GetDBInstanceDatabaseNames())...,
	)

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_MONGODB)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")
}

func TestMongoDbProvisionRequestAppenderWithTypeDetails(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

	baseRequest.ActionArguments = append(baseRequest.ActionArguments, []ActionArgument{
		{Name: "listener_port", Value: "27017"},
		{Name: "log_size", Value: "100"},
		{Name: "journal_size", Value: "100"},
	}...)

	expectedActionArgs := []ActionArgument{
		{Name: "listener_port", Value: "27017"},
		{Name: "log_size", Value: "100"},
		{Name: "journal_size", Value: "100"},
		{Name: "restart_mongod", Value: "true"},
		{Name: "working_dir", Value: "/tmp"},
		{Name: "db_user", Value: "admin"},
		{Name: "backup_policy", Value: "primary_only"},
		{Name: "db_password", Value: TEST_PASSWORD},
		{Name: "database_names", Value: mockDatabase.GetDBInstanceDatabaseNames()},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_MONGODB)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")

}

// Tests if MySqlProvisionRequestAppender() function appends requests correctly without typeDetails
func TestMySqlProvisionRequestAppenderWithoutTypeDetails(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

	var expectedActionArgs []ActionArgument
	expectedActionArgs = append(
		expectedActionArgs,
		mysqlReplacableActionArgs()...,
	)
	expectedActionArgs = append(
		expectedActionArgs,
		mysqlDefaultActionArgs(TEST_PASSWORD, mockDatabase.GetDBInstanceDatabaseNames())...,
	)

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_MYSQL)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")
}

// Tests if MySqlProvisionRequestAppender() function appends requests correctly with type details
func TestMySqlProvisionRequestAppenderWithTypeDetails(t *testing.T) {

	baseRequest := &DatabaseProvisionRequest{}
	// Create a mock implementation of DatabaseInterface
	mockDatabase := &MockDatabaseInterface{}

	reqData := map[string]interface{}{
		common.NDB_PARAM_SSH_PUBLIC_KEY: TEST_SSHKEY,
		common.NDB_PARAM_PASSWORD:       TEST_PASSWORD,
	}

	// Mock required Mock Database Interface methods
	mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

	baseRequest.ActionArguments = append(baseRequest.ActionArguments, []ActionArgument{
		{Name: "listener_port", Value: "3306"},
	}...)

	expectedActionArgs := []ActionArgument{
		{Name: "listener_port", Value: "3306"},
		{Name: "db_password", Value: TEST_PASSWORD},
		{Name: "database_names", Value: mockDatabase.GetDBInstanceDatabaseNames()},
	}

	// Get specific implementation of RequestAppender
	requestAppender, _ := GetDbProvRequestAppender(common.DATABASE_TYPE_MYSQL)

	// Call function being tested
	resultRequest := requestAppender.appendRequest(baseRequest, mockDatabase, reqData)

	// Assert expected results
	if resultRequest.SSHPublicKey != reqData[common.NDB_PARAM_SSH_PUBLIC_KEY] {
		t.Errorf("Unexpected SSHPublicKey value. Expected: %s, Got: %s", reqData[common.NDB_PARAM_SSH_PUBLIC_KEY], resultRequest.SSHPublicKey)
	}

	// Sort action args
	sortActionArgs(expectedActionArgs)
	sortActionArgs(resultRequest.ActionArguments)

	// Check if the lengths of the slices are equal
	if len(expectedActionArgs) != len(resultRequest.ActionArguments) {
		t.Errorf("Unexpected ActionArguments length. Expected: %d, Got: %d", len(expectedActionArgs), len(resultRequest.ActionArguments))
		return
	}

	// Iterate over the sorted slices and compare each element
	for i := range expectedActionArgs {
		if expectedActionArgs[i] != resultRequest.ActionArguments[i] {
			t.Errorf("Unexpected ActionArgument at index %d. Expected: %v, Got: %v", i, expectedActionArgs[i], resultRequest.ActionArguments[i])
		}
	}

	// Verify that the mock method was called with the expected arguments
	mockDatabase.AssertCalled(t, "GetDBInstanceDatabaseNames")

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
		mockDatabase.On("GetDBInstanceName").Return("db_instance_name")
		mockDatabase.On("GetDBInstanceType").Return("db_instance_type")
		mockDatabase.On("GetTMDetails").Return("tm_name", "rm_description", tc.slaName)
		mockDatabase.On("GetTMSchedule").Return(tc.tmSchedule, tc.tmScheduleErr)

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
		mockDatabase.On("GetDBInstanceName").Return("db_instance_name")
		mockDatabase.On("GetDBInstanceType").Return(instanceType)
		mockDatabase.On("GetTMDetails").Return("tm_name", "rm_description", "SLA 1")
		mockDatabase.On("GetTMSchedule").Return(Schedule{}, nil)
		mockDatabase.On("GetProfileResolvers").Return(profileResolvers)

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
		mockDatabase.On("GetDBInstanceName").Return("db_instance_name")
		mockDatabase.On("GetDBInstanceDescription").Return("db_instance_description")
		mockDatabase.On("GetDBInstanceType").Return(instanceType)
		mockDatabase.On("GetDBInstanceTypeDetails").Return([]ActionArgument{})
		mockDatabase.On("GetTMDetails").Return("tm_name", "rm_description", "SLA 1")
		mockDatabase.On("GetTMSchedule").Return(Schedule{}, nil)
		mockDatabase.On("GetProfileResolvers").Return(profileResolvers)
		mockDatabase.On("GetDBInstanceTimeZone").Return(TEST_TIMEZONE)
		mockDatabase.On("GetNDBClusterId").Return(TEST_CLUSTER_ID)
		mockDatabase.On("GetDBInstanceSize").Return(TEST_INSTANCE_SIZE)
		mockDatabase.On("GetDBInstanceDatabaseNames").Return(TEST_DB_NAMES)

		// Test
		_, err := GenerateProvisioningRequest(context.Background(), ndb_client, &mockDatabase, tc.reqData)

		// Assert
		if err != tc.expectedError && err.Error() != tc.expectedError.Error() {
			t.Fatalf("expected: %v, got: %v", tc.expectedError, err)
		}
	}
}
