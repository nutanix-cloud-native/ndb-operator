package automation

// Basic imports
import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type PostgresqlSingleInstanceTestSuite struct {
	suite.Suite
	config            *rest.Config
	v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	clientset         *kubernetes.Clientset
	kubeconfig        string
	logFile           string
	setupPath         SetupPath
}

// before the suite
func (suite *PostgresqlSingleInstanceTestSuite) SetupSuite() {
	var err error
	var config *rest.Config
	var v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client
	var clientset *kubernetes.Clientset
	kubeconfig := os.Getenv("KUBECONFIG")
	logFile := "./PostgresqlSingleInstanceTestSuite.log"

	// Setup output log file
	if _, err := os.Stat(logFile); err == nil {
		_ = os.Remove(logFile)
	}
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	// Setup kubeconfig
	if kubeconfig == "" {
		log.Println("using in-cluster configuration")
		config, err = rest.InClusterConfig()
	} else {
		log.Printf("using configuration from '%s'\n", kubeconfig)
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
	setupInfo := SetupInfo{
		dbSecretPath:  "./files/db-secret-pg-si.yaml",
		ndbSecretPath: "./files/ndb-secret-pg-si.yaml",
		dbPath:        "./files/database-pg-si.yaml",
		appPodPath:    "./files/pod-pg-si.yaml",
		appSvcPath:    "./files/service-pg-si.yaml",
	}
	dbSecret, err := setupInfo.getDbSecret()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getDbSecret()", err)
		suite.T().FailNow()
	}
	ndbSecret, err := setupInfo.getNdbSecret()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getNdbSecret()", err)
		suite.T().FailNow()
	}
	database, err := setupInfo.getDatabase()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getDatabase()", err)
		suite.T().FailNow()
	}
	appSvc, err := setupInfo.getAppService()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getAppService()", err)
		suite.T().FailNow()
	}
	appPod, err := setupInfo.getAppPod()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getAppPod()", err)
		suite.T().FailNow()
	}
	err = test_setup(dbSecret, ndbSecret, database, appSvc, appPod, clientset, v1alpha1ClientSet, suite.T())
	if err != nil {
		log.Printf("Error occurred: %s\n", err)
		log.Println("Setup failed")
		suite.T().FailNow()
	} else {
		log.Println("Setup completed")
	}

	// Set variables for the entire suite
	suite.kubeconfig = kubeconfig
	suite.logFile = logFile
	suite.v1alpha1ClientSet = v1alpha1ClientSet
	suite.clientset = clientset
	suite.setupPath = setupInfo
	suite.config = config
}

func (suite *PostgresqlSingleInstanceTestSuite) TearDownSuite() {
	setupInfo := suite.setupPath
	dbSecret, err := setupInfo.getDbSecret()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getDbSecret()", err)
		suite.T().FailNow()
	}
	ndbSecret, err := setupInfo.getNdbSecret()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getNdbSecret()", err)
		suite.T().FailNow()
	}
	database, err := setupInfo.getDatabase()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getDatabase()", err)
		suite.T().FailNow()
	}
	appSvc, err := setupInfo.getAppService()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getAppService()", err)
		suite.T().FailNow()
	}
	appPod, err := setupInfo.getAppPod()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "setupInfo.getAppPod()", err)
		suite.T().FailNow()
	}
	err = test_teardown(dbSecret, ndbSecret, database, appSvc, appPod, suite.clientset, suite.v1alpha1ClientSet, suite.T())
	if err != nil {
		log.Printf("Error occurred: %s\n", err)
		log.Println("teardown failed")
	} else {
		log.Println("teardown completed")
	}

	suite.v1alpha1ClientSet.Databases(database.Namespace)
}
func (suite *PostgresqlSingleInstanceTestSuite) BeforeTest(suiteName, testName string) {
	log.Printf("******************** RUNNING TEST %s %s ********************\n", suiteName, testName)
}

func (suite *PostgresqlSingleInstanceTestSuite) AfterTest(suiteName, testName string) {
	log.Printf("******************** END TEST %s %s ********************\n", suiteName, testName)
}

func (suite *PostgresqlSingleInstanceTestSuite) TestProvisioningSuccess() {
	database, err := suite.setupPath.getDatabase()
	if err != nil {
		log.Printf("Error occurred while executing %s, err: %v\n", "suite.setupInfo.getDatabase()", err)
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
}

func (suite *PostgresqlSingleInstanceTestSuite) TestAppConnectivity() {
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
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPostgresqlSingleInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresqlSingleInstanceTestSuite))
}
