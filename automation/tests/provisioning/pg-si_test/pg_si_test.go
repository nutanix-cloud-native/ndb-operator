package postgres_si_provisoning

// Basic imports
import (
	"context"
	"testing"

	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	util "github.com/nutanix-cloud-native/ndb-operator/automation/util"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// A test suite is a collection of related test cases that are grouped together for testing a specific package or functionality.
// The testify package builds on top of Go's built-in testing package and enhances it with additional features like assertions and test suite management.
// PostgresqlSingleInstanceTestSuite is a test suite struct that embeds testify's suite.Suite
type PostgresqlSingleInstanceTestSuite struct {
	suite.Suite
	ctx               context.Context
	setupTypes        *util.SetupTypes
	v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	clientset         *kubernetes.Clientset
}

// SetupSuite is called once before running the tests in the suite
func (suite *PostgresqlSingleInstanceTestSuite) SetupSuite() {
	var err error
	var config *rest.Config

	var ctx context.Context
	var v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	var clientset *kubernetes.Clientset

	// Setup logger and context
	logger, err := util.SetupLogger("./pg-si_test.log")
	if err != nil {
		suite.T().FailNow()
	}
	ctx = util.SetupContext(context.Background(), logger)

	logger.Println("SetupSuite() starting...")
	errBaseMsg := "Error: SetupSuite() ended"

	// Setup env
	if err = util.LoadEnv(ctx); err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Setup kubeconfig
	config, err = util.SetupKubeconfig(ctx)
	if err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Setup scheme and clientsets
	if v1alpha1ClientSet, clientset, err = util.SetupSchemeAndClientSet(ctx, config); err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Setup yaml types
	setupTypes, err := util.SetupTypeTemplates(ctx)
	if err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Provision database and wait for database and pod to be ready
	if err := util.ProvisioningTestSetup(ctx, setupTypes, clientset, v1alpha1ClientSet, suite.T()); err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Set variables for the entire suite
	suite.ctx = ctx
	suite.setupTypes = setupTypes
	suite.v1alpha1ClientSet = v1alpha1ClientSet
	suite.clientset = clientset

	logger.Println("SetupSuite() ended!")
}

// TearDownSuite is called once after running the tests in the suite
func (suite *PostgresqlSingleInstanceTestSuite) TearDownSuite() {
	var err error

	logger := util.GetLogger(suite.ctx)
	logger.Println("TearDownSuite() starting...")
	errBaseMsg := "Error: SetupSuite() ended"

	// Setup yaml types
	setupTypes, err := util.SetupTypeTemplates(suite.ctx)
	if err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Delete resources and de-provision database
	if err = util.ProvisioningTestTeardown(suite.ctx, setupTypes, suite.clientset, suite.v1alpha1ClientSet, suite.T()); err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	logger.Println("TearDownSuite() ended!")
}

// This will run right before the test starts and receives the suite and test names as input
func (suite *PostgresqlSingleInstanceTestSuite) BeforeTest(suiteName, testName string) {
	util.GetLogger(suite.ctx).Printf("******************** RUNNING TEST %s %s ********************\n", suiteName, testName)
}

// This will run after test finishes and receives the suite and test names as input
func (suite *PostgresqlSingleInstanceTestSuite) AfterTest(suiteName, testName string) {
	util.GetLogger(suite.ctx).Printf("******************** END TEST %s %s ********************\n", suiteName, testName)
}

// Tests if provisioning is succesful by checking if database status is 'READY'
func (suite *PostgresqlSingleInstanceTestSuite) TestProvisioningSuccess() {
	logger := util.GetLogger(suite.ctx)

	databaseResponse, err := util.GetDatabaseResponse(suite.ctx, suite.clientset, suite.v1alpha1ClientSet, suite.setupTypes)
	if err != nil {
		logger.Printf("Error: TestProvisioningSuccess() failed! %v", err)
	} else {
		logger.Println("Database response retrieved.")
	}

	assert := assert.New(suite.T())
	assert.Equal(common.DATABASE_CR_STATUS_READY, databaseResponse.Status, "The database status should be ready.")
}

// Tests if app is able to connect to database via GET request
func (suite *PostgresqlSingleInstanceTestSuite) TestAppConnectivity() {
	logger := util.GetLogger(suite.ctx)

	resp, err := util.GetAppResponse(suite.ctx, suite.clientset, suite.setupTypes.AppPod, "3000")
	if err != nil {
		logger.Printf("Error: TestAppConnectivity failed! %v", err)
	} else {
		logger.Println("App response retrieved.")
	}

	assert := assert.New(suite.T())
	assert.Equal(200, resp.StatusCode, "The response status should be 200.")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPostgresqlSingleInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresqlSingleInstanceTestSuite))
}
