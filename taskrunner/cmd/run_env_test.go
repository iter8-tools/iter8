package cmd

import (
	"context"
	"os"

	"github.com/iter8-tools/etc3/taskrunner/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Experiment's handler field", func() {
	var exp *core.Experiment
	var err error
	var head = func() {
		By("reading the experiment from file")
		exp, err = (&core.Builder{}).FromFile(core.CompletePath("../", "testdata/experiment6.yaml")).Build()
		Expect(err).ToNot(HaveOccurred())
	}
	var create = func() {
		By("creating experiment in cluster")
		Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())
		Expect(k8sClient.Status().Update(context.Background(), exp)).To(Succeed())
	}
	var runhandler = func(by string) {
		By(by)
		os.Setenv("EXPERIMENT_NAME", "sklearn-iris-experiment-6")
		os.Setenv("EXPERIMENT_NAMESPACE", "default")
		nn, err := getExperimentNN()
		Expect(err).ToNot(HaveOccurred())
		Expect("sklearn-iris-experiment-6").To(Equal(nn.Name))
		Expect("default").To(Equal(nn.Namespace))

		action = "start"
		runCmd.Run(nil, nil)
	}
	var tail = func(runby string) {
		create()
		runhandler(runby)
		Expect(k8sClient.Delete(context.Background(), exp)).To(Succeed())
	}

	Context("when containing handler actions", func() {
		It("should run handler", func() {
			head()
			tail("as a normal run")
		})
	})
	Context("when not containing the specified action", func() {
		It("should exit gracefully", func() {
			head()
			delete(exp.Spec.Strategy.Actions, "start")
			tail("with an error log")
		})
		It("should exit gracefully when ActionMap is nil", func() {
			head()
			exp.Spec.Strategy.Actions = nil
			tail("with an error log")
		})
	})
})
