package automation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"sigs.k8s.io/yaml"
)

func readYAMLFile(filename string) ([]byte, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read YAML file: %v", err)
	}

	return yamlFile, nil
}

// Reads a file path, converts to json, unmarshals to generic
func createGeneric(generic interface{}, path string) (err error) {
	if generic == nil {
		return err
	}

	data, err := readYAMLFile(path)
	if err != nil {
		return err
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &generic)
	if err != nil {
		return err
	}

	return nil
}

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
