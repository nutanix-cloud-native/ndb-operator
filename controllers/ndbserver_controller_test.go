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

package controllers

import (
	"context"
	b64 "encoding/base64"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("NDBServer controller", func() {
	Context("NDBServer controller test", func() {
		var namespace *corev1.Namespace
		var ndbSecret *corev1.Secret

		const username = "test-username"
		const password = "test-password"
		const namespaceName = "test-namespace-ndb-controller"
		const ndbSecretName = "test-ndb-secret-name"

		ndbSecretTypeNamespaceName := types.NamespacedName{Name: ndbSecretName, Namespace: namespaceName}

		ctx := context.Background()

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

			By("Creating the secrets")
			err = k8sClient.Get(ctx, ndbSecretTypeNamespaceName, ndbSecret)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, ndbSecret)
				Expect(err).To(Not(HaveOccurred()))
			}

		})

		AfterEach(func() {
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("Should reconcile a NDBServer CR", func() {
			By("Creating a NDBServer CR")
			ndbServer := &v1alpha1.NDBServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ndbSecretName,
					Namespace: namespaceName,
				},
				Spec: v1alpha1.NDBServerSpec{
					Server:           "https:111.222.333.444",
					CredentialSecret: ndbSecretName,
				},
				Status: v1alpha1.NDBServerStatus{},
			}
			err := k8sClient.Create(ctx, ndbServer)
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
