package automation

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	clientsetv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/automation/clientset/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const namespace_default = "default"

// This function is called from the SetupSuite() function of all testsuites.
// It loads environment variables, instantiate resources, waits for db to be ready, and pod to start.
func TestSetup(dbSecret, ndbSecret *corev1.Secret, database *ndbv1alpha1.Database, appPod *corev1.Pod, appSvc *corev1.Service, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	log.Println("test_setup() starting...")

	ns := namespace_default
	if database != nil && database.Namespace != "" {
		ns = database.Namespace
	}

	// Create Secrets
	if dbSecret != nil {
		dbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("DB_SECRET_PASSWORD")
		_, err = clientset.CoreV1().Secrets(ns).Create(context.TODO(), dbSecret, metav1.CreateOptions{})
		if err == nil {
			log.Printf("Secret %s created\n", dbSecret.Name)
		} else {
			log.Printf("Error while creating secret %s: %s\n", dbSecret.Name, err)
		}
	}
	if ndbSecret != nil {
		ndbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("NDB_SECRET_USERNAME")
		ndbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("NDB_SECRET_PASSWORD")
		_, err = clientset.CoreV1().Secrets(ns).Create(context.TODO(), ndbSecret, metav1.CreateOptions{})
		if err == nil {
			log.Printf("Secret %s created\n", ndbSecret.Name)
		} else {
			log.Printf("Error while creating secret %s: %s\n", ndbSecret.Name, err)
		}
	}

	// Create Database
	if database != nil {
		// log.Printf(database.Spec.Instance.DatabaseInstanceName + ", " + database.Spec.NDB.ClusterId)
		database.Spec.NDB.Server = os.Getenv("NDB_SERVER")
		database.Spec.NDB.ClusterId = os.Getenv("NDB_CLUSTER_ID")
		database, err = v1alpha1ClientSet.Databases(database.Namespace).Create(database)
		if err != nil {
			log.Printf("Error while creating Database %s: %s\n", database.Name, err)
		} else {
			log.Printf("Database %s created\n", database.Name)
		}
	}

	// Create Application
	if appPod != nil {
		appPod, err = clientset.CoreV1().Pods(ns).Create(context.TODO(), appPod, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Error while creating Pod %s: %s\n", appPod.Name, err)
		} else {
			log.Printf("Pod %s created\n", appPod.Name)
		}
	}
	if appSvc != nil {
		appSvc, err = clientset.CoreV1().Services(ns).Create(context.TODO(), appSvc, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Error while creating Svc %s: %s\n", appSvc.Name, err)
		} else {
			log.Printf("Svc %s created\n", appSvc.Name)
		}
	}

	// Wait for DB to get Ready
	if database != nil {
		err = waitAndRetryOperation(time.Minute, 60, func() (err error) {
			database, err = v1alpha1ClientSet.Databases(database.Namespace).Get(database.Name, metav1.GetOptions{})
			if err != nil {
				return
			}
			statusMessage := "DB " + database.Name + " is in '" + database.Status.Status + "' status."
			if database.Status.Status == common.DATABASE_CR_STATUS_READY {
				log.Println(statusMessage)
				return
			}
			err = errors.New(statusMessage)
			return
		})
		if err == nil {
			log.Println("Database is ready")
		} else {
			log.Println(err)
		}
	}
	// Wait for Application Pod to start
	if appPod != nil {
		err = waitAndRetryOperation(time.Second, 300, func() (err error) {
			appPod, err = clientset.CoreV1().Pods(ns).Get(context.TODO(), appPod.Name, metav1.GetOptions{})
			if err != nil {
				return
			}
			statusMessage := "Pod " + appPod.Name + " is in '" + string(appPod.Status.Phase) + "' status."
			if appPod.Status.Phase == "Running" {
				log.Println(statusMessage)
				return
			}
			err = errors.New(statusMessage)
			return
		})
		if err == nil {
			log.Println("Pod is ready")
		} else {
			log.Println(err)
			return
		}
	}

	log.Println("test_setup() ended.")

	return
}

// This function is called from the TeardownSuite() function of all testsuites.
// Delete resources and de-provision database.
func TestTeardown(dbSecret, ndbSecret *corev1.Secret, database *ndbv1alpha1.Database, appPod *corev1.Pod, appSvc *corev1.Service, clientset *kubernetes.Clientset, v1alpha1ClientSet *clientsetv1alpha1.V1alpha1Client, t *testing.T) (err error) {
	log.Println("test_teardown() starting...")

	ns := namespace_default
	if database != nil && database.Namespace != "" {
		ns = database.Namespace
	}

	// Delete Database
	if database != nil {
		database.Spec.NDB.Server = os.Getenv("NDB-SERVER")
		database.Spec.NDB.ClusterId = os.Getenv("NDB-CLUSTER-ID")
		err := v1alpha1ClientSet.Databases(database.Namespace).Delete(database.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Error while deleting Database %s: %s\n", database.Name, err)
		} else {
			log.Printf("Database %s deleted\n", database.Name)
		}
		waitAndRetryOperation(time.Minute, 10, func() (err error) {
			database, err = v1alpha1ClientSet.Databases(database.Namespace).Get(database.Name, metav1.GetOptions{})
			if err != nil {
				return nil
			}
			if (database == &ndbv1alpha1.Database{}) {
				log.Println("Received empty database")
				return nil
			}
			statusMessage := "DB " + database.Name + " is not yet deleted"
			log.Println(statusMessage)
			err = errors.New(statusMessage)
			return
		})
	}

	// Delete Secrets
	if dbSecret != nil {
		dbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("DB-SECRET-USERNAME")
		dbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("DB-SECRET-PASSWORD")
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), dbSecret.Name, metav1.DeleteOptions{})
		if err == nil {
			log.Printf("Secret %s deleted\n", dbSecret.Name)
		} else {
			log.Printf("Error while deleting secret %s: %s\n", dbSecret.Name, err)
		}
	}
	if ndbSecret != nil {
		ndbSecret.StringData[common.SECRET_DATA_KEY_USERNAME] = os.Getenv("NDB-SECRET-USERNAME")
		ndbSecret.StringData[common.SECRET_DATA_KEY_PASSWORD] = os.Getenv("NDB-SECRET-PASSWORD")
		err = clientset.CoreV1().Secrets(ns).Delete(context.TODO(), ndbSecret.Name, metav1.DeleteOptions{})
		if err == nil {
			log.Printf("Secret %s deleted\n", ndbSecret.Name)
		} else {
			log.Printf("Error while deleting secret %s: %s\n", ndbSecret.Name, err)
		}
	}

	// Delete Application
	if appPod != nil {
		err := clientset.CoreV1().Pods(ns).Delete(context.TODO(), appPod.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Error while deleting Pod %s: %s\n", appPod.Name, err)
		} else {
			log.Printf("Pod %s deleted\n", appPod.Name)
		}
	}
	if appSvc != nil {
		err = clientset.CoreV1().Services(ns).Delete(context.TODO(), appSvc.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Printf("Error while deleting Svc %s: %s\n", appSvc.Name, err)
		} else {
			log.Printf("Svc %s deleted\n", appSvc.Name)
		}
	}

	log.Println("test_teardown() ended.")

	return
}
