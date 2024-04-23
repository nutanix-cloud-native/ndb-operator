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

const (
	NAMESPACE         = "default"
	NDB_REF           = "ndbRef"
	NAME              = "name"
	DEFAULT_UUID      = "6381eb1f-1114-4837-b19f-87d3db8ebfde"
	CREDENTIAL_SECRET = "database-secret"
	TIMEZONE          = "UTC"
	SIZE              = 10
	HA                = false
)

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
	Context("Database checks", func() {
		It("Should check for missing Database Instance Name", func() {
			database := createDefaultDatabase("db1")
			database.Spec.Instance.Name = ""

			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid Database Instance Name must be specified"))
		})

		It("Should check for missing ClusterId", func() {
			database := createDefaultDatabase("db2")
			database.Spec.Instance.ClusterId = ""

			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("ClusterId field must be a valid UUID"))
		})

		It("Should check for missing CredentialSecret", func() {
			database := createDefaultDatabase("db3")
			database.Spec.Instance.CredentialSecret = ""

			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("CredentialSecret must be provided in the Instance Spec"))
		})

		It("Should check for size < 10'", func() {
			database := createDefaultDatabase("db4")
			database.Spec.Instance.Size = 1

			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("Initial Database size must be specified with a value 10 GBs or more"))
		})

		It("Should check for missing Type", func() {
			database := createDefaultDatabase("db5")
			database.Spec.Instance.Type = ""

			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid database type must be specified. Valid values are: "))
		})

		It("Should check for invalid Type'", func() {
			database := createDefaultDatabase("db6")
			database.Spec.Instance.Type = "invalid"

			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid database type must be specified. Valid values are: "))
		})

		When("Profiles missing", func() {
			It("Should not error out for missing Profiles: Open-source engines", func() {
				database := createDefaultDatabase("db7")

				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Should error out for missing Profiles: Closed-source engines", func() {
				database := createDefaultDatabase("db8")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MSSQL

				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("Software Profile must be provided for the closed-source database engines"))
			})
		})

		When("AdditionalArguments for MYSQL specified", func() {
			It("Should not error for valid MYSQL additionalArguments", func() {
				database := createDefaultDatabase("db9")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MYSQL
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"listener_port": "3306",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Should error out for invalid MYSQL additionalArguments", func() {
				database := createDefaultDatabase("db10")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MYSQL
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"invalid": "invalid",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_MYSQL)))
			})
		})

		When("AdditionalArguments for PostGres specified", func() {
			It("Should not error for valid Postgres additionalArguments", func() {
				database := createDefaultDatabase("db11")
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"listener_port": "5432",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should error out for invalid Postgres additionalArguments", func() {
				database := createDefaultDatabase("db12")
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"listener_port": "5432",
					"invalid":       "invalid",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_POSTGRES)))
			})
		})

		When("AdditionalArguments for MongoDB specified", func() {
			It("Should not error for valid MongoDB additionalArguments", func() {
				database := createDefaultDatabase("db13")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MONGODB
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"listener_port": "5432",
					"log_size":      "10",
					"journal_size":  "10",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should error out for invalid MongoDB additionalArguments", func() {
				database := createDefaultDatabase("db14")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MONGODB
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"listener_port": "5432",
					"log_size":      "10",
					"invalid":       "invalid",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_MONGODB)))
			})
		})

		When("AdditionalArguments for MSSQL specified", func() {
			It("Should not error for valid MSSQL additionalArguments", func() {
				database := createDefaultDatabase("db15")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MSSQL
				database.Spec.Instance.Profiles = &Profiles{
					Software: Profile{
						Name: "MAZIN-MSSQL2",
					},
				}
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"server_collation":           "SQL_Latin1_General_CPI_CI_AS",
					"database_collation":         "SQL_Latin1_General_CPI_CI_AS",
					"vm_win_license_key":         "XXXX-XXXXX-XXXXX-XXXXX-XXXXX",
					"vm_dbserver_admin_password": "<password>",
					"authentication_mode":        "mixed",
					"sql_user_name":              "sa",
					"sql_user_password":          "<password>",
					"windows_domain_profile_id":  "<XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
					"vm_db_server_user":          "<prod.cdm.com\\<user>",
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Should error out for invalid MSSQL additionalArguments", func() {
				database := createDefaultDatabase("db16")
				database.Spec.Instance.Type = common.DATABASE_TYPE_MSSQL
				database.Spec.Instance.AdditionalArguments = map[string]string{
					"invalid":  "invalid",
					"invalid2": "invalid2",
				}
				database.Spec.Instance.Profiles = &Profiles{
					Software: Profile{
						Name: "MAZIN-MSSQL2",
					},
				}

				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_MSSQL)))
			})
		})
	})

	Context("Clone checks", func() {
		It("Should check for missing Clone Name", func() {
			clone := createDefaultClone("clone1")
			clone.Spec.Clone.Name = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid Clone Name must be specified"))
		})

		It("Should check for missing ClusterId", func() {
			clone := createDefaultClone("clone2")
			clone.Spec.Clone.ClusterId = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("ClusterId field must be a valid UUID"))
		})

		It("Should check for missing CredentialSecret", func() {
			clone := createDefaultClone("clone3")
			clone.Spec.Clone.CredentialSecret = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("CredentialSecret must be provided in the Clone Spec"))
		})

		It("Should check for missing TimeZone", func() {
			clone := createDefaultClone("clone4")
			clone.Spec.Clone.TimeZone = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("TimeZone must be provided in Clone Spec"))
		})

		It("Should check for missing/invalid Type", func() {
			clone := createDefaultClone("clone5")
			clone.Spec.Clone.Type = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid clone type must be specified. Valid values are:"))
		})

		It("Should check for sourceDatabaseId", func() {
			clone := createDefaultClone("clone6")
			clone.Spec.Clone.SourceDatabaseId = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("sourceDatabaseId must be a valid UUID"))
		})

		It("Should check for snapshotId", func() {
			clone := createDefaultClone("clone7")
			clone.Spec.Clone.SnapshotId = ""

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("snapshotId must be a valid UUID"))
		})

		It("Should check for invalid Type'", func() {
			clone := createDefaultClone("clone8")
			clone.Spec.Clone.Type = "invalid"

			err := k8sClient.Create(context.Background(), clone)
			Expect(err).To(HaveOccurred())
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid clone type must be specified. Valid values are: "))
		})

		When("Profiles missing", func() {
			It("Should not error out for missing Profiles: Open-source engines", func() {
				clone := createDefaultClone("clone9")

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Should error out for missing Profiles: Closed-source engines", func() {
				clone := createDefaultClone("clone10")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MSSQL

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("Software Profile must be provided for the closed-source database engines"))
			})
		})

		When("AdditionalArguments for MYSQL specified", func() {
			It("Should not error for valid MYSQL additionalArguments", func() {
				clone := createDefaultClone("clone11")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MYSQL
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"expireInDays":        "30",
					"expiryDateTimezone":  common.TIMEZONE_UTC,
					"deleteDatabase":      "true",
					"refreshInDays":       "7",
					"refreshTime":         "00:00:00",
					"refreshDateTimezone": common.TIMEZONE_UTC,
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Should error out for invalid MYSQL additionalArguments", func() {
				clone := createDefaultClone("clone12")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MYSQL
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"invalid": "invalid",
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_MYSQL)))
			})
		})

		When("AdditionalArguments for PostGres specified", func() {
			It("Should not error for valid Postgres additionalArguments", func() {
				clone := createDefaultClone("clone13")
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"expireInDays":        "30",
					"expiryDateTimezone":  common.TIMEZONE_UTC,
					"deleteDatabase":      "true",
					"refreshInDays":       "7",
					"refreshTime":         "00:00:00",
					"refreshDateTimezone": common.TIMEZONE_UTC,
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should error out for invalid Postgres additionalArguments", func() {
				clone := createDefaultClone("clone14")
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"invalid": "invalid",
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_POSTGRES)))
			})
		})

		When("AdditionalArguments for MongoDB specified", func() {
			It("Should not error for valid MongoDB additionalArguments", func() {
				clone := createDefaultClone("clone15")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MONGODB
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"expireInDays":        "30",
					"expiryDateTimezone":  common.TIMEZONE_UTC,
					"deleteDatabase":      "true",
					"refreshInDays":       "7",
					"refreshTime":         "00:00:00",
					"refreshDateTimezone": common.TIMEZONE_UTC,
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should error out for invalid MongoDB additionalArguments", func() {
				clone := createDefaultClone("clone16")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MONGODB
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"invalid": "invalid",
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_MONGODB)))
			})
		})

		When("AdditionalArguments for MSSQL specified", func() {
			It("Should not error for valid MSSQL additionalArguments", func() {
				clone := createDefaultClone("clone17")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MSSQL
				clone.Spec.Clone.Profiles = &Profiles{
					Software: Profile{
						Name: "MAZIN-MSSQL2",
					},
				}
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"vm_name":                     "",
					"database_name":               "",
					"vm_dbserver_admin_password":  "",
					"dbserver_description":        "",
					"sql_user_name":               "",
					"authentication_mode":         "",
					"instance_name":               "",
					"windows_domain_profile_id":   "",
					"era_worker_service_user":     "",
					"sql_service_startup_account": "",
					"vm_win_license_key":          "",
					"target_mountpoints_location": "",
					"expireInDays":                "30",
					"expiryDateTimezone":          common.TIMEZONE_UTC,
					"deleteDatabase":              "true",
					"refreshInDays":               "7",
					"refreshTime":                 "00:00:00",
					"refreshDateTimezone":         common.TIMEZONE_UTC,
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Should error out for invalid MSSQL additionalArguments", func() {
				clone := createDefaultClone("clone18")
				clone.Spec.Clone.Type = common.DATABASE_TYPE_MSSQL
				clone.Spec.Clone.AdditionalArguments = map[string]string{
					"invalid": "invalid",
				}
				clone.Spec.Clone.Profiles = &Profiles{
					Software: Profile{
						Name: "MAZIN-MSSQL2",
					},
				}

				err := k8sClient.Create(context.Background(), clone)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring(fmt.Sprintf("additional arguments validation for type: %s failed!", common.DATABASE_TYPE_MSSQL)))
			})
		})
	})
})

