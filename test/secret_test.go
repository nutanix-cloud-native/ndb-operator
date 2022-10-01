package test

import (
	"context"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Test Secret Utility", func() {
	Context("util.get_data_from_secret test", func() {
		const ndbSecretName = "test-ndb-secret-name"
		const instanceSecretName = "test-instance-secret-name"
		const namespaceName = "test-namespace"

		const username = "test-username"
		const password = "test-password"
		const sshPublicKey = "test-ssh-key"

		ctx := context.Background()

		var namespace *corev1.Namespace
		var ndbSecret *corev1.Secret
		var instanceSecret *corev1.Secret

		ndbSecretTypeNamespaceName := types.NamespacedName{Name: ndbSecretName, Namespace: namespaceName}
		instanceSecretTypeNamespaceName := types.NamespacedName{Name: instanceSecretName, Namespace: namespaceName}

		BeforeEach(func() {
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namespaceName,
					Namespace: namespaceName,
				},
			}
			ndbSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ndbSecretName,
					Namespace: namespaceName,
				},
				Type: corev1.SecretTypeOpaque,
				StringData: map[string]string{
					ndbv1alpha1.NDB_PARAM_USERNAME: username,
					ndbv1alpha1.NDB_PARAM_PASSWORD: password,
				},
			}
			instanceSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instanceSecretName,
					Namespace: namespaceName,
				},
				Type: corev1.SecretTypeOpaque,
				StringData: map[string]string{
					ndbv1alpha1.NDB_PARAM_PASSWORD:       password,
					ndbv1alpha1.NDB_PARAM_SSH_PUBLIC_KEY: sshPublicKey,
				},
			}
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

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

		It("should get the data", func() {
			By("fetching and checking the NDB secrets")
			foundNDBSecret := &corev1.Secret{}
			err := k8sClient.Get(ctx, ndbSecretTypeNamespaceName, foundNDBSecret)
			Expect(err).To(Not(HaveOccurred()))
			//Checking data from ndb secret
			// username
			data, err := util.GetDataFromSecret(ctx, k8sClient, ndbSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_USERNAME)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data).To(Equal(username))
			// password
			data, err = util.GetDataFromSecret(ctx, k8sClient, ndbSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_PASSWORD)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data).To(Equal(password))

			By("fetching and checking the instance secrets")
			foundInstanceSecret := &corev1.Secret{}
			err = k8sClient.Get(ctx, instanceSecretTypeNamespaceName, foundInstanceSecret)
			Expect(err).To(Not(HaveOccurred()))
			//Checking data from instance secret
			// password
			data, err = util.GetDataFromSecret(ctx, k8sClient, instanceSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_PASSWORD)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data).To(Equal(password))
			// ssh key
			data, err = util.GetDataFromSecret(ctx, k8sClient, instanceSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_SSH_PUBLIC_KEY)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data).To(Equal(sshPublicKey))

			By("returning error when secrets are not present")
			// To simulate the situation when no secrets are present
			_ = k8sClient.Delete(ctx, ndbSecret)
			_ = k8sClient.Delete(ctx, instanceSecret)
			_, err = util.GetDataFromSecret(ctx, k8sClient, ndbSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_USERNAME)
			Expect(err).To(HaveOccurred())
			_, err = util.GetDataFromSecret(ctx, k8sClient, ndbSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_PASSWORD)
			Expect(err).To(HaveOccurred())
			_, err = util.GetDataFromSecret(ctx, k8sClient, instanceSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_PASSWORD)
			Expect(err).To(HaveOccurred())
			_, err = util.GetDataFromSecret(ctx, k8sClient, instanceSecretName, namespaceName, ndbv1alpha1.SECRET_DATA_KEY_SSH_PUBLIC_KEY)
			Expect(err).To(HaveOccurred())
		})
	})
})
