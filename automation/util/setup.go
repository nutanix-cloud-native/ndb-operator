package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	}
	missingRequiredEnvs := []string{}
	for _, env := range requiredEnvs {
		if _, ok := os.LookupEnv(env); !ok {
			missingRequiredEnvs = append(missingRequiredEnvs, env)
		}
	}
	// Check that at least one of NX_CLUSTER_ID or NX_CLUSTER_NAME is provided
	nxClusterId := os.Getenv(automation.NX_CLUSTER_ID_ENV)
	nxClusterName := os.Getenv(automation.NX_CLUSTER_NAME_ENV)
	if nxClusterId == "" && nxClusterName == "" {
		missingRequiredEnvs = append(missingRequiredEnvs, fmt.Sprintf("%s or %s", automation.NX_CLUSTER_ID_ENV, automation.NX_CLUSTER_NAME_ENV))
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

// YAML Resource types
type SetupTypes struct {
	NdbServer *ndbv1alpha1.NDBServer
	Database  *ndbv1alpha1.Database
	NdbSecret *corev1.Secret
	DbSecret  *corev1.Secret
	AppPod    *corev1.Pod
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
