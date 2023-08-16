package automation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const namespace_default = "default"

// This function is called from the SetupSuite() function of all testsuites.
// It loads environment variables, instantiate resources, waits for db to be ready, and pod to start.
func ProvisioningTestSetup(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("ProvisioningTestSetup() starting...")

	// Nil check
	if st == nil || clientset == nil || v1alpha1ClientSet == nil {
		errMsg := "Error: ProvisioningTestSetup() starting ended! "
		if st == nil {
			errMsg += "st is nil! "
		}
		if clientset == nil {
			errMsg += "clientset is nil! "
		}
		if v1alpha1ClientSet == nil {
			errMsg += "v1alpha1ClientSet is nil! "
		}

		return errors.New(errMsg)
	}

	ns := namespace_default
	if st.Database != nil && st.Database.Namespace != "" {
		ns = st.Database.Namespace
	}

	// Create Secrets
	if st.DbSecret != nil {
		st.DbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("DB_SECRET_PASSWORD")
		_, err = clientset.CoreV1().Secrets(ns).Create(ctx, st.DbSecret, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating secret %s: %s\n", st.DbSecret.Name, err)
		} else {
			logger.Printf("DB Secret %s created.\n", st.DbSecret.Name)
		}
	}
	if st.NdbSecret != nil {
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("NDB_SECRET_USERNAME")
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("NDB_SECRET_PASSWORD")
		_, err = clientset.CoreV1().Secrets(ns).Create(context.TODO(), st.NdbSecret, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating secret %s: %s\n", st.NdbSecret.Name, err)
		} else {
			logger.Printf("NDB Secret %s created.\n", st.NdbSecret.Name)
		}
	}

	// Create Database
	if st.Database != nil {
		st.Database.Spec.NDB.Server = os.Getenv("NDB_SERVER")
		st.Database.Spec.NDB.ClusterId = os.Getenv("NDB_CLUSTER_ID")
		st.Database, err = v1alpha1ClientSet.Databases(st.Database.Namespace).Create(st.Database)
		if err != nil {
			logger.Printf("Error while creating Database %s: %s\n", st.Database.Name, err)
		} else {
			logger.Printf("Database %s created.\n", st.Database.Name)
		}
	}

	// Create Application
	if st.AppPod != nil {
		st.AppPod, err = clientset.CoreV1().Pods(ns).Create(context.TODO(), st.AppPod, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating Pod %s: %s\n", st.AppPod.Name, err)
		} else {
			logger.Printf("Pod %s created.\n", st.AppPod.Name)
		}
	}
	if st.AppSvc != nil {
		st.AppSvc, err = clientset.CoreV1().Services(ns).Create(context.TODO(), st.AppSvc, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating Svc %s: %s\n", st.AppSvc.Name, err)
		} else {
			logger.Printf("Svc %s created.\n", st.AppSvc.Name)
		}
	}

	// Wait for DB to get Ready
	if st.Database != nil {
		err = waitAndRetryOperation(ctx, time.Minute, 60, func() (err error) {
			st.Database, err = v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
			if err != nil {
				return
			}
			statusMessage := "DB " + st.Database.Name + " is in '" + st.Database.Status.Status + "' status."
			if st.Database.Status.Status == common.DATABASE_CR_STATUS_READY {
				logger.Println(statusMessage)
				return
			}
			err = errors.New(statusMessage)
			return
		})
		if err == nil {
			logger.Println("Database is ready")
		} else {
			logger.Println(err)
		}
	}
	// Wait for Application Pod to start
	if st.AppPod != nil {
		err = waitAndRetryOperation(ctx, time.Second, 300, func() (err error) {
			st.AppPod, err = clientset.CoreV1().Pods(ns).Get(context.TODO(), st.AppPod.Name, metav1.GetOptions{})
			if err != nil {
				return
			}
			statusMessage := "Pod " + st.AppPod.Name + " is in '" + string(st.AppPod.Status.Phase) + "' status."
			if st.AppPod.Status.Phase == "Running" {
				logger.Println(statusMessage)
				return
			}
			err = errors.New(statusMessage)
			return
		})
		if err == nil {
			logger.Println("Pod is ready")
		} else {
			logger.Println(err)
			return
		}
	}

	logger.Println("test_setup() ended.")

	return
}

