package util

import (
	"context"
	"net/http"
	"testing"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"k8s.io/client-go/kubernetes"
)

func GetTestSuiteManager(database ndbv1alpha1.Database) (manager TestSuiteManager) {
	if database.Spec.IsClone {
		manager = &CloneTestSuiteManager{}
	} else {
		manager = &DatabaseTestSuiteManager{}
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

type CloneTestSuiteManager struct{}

type DatabaseTestSuiteManager struct{}
