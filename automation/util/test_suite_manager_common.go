package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

// This function is called by TestSuiteManager.Setup in all Setup test suites.
// It loads environment variables, instantiate resources, waits for db/clone to be ready, and pod to start.
func ProvisionOrClone(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("ProvisionOrClone() starting. Attempting to initialize properties...")

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
		st.DbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv(automation.DB_SECRET_PASSWORD_ENV)
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
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv(automation.NDB_SECRET_USERNAME_ENV)
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv(automation.NDB_SECRET_PASSWORD_ENV)
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
		st.NdbServer.Spec.Server = os.Getenv(automation.NDB_SERVER_ENV)
		st.NdbServer, err = v1alpha1ClientSet.NDBServers(st.NdbServer.Namespace).Create(st.NdbServer)
		if err != nil {
			logger.Printf("Error while creating NDBServer %s: %s\n", st.NdbServer.Name, err)
		} else {
			logger.Printf("NDBServer %s created.\n", st.NdbServer.Name)
		}
	} else {
		logger.Printf("Error while fetching NDBServer type %s. NDBServer is nil.\n", st.DbSecret.Name)
	}

	// Create Database or Clone
	if st.Database != nil {
		clusterId := os.Getenv(automation.NX_CLUSTER_ID_ENV)
		if st.Database.Spec.IsClone {
			st.Database.Spec.Clone.ClusterId = clusterId
		} else {
			st.Database.Spec.Instance.ClusterId = clusterId
		}
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

// This function is called by TestSuiteManager.TearDown in all TearDown test suites.
// Delete resources and de-provision database/clone.
func DeprovisionOrDeclone(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
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
		err := v1alpha1ClientSet.NDBServers(st.NdbServer.Namespace).Delete(st.NdbServer.Name, &metav1.DeleteOptions{})
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

// This function is called by TestSuiteManager.GetDatabaseOrCloneResponse in all GetDatabase/GetCloneResponse test suites
// Returns a DatabaseResponse indicating if provisoning or cloning was succesful
func GetDatabaseOrCloneResponse(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (databaseOrCloneResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetDatabaseOrCloneResponse() starting...")
	errBaseMsg := "Error: GetDatabaseOrCloneResponse() ended"

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

	// Create ndbClient and getting databaseOrCloneResponse
	ndbClient := ndb_client.NewNDBClient(username, password, ndbServer.Spec.Server, "", true)
	if st.Database.Spec.IsClone {
		databaseOrCloneResponse, err = ndb_api.GetCloneById(context.TODO(), ndbClient, database.Status.Id)
	} else {
		databaseOrCloneResponse, err = ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	}

	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("%s! Database response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	logger.Printf("Database response.status: %s.\n", databaseOrCloneResponse.Status)
	logger.Println("GetDatabaseOrCloneResponse() ended!")

	return databaseOrCloneResponse, nil
}

// Tests if pod is able to connect to database
func GetAppResponse(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetAppResponse() started...")
	errBaseMsg := "GetAppResponse() ended"

	// Retrieve the pod name and targetPort
	podName := st.AppPod.Name
	podTargetPort := st.AppPod.Spec.Containers[0].Ports[0].ContainerPort

	// Run port-forward command using kubectl
	cmd := exec.Command("kubectl", "port-forward", podName, fmt.Sprintf("%s:%d", localPort, podTargetPort))
	err = cmd.Start()
	if err != nil {
		return http.Response{}, fmt.Errorf("%s! kubectl port-forward %s %s:%d failed! %v. ", errBaseMsg, podName, localPort, podTargetPort, err)
	} else {
		logger.Printf("kubectl port-forward %s %s:%d succesful.", podName, localPort, podTargetPort)
	}

	// Wait for a brief period to let port-forwarding start
	time.Sleep(2 * time.Second)

	// Verify the forwarded port by making an HTTP request
	url := fmt.Sprintf("http://localhost:%s", localPort)
	resp, err := http.Get(url)

	if err != nil {
		return http.Response{}, fmt.Errorf("http://localhost:%s failed! %v,", localPort, err)
	} else {
		logger.Printf("http://localhost:%s succesful.", localPort)
	}

	defer resp.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.Response{}, errors.New(fmt.Sprintf("GetAppResponse() ended! Error while reading response body: %s", err))
	} else {
		logger.Println("Response: ", string(body))
	}

	logger.Println(fmt.Sprintf("%s!", errBaseMsg))

	return *resp, nil
}
