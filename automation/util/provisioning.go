package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/automation"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// This function is called from the SetupSuite() function of all testsuites.
// It loads environment variables, instantiate resources, waits for db to be ready, and pod to start.
func ProvisioningTestSetup(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("ProvisioningTestSetup() starting. Attempting to initialize properties...")

	// Checking if setupTypes, clientSet, or v1alpha1ClientSet is nil
	if st == nil || clientset == nil || v1alpha1ClientSet == nil {
		errMsg := "Error: ProvisioningTestSetup() ended! Initialization Failed! "
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

	ns := automation.NAMESPACE_DEFAULT
	if st.Database != nil && st.Database.Namespace != "" {
		ns = st.Database.Namespace
	}

	// Create Secrets
	if st.DbSecret != nil {
		st.DbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("DB_SECRET_PASSWORD")
		_, err = clientset.CoreV1().Secrets(ns).Create(ctx, st.DbSecret, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating db secret %s: %s\n", st.DbSecret.Name, err)
		} else {
			logger.Printf("DB Secret %s created.\n", st.DbSecret.Name)
		}
	} else {
		logger.Printf("Error while fetching db secret type %s. Db Secret is nil.\n", st.DbSecret.Name)
	}

	if st.NdbSecret != nil {
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("NDB_SECRET_USERNAME")
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("NDB_SECRET_PASSWORD")
		_, err = clientset.CoreV1().Secrets(ns).Create(context.TODO(), st.NdbSecret, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating ndb secret %s: %s\n", st.NdbSecret.Name, err)
		} else {
			logger.Printf("NDB Secret %s created.\n", st.NdbSecret.Name)
		}
	} else {
		logger.Printf("Error while fetching ndb secret type %s. Ndb Secret is nil.\n", st.DbSecret.Name)
	}

	// Create NDBServer
	if st.NdbServer != nil {
		st.NdbServer.Spec.Server = os.Getenv("NDB_SERVER")
		st.NdbServer, err = v1alpha1ClientSet.NDBServers(st.NdbServer.Namespace).Create(st.NdbServer)
		if err != nil {
			logger.Printf("Error while creating NDBServer %s: %s\n", st.Database.Name, err)
		} else {
			logger.Printf("NDBServer %s created.\n", st.Database.Name)
		}
	} else {
		logger.Printf("Error while fetching NDBServer type %s. NDBServer is nil.\n", st.DbSecret.Name)
	}

	// Create Database
	if st.Database != nil {
		st.Database.Spec.Instance.ClusterId = os.Getenv("CLUSTER_ID")
		st.Database, err = v1alpha1ClientSet.Databases(st.Database.Namespace).Create(st.Database)
		if err != nil {
			logger.Printf("Error while creating Database %s: %s\n", st.Database.Name, err)
		} else {
			logger.Printf("Database %s created.\n", st.Database.Name)
		}
	} else {
		logger.Printf("Error while fetching database type %s. Database is nil.\n", st.DbSecret.Name)
	}

	// Create Application
	if st.AppPod != nil {
		st.AppPod, err = clientset.CoreV1().Pods(ns).Create(context.TODO(), st.AppPod, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating Pod %s: %s\n", st.AppPod.Name, err)
		} else {
			logger.Printf("Pod %s created.\n", st.AppPod.Name)
		}
	} else {
		logger.Printf("Error while fetching app pod type %s. AppPod is nil.\n", st.DbSecret.Name)
	}

	// Wait for DB to get Ready
	if st.Database != nil {
		err = waitAndRetryOperation(ctx, time.Minute, 80, func() (err error) {
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

	logger.Println("ProvisioningTestSetup() ended. Initialization complete.")

	return
}

// This function is called from the TeardownSuite() function of all testsuites.
// Delete resources and de-provision database.
func ProvisioningTestTeardown(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("ProvisioningTestTeardown() starting...")

	ns := automation.NAMESPACE_DEFAULT
	if st.Database != nil && st.Database.Namespace != "" {
		ns = st.Database.Namespace
	}

	// Delete Service
	svcName := st.Database.Name + "-svc"
	logger.Printf("Attempting to delete service: %s...", svcName)
	err = clientset.CoreV1().Services(ns).Delete(context.TODO(), svcName, metav1.DeleteOptions{})
	if err != nil {
		logger.Printf("Error while deleting service %s: %s\n", svcName, err)
	} else {
		logger.Printf("Service %s deleted.\n", svcName)
	}

	// Delete Database
	if st.Database != nil {
		logger.Printf("Attempting to delete database: %s...", st.Database.Name)
		err := v1alpha1ClientSet.Databases(st.Database.Namespace).Delete(st.Database.Name, &metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting Database %s: %s!\n", st.Database.Name, err)
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
	} else {
		logger.Printf("Error while fetching database type %s. Database is nil.\n", st.DbSecret.Name)
	}

	// Delete NDB Server
	if st.NdbServer != nil {
		logger.Printf("Attempting to delete ndb server: %s...", st.NdbServer.Name)
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), st.NdbServer.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting ndb server %s: %s!\n", st.NdbServer.Name, err)
		} else {
			logger.Printf("Ndb server %s deleted.\n", st.DbSecret.Name)
		}
	} else {
		logger.Printf("Error while fetching NDBServer type %s. NDBServer is nil.\n", st.DbSecret.Name)
	}

	// Delete Secrets
	if st.DbSecret != nil {
		logger.Printf("Attempting to delete db secret: %s...", st.DbSecret.Name)
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), st.DbSecret.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting secret %s: %s!\n", st.DbSecret.Name, err)
		} else {
			logger.Printf("Secret %s deleted.\n", st.DbSecret.Name)
		}
	} else {
		logger.Printf("Error while fetching db secret type %s. Db Secret is nil.\n", st.DbSecret.Name)
	}
	if st.NdbSecret != nil {
		logger.Printf("Attempting to delete ndb secret: %s...", st.NdbSecret.Name)
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), st.NdbSecret.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting secret %s: %s!\n", st.NdbSecret.Name, err)
		} else {
			logger.Printf("Secret %s deleted.\n", st.NdbSecret.Name)
		}
	} else {
		logger.Printf("Error while fetching ndb secret type %s. NDB Secret is nil.\n", st.DbSecret.Name)
	}

	// Delete Application
	if st.AppPod != nil {
		logger.Printf("Attempting to delete application: %s...", st.AppPod.Name)
		err := clientset.CoreV1().Pods(ns).Delete(context.TODO(), st.AppPod.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Printf("Error while deleting pod %s: %s!\n", st.AppPod.Name, err)
		} else {
			logger.Printf("Pod %s deleted.\n", st.AppPod.Name)
		}
	} else {
		logger.Printf("Error while fetching app pod type %s. AppPod is nil.\n", st.DbSecret.Name)
	}

	logger.Println("ProvisioningTestTeardown() ended. Initialization complete.")

	return
}

