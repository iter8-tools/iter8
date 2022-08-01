package watcher

import (
	"testing"

	"github.com/iter8-tools/iter8/driver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewInformer(t *testing.T) {
	kd := driver.NewFakeKubeDriver(cli.New())
	// byteArray, _ := ioutil.ReadFile(base.CompletePath("../../testdata", "abninputs/readtest.yaml"))
	// s, _ := kd.Clientset.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      app,
	// 		Namespace: "default",
	// 	},
	// 	StringData: map[string]string{"versionData.yaml": string(byteArray)},
	// }, metav1.CreateOptions{})
	// s.ObjectMeta.Labels = map[string]string{"foo": "bar"}
	// kd.Clientset.CoreV1().Secrets("default").Update(context.TODO(), s, metav1.UpdateOptions{})

	informer := NewInformer(
		kd,
		[]schema.GroupVersionResource{{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		}, {
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}},
		[]string{"default", "foo"},
	)
	assert.NotNil(t, informer)
	// 2 resource types for 2 namespaces
	assert.Equal(t, 4, len(informer.informersByKey))
}
