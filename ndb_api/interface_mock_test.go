package ndb_api

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockDatabaseInterface is a mock implementation of the DatabaseInterface interface
type MockDatabaseInterface struct {
	mock.Mock
}

// MockProfileResolverInterface is a mock implementation of the ProfileResolver interface
type MockProfileResolverInterface struct {
	mock.Mock
}

// GetDBInstanceName is a mock implementation of the GetDBInstanceName method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceName() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceDescription is a mock implementation of the GetDBInstanceDescription method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceDescription() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceType is a mock implementation of the GetDBInstanceType method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceType() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceDatabaseNames is a mock implementation of the GetDBInstanceDatabaseNames method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceDatabaseNames() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceTimeZone is a mock implementation of the GetDBInstanceTimeZone method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceTimeZone() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceSize is a mock implementation of the GetDBInstanceSize method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceSize() int {
	args := m.Called()
	return args.Int(0)
}

// GetNDBClusterId is a mock implementation of the GetNDBClusterId method in the Database interface
func (m *MockDatabaseInterface) GetNDBClusterId() string {
	args := m.Called()
	return args.String(0)
}

// GetProfileResolvers is a mock implementation of the GetProfileResolvers method in the Database interface
func (m *MockDatabaseInterface) GetProfileResolvers() ProfileResolvers {
	args := m.Called()
	return args.Get(0).(ProfileResolvers)
}

// GetTMDetails is a mock implementation of the GetTMDetails method in the Database interface
func (m *MockDatabaseInterface) GetTMDetails() (string, string, string) {
	args := m.Called()
	return args.String(0), args.String(1), args.String(2)
}

// GetTMSchedule is a mock implementation of the GetTMSchedule method in the Database interface
func (m *MockDatabaseInterface) GetTMSchedule() (Schedule, error) {
	args := m.Called()
	return args.Get(0).(Schedule), args.Error(1)
}

// GetName is a mock implementation of the GetName method defined in the ProfileResolver interface
func (m *MockProfileResolverInterface) GetName() string {
	args := m.Called()
	return args.String(0)
}

// GetId is a mock implementation of the GetId method defined in the ProfileResolver interface
func (m *MockProfileResolverInterface) GetId() string {
	args := m.Called()
	return args.String(0)
}

// Resolve is a mock implementation of the Resolve method defined in the ProfileResolver interface
func (m *MockProfileResolverInterface) Resolve(ctx context.Context, allProfiles []ProfileResponse, filter func(p ProfileResponse) bool) (ProfileResponse, error) {
	args := m.Called()
	return args.Get(0).(ProfileResponse), args.Error(1)
}

// GetDBInstanceAdditionalArguments is a mock implementation of the GetDBInstanceTypeDetails method in the Database interface
func (m *MockDatabaseInterface) GetDBInstanceAdditionalArguments() map[string]string {
	args := m.Called()

	// Perform a type assertion to convert the value to map[string]string
	if result, ok := args.Get(0).(map[string]string); ok {
		return result
	}

	// If the type assertion fails, return default
	return map[string]string{}
}
