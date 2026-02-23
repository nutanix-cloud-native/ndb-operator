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
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// This function is called by TestSuiteManager.Setup in all Setup test suites.
// It loads environment variables, instantiate resources, waits for db/clone to be ready, and pod to start.
func provisionOrClone(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("provisionOrClone() starting. Attempting to initialize properties...")

	// Checking if setupTypes, clientSet, or v1alpha1ClientSet is nil
	if st == nil || clientset == nil || v1alpha1ClientSet == nil {
		errMsg := "Error: provisionOrClone() ended! Initialization Failed! "
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

	// Database (and DB instance secret, app pod) live in this namespace
	ns := automation.NAMESPACE_DEFAULT
	if st.Database != nil && st.Database.Namespace != "" {
		ns = st.Database.Namespace
	}
	ndbCredsNs := automation.NDB_CREDENTIALS_NAMESPACE

	// Ensure dedicated namespace for NDB API credentials exists
	_, err = clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ndbCredsNs}}, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace %s: %w", ndbCredsNs, err)
	}
	if err == nil {
		logger.Printf("Namespace %s created for NDB credentials.\n", ndbCredsNs)
	}

	// Create DB instance secret in database namespace
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

	// Create NDB API credential secret in dedicated namespace (not visible to devs in database namespace)
	if st.NdbSecret != nil {
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv(automation.NDB_SECRET_USERNAME_ENV)
		st.NdbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv(automation.NDB_SECRET_PASSWORD_ENV)
		_, err = clientset.CoreV1().Secrets(ndbCredsNs).Create(context.TODO(), st.NdbSecret, metav1.CreateOptions{})
		if err != nil {
			logger.Printf("Error while creating ndb secret %s: %s\n", st.NdbSecret.Name, err)
		} else {
			logger.Printf("NDB Secret %s created in namespace %s.\n", st.NdbSecret.Name, ndbCredsNs)
		}
	} else {
		logger.Printf("Error while fetching ndb secret type %s. Ndb Secret is nil.\n", st.DbSecret.Name)
	}

	// Create NDBServer (cluster-scoped) with credential ref pointing to dedicated namespace
	if st.NdbServer != nil {
		st.NdbServer.Spec.Server = os.Getenv(automation.NDB_SERVER_ENV)
		st.NdbServer.Spec.CredentialSecretRef.Namespace = ndbCredsNs
		if st.NdbServer.Spec.CredentialSecretRef.Name == "" && st.NdbSecret != nil {
			st.NdbServer.Spec.CredentialSecretRef.Name = st.NdbSecret.Name
		}
		st.NdbServer.Namespace = "" // cluster-scoped has no namespace
		st.NdbServer, err = v1alpha1ClientSet.NDBServers("").Create(st.NdbServer)
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
		// Check if clusterName is provided, otherwise use clusterId
		nxClusterName := os.Getenv(automation.NX_CLUSTER_NAME_ENV)
		nxClusterId := os.Getenv(automation.NX_CLUSTER_ID_ENV)
		if st.Database.Spec.IsClone {
			// For clones, set clusterName from env if provided
			if nxClusterName != "" {
				st.Database.Spec.Clone.ClusterName = nxClusterName
				// Clear clusterId if clusterName is provided
				st.Database.Spec.Clone.ClusterId = ""
			} else if nxClusterId != "" {
				st.Database.Spec.Clone.ClusterId = nxClusterId
			}

			// Get source database name from environment and set it in the spec
			// This enables testing of name-based resolution for source databases
			sourceDatabaseName, dbErr := getDatabaseName(ctx, st.Database)
			if dbErr == nil && sourceDatabaseName != "" {
				st.Database.Spec.Clone.SourceDatabaseName = sourceDatabaseName
				// Clear sourceDatabaseId if sourceDatabaseName is provided
				st.Database.Spec.Clone.SourceDatabaseId = ""
			}

			if err = updateClone(ctx, st.Database, st.NdbServer, st.NdbSecret); err != nil {
				return
			}
		} else {
			if nxClusterName != "" {
				st.Database.Spec.Instance.ClusterName = nxClusterName
				// Clear clusterId if clusterName is provided
				st.Database.Spec.Instance.ClusterId = ""
			} else if nxClusterId != "" {
				st.Database.Spec.Instance.ClusterId = nxClusterId
			}
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
			// Check if pod is ready (not just running) - this ensures init containers have completed
			podReady := false
			for _, condition := range st.AppPod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					podReady = true
					break
				}
			}
			statusMessage := "Pod " + st.AppPod.Name + " is in '" + string(st.AppPod.Status.Phase) + "' status."
			if podReady {
				logger.Println(statusMessage + " Pod is ready.")
				return
			}
			err = errors.New(statusMessage)
			return
		})
		if err == nil {
			logger.Println("Pod is ready")
			// Add a small grace period to ensure the app inside the pod
			// has fully initialized and bound to its port before tests start
			logger.Println("Waiting 5 seconds for application to fully initialize...")
			time.Sleep(5 * time.Second)
		} else {
			logger.Println(err)
			return
		}
	}

	logger.Println("provisionOrClone() ended. Initialization complete.")

	return
}

