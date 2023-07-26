package automation

import (
	"log"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type SetupPath struct {
	dbSecretPath  string
	ndbSecretPath string
	dbPath        string
	appPodPath    string
	appSvcPath    string
}

func (sp *SetupPath) getDbSecret() (*corev1.Secret, error) {
	dbSecretbytes, err := ReadYAMLFile(sp.dbSecretPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", sp.dbSecretPath)
		return nil, err
	}
	dbSecret, err := ConvertBytesToSecret(dbSecretbytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to secret")
		return nil, err
	}
	return dbSecret, nil
}

func (sp *SetupPath) getNdbSecret() (*corev1.Secret, error) {
	ndbSecretbytes, err := ReadYAMLFile(sp.ndbSecretPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", sp.ndbSecretPath)
		return nil, err
	}
	ndbSecret, err := ConvertBytesToSecret(ndbSecretbytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to NdbSecret")
		return nil, err
	}
	return ndbSecret, nil
}

func (sp *SetupPath) getDatabase() (*ndbv1alpha1.Database, error) {
	databaseBytes, err := ReadYAMLFile(sp.dbPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", sp.dbPath)
		return nil, err
	}
	database, err := ConvertBytesToDatabase(databaseBytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to Database")
		return nil, err
	}
	return database, nil
}

func (sp *SetupPath) getAppPod() (*corev1.Pod, error) {
	appPodBytes, err := ReadYAMLFile(sp.appPodPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", sp.appPodPath)
		return nil, err
	}
	appPod, err := ConvertBytesToPod(appPodBytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to AppPod")
		return nil, err
	}
	return appPod, nil
}

func (sp *SetupPath) getAppService() (*corev1.Service, error) {
	appSvcBytes, err := ReadYAMLFile(sp.appSvcPath)
	if err != nil {
		log.Printf("Error occurred while reading bytes from %s", sp.appSvcPath)
		return nil, err
	}
	appSvc, err := ConvertBytesToService(appSvcBytes)
	if err != nil {
		log.Printf("Error occurred while converting bytes to AppService")
		return nil, err
	}
	return appSvc, nil
}
