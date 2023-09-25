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
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	//+kubebuilder:scaffold:imports

	"github.com/nutanix-cloud-native/ndb-operator/api"

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

	Describe("NDB Validation", func() {
		When("Spec field is missing", func() {
			It("Throws an rrror", func() {
				database := ndbSpecMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("NDB server spec must be provided!"))
			})
		})
		When("Cluster Id is missing", func() {
			It("Throws an error", func() {
				database := ndbClusterIdMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("NDB ClusterId must be provided and be a valid UUID!"))

			})
		})
		When("CredentialSecret missing", func() {
			It("Throws an error", func() {
				database := ndbCredentialSecretMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("NDB CredentialSecret must be provided!"))

			})
		})

		When("Server URL missing", func() {
			It("Throws an error", func() {
				database := ndbServerURLMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("NDB Server URL must be provided and be a valid URL!"))
			})
		})
	})

	Context("Validation Tests", func() {

		It("Database Instance Name missing", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "",
						Type:                 "postgres",
						Size:                 10,
						TimeZone:             "UTC",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid Database Instance Name must be specified"))

		})

		It("Instance CredentialSecret missing", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						DatabaseInstanceName: "db-instance-name",
						Type:                 "postgres",
						Size:                 10,
						TimeZone:             "UTC",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("CredentialSecret must be provided in the Instance Spec"))

		})

		It("Type missing", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("A valid database type must be specified"))

		})

		It("Database Size < 10 GB", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Type:                 "postgres",
						Size:                 8,
						TimeZone:             "UTC",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("Initial Database size must be specified with a value 10 GBs or more"))

		})

		It("Profiles missing: Open-source engines", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "postgres",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Profiles missing: Closed-source engines", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mssql",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring("Software Profile must be provided for the closed-source database engines"))
		})

		It("Time Machine missing", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					// need to provide a unique name to avoid getting " "db" already exists"
					// this is because the previous test was a happy case and a db CR was successfully created
					Name:      "db1",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "postgres",
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TypeDetails for mysql valid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db2",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mysql",
						/* Valid arguments. These are optional */
						TypeDetails: map[string]string{
							"listener_port": "3306",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TypeDetails for mysql invalid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db3",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mysql",
						TypeDetails: map[string]string{
							"invalid": "invalid",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details may optionally be provided. Valid values are: %s", reflect.ValueOf(api.AllowedMySqlTypeDetails).MapKeys())))
		})

		It("TypeDetails for postgres valid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db4",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "postgres",
						/* Valid arguments. These are optional */
						TypeDetails: map[string]string{
							"listener_port": "5432",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TypeDetails for postgres invalid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db5",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "postgres",
						TypeDetails: map[string]string{
							"invalid": "invalid",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details may optionally be provided. Valid values are: %s", reflect.ValueOf(api.AllowedPostGresTypeDetails).MapKeys())))
		})

		It("TypeDetails for mongodb valid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db6",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mongodb",
						/* Valid arguments. These are optional */
						TypeDetails: map[string]string{
							"listener_port": "5432",
							"log_size":      "10",
							"journal_size":  "10",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TypeDetails for mongodb invalid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db7",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mongodb",
						TypeDetails: map[string]string{
							"invalid": "invalid",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details may optionally be provided. Valid values are: %s", sortKeys(api.AllowedMongoDBTypeDetails))))
		})

		It("TypeDetails for mssql valid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db8",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mssql",
						Profiles: &Profiles{
							Software: Profile{
								Name: "MSSQL-MAZIN2",
							},
						},
						/* Valid arguments. These are optional */
						TypeDetails: map[string]string{
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
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			fmt.Print(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TypeDetails for mssql invalid", func() {

			database := &Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "db9",
					Namespace: "default",
				},
				Spec: DatabaseSpec{
					NDB: NDB{
						ClusterId:                   "27bcce67-7b83-42c2-a3fe-88154425c170",
						SkipCertificateVerification: true,
						CredentialSecret:            "ndb-secret",
						Server:                      "https://10.51.140.43:8443/era/v0.9",
					},
					Instance: Instance{
						CredentialSecret:     "db-instance-secret",
						DatabaseInstanceName: "db-instance-name",
						Size:                 10,
						TimeZone:             "UTC",
						Type:                 "mssql",
						Profiles: &Profiles{
							Software: Profile{
								Name: "MSSQL-MAZIN2",
							},
						},
						TypeDetails: map[string]string{
							"invalid": "invalid",
						},
					},
				},
			}
			err := k8sClient.Create(context.Background(), database)
			Expect(err).To(HaveOccurred())

			// Extract the error message from the error object
			errMsg := err.(*errors.StatusError).ErrStatus.Message
			Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details may optionally be provided. Valid values are: %s", sortKeys(api.AllowedMsSqlTypeDetails))))
		})

	})
})
