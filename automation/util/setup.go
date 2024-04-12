package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/joho/godotenv"
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	automation "github.com/nutanix-cloud-native/ndb-operator/automation"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

type key int

const loggerKey key = iota

// Setup up Context with Logger
func SetupContext(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Setup a logger with a unique file path
func SetupLogger(path string, rootName string) (*log.Logger, error) {

	// Deletes the old logging file if it exists
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path)
	}

	// Get the directory of the path
	dir := filepath.Dir(path)

	// Create the directory and all parent directories if they do not exist
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	
	// Creates the file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	

	// Links the logger to the file and returns the logger
	return log.New(file, rootName, log.Ldate|log.Ltime|log.Lshortfile), nil
}

// Gets logger from context
func GetLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(loggerKey).(*log.Logger)
	if !ok {
		return log.New(os.Stdout, "", log.Ldate|log.Ltime)
	}
	return logger
}

// Check if required environment variables are present
func CheckRequiredEnv(ctx context.Context) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CheckRequiredEnv() started...")

	// Loading env variables
	err = godotenv.Load("../../.env")
	if err != nil {
		return fmt.Errorf("error: loadEnv() ended! %s", err)
	} else {
		logger.Print("Loaded .env file!")
	}

	logger.Print("Checking for missing required env variables...")
	requiredEnvs := []string{
		automation.DB_SECRET_PASSWORD_ENV,
		automation.NDB_SECRET_USERNAME_ENV,
		automation.NDB_SECRET_PASSWORD_ENV,
		automation.NDB_SERVER_ENV,
		automation.NX_CLUSTER_ID_ENV,
	}
	missingRequiredEnvs := []string{}
	for _, env := range requiredEnvs {
		if _, ok := os.LookupEnv(env); !ok {
			missingRequiredEnvs = append(missingRequiredEnvs, env)
		}
	}
	if len(missingRequiredEnvs) != 0 {
		return fmt.Errorf("error: loadEnv() ended! Missing the following required env variables: %s", missingRequiredEnvs)
	} else {
		logger.Print("Found no missing required env variables!")
	}

	logger.Println("CheckRequiredEnv() exited!")

	return nil
}

// Setup kubeconfig
func SetupKubeconfig(ctx context.Context) (config *rest.Config, err error) {
	logger := GetLogger(ctx)
	logger.Println("SetupKubeconfig() started...")

	logger.Println("Looking up environment variable 'KUBECONFIG'...")
	kubeconfig, ok := os.LookupEnv(automation.KUBECONFIG_ENV)
	if ok {
		logger.Printf("Using configuration from '%s'\n", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		logger.Println("Using in-cluster configuration")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("error: SetupKubeconfig() ended! %s", err)
	}

	logger.Println("SetupKubeconfig() ended!")

	return
}

// Setup scheme and clientsets
func SetupSchemeAndClientSet(ctx context.Context, config *rest.Config) (v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, clientset *kubernetes.Clientset, err error) {
	logger := GetLogger(ctx)
	logger.Println("SetupSchemeAndClientSet() started...")

	ndbv1alpha1.AddToScheme(scheme.Scheme)
	logger.Printf("Added scheme to ndbv1alpha1.")

	v1alpha1ClientSet, err = clientsetv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error: SetupSchemeAndClientSet() ended! %s", err)
	}
	logger.Printf("Created v1alpha1Client.")

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error: SetupSchemeAndClientSet() ended! %s", err)
	}
	logger.Printf("Created clientset.")

	logger.Println("SetupSchemeAndClientSet() ended!")

	return v1alpha1ClientSet, clientset, err
}