func createDefaultDatabase(metadataName string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metadataName,
			Namespace: NAMESPACE,
		},
		Spec: DatabaseSpec{
			NDBRef:  NDB_REF,
			IsClone: false,
			Instance: &Instance{
				Name:                NAME,
				ClusterId:           DEFAULT_UUID,
				CredentialSecret:    CREDENTIAL_SECRET,
				Size:                SIZE,
				Type:                common.DATABASE_TYPE_POSTGRES,
				Profiles:            &(Profiles{}),
				AdditionalArguments: map[string]string{},
				IsHighAvailability:  HA,
			},
		},
	}
}

func createDefaultClone(metadataName string) *Database {
	return &Database{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metadataName,
			Namespace: NAMESPACE,
		},
		Spec: DatabaseSpec{
			NDBRef:  NDB_REF,
			IsClone: true,
			Clone: &Clone{
				Name:                NAME,
				Type:                common.DATABASE_TYPE_POSTGRES,
				ClusterId:           DEFAULT_UUID,
				CredentialSecret:    CREDENTIAL_SECRET,
				TimeZone:            common.TIMEZONE_UTC,
				SourceDatabaseId:    DEFAULT_UUID,
				SnapshotId:          DEFAULT_UUID,
				Profiles:            &(Profiles{}),
				AdditionalArguments: map[string]string{},
			},
		},
	}
}