// This function is called by TestSuiteManager.TearDown in all TearDown test suites.
// Delete resources and de-provision database/clone.
func deprovisionOrDeclone(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("deprovisionOrDeclone() starting...")

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
		err := v1alpha1ClientSet.NDBServers("").Delete(st.NdbServer.Name, &metav1.DeleteOptions{})
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
		logger.Printf("Attempting to delete ndb secret: %s (from %s)...", st.NdbSecret.Name, automation.NDB_CREDENTIALS_NAMESPACE)
		err = clientset.CoreV1().Secrets(automation.NDB_CREDENTIALS_NAMESPACE).Delete(context.TODO(), st.NdbSecret.Name, metav1.DeleteOptions{})
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

	logger.Println("deprovisionOrDeclone() ended. Initialization complete.")

	return
}

// This function is called by TestSuiteManager.getDatabaseOrCloneResponse in all GetDatabase/GetCloneResponse test suites
// Returns a DatabaseResponse indicating if provisoning or cloning was succesful
func getDatabaseOrCloneResponse(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client) (databaseOrCloneResponse *ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("getDatabaseOrCloneResponse() starting...")
	errBaseMsg := "Error: getDatabaseOrCloneResponse() ended"

	// Get NDBServer CR
	ndbServer, err := v1alpha1ClientSet.NDBServers("").Get(st.NdbServer.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s! Could not fetch ndbServer '%s' CR! %s", errBaseMsg, st.NdbServer.Name, err)
	} else {
		logger.Printf("Retrieved ndbServer '%s' CR from v1alpha1ClientSet", ndbServer.Name)
	}

	// Get Database CR
	database, err := v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s! Could not fetch database '%s' CR! %s", errBaseMsg, database.Name, err)
	} else {
		logger.Printf("Retrieved database '%s' CR from v1alpha1ClientSet", database.Name)
	}

	// Get NDB username and password from NDB CredentialSecretRef
	ref := ndbServer.Spec.CredentialSecretRef
	secret, err := clientset.CoreV1().Secrets(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("%s! Could not fetch data from secret! %s", errBaseMsg, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("%s! Secret %s/%s is nil", errBaseMsg, ref.Namespace, ref.Name)
	}
	username, password := string(secret.Data[common.SECRET_DATA_KEY_USERNAME]), string(secret.Data[common.SECRET_DATA_KEY_PASSWORD])
	if username == "" || password == "" {
		return nil, fmt.Errorf("%s! Secret %s/%s has empty username or password", errBaseMsg, ref.Namespace, ref.Name)
	}

	// Create ndbClient and getting databaseOrCloneResponse
	ndbClient := ndb_client.NewNDBClient(username, password, ndbServer.Spec.Server, "", true)
	if st.Database.Spec.IsClone {
		databaseOrCloneResponse, err = ndb_api.GetCloneById(context.TODO(), ndbClient, database.Status.Id)
	} else {
		databaseOrCloneResponse, err = ndb_api.GetDatabaseById(context.TODO(), ndbClient, database.Status.Id)
	}

	if err != nil {
		return nil, fmt.Errorf("%s! Database response from ndb_api failed! %s", errBaseMsg, err)
	}

	logger.Printf("Database response.status: %s.\n", databaseOrCloneResponse.Status)
	logger.Println("getDatabaseOrCloneResponse() ended!")

	return databaseOrCloneResponse, nil
}

// Tests if pod is able to connect to database
func getAppResponse(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("getAppResponse() started...")
	errBaseMsg := "getAppResponse() ended"

	// Retrieve the pod name and targetPort
	podName := st.AppPod.Name
	podTargetPort := st.AppPod.Spec.Containers[0].Ports[0].ContainerPort

	// Run port-forward command using kubectl
	cmd := exec.Command("kubectl", "port-forward", podName, fmt.Sprintf("%s:%d", localPort, podTargetPort))
	err = cmd.Start()
	if err != nil {
		return http.Response{}, fmt.Errorf("%s! kubectl port-forward %s %s:%d failed! %v. ", errBaseMsg, podName, localPort, podTargetPort, err)
	} else {
		logger.Printf("kubectl port-forward %s %s:%d started.", podName, localPort, podTargetPort)
	}

	// Wait for port-forwarding to establish
	// Increased from 2s to 10s to ensure port-forward fully establishes before HTTP attempts
	logger.Println("Waiting 10 seconds for port-forward to establish...")
	time.Sleep(10 * time.Second)

	// Verify the forwarded port by making an HTTP request with retry logic
	url := fmt.Sprintf("http://localhost:%s", localPort)
	maxRetries := 10
	retryDelay := 2 * time.Second
	var resp *http.Response

	for i := 0; i < maxRetries; i++ {
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			logger.Printf("http://localhost:%s successful on attempt %d.", localPort, i+1)
			break
		}
		if i < maxRetries-1 {
			logger.Printf("Attempt %d failed (err: %v), retrying in %v...", i+1, err, retryDelay)
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(retryDelay)
		}
	}

	if err != nil || resp.StatusCode != 200 {
		return http.Response{}, fmt.Errorf("http://localhost:%s failed after %d attempts! Last error: %v", localPort, maxRetries, err)
	}

	defer resp.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.Response{}, fmt.Errorf("getAppResponse() ended! Error while reading response body: %s", err)
	} else {
		logger.Println("Response: ", string(body))
	}

	logger.Printf("%s!", errBaseMsg)

	return *resp, nil
}
