package postgres_provisoning_ha

// Basic imports
import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/automation"
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
// PostgresProvisioningHightAvailabilityInstanceTestSuite is a test suite struct that embeds testify's suite.Suite
type PostgresProvisioningHighAvailabilityTestSuite struct {
	suite.Suite
	ctx               context.Context
	setupTypes        *util.SetupTypes
	v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	clientset         *kubernetes.Clientset
	tsm               util.TestSuiteManager
}

// SetupSuite is called once before running the tests in the suite
func (suite *PostgresProvisioningHighAvailabilityTestSuite) SetupSuite() {
	var err error
	var config *rest.Config
	var ctx context.Context
	var v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	var clientset *kubernetes.Clientset
	var tsm util.TestSuiteManager

	// Setup logger and context
	logger, err := util.SetupLogger(fmt.Sprintf("%s/pg-provisioning-ha_test.log", automation.PROVISIONING_LOG_PATH), "pg-provisioning-ha: ")
	if err != nil {
		suite.T().FailNow()
	}
	ctx = util.SetupContext(context.Background(), logger)

	logger.Println("SetupSuite() starting...")
	errBaseMsg := "Error: SetupSuite() ended"

	// Setup env
	if err = util.CheckRequiredEnv(ctx); err != nil {
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

	// Get test suite manager
	tsm = util.GetTestSuiteManager(ctx, *setupTypes)

	// Provision database and wait for database and pod to be ready
	if err := tsm.Setup(ctx, setupTypes, clientset, v1alpha1ClientSet, suite.T()); err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	// Set variables for the entire suite
	suite.ctx = ctx
	suite.setupTypes = setupTypes
	suite.v1alpha1ClientSet = v1alpha1ClientSet
	suite.clientset = clientset
	suite.tsm = tsm

	logger.Println("SetupSuite() ended!")
}

// TearDownSuite is called once after running the tests in the suite
func (suite *PostgresProvisioningHighAvailabilityTestSuite) TearDownSuite() {
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
	if err = suite.tsm.TearDown(suite.ctx, setupTypes, suite.clientset, suite.v1alpha1ClientSet, suite.T()); err != nil {
		logger.Printf("%s! %s\n", errBaseMsg, err)
		suite.T().FailNow()
	}

	logger.Println("HA TearDownSuite() completed!")
}

// This will run right before the test starts and receives the suite and test names as input
func (suite *PostgresProvisioningHighAvailabilityTestSuite) BeforeTest(suiteName, testName string) {
	util.GetLogger(suite.ctx).Printf("******************** RUNNING HA TEST %s %s ********************\n", suiteName, testName)
}

// This will run after test finishes and receives the suite and test names as input
func (suite *PostgresProvisioningHighAvailabilityTestSuite) AfterTest(suiteName, testName string) {
	util.GetLogger(suite.ctx).Printf("******************** END HA TEST %s %s ********************\n", suiteName, testName)
}

// Tests if provisioning is succesful by checking if database status is 'READY'
func (suite *PostgresProvisioningHighAvailabilityTestSuite) TestProvisioningSuccess() {
	logger := util.GetLogger(suite.ctx)

	databaseResponse, err := suite.tsm.GetDatabaseOrCloneResponse(suite.ctx, suite.setupTypes, suite.clientset, suite.v1alpha1ClientSet)
	if err != nil {
		logger.Printf("Error: TestProvisioningSuccess() failed! %v", err)
	} else {
		logger.Println("Database response retrieved.")
	}

	assert := assert.New(suite.T())
	assert.Equal(common.DATABASE_CR_STATUS_READY, databaseResponse.Status, "The database status should be ready.")
}

// Tests if app is able to connect to database via GET request
func (suite *PostgresProvisioningHighAvailabilityTestSuite) TestAppConnectivity() {
	logger := util.GetLogger(suite.ctx)

	resp, err := suite.tsm.GetAppResponse(suite.ctx, suite.setupTypes, suite.clientset, automation.POSTGRES_SI_PROVISONING_LOCAL_PORT)
	if err != nil {
		logger.Printf("Error: TestAppConnectivity failed! %v", err)
	} else {
		logger.Println("App response retrieved.")
	}

	assert := assert.New(suite.T())
	assert.Equal(200, resp.StatusCode, "The response status should be 200.")
}

// Tests if creation of time machine is succesful
func (suite *PostgresProvisioningHighAvailabilityTestSuite) TestTimeMachineSuccess() {
	logger := util.GetLogger(suite.ctx)
	assert := assert.New(suite.T())

	if suite.setupTypes.Database.Spec.Instance.TMInfo.SLAName == "" || suite.setupTypes.Database.Spec.Instance.TMInfo.SLAName == "NONE" {
		logger.Println("No time machine specified, test automatically passing.")
		return
	}

	tm, err := suite.tsm.GetTimemachineResponseByDatabaseId(suite.ctx, suite.setupTypes, suite.clientset, suite.v1alpha1ClientSet)
	if err != nil {
		logger.Printf("Error: TestTimeMachineSuccess() failed! %v", err)
		assert.FailNow("Error: TestTimeMachineSuccess() failed! %v", err)
	} else {
		logger.Println("Timemachine response retrieved.")
	}

	err = util.CheckTmInfo(suite.ctx, suite.setupTypes.Database, tm)
	if err != nil {
		logger.Printf("Error: TestTimeMachineSuccess() failed! %v", err)
		assert.FailNow("Error: TestTimeMachineSuccess() failed! %v", err)
	} else {
		logger.Println("CheckTmInfo succesful")
	}

	assert.Equal(common.DATABASE_CR_STATUS_READY, tm.Status, "The tm status should be ready.")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPostgresProvisioningHighAvailabilityTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresProvisioningHighAvailabilityTestSuite))
}

// BeforeTestLogTime will run right before the test starts and logs the start time of the test.
func (suite *PostgresProvisioningHighAvailabilityTestSuite) BeforeTestLogTime(suiteName, testName string) {
	logger := util.GetLogger(suite.ctx)
	startTime := time.Now()
	// Store the start time in the context for use in AfterTestLogTime
	ctx := context.WithValue(suite.ctx, "startTime", startTime)
	suite.ctx = ctx
	logger.Printf("******************** STARTING HA TEST %s %s at %v ********************\n", suiteName, testName, startTime)
}

// AfterTestLogTime will run after the test finishes and calculates the duration of the test.
func (suite *PostgresProvisioningHighAvailabilityTestSuite) AfterTestLogTime(suiteName, testName string) {
	logger := util.GetLogger(suite.ctx)
	startTime := suite.ctx.Value("startTime").(time.Time)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	logger.Printf("******************** ENDING HA TEST %s %s at %v (Duration: %v) ********************\n", suiteName, testName, endTime, duration)
}
