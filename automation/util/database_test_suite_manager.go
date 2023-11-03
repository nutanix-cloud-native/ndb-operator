package util

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (dm *DatabaseTestSuiteManager) Setup(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("DatabaseTestSuiteManager.Setup() starting...")

	err = ProvisionOrClone(ctx, st, clientset, v1alpha1ClientSet, t)

	logger.Println("DatabaseTestSuiteManager.Setup() ended!")

	return
}

func (dm *DatabaseTestSuiteManager) TearDown(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("DatabaseTestSuiteManager.TearDown() starting...")

	err = DeprovisionOrDeclone(ctx, st, clientset, v1alpha1ClientSet, t)

	logger.Println("DatabaseTestSuiteManager.TearDown() ended!")

	return
}

func (dm *DatabaseTestSuiteManager) GetDatabaseOrCloneResponse(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (databaseResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("DatabaseTestSuiteManager.GetDatabaseResponse() starting...")

	databaseResponse, err = GetDatabaseOrCloneResponse(ctx, clientset, v1alpha1ClientSet, st)

	logger.Println("DatabaseTestSuiteManager.GetDatabaseResponse() ended!")

	return
}

func (dm *DatabaseTestSuiteManager) GetAppResponse(ctx context.Context, clientset *kubernetes.Clientset, pod *corev1.Pod, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("DatabaseTestSuiteManager.GetAppResponse() starting...")

	res, err = GetAppResponse(ctx, clientset, pod, localPort)

	logger.Println("DatabaseTestSuiteManager.GetAppResponse() ended!")

	return
}

// Tests TM Response
func (dm *DatabaseTestSuiteManager) GetTimemachineResponseByDatabaseId(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (timemachineResponse ndb_api.TimeMachineResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetTimemachineResponse() starting...")
	errBaseMsg := "Error: GetTimemachineResponse() ended"

	// Get NDBServer CR
	ndbServer, err := v1alpha1ClientSet.NDBServers(st.NdbServer.Namespace).Get(st.NdbServer.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Could not fetch ndbServer '%s' CR! %s\n", errBaseMsg, ndbServer.Name, err)
	} else {
		logger.Printf("Retrieved ndbServer '%s' CR from v1alpha1ClientSet", ndbServer.Name)
	}

	// Get Database CR
	database, err := v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Could not fetch database '%s' CR! %s\n", errBaseMsg, database.Name, err)
	} else {
		logger.Printf("Retrieved database '%s' CR from v1alpha1ClientSet", database.Name)
	}

	// Get NDB username and password from NDB CredentialSecret
	ndb_secret_name := ndbServer.Spec.CredentialSecret
	secret, err := clientset.CoreV1().Secrets(database.Namespace).Get(context.TODO(), ndb_secret_name, metav1.GetOptions{})
	username, password := string(secret.Data[common.SECRET_DATA_KEY_USERNAME]), string(secret.Data[common.SECRET_DATA_KEY_PASSWORD])
	if err != nil || username == "" || password == "" {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Could not fetch data from secret! %s\n", errBaseMsg, err)
	}

	// Create ndbClient and getting databaseResponse so we can get timemachine id
	ndbClient := ndb_client.NewNDBClient(username, password, ndbServer.Spec.Server, "", true)
	databaseResponse, err := ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! Database response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	// Get timemachine
	timemachineResponse, err = ndb_api.GetTimeMachineById(context.TODO(), ndbClient, databaseResponse.TimeMachineId)
	if err != nil {
		return ndb_api.TimeMachineResponse{}, fmt.Errorf("%s! time machine response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	logger.Printf("Timemachine response.status: %s.\n", timemachineResponse.Status)
	logger.Println("GetTimemachineResponse() ended!")

	return timemachineResponse, nil
}
