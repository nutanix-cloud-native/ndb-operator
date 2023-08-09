package postgressi

// Basic imports
import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/automation"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// A test suite is a collection of related test cases that are grouped together for testing a specific package or functionality.
// The testify package builds on top of Go's built-in testing package and enhances it with additional features like assertions and test suite management.
// PostgresqlSingleInstanceTestSuite is a test suite struct that embeds testify's suite.Suite
type PostgresqlSingleInstanceTestSuite struct {
	suite.Suite
	config            *rest.Config
	v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	clientset         *kubernetes.Clientset
	kubeconfig        string
	logFile           string
	setupPath         automation.SetupPaths
}

// SetupSuite is called once before running the tests in the suite
func (suite *PostgresqlSingleInstanceTestSuite) SetupSuite() {
	var err error
	var config *rest.Config
	var v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	var clientset *kubernetes.Clientset
	kubeconfig := os.Getenv("KUBECONFIG")
	logFile := "./pg-si-test-suite.log"

	// Setup output log file
	if _, err := os.Stat(logFile); err == nil {
		_ = os.Remove(logFile)
	}
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	log.Printf("******************** RUNNING PostgresqlSingleInstanceTestSuite SETUPSUITE() ********************\n")

	// Setup kubeconfig
	if kubeconfig == "" {
		log.Println("Using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("Using configuration from '%s'\n", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		log.Printf("Error: %s\n", err)
		suite.T().FailNow()
	}

	// Setup scheme and clientsets
	ndbv1alpha1.AddToScheme(scheme.Scheme)
	v1alpha1ClientSet, err = clientsetv1alpha1.NewForConfig(config)
	if err != nil {
		log.Printf("Error: %s\n", err)
		suite.T().FailNow()
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error: %s\n", err)
		suite.T().FailNow()
	}

	// Create base setup for all tests in this suite
	setupPaths := automation.SetupPaths{
		DbPath:        "./config/db-pg-si.yaml",
		DbSecretPath:  "./config/db-secret-pg-si.yaml",
		NdbSecretPath: "./config/ndb-secret-pg-si.yaml",
		AppPodPath:    "./config/pod-pg-si.yaml",
		AppSvcPath:    "./config/svc-pg-si.yaml",
	}

	// Create dbSecret template from setupPaths.dbSecretPath
	dbSecret := &v1.Secret{}
	if err := automation.CreateTypeFromPath(dbSecret, setupPaths.DbSecretPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for dbSecret with path %s failed! %v\n", setupPaths.DbSecretPath, err)
		suite.T().FailNow()
	}

	// Create ndbSecret template from setupPaths.ndbSecretPath
	ndbSecret := &v1.Secret{}
	if err := automation.CreateTypeFromPath(ndbSecret, setupPaths.NdbSecretPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for ndbSecret with path %s failed! %v\n", setupPaths.NdbSecretPath, err)
		suite.T().FailNow()
	}

	// Create database template from setupPaths.dbPath
	database := &ndbv1alpha1.Database{}
	if err := automation.CreateTypeFromPath(database, setupPaths.DbPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for database with path %s failed! %v\n", setupPaths.DbPath, err)
		suite.T().FailNow()
	}

	// Create appPod template from setupPaths.appPodPath
	appPod := &v1.Pod{}
	if err := automation.CreateTypeFromPath(appPod, setupPaths.AppPodPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for pod with path %s failed! %v\n", setupPaths.AppPodPath, err)
		suite.T().FailNow()
	}

	// Create appSvc template from setupPaths.appSvcPath
	appSvc := &v1.Service{}
	if err := automation.CreateTypeFromPath(appSvc, setupPaths.AppSvcPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for service with path %s failed! %v\n", setupPaths.AppSvcPath, err)
		suite.T().FailNow()
	}

	// Create resources, wait for db to be ready, and pod to start
	if err := automation.TestSetup(dbSecret, ndbSecret, database, appPod, appSvc, clientset, v1alpha1ClientSet, suite.T()); err != nil {
		log.Printf(err.Error())
		log.Printf("******************** FAILED PostgresqlSingleInstanceTestSuite SETUPSUITE() ********************\n")
		suite.T().FailNow()
	}

	// Set variables for the entire suite
	suite.kubeconfig = kubeconfig
	suite.logFile = logFile
	suite.v1alpha1ClientSet = v1alpha1ClientSet
	suite.clientset = clientset
	suite.config = config
	suite.setupPath = setupPaths

	log.Printf("******************** END PostgresqlSingleInstanceTestSuite SETUPSUITE() ********************\n")
}

// TearDownSuite is called once after running the tests in the suite
func (suite *PostgresqlSingleInstanceTestSuite) TearDownSuite() {
	log.Printf("******************** RUNNING PostgresqlSingleInstanceTestSuite TEARDOWNSUITE() ********************\n")

	var err error
	setupPaths := suite.setupPath

	// Create dbSecret template from dbSecretPath
	dbSecret := &v1.Secret{}
	if err = automation.CreateTypeFromPath(dbSecret, setupPaths.DbSecretPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for dbSecret with path %s failed! %v\n", setupPaths.DbSecretPath, err)
		suite.T().FailNow()
	}

	// Create ndbSecret template from setupPaths.NdbSecretPath
	ndbSecret := &v1.Secret{}
	if err = automation.CreateTypeFromPath(ndbSecret, setupPaths.NdbSecretPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for ndbSecret with path %s failed! %v\n", setupPaths.NdbSecretPath, err)
		suite.T().FailNow()
	}

	// Create database template from setupPaths.DbPath
	database := &ndbv1alpha1.Database{}
	if err = automation.CreateTypeFromPath(database, setupPaths.DbPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for database with path %s failed! %v\n", setupPaths.DbPath, err)
		suite.T().FailNow()
	}

	// Create appPod template from setupPaths.appPodPath
	appPod := &v1.Pod{}
	if err = automation.CreateTypeFromPath(appPod, setupPaths.AppPodPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for pod with path %s failed! %v\n", setupPaths.AppPodPath, err)
		suite.T().FailNow()
	}

	// Create appSvc template from setupPaths.appSvc
	appSvc := &v1.Service{}
	if err = automation.CreateTypeFromPath(appSvc, setupPaths.AppSvcPath); err != nil {
		log.Printf("Error: utils.CreateTypeFromPath() for service with path %s failed! %v\n", setupPaths.AppSvcPath, err)
		suite.T().FailNow()
	}

	// Delete resources and de-provision database
	if err = automation.TestTeardown(dbSecret, ndbSecret, database, appPod, appSvc, suite.clientset, suite.v1alpha1ClientSet, suite.T()); err != nil {
		log.Printf(err.Error())
		log.Printf("******************** FAILED PostgresqlSingleInstanceTestSuite TEARDOWNSUITE() ********************\n")
		suite.T().FailNow()
	}

	suite.v1alpha1ClientSet.Databases(database.Namespace)

	log.Printf("******************** END PostgresqlSingleInstanceTestSuite() ********************\n")
}

// This will run right before the test starts and receives the suite and test names as input
func (suite *PostgresqlSingleInstanceTestSuite) BeforeTest(suiteName, testName string) {
	log.Printf("******************** RUNNING TEST %s %s ********************\n", suiteName, testName)
}

// This will run after test finishes and receives the suite and test names as input
func (suite *PostgresqlSingleInstanceTestSuite) AfterTest(suiteName, testName string) {
	log.Printf("******************** END TEST %s %s ********************\n", suiteName, testName)
}

// Tests if provisioning is succesful by checking if database stauts is 'READY'
func (suite *PostgresqlSingleInstanceTestSuite) TestProvisioningSuccess() {
	log.Printf("Start TestProvisioningSuccess()...\n")

	database := &v1alpha1.Database{}
	err := automation.CreateTypeFromPath(database, suite.setupPath.DbPath)

	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "utils.CreateTypeFromPath()", err)
		suite.T().FailNow()
	}
	database, err = suite.v1alpha1ClientSet.Databases(database.Namespace).Get(database.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("error while fetching database CR: %s\n", err)
		suite.T().FailNow()
	}
	ndb_secret_name := database.Spec.NDB.CredentialSecret
	secret, err := suite.clientset.CoreV1().Secrets(database.Namespace).Get(context.TODO(), ndb_secret_name, metav1.GetOptions{})
	username, password := string(secret.Data[common.SECRET_DATA_KEY_USERNAME]), string(secret.Data[common.SECRET_DATA_KEY_PASSWORD])
	if err != nil || username == "" || password == "" {
		log.Printf("error while fetching data from secret: %s\n", err)
		suite.T().FailNow()
	}
	ndbClient := ndb_client.NewNDBClient(username, password, database.Spec.NDB.Server, "", true)
	databaseResponse, err := ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	if err != nil {
		log.Printf("error while getting database response from ndb_api: %s\n", err)
		suite.T().FailNow()
	}
	log.Printf("Database response.status: %s\n", databaseResponse.Status)
	assert := assert.New(suite.T())
	assert.Equal(common.DATABASE_CR_STATUS_READY, databaseResponse.Status, "The database status should be ready.")

	log.Printf("End TestProvisioningSuccess()!**\n")
}

// Tests if 'manavrajvanshinx/best-app:latest' is able to connect to database
func (suite *PostgresqlSingleInstanceTestSuite) TestAppConnectivity() {
	log.Printf("Start TestAppConnectivity() started. **\n")

	client := http.Client{}
	// Send GET request
	resp, err := client.Get("http://localhost:30000")
	if err != nil {
		log.Println("Error while performing GET:", err)
		suite.T().FailNow()
	}
	defer resp.Body.Close()
	log.Println("Response status:", string(resp.Status))
	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading response:", err)
		suite.T().FailNow()
	}
	// Print the response body
	log.Println("Response:", string(body))
	assert := assert.New(suite.T())
	assert.Equal(200, resp.StatusCode, "The response status should be 200.")

	log.Printf("End TestAppConnectivity()!\n")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPostgresqlSingleInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresqlSingleInstanceTestSuite))
}