// Setup yaml types. Uses paths specified in automation.constants.go
func SetupTypeTemplates(ctx context.Context) (setupTypes *SetupTypes, err error) {
	logger := GetLogger(ctx)
	logger.Println("SetupTypeTemplates() started...")

	var logMsg string
	var errMsg string

	// Create ndbServer template from automation.NDBSERVER_PATH
	ndbServer := &ndbv1alpha1.NDBServer{}
	if err := CreateTypeFromPath(ndbServer, automation.NDBSERVER_PATH); err != nil {
		errMsg += fmt.Sprintf("NdbServer with path %s failed! %v. ", automation.NDBSERVER_PATH, err)
	} else {
		logMsg += fmt.Sprintf("NdbServer with path %s created. ", automation.NDBSERVER_PATH)
	}

	// Create database template from automation.DATABASE_PATH
	database := &ndbv1alpha1.Database{}
	if err := CreateTypeFromPath(database, automation.DATABASE_PATH); err != nil {
		errMsg += fmt.Sprintf("Database with path %s failed! %v. ", automation.DATABASE_PATH, err)
	} else {
		logMsg += fmt.Sprintf("Database with path %s created. ", automation.DATABASE_PATH)
	}

	// Create ndbSecret template automation.NDB_SECRET_PATH
	ndbSecret := &corev1.Secret{}
	if err := CreateTypeFromPath(ndbSecret, automation.NDB_SECRET_PATH); err != nil {
		errMsg += fmt.Sprintf("NdbSecret with path %s failed! %v. ", automation.NDB_SECRET_PATH, err)
	} else {
		logMsg += fmt.Sprintf("NdbSecret with path %s created. ", automation.NDB_SECRET_PATH)
	}

	// Create dbSecret template from automation.DB_SECRET_PATH
	dbSecret := &corev1.Secret{}
	if err := CreateTypeFromPath(dbSecret, automation.DB_SECRET_PATH); err != nil {
		errMsg += fmt.Sprintf("DbSecret with path %s failed! %v. ", automation.DB_SECRET_PATH, err)
	} else {
		logMsg += fmt.Sprintf("DbSecret with path %s created. ", automation.DB_SECRET_PATH)
	}

	// Create appPod template from automation.APP_POD_PATH
	appPod := &corev1.Pod{}
	if err := CreateTypeFromPath(appPod, automation.APP_POD_PATH); err != nil {
		errMsg += fmt.Sprintf("App Pod with path %s failed! %v. ", automation.APP_POD_PATH, err)
	} else {
		logMsg += fmt.Sprintf("App Pod with path %s created. ", automation.APP_POD_PATH)
	}

	setupTypes = &SetupTypes{
		NdbServer: ndbServer,
		Database:  database,
		DbSecret:  dbSecret,
		NdbSecret: ndbSecret,
		AppPod:    appPod,
	}

	if errMsg == "" {
		logger.Println(logMsg)
		err = nil
	} else {
		err = errors.New("Error: SetupResourceTypes ended! " + errMsg)
	}

	logger.Println("SetupTypeTemplates() ended!")

	return setupTypes, err
}

// YAML Resource types
type SetupTypes struct {
	NdbServer *ndbv1alpha1.NDBServer
	Database  *ndbv1alpha1.Database
	NdbSecret *corev1.Secret
	DbSecret  *corev1.Secret
	AppPod    *corev1.Pod
}

// CreateTypeFromPath reads a file path, converts it to json, and unmarshals json to a pointer.
// Ensure that theType is a pointer.
func CreateTypeFromPath(theType any, path string) (err error) {
	if theType == nil {
		return errors.New("theType is nil")
	}

	// Check if theType is not a pointer
	if reflect.ValueOf(theType).Kind() != reflect.Ptr {
		return errors.New("theType is not a pointer")
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

// Performs an operation a certain number of times with a given interval
func waitAndRetryOperation(ctx context.Context, interval time.Duration, retries int, operation func() error) (err error) {
	logger := GetLogger(ctx)
	logger.Println("waitAndRetryOperation() starting...")

	for i := 0; i < retries; i++ {
		if i != 0 {
			logger.Printf("Retrying, attempt # %d\n", i)
		}
		err = operation()
		if err == nil {
			return nil
		} else {
			logger.Printf("Error: %s\n", err)
		}
		time.Sleep(interval)
	}

	logger.Println("waitAndRetryOperation() ended!")

	// Operation failed after all retries, return the last error received
	return err
}
