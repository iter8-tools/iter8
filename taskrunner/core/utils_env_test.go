package core

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GetSecretToken", func() {
	Context("when proper secret is present in the cluster", func() {
		It("should retrieve token properly", func() {
			By("building a secret")
			secret := corev1.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"token": []byte("abc123"),
				},
			}

			By("creating secret in cluster")
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())

			By("fetching secret from cluster secret")
			s, err := GetSecret("default/mysecret")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).ToNot(BeNil())
			token, err := GetTokenFromSecret(s)
			Expect(err).ToNot(HaveOccurred())
			Expect(token).To(Equal("abc123"))
		})
	})
})
