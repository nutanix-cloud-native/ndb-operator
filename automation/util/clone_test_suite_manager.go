package util

import (
	"context"
	"net/http"
	"testing"

	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func (cm *CloneTestSuiteManager) Setup(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.Setup() starting...")

	err = ProvisionOrClone(ctx, st, clientset, v1alpha1ClientSet, t)

	logger.Println("CloneTestSuiteManager.Setup() ended!")

	return
}

func (cm *CloneTestSuiteManager) TearDown(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.TearDown() starting...")

	err = DeprovisionOrDeclone(ctx, st, clientset, v1alpha1ClientSet, t)

	logger.Println("CloneTestSuiteManager.TearDown() ended!")

	return
}

func (cm *CloneTestSuiteManager) GetDatabaseOrCloneResponse(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (databaseResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.GetDatabaseResponse() starting...")

	databaseResponse, err = GetDatabaseOrCloneResponse(ctx, clientset, v1alpha1ClientSet, st)

	logger.Println("CloneTestSuiteManager.GetDatabaseResponse() ended!")

	return
}

func (cm *CloneTestSuiteManager) GetAppResponse(ctx context.Context, clientset *kubernetes.Clientset, pod *corev1.Pod, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("CloneTestSuiteManager.GetAppResponse() starting...")

	res, err = GetAppResponse(ctx, clientset, pod, localPort)

	logger.Println("CloneTestSuiteManager.GetAppResponse() ended!")

	return
}

// EMPTY STUB
func (dm *CloneTestSuiteManager) GetTimemachineResponseByDatabaseId(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (timemachineResponse ndb_api.TimeMachineResponse, err error) {
	return ndb_api.TimeMachineResponse{}, nil
}
