package automation

import (
	"log"

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

func (i *SetupInfo) getDbSecret() (*corev1.Secret, error) {
	dbSecretbytes, err := readYAMLFile(i.dbSecretPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", i.dbSecretPath)
		return nil, err
	}
	dbSecret, err := ConvertToSecret(dbSecretbytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to secret")
		return nil, err
	}
	return dbSecret, nil
}

func (i *SetupInfo) getNdbSecret() (*corev1.Secret, error) {
	ndbSecretbytes, err := readYAMLFile(i.ndbSecretPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", i.ndbSecretPath)
		return nil, err
	}
	ndbSecret, err := ConvertToSecret(ndbSecretbytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to NdbSecret")
		return nil, err
	}
	return ndbSecret, nil
}

func (i *SetupInfo) getDatabase() (*ndbv1alpha1.Database, error) {
	databaseBytes, err := readYAMLFile(i.dbPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", i.dbPath)
		return nil, err
	}
	database, err := ConvertToDatabase(databaseBytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to Database")
		return nil, err
	}
	return database, nil
}

func (i *SetupInfo) getAppPod() (*corev1.Pod, error) {
	appPodBytes, err := readYAMLFile(i.appPodPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", i.appPodPath)
		return nil, err
	}
	appPod, err := ConvertToPod(appPodBytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to AppPod")
		return nil, err
	}
	return appPod, nil
}

func (i *SetupInfo) getAppService() (*corev1.Service, error) {
	appSvcBytes, err := readYAMLFile(i.appSvcPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", i.appSvcPath)
		return nil, err
	}
	appSvc, err := ConvertToService(appSvcBytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to AppService")
		return nil, err
	}
	return appSvc, nil
}
