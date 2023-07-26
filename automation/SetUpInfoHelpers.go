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

// Reads and stores bytes of yaml file
func ReadYAMLFile(filename string) ([]byte, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read YAML file: %v", err)
	}

	return yamlFile, nil
}

// Converts bytes to Databaase, secret, pod, or service type
func ConvertBytesToType(data []byte, t string) (typ interface{}, err error) {
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
		err = fmt.Errorf("Unrecognized type: %s", t)
		return
	}
	err = json.Unmarshal(jsonData, typ)
	if err != nil {
		log.Println(err)
	}
	return
}

func ConvertBytesToSecret(data []byte) (secret *corev1.Secret, err error) {
	typ, err := ConvertBytesToType(data, "secret")
	secret = typ.(*corev1.Secret)
	return
}

func ConvertBytesToDatabase(data []byte) (database *ndbv1alpha1.Database, err error) {
	typ, err := ConvertBytesToType(data, "database")
	database = typ.(*ndbv1alpha1.Database)
	return
}

func ConvertBytesToPod(data []byte) (pod *corev1.Pod, err error) {
	typ, err := ConvertBytesToType(data, "pod")
	pod = typ.(*corev1.Pod)
	return
}

func ConvertBytesToService(data []byte) (service *corev1.Service, err error) {
	typ, err := ConvertBytesToType(data, "service")
	service = typ.(*corev1.Service)
	return
}

func WaitAndRetryOperation(interval time.Duration, retries int, operation func() error) (err error) {
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
