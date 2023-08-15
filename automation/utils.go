package automation

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"

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

// CreateTypeFromPath reads a file path, converts it to json, and unmarshals json to a pointer.
// Ensure that theType is a pointer.
func CreateTypeFromPath(theType any, path string) (err error) {
	if theType == nil {
		return errors.New("theType is nil! Ensure you are passing in a non-nil value!")
	}

	// Check if theType is not a pointer
	if reflect.ValueOf(theType).Kind() != reflect.Ptr {
		return errors.New("theTyper is not a pointer! Ensure you are passing in a pointer for unmarshalling to work correctly!")
	}

	// Reads file path
	data, err := os.ReadFile(path)
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
