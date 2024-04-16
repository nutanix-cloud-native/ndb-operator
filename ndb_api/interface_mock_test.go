package ndb_api

import (
	"context"
	"net/http"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
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

// MockNDBClientHTTPInterface is a mock implementation of the NDBClientHTTP interface defined in the ndb_client package
type MockNDBClientHTTPInterface struct {
	mock.Mock
}

// IsClone is a mock implementation of the IsClone method defined in the Database interface
func (m *MockDatabaseInterface) IsClone() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetName is a mock implementation of the GetName method in the Database interface
func (m *MockDatabaseInterface) GetName() string {
	args := m.Called()
	return args.String(0)
}

// GetDescription is a mock implementation of the GetDescription method in the Database interface
func (m *MockDatabaseInterface) GetDescription() string {
	args := m.Called()
	return args.String(0)
}

// GetClusterId is a mock implementation of the GetClusterId method in the Database interface
func (m *MockDatabaseInterface) GetClusterId() string {
	args := m.Called()
	return args.String(0)
}

// GetProfileResolvers is a mock implementation of the GetProfileResolvers method in the Database interface
func (m *MockDatabaseInterface) GetProfileResolvers() ProfileResolvers {
	args := m.Called()
	return args.Get(0).(ProfileResolvers)
}

// GetInstanceSize is a mock implementation of the GetInstanceSize method in the Database interface
func (m *MockDatabaseInterface) GetCredentialSecret() string {
	args := m.Called()
	return args.String(0)
}

// GetTimeZone is a mock implementation of the GetTimeZone method in the Database interface
func (m *MockDatabaseInterface) GetTimeZone() string {
	args := m.Called()
	return args.String(0)
}

// GetInstanceType is a mock implementation of the GetInstanceType method in the Database interface
func (m *MockDatabaseInterface) GetInstanceType() string {
	args := m.Called()
	return args.String(0)
}

// GetInstanceDatabaseNames is a mock implementation of the GetInstanceDatabaseNames method in the Database interface
func (m *MockDatabaseInterface) GetInstanceDatabaseNames() string {
	args := m.Called()
	return args.String(0)
}

// GetInstanceSize is a mock implementation of the GetInstanceSize method in the Database interface
func (m *MockDatabaseInterface) GetInstanceSize() int {
	args := m.Called()
	return args.Int(0)
}

// GetInstanceTMDetails is a mock implementation of the GetInstanceTMDetails method in the Database interface
func (m *MockDatabaseInterface) GetInstanceTMDetails() (string, string, string) {
	args := m.Called()
	return args.String(0), args.String(1), args.String(2)
}

// GetTMScheduleForInstance is a mock implementation of the GetTMScheduleForInstance method in the Database interface
func (m *MockDatabaseInterface) GetTMScheduleForInstance() (Schedule, error) {
	args := m.Called()
	return args.Get(0).(Schedule), args.Error(1)
}

// GetCloneSourceDBId is a mock implementation of the GetCloneSourceDBId method in the Database interface
func (m *MockDatabaseInterface) GetCloneSourceDBId() string {
	args := m.Called()
	return args.String(0)
}

// GetCloneTMName is a mock implementation of the GetCloneTMName method in the Database interface
func (m *MockDatabaseInterface) GetCloneTMName() string {
	args := m.Called()
	return args.String(0)
}

// GetCloneSnapshotId is a mock implementation of the GetCloneSnapshotId method in the Database interface
func (m *MockDatabaseInterface) GetCloneSnapshotId() string {
	args := m.Called()
	return args.String(0)
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

// GetAdditionalArguments is a mock implementation of the GetAdditionalArguments method in the Database interface
func (m *MockDatabaseInterface) GetAdditionalArguments() map[string]string {
	args := m.Called()

	// Perform a type assertion to convert the value to map[string]string
	if result, ok := args.Get(0).(map[string]string); ok {
		return result
	}

	// If the type assertion fails, return default
	return map[string]string{}
}

func (m *MockNDBClientHTTPInterface) NewRequest(method, endpoint string, requestBody interface{}) (*http.Request, error) {
	args := m.Called(method, endpoint, requestBody)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Request), args.Error(1)
}

func (m *MockNDBClientHTTPInterface) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// GetInstanceIsHighAvailability is a mock implementation of the GetInstanceIsHighAvailability method in the Database interface
func (m *MockDatabaseInterface) GetInstanceIsHighAvailability() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDatabaseInterface) GetInstanceNodes() []*v1alpha1.Node {
	args := m.Called()
	return args.Get(0).([]*v1alpha1.Node)
}
