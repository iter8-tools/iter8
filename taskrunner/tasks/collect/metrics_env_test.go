package collect

import (
	"context"
	"encoding/json"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("metrics library", func() {
	Context("when running a metrics/collect task", func() {
		var exp *core.Experiment
		var err error

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(v2alpha2.GroupVersion.WithKind("experiment"))
		BeforeEach(func() {
			k8sClient.DeleteAllOf(context.Background(), u, client.InNamespace("default"))
		})

		It("should initialize an experiment", func() {
			By("reading the experiment from file")
			exp, err = (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/metricscollect/metricscollect.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())

			By("getting the experiment from the cluster")
			exp2 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "metrics-collect-exp",
			}, exp2)).To(Succeed())

			By("populating context with the experiment")
			ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp2)
			ctx = context.WithValue(ctx, core.ContextKey("action"), "start")

			By("creating a metrics/collect task")
			ct := CollectTask{
				TaskMeta: core.TaskMeta{
					Task: core.StringPointer(TaskName),
				},
				With: CollectInputs{
					Versions: []Version{
						{
							Name: "default",
							URL:  "https://httpbin.org",
						},
						{
							Name: "canary",
							URL:  "https://httpbin.org/stream/1",
						},
					},
				},
			}
			ct.InitializeDefaults()

			By("running the metrics/collect task")
			Expect(ct.Run(ctx)).ToNot(HaveOccurred())

			By("getting the experiment from cluster")
			exp3 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "metrics-collect-exp",
			}, exp3)).To(Succeed())

			By("confirming that the experiment looks right")
			Expect(exp3.Status.Analysis.AggregatedBuiltinHists).ToNot(BeNil())

			By("running the metrics/collect task again")
			Expect(ct.Run(ctx)).ToNot(HaveOccurred())

			By("getting the experiment from cluster")
			exp4 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "metrics-collect-exp",
			}, exp4)).To(Succeed())

			By("confirming that the experiment looks right")
			fortioData := make(map[string]*Result)

			Expect(exp4.Status.Analysis.AggregatedBuiltinHists).ToNot(BeNil())
			jsonBytes, err := json.Marshal(exp4.Status.Analysis.AggregatedBuiltinHists.Data)
			// convert jsonBytes to fortioData
			Expect(err).ShouldNot(HaveOccurred())
			err = json.Unmarshal(jsonBytes, &fortioData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fortioData).ToNot(BeNil())
			Expect(fortioData["default"]).ToNot(BeNil())
			Expect(fortioData["canary"]).ToNot(BeNil())
			Expect(fortioData["canary"].DurationHistogram.Count).To(Equal(int(2 * DefaultNumQueries)))
		}) // it

		It("should initialize an experiment", func() {
			By("reading the experiment from file")
			exp, err = (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/metricscollect/loadgen.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())

			By("getting the experiment from the cluster")
			exp2 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "loadgen-exp",
			}, exp2)).To(Succeed())

			By("populating context with the experiment")
			ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp2)
			ctx = context.WithValue(ctx, core.ContextKey("action"), "start")

			By("creating a metrics/collect task")
			ct := CollectTask{
				TaskMeta: core.TaskMeta{
					Task: core.StringPointer(TaskName),
				},
				With: CollectInputs{
					LoadOnly: core.BoolPointer(true),
					Versions: []Version{
						{
							Name: "default",
							URL:  "https://httpbin.org",
						},
						{
							Name: "canary",
							URL:  "https://httpbin.org/stream/1",
						},
					},
				},
			}
			ct.InitializeDefaults()

			By("running the metrics/collect task")
			Expect(ct.Run(ctx)).ToNot(HaveOccurred())

			By("getting the experiment from cluster")
			exp3 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "loadgen-exp",
			}, exp3)).To(Succeed())

			By("confirming that the experiment looks right")
			Expect(exp3.Status.Analysis).To(BeNil())

			By("running the metrics/collect task again")
			Expect(ct.Run(ctx)).ToNot(HaveOccurred())

			By("getting the experiment from cluster")
			exp4 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "loadgen-exp",
			}, exp4)).To(Succeed())

			Expect(exp4.Status.Analysis).To(BeNil())
		}) // it

	})
})
