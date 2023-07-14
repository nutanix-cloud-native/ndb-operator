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

/*
GENERATED by operator-sdk
Changes added
*/

package controllers

import (
	"context"
	"time"

	b64 "encoding/base64"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Database controller", func() {
	Context("Database controller test", func() {

		DatabaseName := "test-database-name"
		timezoneUTC := "UTC"
		typePostgres := "postgres"
		databaseSize := 10
		const namespaceName = "test-namespace"
		const testNDBServer = "http://123.123.123.123:123"
		const testDatabaseIP = "111.222.123.234"

		ctx := context.Background()

		const ndbSecretName = "test-ndb-secret-name"
		instanceSecretName := "test-instance-secret-name"
		const username = "test-username"
		const password = "test-password"
		const sshPublicKey = "test-ssh-key"

		var namespace *corev1.Namespace
		var ndbSecret *corev1.Secret
		var instanceSecret *corev1.Secret
		database := &ndbv1alpha1.Database{}

		typeNamespaceName := types.NamespacedName{Name: DatabaseName, Namespace: namespaceName}
		typeNamespaceNameForService := types.NamespacedName{Name: DatabaseName + "-svc", Namespace: namespaceName}
		ndbSecretTypeNamespaceName := types.NamespacedName{Name: ndbSecretName, Namespace: namespaceName}
		instanceSecretTypeNamespaceName := types.NamespacedName{Name: instanceSecretName, Namespace: namespaceName}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namespaceName,
					Namespace: namespaceName,
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			database = &ndbv1alpha1.Database{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DatabaseName,
					Namespace: namespaceName,
				},
				Spec: ndbv1alpha1.DatabaseSpec{
					NDB: ndbv1alpha1.NDB{
						ClusterId:        "abcd",
						CredentialSecret: ndbSecretName,
						Server:           testNDBServer,
					},
					Instance: ndbv1alpha1.Instance{
						DatabaseInstanceName: &DatabaseName,
						DatabaseNames:        []string{"database_1"},
						CredentialSecret:     &instanceSecretName,
						Size:                 &databaseSize,
						TimeZone:             &timezoneUTC,
						Type:                 &typePostgres,
						TMInfo: &ndbv1alpha1.DBTimeMachineInfo{
							QuarterlySnapshotMonth: "Jan",
							SnapshotsPerDay:        4,
							LogCatchUpFrequency:    90,
							WeeklySnapshotDay:      "WEDNESDAY",
							MonthlySnapshotDay:     24,
							DailySnapshotTime:      "12:34:56",
						},
					},
				},
			}

			ndbSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ndbSecretName,
					Namespace: namespaceName,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					common.NDB_PARAM_USERNAME: []byte(b64.StdEncoding.EncodeToString([]byte(username))),
					common.NDB_PARAM_PASSWORD: []byte(b64.StdEncoding.EncodeToString([]byte(password))),
				},
			}
			instanceSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instanceSecretName,
					Namespace: namespaceName,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					common.NDB_PARAM_PASSWORD:       []byte(b64.StdEncoding.EncodeToString([]byte(password))),
					common.NDB_PARAM_SSH_PUBLIC_KEY: []byte(b64.StdEncoding.EncodeToString([]byte(sshPublicKey))),
				},
			}

			By("Creating the secrets")
			err = k8sClient.Get(ctx, ndbSecretTypeNamespaceName, ndbSecret)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, ndbSecret)
				Expect(err).To(Not(HaveOccurred()))
			}
			err = k8sClient.Get(ctx, instanceSecretTypeNamespaceName, instanceSecret)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, instanceSecret)
				Expect(err).To(Not(HaveOccurred()))
			}

		})

		AfterEach(func() {
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("should reconcile a custom resource for Database", func() {
			By("Creating the custom resource for the Kind Database")
			err := k8sClient.Get(ctx, typeNamespaceName, database)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples

				err = k8sClient.Create(ctx, database)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &ndbv1alpha1.Database{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			// Subsequent reconciles will need to make HTTP calls to the NDB server
			// that will cause an error. Even if we use a mocked HTTP Server (httptest),
			// that mock server will still not be reachable from the controller as the
			// controller test runs with a mock kubernetes api server (provided by envtest)
			// and the mock testhttp server will not be reachable from envtest k8s api server.
			By("Reconciling the custom resource created once")
			DatabaseReconciler := &DatabaseReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err = DatabaseReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Setting up network connectivity")
			// Updating status to mock successful provisioning of a database on NDB
			database.Status.DatabaseServerId = "mock-database-server-id"
			database.Status.Status = "READY"
			database.Status.IPAddress = testDatabaseIP
			k8sClient.Status().Update(ctx, database)
			// Setting up network connectivity
			err = DatabaseReconciler.setupConnectivity(ctx, database, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			// Check service is created
			createdService := &corev1.Service{}
			err = k8sClient.Get(ctx, typeNamespaceNameForService, createdService)
			Expect(err).To(Not(HaveOccurred()))
			Expect(createdService.Name).To(Equal(typeNamespaceNameForService.Name))

			// Check if endpoint is created
			createdEndpoints := &corev1.Endpoints{}
			err = k8sClient.Get(ctx, typeNamespaceNameForService, createdEndpoints)
			Expect(err).To(Not(HaveOccurred()))
			Expect(createdEndpoints.Name).To(Equal(typeNamespaceNameForService.Name))
			// Check if endpoint maps to the same IP as the Database.Status.IPAddress
			Expect(createdEndpoints.Subsets[0].Addresses[0].IP).To(Equal(testDatabaseIP))
			// Check if service name is same as endpoints name
			Expect(createdService.Name).To(Equal(createdEndpoints.Name))
		})
	})
})
