package core

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Experiment's handler field", func() {
	Context("when containing handler actions", func() {
		var exp *Experiment
		var err error
		It("should retrieve handler info properly", func() {
			By("reading the experiment from file")
			exp, err = (&Builder{}).FromFile(CompletePath("../", "testdata/experiment1.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("creating experiment in cluster")
			Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())

			By("fetching experiment from cluster")
			b := &Builder{}
			exp2, err := b.FromCluster(&types.NamespacedName{
				Name:      "sklearn-iris-experiment-1",
				Namespace: "default",
			}).Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(exp2.Spec).To(Equal(exp.Spec))
		})

		It("should handle non-existing experiments properly", func() {
			By("signaling error")
			b := &Builder{}
			// store k8sclient defaults
			numAttempts := NumAttempt
			period := Period
			// change k8sclient defaults
			NumAttempt = 2
			Period = 2
			_, err := b.FromCluster(&types.NamespacedName{
				Name:      "non-existent",
				Namespace: "default",
			}).Build()
			// change k8sclient defaults
			NumAttempt = numAttempts
			Period = period
			Expect(err).To(HaveOccurred())
		})

		// It("should run handler", func() {
		// 	By("reading the experiment from file")
		// 	exp, err = (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment6.yaml")).Build()
		// 	Expect(err).ToNot(HaveOccurred())

		// 	By("creating experiment in cluster")
		// 	Expect(k8sClient.Create(context.Background(), exp)).To(Succeed())
		// 	Expect(k8sClient.Status().Update(context.Background(), exp)).To(Succeed())

		// 	By("fetching experiment from cluster")
		// 	b := &Builder{}
		// 	exp2, err := b.FromCluster("sklearn-iris-experiment-6", "default", k8sClient).Build()
		// 	Expect(err).ToNot(HaveOccurred())
		// 	Expect(exp2.Spec).To(Equal(exp.Spec))

		// 	By("running the experiment")
		// 	err = exp2.Run("start")
		// 	Expect(err).NotTo(HaveOccurred())
		// 	action := (*exp2.Spec.Strategy.Handlers.Actions)["start"]
		// 	execTask := (*action)[0].(*def.ExecTask)
		// 	arg := execTask.With.Args[1]
		// 	Expect("hello revision1 world").To(Equal(arg))
		// })

		// It("should deal with extrapolation errors", func() {
		// 	By("reading the experiment from file")
		// 	exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment7.yaml")).Build()
		// 	Expect(err).ToNot(HaveOccurred())

		// 	By("running and gracefully exiting")
		// 	err = exp.Run("start")
		// 	Expect(err).To(HaveOccurred())
		// })
	})
})

var _ = Describe("GetSecret", func() {
	Context("When call GetSecret for a valid secret", func() {
		It("should read the secret", func() {
			By("Creating a secret")
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"secretName": []byte("tester"),
				},
			}
			k8sClient.Create(context.Background(), &secret)
			By("Calling GetSecret")
			s, err := GetSecret("default/test-secret")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(s.Data["secretName"])).To(Equal("tester"))
		})
	})
})
