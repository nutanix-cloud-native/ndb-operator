package ndb_api

import (
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

// GetDBInstanceName is a mock implementation of the GetDBInstanceName method
func (m *MockDatabaseInterface) GetDBInstanceName() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceType is a mock implementation of the GetDBInstanceType method
func (m *MockDatabaseInterface) GetDBInstanceType() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceDatabaseNames is a mock implementation of the GetDBInstanceDatabaseNames method
func (m *MockDatabaseInterface) GetDBInstanceDatabaseNames() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceTimeZone is a mock implementation of the GetDBInstanceTimeZone method
func (m *MockDatabaseInterface) GetDBInstanceTimeZone() string {
	args := m.Called()
	return args.String(0)
}

// GetDBInstanceSize is a mock implementation of the GetDBInstanceSize method
func (m *MockDatabaseInterface) GetDBInstanceSize() int {
	args := m.Called()
	return args.Int(0)
}

// GetNDBClusterId is a mock implementation of the GetNDBClusterId method
func (m *MockDatabaseInterface) GetNDBClusterId() string {
	args := m.Called()
	return args.String(0)
}

// GetProfileResolvers is a mock implementation of the GetProfileResolvers method
func (m *MockDatabaseInterface) GetProfileResolvers() ProfileResolvers {
	args := m.Called()
	return args.Get(0).(ProfileResolvers)
}

// GetTMDetails is a mock implementation of the GetTMDetails method
func (m *MockDatabaseInterface) GetTMDetails() (string, string, string) {
	args := m.Called()
	return args.String(0), args.String(1), args.String(2)
}

// GetTMDetails is a mock implementation of the GetTMDetails method
func (m *MockDatabaseInterface) GetTMSchedule() (Schedule, error) {
	args := m.Called()
	return args.Get(0).(Schedule), args.Get(1).(error)
}
