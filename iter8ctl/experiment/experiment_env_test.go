package experiment

import (
	"context"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	tasks "github.com/iter8-tools/etc3/taskrunner/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("iter8ctl", func() {
	// cleanup cluster
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(v2alpha2.GroupVersion.WithKind("experiment"))
	BeforeEach(func() {
		k8sClient.DeleteAllOf(context.Background(), u, client.InNamespace("default"))
	})

	Context("when cluster has an experiment", func() {
		var exp *tasks.Experiment
		var err error
		It("should retrieve experiment properly", func() {
			By("reading the experiment from file")
			exp, err = (&tasks.Builder{}).FromFile(tasks.CompletePath("../", "testdata/experiment1.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), &exp.Experiment)).To(Succeed())

			By("fetching experiment from cluster using latest flag")
			exp2, err := GetExperiment(true, "dummy", "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(exp2.Spec).To(Equal(exp.Spec))

			By("fetching experiment from cluster using name and namespace")
			exp2, err = GetExperiment(false, "sklearn-iris-experiment-1", "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(exp2.Spec).To(Equal(exp.Spec))

			By("failing to fetch experiment from cluster when name is wrong")
			exp2, err = GetExperiment(false, "so-wrong", "default")
			Expect(err).To(HaveOccurred())
			Expect(exp2).To(BeNil())

			By("failing to fetch experiment from cluster when namespace is wrong")
			exp2, err = GetExperiment(false, "sklearn-iris-experiment-1", "so-wrong")
			Expect(err).To(HaveOccurred())
			Expect(exp2).To(BeNil())
		})
	})

	Context("when cluster does not have any experiment", func() {
		It("should fail when", func() {
			By("fetching experiment from cluster using latest flag")
			exp, err := GetExperiment(true, "dummy", "dummy")
			Expect(err).To(HaveOccurred())
			Expect(exp).To(BeNil())

			By("fetching experiment from cluster using name and namespace")
			exp, err = GetExperiment(false, "dummy", "default")
			Expect(err).To(HaveOccurred())
			Expect(exp).To(BeNil())
		})
	})

})
