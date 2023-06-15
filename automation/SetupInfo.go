package automation

import (
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type SetupInfo struct {
	dbSecretPath  string
	ndbSecretPath string
	dbPath        string
	appPodPath    string
	appSvcPath    string
}

func (i *SetupInfo) getDbSecret() *corev1.Secret {
	dbSecretbytes, _ := ReadYAMLFile(i.dbSecretPath)
	dbSecret, _ := ConvertToSecret(dbSecretbytes)
	return dbSecret
}

func (i *SetupInfo) getNdbSecret() *corev1.Secret {
	ndbSecretbytes, _ := ReadYAMLFile(i.ndbSecretPath)
	ndbSecret, _ := ConvertToSecret(ndbSecretbytes)
	return ndbSecret
}

func (i *SetupInfo) getDatabase() *ndbv1alpha1.Database {
	databaseBytes, _ := ReadYAMLFile(i.dbPath)
	database, _ := ConvertToDatabase(databaseBytes)
	return database
}

func (i *SetupInfo) getAppPod() *corev1.Pod {
	appPodBytes, _ := ReadYAMLFile(i.appPodPath)
	appPod, _ := ConvertToPod(appPodBytes)
	return appPod
}

func (i *SetupInfo) getAppService() *corev1.Service {
	appSvcBytes, _ := ReadYAMLFile(i.appSvcPath)
	appSvc, _ := ConvertToService(appSvcBytes)
	return appSvc
}
