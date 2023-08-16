package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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
func SetupLogger(path string) (*log.Logger, error) {

	// Deletes the old logging file if it exists
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path)
	}

	// Creates the file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Links the logger to the file and returns the logger
	return log.New(file, "pg-si: ", log.Ldate|log.Ltime|log.Lshortfile), nil
}

// Gets logger form context
func GetLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(loggerKey).(*log.Logger)
	if !ok {
		return log.New(os.Stdout, "", log.Ldate|log.Ltime)
	}
	return logger
}

// Setup kubeconfig
func SetupKubeconfig(ctx context.Context) (config *rest.Config, err error) {
	logger := GetLogger(ctx)
	logger.Println("SetupKubeconfig() started...")

	logger.Println("Looking up environment variable 'KUBECONFIG'...")
	kubeconfig, ok := os.LookupEnv("KUBECONFIG")
	if ok {
		logger.Printf("Using configuration from '%s'\n", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		logger.Println("Using in-cluster configuration")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("Error: SetupKubeconfig() ended! %s", err)
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
		return nil, nil, fmt.Errorf("Error: SetupSchemeAndClientSet() ended! %s", err)
	}
	logger.Printf("Created v1alpha1Client.")

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: SetupSchemeAndClientSet() ended! %s", err)
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

	// Create dbSecret template from automation.DB_SECRET_PATH
	dbSecret := &v1.Secret{}
	if err := CreateTypeFromPath(dbSecret, DB_SECRET_PATH); err != nil {
		errMsg += fmt.Sprintf("DbSecret with path %s failed! %v. ", DB_SECRET_PATH, err)
	} else {
		logMsg += fmt.Sprintf("DbSecret with path %s created. ", DB_SECRET_PATH)
	}

	// Create ndbSecret template automation.NDB_SECRET_PATH
	ndbSecret := &v1.Secret{}
	if err := CreateTypeFromPath(ndbSecret, NDB_SECRET_PATH); err != nil {
		errMsg += fmt.Sprintf("NdbSecret with path %s failed! %v. ", NDB_SECRET_PATH, err)
	} else {
		logMsg += fmt.Sprintf("NdbSecret with path %s created. ", NDB_SECRET_PATH)
	}

	// Create database template from automation.DATABASE_PATH
	database := &ndbv1alpha1.Database{}
	if err := CreateTypeFromPath(database, DATABASE_PATH); err != nil {
		errMsg += fmt.Sprintf("Database with path %s failed! %v. ", DATABASE_PATH, err)
	} else {
		logMsg += fmt.Sprintf("database with path %s created. ", DATABASE_PATH)
	}

	// Create appPod template from automation.APP_POD_PATH
	appPod := &v1.Pod{}
	if err := CreateTypeFromPath(appPod, APP_POD_PATH); err != nil {
		errMsg += fmt.Sprintf("App Pod with path %s failed! %v. ", APP_POD_PATH, err)
	} else {
		logMsg += fmt.Sprintf("App Pod with path %s created. ", APP_POD_PATH)
	}

	// Create appSvc template from automation.APP_SVC_PATH
	appSvc := &v1.Service{}
	if err := CreateTypeFromPath(appSvc, APP_SVC_PATH); err != nil {
		errMsg += fmt.Sprintf("App Service with path %s failed! %v. ", APP_SVC_PATH, err)
	} else {
		logMsg += fmt.Sprintf("App Service with path %s created. ", APP_SVC_PATH)
	}

	// The yaml types
	setupTypes = &SetupTypes{
		DbSecret:  dbSecret,
		NdbSecret: ndbSecret,
		Database:  database,
		AppPod:    appPod,
		AppSvc:    appSvc,
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

// Yaml types used for testing
type SetupTypes struct {
	DbSecret  *corev1.Secret
	NdbSecret *corev1.Secret
	Database  *ndbv1alpha1.Database
	AppPod    *corev1.Pod
	AppSvc    *corev1.Service
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
