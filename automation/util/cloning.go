package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/automation"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CloningTestSetup(ctx context.Context, st *SetupTypes, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	logger := GetLogger(ctx)
	logger.Println("CloningTestSetup() starting. Attempting to initialize properties...")

	// Checking if setupTypes, clientSet, or v1alpha1ClientSet is nil
	if st == nil || clientset == nil || v1alpha1ClientSet == nil {
		errMsg := "Error: CloningTestSetup() ended! Initialization Failed! "
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

	// 1) TODO: If sourceDatabaseId is not found, fail test
	// 2) TODO: If snapshotId is missing, create a new snapshot and wait for it to finish creating
	// -> 1) Get TM ID from ndb_api.GetDatabaseById
	// -> 2) Create snapshot on TM using ndb_api.TakeSnapshotForTM (ensure unique snapshot name)
	// -> 3) When snapshot is finished, update database CR snapshotID and create db

	// Create Clone
	if st.Database != nil {
		st.Database.Spec.Clone.ClusterId = os.Getenv("NDB_CLUSTER_ID")
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

	// Wait for Clone to Get Ready
	if st.Database != nil {
		err = waitAndRetryOperation(ctx, time.Minute, 80, func() (err error) {
			st.Database, err = v1alpha1ClientSet.Databases(st.Database.Namespace).Get(st.Database.Name, metav1.GetOptions{})
			if err != nil {
				return
			}
			statusMessage := "Clone " + st.Database.Name + " is in '" + st.Database.Status.Status + "' status."
			if st.Database.Status.Status == common.DATABASE_CR_STATUS_READY {
				logger.Println(statusMessage)
				return
			}
			err = errors.New(statusMessage)
			return
		})
		if err == nil {
			logger.Println("Clone is ready")
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

	logger.Println("CloningTestSetup() ended. Initialization complete.")

	return
}

func GetCloneResponse(ctx context.Context, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, st *SetupTypes) (cloneResponse ndb_api.DatabaseResponse, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetCloneResponse() starting...")
	errBaseMsg := "Error: GetCloneResponse() ended"

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
	cloneResponse, err = ndb_api.GetCloneById(context.TODO(), ndbClient, database.Status.Id)
	if err != nil {
		return ndb_api.DatabaseResponse{}, fmt.Errorf("%s! Clone response from ndb_api failed! %s\n", errBaseMsg, err)
	}

	logger.Printf("Clone response.status: %s.\n", cloneResponse.Status)
	logger.Println("GetCloneResponse() ended!")

	return cloneResponse, nil
}
