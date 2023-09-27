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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	//+kubebuilder:scaffold:imports

	"k8s.io/apimachinery/pkg/api/errors"
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

	Describe("DB Validation", func() {
		When("'databaseInstanceName' is missing", func() {
			It("Throws an error", func() {
				database := dbInstanceNameMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("A valid Database Instance Name must be specified in the Instance Spec!"))
			})
		})

		When("'description' not specified", func() {
			It("Sets a default 'description'", func() {
				database := dbDescriptionNotSpecified("db1")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
				// TODO: Check if 'description' was defaulted
			})
		})

		When("'databaseNames' not specified", func() {
			It("Sets a default 'databaseNames'", func() {
				database := dbDescriptionNotSpecified("db2")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
				// TODO: Check if 'databaseNames' were defaulted
			})
		})

		When("'credentialSecret' is missing", func() {
			It("Throws an error'", func() {
				database := dbCredentialSecretMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("CredentialSecret must be provided in the Instance Spec!"))
			})
		})

		When("'size' is less than 10", func() {
			It("Throws an error'", func() {
				database := dbSizeLessThan10()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("Initial Database size must be specified with a value 10 GBs or more in the Instance Spec!"))
			})
		})

		When("'timeZone' not specified", func() {
			It("Sets a default 'timeZone'", func() {
				database := dbDescriptionNotSpecified("db3")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
				// TODO: Check if 'timeZone' was defaulted
			})
		})

		When("'type' missing", func() {
			It("Throws an error'", func() {
				database := dbTypeMissing()
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("A valid database type must be specified in the Instance Spec!"))
			})
		})

		When("'type' invalid", func() {
			It("Throws an error'", func() {
				database := dbWithType("db", "invalid")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("A valid database type must be specified in the Instance Spec!"))
			})
		})

		When("'profiles' missing", func() {
			It("Passes because open-source engine was specified", func() {
				database := dbWithType("db4", "postgres")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Throws error because closed-source engine was specified with no software id", func() {
				database := dbWithType("db", "mssql")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).To(HaveOccurred())
				errMsg := err.(*errors.StatusError).ErrStatus.Message
				Expect(errMsg).To(ContainSubstring("Software Profile must be provided for the closed-source database engines in the Instance Spec!"))
			})
		})

		When("'timeMachine' not specified", func() {
			It("Sets default 'timeMachineInfo", func() {
				database := dbTimeMachineNotSpecified("db5")
				err := k8sClient.Create(context.Background(), database)
				Expect(err).ToNot(HaveOccurred())
				// TODO: Check if 'timeMachine' was defaulted
			})
		})

		Context("'typeDetails' is specified", func() {
			When("'type' is mysql", func() {
				It("Valid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db6",
						"mysql",
						map[string]string{
							"listener_port": "3306",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Invalid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db",
						"mysql",
						map[string]string{
							"invalid": "invalid",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).To(HaveOccurred())
					errMsg := err.(*errors.StatusError).ErrStatus.Message
					Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details for %s are invalid! Valid values are: ", "mysql")))
				})
			})

			When("'type' is postgres", func() {
				It("Valid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db7",
						"postgres",
						map[string]string{
							"listener_port": "5432",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Invalid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db",
						"postgres",
						map[string]string{
							"invalid": "invalid",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).To(HaveOccurred())
					errMsg := err.(*errors.StatusError).ErrStatus.Message
					Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details for %s are invalid! Valid values are: ", "postgres")))
				})
			})

			When("'type' is mongodb", func() {
				It("Valid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db8",
						"mongodb",
						map[string]string{
							"listener_port": "5432",
							"log_size":      "10",
							"journal_size":  "10",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Invalid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db",
						"mongodb",
						map[string]string{
							"invalid": "invalid",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).To(HaveOccurred())
					errMsg := err.(*errors.StatusError).ErrStatus.Message
					Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details for %s are invalid! Valid values are: ", "mongodb")))
				})
			})

			When("'type' is mssql", func() {
				It("Valid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db9",
						"mssql",
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
				It("Invalid typeDetails specified", func() {
					database := dbWithAdditionalArguments(
						"db",
						"mssql",
						map[string]string{
							"invalid": "invalid",
						},
					)
					err := k8sClient.Create(context.Background(), database)
					Expect(err).To(HaveOccurred())
					errMsg := err.(*errors.StatusError).ErrStatus.Message
					Expect(errMsg).To(ContainSubstring(fmt.Sprintf("Type Details for %s are invalid! Valid values are: ", "mssql")))
				})
			})
		})
	})
})