// This function is called from the TeardownSuite() function of all testsuites.
// Delete resources and de-provision database.
func ProvisioningTestTeardown(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("ProvisioningTestTeardown() starting...")

	ns := namespace_default
	if st.Database != nil && st.Database.Namespace != "" {
		ns = st.Database.Namespace
	}

	// Delete Database
	if st.Database != nil {
		st.Database.Spec.NDB.Server = os.Getenv("NDB-SERVER")
		st.Database.Spec.NDB.ClusterId = os.Getenv("NDB-CLUSTER-ID")
		err := v1alpha1ClientSet.Databases(st.Database.Namespace).Delete(st.Database.Name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting Database %s: %s\n", st.Database.Name, err)
		} else {
			logger.Printf("Database %s deleted\n", st.Database.Name)
		}
		waitAndRetryOperation(ctx, time.Minute, 10, func() (err error) {
			st.Database, err = v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
			if err != nil {
				return nil
			}
			if (st.Database == &ndbv1alpha1.Database{}) {
				logger.Println("Received empty database")
				return nil
			}
			statusMessage := "DB " + st.Database.Name + " is not yet deleted"
			logger.Println(statusMessage)
			err = errors.New(statusMessage)
			return
		})
	}

	// Delete Secrets
	if st.DbSecret != nil {
		st.DbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("DB-SECRET-USERNAME")
		st.DbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("DB-SECRET-PASSWORD")
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), st.DbSecret.Name, metav1.DeleteOptions{})
		if err == nil {
			logger.Printf("Secret %s deleted\n", st.DbSecret.Name)
		} else {
			logger.Printf("Error while deleting secret %s: %s\n", st.DbSecret.Name, err)
		}
	}
	if st.NdbSecret != nil {
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("NDB-SECRET-USERNAME")
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("NDB-SECRET-PASSWORD")
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), st.NdbSecret.Name, metav1.DeleteOptions{})
		if err == nil {
			logger.Printf("Secret %s deleted\n", st.NdbSecret.Name)
		} else {
			logger.Printf("Error while deleting secret %s: %s\n", st.NdbSecret.Name, err)
		}
	}

	// Delete Application
	if st.AppPod != nil {
		err := clientset.CoreV1().Pods(ns).Delete(context.TODO(), st.AppPod.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting Pod %s: %s\n", st.AppPod.Name, err)
		} else {
			logger.Printf("Pod %s deleted\n", st.AppPod.Name)
		}
	}
	if st.AppSvc != nil {
		err = clientset.CoreV1().Services(ns).Delete(context.TODO(), st.AppSvc.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting Svc %s: %s\n", st.AppSvc.Name, err)
		} else {
			logger.Printf("Svc %s deleted\n", st.AppSvc.Name)
		}
	}

	logger.Println("ProvisioningTestTeardown() ended!")

	return
}

// Wrapper function called in all TestSuite GetDatabaseResponse methods. Returns a DatabaseResponse which indicates if provison was succesful
func GetDatabaseResponseFromCR(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (databaseResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetDatabaseResponseFromCR() starting...")

	// Get db template from yaml to acquire database name
	database := &v1alpha1.Database{}
	err = CreateTypeFromPath(database, DATABASE_PATH)
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("Error: GetDatabaseResponseFromCR() ended! Database with path %s failed! %v. ", DATABASE_PATH, err)
	} else {
		logger.Printf("Database with path %s created.", DATABASE_PATH)
	}

	// Get database CR from above database name
	database, err = v1alpha1ClientSet.Databases(database.Namespace).Get(database.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("Error: GetDatabaseResponseFromCR() ended! Could not fetch database '%s' CR! %s\n", database.Name, err)
	} else {
		logger.Printf("Retrieved database '%s' CR from v1alpha1ClientSet", database.Name)
	}

	// Get NDB username and password from NDB CredentialSecret
	ndb_secret_name := database.Spec.NDB.CredentialSecret
	secret, err := clientset.CoreV1().Secrets(database.Namespace).Get(context.TODO(), ndb_secret_name, metav1.GetOptions{})
	username, password := string(secret.Data[common.SECRET_DATA_KEY_USERNAME]), string(secret.Data[common.SECRET_DATA_KEY_PASSWORD])
	if err != nil || username == "" || password == "" {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("Error: GetDatabaseResponseFromCR() ended! Could not fetch data from secret! %s\n", err)
	}

	// Create ndbClient and getting databaseResponse
	ndbClient := ndb_client.NewNDBClient(username, password, database.Spec.NDB.Server, "", true)
	databaseResponse, err = ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("Error: GetDatabaseResponseFromCR() ended! Database response from ndb_api failed! %s\n", err)
	}

	logger.Printf("Database response.status: %s.\n", databaseResponse.Status)
	logger.Println("GetDatabaseResponseFromCR() ended!")

	return databaseResponse, nil
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
