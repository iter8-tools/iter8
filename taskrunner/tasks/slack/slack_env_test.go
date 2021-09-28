package slack

import (
	"context"

	"github.com/iter8-tools/etc3/taskrunner/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GetSecretToken", func() {
	Context("when proper secret is present in the cluster", func() {
		It("should retrieve token properly if present", func() {
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
			st := &Task{
				TaskMeta: core.TaskMeta{},
				With: Inputs{
					Secret: "default/mysecret",
				},
			}
			token := st.getToken()
			Expect(*token).To(Equal("abc123"))

			Expect(k8sClient.Delete(context.Background(), &secret)).To(Succeed())
		})

		It("should retrieve no token if not present", func() {
			By("building a secret")
			secret := corev1.Secret{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"nontoken": []byte("abc123"),
				},
			}

			By("updating secret in cluster")
			Expect(k8sClient.Create(context.Background(), &secret)).To(Succeed())

			By("fetching secret from cluster secret")
			st := &Task{
				TaskMeta: core.TaskMeta{},
				With: Inputs{
					Secret: "default/mysecret",
				},
			}
			token := st.getToken()
			Expect(token).To(BeNil())

			Expect(k8sClient.Delete(context.Background(), &secret)).To(Succeed())
		})

	})
})
