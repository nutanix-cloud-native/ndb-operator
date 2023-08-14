package automation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"sigs.k8s.io/yaml"
)

// Used in TestSuites to create resource objects
type SetupPaths struct {
	DbPath        string
	DbSecretPath  string
	NdbSecretPath string
	AppPodPath    string
	AppSvcPath    string
}

// Reads a file path and stores in bytes representation
func readYAMLFile(path string) ([]byte, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read YAML file: %v", err)
	}

	return yamlFile, nil
}

// CreateTypeFromPath reads a file path, converts it to json, and unmarshals json to a pointer.
// Ensure that theType is a pointer.
func CreateTypeFromPath(theType any, path string) (err error) {
	if theType == nil {
		return errors.New("theType is nil!")
	}

	// Check if theType is not a pointer
	if reflect.ValueOf(theType).Kind() != reflect.Ptr {
		return errors.New("theType is not a pointer!")
	}

	// Reads file path
	data, err := readYAMLFile(path)
	if err != nil {
		return err
	}

	// Converts byte data to json
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	// Creates 'type' object by unmarshalling data
	err = json.Unmarshal(jsonData, theType)
	if err != nil {
		return err
	}

	return nil
}

// Performs an operation a certain number of times with a given interval
func waitAndRetryOperation(interval time.Duration, retries int, operation func() error) (err error) {
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