// Wrapper function called in all TestSuite TestProvisioningSuccess methods. Returns a DatabaseResponse which indicates if provison was succesful
func GetDatabaseResponse(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (databaseResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetDatabaseResponse() starting...")
	errBaseMsg := "Error: GetDatabaseResponse() ended"

	// Get NDBServer CR
	ndbServer, err := v1alpha1ClientSet.NDBServers(st.NdbServer.Namespace).Get(st.NdbServer.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("%s! Could not fetch ndbServer '%s' CR! %s\n", errBaseMsg, ndbServer.Name, err)
	} else {
		logger.Printf("Retrieved ndbServer '%s' CR from v1alpha1ClientSet", ndbServer.Name)
	}

	// Get Database CR
	database, err := v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("%s! Could not fetch database '%s' CR! %s\n", errBaseMsg, database.Name, err)
	} else {
		logger.Printf("Retrieved database '%s' CR from v1alpha1ClientSet", database.Name)
	}

	// Get NDB username and password from NDB CredentialSecret
	ndb_secret_name := ndbServer.Spec.CredentialSecret
	secret, err := clientset.CoreV1().Secrets(database.Namespace).Get(context.TODO(), ndb_secret_name, metav1.GetOptions{})
	username, password := string(secret.Data[common.SECRET_DATA_KEY_USERNAME]), string(secret.Data[common.SECRET_DATA_KEY_PASSWORD])
	if err != nil || username == "" || password == "" {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("%s! Could not fetch data from secret! %s\n", errBaseMsg, err)
	}

	// Create ndbClient and getting databaseResponse
	ndbClient := ndb_client.NewNDBClient(username, password, ndbServer.Spec.Server, "", true)
	databaseResponse, err = ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("%s! Database response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	logger.Printf("Database response.status: %s.\n", databaseResponse.Status)
	logger.Println("GetDatabaseResponse() ended!")

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
