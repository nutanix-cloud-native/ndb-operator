package util

import (
	"context"
	"net/http"
	"testing"

	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"k8s.io/client-go/kubernetes"
)

func (cm *CloningTestSuiteManager) Setup(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.Setup() starting...")

	err = provisionOrClone(ctx, st, clientset, v1alpha1ClientSet, t)

	logger.Println("CloneTestSuiteManager.Setup() ended!")

	return
}

func (cm *CloningTestSuiteManager) TearDown(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.TearDown() starting...")

	err = deprovisionOrDeclone(ctx, st, clientset, v1alpha1ClientSet, t)

	logger.Println("CloneTestSuiteManager.TearDown() ended!")

	return
}

func (cm *CloningTestSuiteManager) GetDatabaseOrCloneResponse(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (databaseResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.GetDatabaseResponse() starting...")

	databaseResponse, err = getDatabaseOrCloneResponse(ctx, st, clientset, v1alpha1ClientSet)

	logger.Println("CloneTestSuiteManager.GetDatabaseResponse() ended!")

	return
}

func (cm *CloningTestSuiteManager) GetAppResponse(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.GetAppResponse() starting...")

	res, err = getAppResponse(ctx, st, clientset, localPort)

	logger.Println("CloneTestSuiteManager.GetAppResponse() ended!")

	return
}

// EMPTY STUB
func (dm *CloningTestSuiteManager) GetTimemachineResponseByDatabaseId(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (timemachineResponse ndb_api.TimeMachineResponse, err error) {
	return ndb_api.TimeMachineResponse{}, nil
}
