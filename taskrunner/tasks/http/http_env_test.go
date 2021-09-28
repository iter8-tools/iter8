package http

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	// "k8s.io/apimachinery/pkg/types"
	// "sigs.k8s.io/controller-runtime/pkg/client"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("notification/http task", func() {
	Context("When authType is Basic", func() {
		It("should the correct authorization header", func() {
			By("creating a context with an experiment")
			exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment1.yaml")).Build()
			Expect(err).NotTo(HaveOccurred())
			ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

			By("creating a secret with username/password")
			s := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "basic-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"username": []byte("user"),
					"password": []byte("pass"),
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(ctx, &s)).NotTo(HaveOccurred())

			By("creating an task with authType Bearer")
			url, _ := json.Marshal("http://test")
			authType, _ := json.Marshal("Basic")
			secret, _ := json.Marshal("default/basic-secret")
			task, err := Make(&v2alpha2.TaskSpec{
				Task: core.StringPointer(TaskName),
				With: map[string]apiextensionsv1.JSON{
					"URL":      {Raw: url},
					"authType": {Raw: authType},
					"secret":   {Raw: secret},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			By("preparing the task")
			req, err := task.(*Task).prepareRequest(ctx)
			Expect(err).ToNot(HaveOccurred())

			By("checking the Authorization header")
			authHeader := req.Header.Get("Authorization")
			Expect(authHeader).To(Equal("Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))))
		})
	})
	Context("When authType is Bearer", func() {
		It("should the correct authorization header", func() {
			By("creating a context with an experiment")
			exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment1.yaml")).Build()
			Expect(err).NotTo(HaveOccurred())
			ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

			By("creating a secret with username/password")
			s := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bearer-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"token": []byte(base64.StdEncoding.EncodeToString([]byte("token"))),
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(ctx, &s)).NotTo(HaveOccurred())

			By("creating an task with authType Bearer")
			url, _ := json.Marshal("http://test")
			authType, _ := json.Marshal("Bearer")
			secret, _ := json.Marshal("default/bearer-secret")
			task, err := Make(&v2alpha2.TaskSpec{
				Task: core.StringPointer(TaskName),
				With: map[string]apiextensionsv1.JSON{
					"URL":      {Raw: url},
					"authType": {Raw: authType},
					"secret":   {Raw: secret},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			By("preparing the task")
			req, err := task.(*Task).prepareRequest(ctx)
			Expect(err).ToNot(HaveOccurred())

			By("checking the Authorization header")
			authHeader := req.Header.Get("Authorization")
			Expect(authHeader).To(Equal("Bearer " + base64.StdEncoding.EncodeToString([]byte("token"))))
		})
	})
})
