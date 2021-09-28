package readiness

import (
	"context"
	"errors"
	"strings"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeCommand struct {
	err  error
	name string
	arg  []string
}

func (f *fakeCommand) Run() error {
	return f.err
}

func (f *fakeCommand) String() string {
	elems := append([]string{f.name}, f.arg...)
	return strings.Join(elems, " ")
}

var _ = Describe("Readiness task", func() {
	Context("when missing specified resources", func() {
		var exp *core.Experiment
		var err error

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(v2alpha2.GroupVersion.WithKind("experiment"))
		BeforeEach(func() {
			k8sClient.DeleteAllOf(context.Background(), u, client.InNamespace("default"))
		})

		It("should initialize a conformance experiment", func() {
			By("reading the experiment from file")
			exp, err = (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/common/readinessexp1.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())

			By("getting the experiment from the cluster")
			exp2 := &core.Experiment{}
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Namespace: "default",
				Name:      "readiness-exp-1",
			}, exp2)).To(Succeed())

			By("populating context with the experiment")
			ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp2)

			By("creating a readiness task")
			taskSpec := exp2.Spec.Strategy.Actions["start"][0]
			readiness, err := Make(&taskSpec)
			Expect(err).ToNot(HaveOccurred())

			By("running the readiness task")
			// first fake the commands...
			getCommand = func(name string, arg ...string) command {
				return &fakeCommand{
					err:  nil,
					name: "my",
					arg:  []string{"fake", "command"},
				}
			}
			// this should succeed... since the command has been faked to succeed
			Expect(readiness.Run(ctx)).ToNot(HaveOccurred())

			// fake the commands again... this time with failure...
			getCommand = func(name string, arg ...string) command {
				return &fakeCommand{
					err:  errors.New("Fake command failures have occurred"),
					name: "my",
					arg:  []string{"fake", "command"},
				}
			}
			// this should fail... since the command has been faked to fail
			Expect(readiness.Run(ctx)).To(HaveOccurred())
		})
	})
})
