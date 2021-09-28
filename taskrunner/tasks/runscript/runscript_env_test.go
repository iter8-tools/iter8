package runscript

import (
	"context"
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

var _ = Describe("run task", func() {
	Context("When Secret with token is specified", func() {
		It("should the correct interpolate script with secret", func() {
			By("creating a context with an experiment")
			exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/common/runexperiment.yaml")).Build()
			Expect(err).NotTo(HaveOccurred())
			ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

			By("creating a secret with token")
			s := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "top-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"token": []byte("great secret"),
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(ctx, &s)).NotTo(HaveOccurred())

			By("creating a run with secret")
			secret, _ := json.Marshal("default/top-secret")
			task, err := Make(&v2alpha2.TaskSpec{
				Run: core.StringPointer(`echo {{ .Secret "token" }}`),
				With: map[string]apiextensionsv1.JSON{
					"secret": {Raw: secret},
				},
			})
			log.Info(*task.(*Task).TaskMeta.Run)
			Expect(err).ToNot(HaveOccurred())

			err = task.Run(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(task.(*Task).With.interpolatedRun).To(Equal("echo great secret"))
		})
	})
})
