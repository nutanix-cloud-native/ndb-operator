package microsoftsqlsi

// Basic imports
import (
	"context"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/automation"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// A test suite is a collection of related test cases that are grouped together for testing a specific package or functionality.
// The testify package builds on top of Go's built-in testing package and enhances it with additional features like assertions and test suite management.
// PostgresqlSingleInstanceTestSuite is a test suite struct that embeds testify's suite.Suite
type MicrosoftSingleInstanceTestSuite struct {
	suite.Suite
	ctx               context.Context
	v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	clientset         *kubernetes.Clientset
}

// SetupSuite is called once before running the tests in the suite
func (suite *MicrosoftSingleInstanceTestSuite) SetupSuite() {
	var err error
	var config *rest.Config

	var ctx context.Context
	var v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	var clientset *kubernetes.Clientset

	// Setup logger and context
	logger, err := automation.SetupLogger("./mssql-si_test.log")
	if err != nil {
		suite.T().FailNow()
	}
	ctx = automation.SetupContext(context.Background(), logger)

	logger.Println("SetupSuite() starting...")

	// Setup kubeconfig
	config, err = automation.SetupKubeconfig(ctx)
	if err != nil {
		logger.Printf("Error: SetupSuite() ended! %s\n", err)
		suite.T().FailNow()
	}

	// Setup scheme and clientsets
	if v1alpha1ClientSet, clientset, err = automation.SetupSchemeAndClientSet(ctx, config); err != nil {
		logger.Printf("Error: SetupSuite() ended! %s\n", err)
		suite.T().FailNow()
	}

	// Setup yaml types
	setupTypes, err := automation.SetupTypeTemplates(ctx)
	if err != nil {
		logger.Printf("Error: SetupSuite() ended! %s\n", err)
		suite.T().FailNow()
	}

	// Provision database and wait for database and pod to be ready
	if err := automation.ProvisioningTestSetup(ctx, setupTypes, clientset, v1alpha1ClientSet, suite.T()); err != nil {
		logger.Printf("Error: SetupSuite() ended! %s\n", err)
		suite.T().FailNow()
	}

	// Set variables for the entire suite
	suite.ctx = ctx
	suite.v1alpha1ClientSet = v1alpha1ClientSet
	suite.clientset = clientset

	logger.Println("SetupSuite() ended!")
}

// TearDownSuite is called once after running the tests in the suite
func (suite *MicrosoftSingleInstanceTestSuite) TearDownSuite() {
	logger := automation.GetLogger(suite.ctx)
	var err error

	logger.Println("TearDownSuite() starting...")

	// Setup yaml types
	setupTypes, err := automation.SetupTypeTemplates(suite.ctx)
	if err != nil {
		logger.Printf("Error: TearDownSuite() ended! %s\n", err)
		suite.T().FailNow()
	}

	// Delete resources and de-provision database
	if err = automation.ProvisioningTestTeardown(suite.ctx, setupTypes, suite.clientset, suite.v1alpha1ClientSet, suite.T()); err != nil {
		logger.Printf("Error: TearDownSuite() ended! %s\n", err)
		suite.T().FailNow()
	}

	logger.Println("TearDownSuite() ended!")
}

// This will run right before the test starts and receives the suite and test names as input
func (suite *MicrosoftSingleInstanceTestSuite) BeforeTest(suiteName, testName string) {
	automation.GetLogger(suite.ctx).Printf("******************** RUNNING TEST %s %s ********************\n", suiteName, testName)
}

// This will run after test finishes and receives the suite and test names as input
func (suite *MicrosoftSingleInstanceTestSuite) AfterTest(suiteName, testName string) {
	automation.GetLogger(suite.ctx).Printf("******************** END TEST %s %s ********************\n", suiteName, testName)
}

// Tests if provisioning is succesful by checking if database status is 'READY'
func (suite *MicrosoftSingleInstanceTestSuite) TestProvisioningSuccess() {
	logger := automation.GetLogger(suite.ctx)

	databaseResponse, err := automation.GetDatabaseResponseFromCR(suite.ctx, suite.clientset, suite.v1alpha1ClientSet)
	if err != nil {
		logger.Printf("TestProvisioningSuccess() failed! %v", err)
	} else {
		logger.Println("DatabaseResponse succesfully retrieved.")
	}

	assert := assert.New(suite.T())
	assert.Equal(common.DATABASE_CR_STATUS_READY, databaseResponse.Status, "The database status should be ready.")
}

// Tests if app is able to connect to database
func (suite *MicrosoftSingleInstanceTestSuite) TestAppConnectivity() {
	logger := automation.GetLogger(suite.ctx)

	resp, err := automation.GetAppResponse(suite.ctx, suite.clientset, "3001")
	if err != nil {
		logger.Printf("TestAppConnectivity failed! %v", err)
	} else {
		logger.Println("App Response succesfully retrieved.")
	}

	assert := assert.New(suite.T())
	assert.Equal(200, resp.StatusCode, "The response status should be 200.")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMicrosoftSingleInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(MicrosoftSingleInstanceTestSuite))
}
