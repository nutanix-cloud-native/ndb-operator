package util

import (
	"context"
	"net/http"
	"testing"

	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"k8s.io/client-go/kubernetes"
)

func GetTestSuiteManager(ctx context.Context, st SetupTypes) (manager TestSuiteManager) {
	logger := GetLogger(ctx)
	logger.Println("GetTestSuiteManager() starting...")

	if st.Database.Spec.IsClone {
		logger.Println("CloneTestSuiteManage() retrieved!")
		manager = &CloningTestSuiteManager{}
	} else {
		logger.Println("DatabaseTestSuiteManager() retrieved!")
		manager = &ProvisioningTestSuiteManager{}
	}
	return
}

type TestSuiteManager interface {
	Setup(
		ctx context.Context,
		st *SetupTypes,
		clientset *kubernetes.Clientset,
		v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client,
		t *testing.T) (err error)
	TearDown(
		ctx context.Context,
		st *SetupTypes,
		clientset *kubernetes.Clientset,
		v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client,
		t *testing.T) (err error)
	GetDatabaseOrCloneResponse(
		ctx context.Context,
		st *SetupTypes,
		clientset *kubernetes.Clientset,
		v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (databaseResponse ndb_api.DatabaseResponse, err error)
	GetAppResponse(
		ctx context.Context,
		st *SetupTypes,
		clientset *kubernetes.Clientset,
		localPort string) (res http.Response, err error)
	GetTimemachineResponseByDatabaseId(
		ctx context.Context,
		st *SetupTypes,
		clientset *kubernetes.Clientset,
		v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (timemachineResponse ndb_api.TimeMachineResponse, err error)
}

type CloningTestSuiteManager struct{}

type ProvisioningTestSuiteManager struct{}
