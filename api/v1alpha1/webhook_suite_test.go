/*
Copyright 2022-2023 Nutanix, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	//+kubebuilder:scaffold:imports

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeEach(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	// CleanUpAfterUse will cause the CRDs listed for installation to be
	// uninstalled when terminating the test environment.
	// Defaults to false.
	testEnv.CRDInstallOptions.CleanUpAfterUse = true

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()
	err = AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	err = (&Database{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:webhook

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())

})

var _ = AfterEach(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Webhook Tests", func() {
	It("Should check for missing Database Instance Name", func() {
		database := &Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db",
				Namespace: "default",
			},
			Spec: DatabaseSpec{
				NDBRef: "ndbRef",
				Instance: Instance{
					CredentialSecret: "db-instance-secret",
					Size:             10,
					TimeZone:         "UTC",
					Type:             common.DATABASE_TYPE_POSTGRES,
				},
			},
		}
		err := k8sClient.Create(context.Background(), database)
		Expect(err).To(HaveOccurred())
		errMsg := err.(*errors.StatusError).ErrStatus.Message
		Expect(errMsg).To(ContainSubstring("A valid Database Instance Name must be specified"))
	})

	It("Should check for missing ClusterId", func() {
		database := &Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db",
				Namespace: "default",
			},
			Spec: DatabaseSpec{
				NDBRef: "ndbRef",
				Instance: Instance{
					CredentialSecret:     "db-instance-secret",
					DatabaseInstanceName: "db-instance-name",
					Type:                 "postgres",
					Size:                 10,
					TimeZone:             "UTC",
				},
			},
		}
		err := k8sClient.Create(context.Background(), database)
		Expect(err).To(HaveOccurred())
		errMsg := err.(*errors.StatusError).ErrStatus.Message
		Expect(errMsg).To(ContainSubstring("ClusterId field must be a valid UUID"))
	})

	It("Should check for missing CredentialSecret", func() {
		database := &Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db",
				Namespace: "default",
			},
			Spec: DatabaseSpec{
				NDBRef: "ndbRef",
				Instance: Instance{
					DatabaseInstanceName: "db-instance-name",
					Size:                 10,
					TimeZone:             "UTC",
					Type:                 common.DATABASE_TYPE_POSTGRES,
				},
			},
		}
		err := k8sClient.Create(context.Background(), database)
		Expect(err).To(HaveOccurred())
		errMsg := err.(*errors.StatusError).ErrStatus.Message
		Expect(errMsg).To(ContainSubstring("CredentialSecret must be provided in the Instance Spec"))
	})

	It("Should check for size < 10'", func() {
		database := &Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db",
				Namespace: "default",
			},
			Spec: DatabaseSpec{
				NDBRef: "ndbRef",
				Instance: Instance{
					DatabaseInstanceName: "db-instance-name",
					CredentialSecret:     "db-instance-secret",
					Size:                 1,
					TimeZone:             "UTC",
					Type:                 common.DATABASE_TYPE_POSTGRES,
				},
			},
		}
		err := k8sClient.Create(context.Background(), database)
		Expect(err).To(HaveOccurred())
		errMsg := err.(*errors.StatusError).ErrStatus.Message
		Expect(errMsg).To(ContainSubstring("Initial Database size must be specified with a value 10 GBs or more"))
	})

	It("Should check for missing Type", func() {
		database := &Database{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "db",
				Namespace: "default",
			},
			Spec: DatabaseSpec{
				NDBRef: "ndbRef",
				Instance: Instance{
					DatabaseInstanceName: "db-instance-name",
					CredentialSecret:     "db-instance-secret",
					Size:                 10,
				},
			},
		}
		err := k8sClient.Create(context.Background(), database)
		Expect(err).To(HaveOccurred())
		errMsg := err.(*errors.StatusError).ErrStatus.Message
		Expect(errMsg).To(ContainSubstring("A valid database type must be specified. Valid values are: "))
	})

	It("Should check for invalid Type'", func() {
		database := createDbWithType("db", "invalid")
		err := k8sClient.Create(context.Background(), database)
		Expect(err).To(HaveOccurred())
		errMsg := err.(*errors.StatusError).ErrStatus.Message
		Expect(errMsg).To(ContainSubstring("A valid database type must be specified. Valid values are: "))
	})

	When("Profiles missing", func() {
		It("Should not error out for missing Profiles: Open-source engines", func() {
			database := createDbWithType("db4", common.DATABASE_TYPE_POSTGRES)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should error out for missing Profiles: Closed-source engines", func() {
			database := createDbWithType("db", common.DATABASE_TYPE_MSSQL)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("Software Profile must be provided for the closed-source database engines"))
		})
	})

	When("AdditionalArguments for MYSQL specified", func() {
		It("Should not error for valid MYSQL additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db6",
				common.DATABASE_TYPE_MYSQL,
				map[string]string{
					"listener_port": "3306",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should error out for invalid MYSQL additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db",
				common.DATABASE_TYPE_MYSQL,
				map[string]string{
					"invalid-key": "invalid-value",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Additional Arguments validation for database type: %s failed!", common.DATABASE_TYPE_MYSQL)))
		})
	})

	When("AdditionalArguments for PostGres specified", func() {
		It("Should not error for valid Postgres additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db7",
				common.DATABASE_TYPE_POSTGRES,
				map[string]string{
					"listener_port": "5432",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should error out for invalid Postgres additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db",
				common.DATABASE_TYPE_POSTGRES,
				map[string]string{
					"listener_port": "5432",
					"invalid-key":   "invalid-value",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Additional Arguments validation for database type: %s failed!", common.DATABASE_TYPE_POSTGRES)))
		})
	})

	When("AdditionalArguments for MongoDB specified", func() {
		It("Should not error for valid MongoDB additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db8",
				common.DATABASE_TYPE_MONGODB,
				map[string]string{
					"listener_port": "5432",
					"log_size":      "10",
					"journal_size":  "10",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should error out for invalid MongoDB additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db",
				common.DATABASE_TYPE_MONGODB,
				map[string]string{
					"listener_port": "5432",
					"log_size":      "10",
					"invalid-key":   "invalid-value",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Additional Arguments validation for database type: %s failed!", common.DATABASE_TYPE_MONGODB)))
		})
	})

	When("AdditionalArguments for MSSQL specified", func() {
		It("Should not error for valid MSSQL additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db9",
				common.DATABASE_TYPE_MSSQL,
				map[string]string{
					"server_collation":           "SQL_Latin1_General_CPI_CI_AS",
					"database_collation":         "SQL_Latin1_General_CPI_CI_AS",
					"vm_win_license_key":         "XXXX-XXXXX-XXXXX-XXXXX-XXXXX",
					"vm_dbserver_admin_password": "<password>",
					"authentication_mode":        "mixed",
					"sql_user_name":              "sa",
					"sql_user_password":          "<password>",
					"windows_domain_profile_id":  "<XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
					"vm_db_server_user":          "<prod.cdm.com\\<user>",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should error out for invalid MSSQL additionalArguments", func() {
			database := createDbWithAdditionalArguments(
				"db",
				"mssql",
				map[string]string{
					"invalid-key":  "invalid-value",
					"invalid-key2": "invalid-value2",
				},
			)
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Additional Arguments validation for database type: %s failed!", common.DATABASE_TYPE_MSSQL)))
		})
	})
})

/* Creates a database CR with 'dbType'. */
func createDbWithType(db string, dbType string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDBRef: "ndbRef",
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				ClusterId:            "27bcce67-7b83-42c2-a3fe-88154425c170",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 dbType,
			},
		},
	}
}

/* Creates a database CR with db with additionalArguments specified */
func createDbWithAdditionalArguments(db string, dbType string, additionalArguments map[string]string) *Database {
	database := &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      db,
			Namespace: "default",
		},
		Spec: DatabaseSpec{
			NDBRef: "ndbRef",
			Instance: Instance{
				DatabaseInstanceName: "db-instance-name",
				ClusterId:            "27bcce67-7b83-42c2-a3fe-88154425c170",
				CredentialSecret:     "db-instance-secret",
				Size:                 10,
				Type:                 dbType,
				AdditionalArguments:  additionalArguments,
			},
		},
	}

	if dbType == common.DATABASE_TYPE_MSSQL {
		database.Spec.Instance.Profiles = &Profiles{
			Software: Profile{Name: "MSSQL_SOFTWARE_PROFILE"},
		}
	}

	return database
}
