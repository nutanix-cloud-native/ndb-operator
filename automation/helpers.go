package automation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func readYAMLFile(filename string) ([]byte, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %v", err)
	}

	return yamlFile, nil
}

func convertBytesToType(data []byte, t string) (typ interface{}, err error) {
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		log.Println(err)
		return
	}
	switch t {
	case "database":
		typ = &ndbv1alpha1.Database{}
	case "secret":
		typ = &corev1.Secret{}
	case "pod":
		typ = &corev1.Pod{}
	case "service":
		typ = &corev1.Service{}
	default:
		err = fmt.Errorf("unrecognized type: %s", t)
		return
	}
	err = json.Unmarshal(jsonData, typ)
	if err != nil {
		log.Println(err)
		return
	}
	return
}

// func ConvertToSecret(data []byte) (secret *corev1.Secret, err error) {
// 	secret = &corev1.Secret{}
// 	jsonData, err := yaml.YAMLToJSON(data)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = json.Unmarshal(jsonData, secret)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	return
// }

// func ConvertToDatabase(data []byte) (database *ndbv1alpha1.Database, err error) {
// 	database = &ndbv1alpha1.Database{}
// 	jsonData, err := yaml.YAMLToJSON(data)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = json.Unmarshal(jsonData, database)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	return
// }

// func ConvertToPod(data []byte) (pod *corev1.Pod, err error) {
// 	pod = &corev1.Pod{}
// 	jsonData, err := yaml.YAMLToJSON(data)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = json.Unmarshal(jsonData, pod)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	return
// }

// func ConvertToService(data []byte) (svc *corev1.Service, err error) {
// 	svc = &corev1.Service{}
// 	jsonData, err := yaml.YAMLToJSON(data)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = json.Unmarshal(jsonData, svc)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	return
// }

func ConvertToSecret(data []byte) (secret *corev1.Secret, err error) {
	typ, err := convertBytesToType(data, "secret")
	secret = typ.(*corev1.Secret)
	return
}

func ConvertToDatabase(data []byte) (database *ndbv1alpha1.Database, err error) {
	typ, err := convertBytesToType(data, "database")
	database = typ.(*ndbv1alpha1.Database)
	return
}

func ConvertToPod(data []byte) (pod *corev1.Pod, err error) {
	typ, err := convertBytesToType(data, "pod")
	pod = typ.(*corev1.Pod)
	return
}

func ConvertToService(data []byte) (service *corev1.Service, err error) {
	typ, err := convertBytesToType(data, "service")
	service = typ.(*corev1.Service)
	return
}

func waitAndRetryFunction(interval time.Duration, retries int, operation func() error) (err error) {

	for i := 0; i < retries; i++ {
		if i != 0 {
			log.Printf("Retrying, attempt # %d\n", i)
		}
		err = operation()
		if err == nil {
			return nil
		} else {
			log.Printf("Error: %s\n", err)
		}
		time.Sleep(interval)
	}
	// Operation failed after all retries, return the last error received
	return err
}
