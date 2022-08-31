package watcher

import (
	"testing"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/abn/k8sdriver"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type scenario struct {
	iter8       string
	application string
	version     string
	track       string
	ready       string
}

func TestAdd(t *testing.T) {
	testcases := map[string]scenario{
		"iter8 not set":       {iter8: "", application: "", version: "", track: "", ready: ""},
		"iter8 not true":      {iter8: "false", application: "ns/app", version: "version", track: "foo", ready: "true"},
		"no application":      {iter8: "true", application: "", version: "", track: "", ready: ""},
		"no version":          {iter8: "true", application: "ns/app", version: "", track: "", ready: ""},
		"w/o track ready":     {iter8: "true", application: "ns/app", version: "version", track: "", ready: "true"},
		"w/o track not ready": {iter8: "true", application: "ns/app", version: "version", track: "", ready: "false"},
		"w/ track ready":      {iter8: "true", application: "ns/app", version: "version", track: "track", ready: "true"},
		"w/ track not ready":  {iter8: "true", application: "ns/app", version: "version", track: "track", ready: ""},
	}
	for label, s := range testcases {
		t.Run(label, func(t *testing.T) {
			setup()
			testAdd(t, s)
		})
	}
}

func testAdd(t *testing.T, s scenario) {
	addObject(wo(s.iter8, s.application, s.version, s.track, s.ready))

	if s.iter8 != "true" ||
		s.application == "" ||
		s.version == "" {
		abnapp.NumApplications(t, 0)
		return
	}

	abnapp.NumApplications(t, 1)

	a, _ := abnapp.Applications.Get(s.application, true)
	assert.NotNil(t, a)

	tracks := []string{}
	if s.track != "" && s.ready == "true" {
		tracks = []string{s.track}
	}
	assertApplication(t, a, applicationAssertion{
		namespace: abnapp.GetNamespaceFromKey(s.application),
		name:      abnapp.GetNameFromKey(s.application),
		tracks:    tracks,
		versions:  []string{s.version},
	})
}

func TestUpdate(t *testing.T) {
	testcases := map[string]scenario{
		"no application":      {iter8: "true", application: "", version: "", track: "", ready: ""},
		"no version":          {iter8: "true", application: "ns/app", version: "", track: "", ready: ""},
		"w/o track ready":     {iter8: "true", application: "ns/app", version: "version", track: "", ready: "true"},
		"w/o track not ready": {iter8: "true", application: "ns/app", version: "version", track: "", ready: "false"},
		"w/ track ready":      {iter8: "true", application: "ns/app", version: "version", track: "track", ready: "true"},
		"w/ track not ready":  {iter8: "true", application: "ns/app", version: "version", track: "track", ready: ""},
	}
	for label, s := range testcases {
		t.Run(label, func(t *testing.T) {
			setup()
			addObject(wo("true", "ns/app", "version", "", ""))
			testUpdate(t, s)
		})
	}
}

func testUpdate(t *testing.T, s scenario) {
	updateObject(wo(s.iter8, s.application, s.version, s.track, s.ready))

	abnapp.NumApplications(t, 1)
	a, _ := abnapp.Applications.Get("ns/app", true)
	assert.NotNil(t, a)

	if s.iter8 != "true" ||
		s.application == "" ||
		s.version == "" {
		assertApplication(t, a, applicationAssertion{
			namespace: "ns",
			name:      "app",
			tracks:    []string{},
			versions:  []string{"version"},
		})
	} else {
		tracks := []string{}
		if s.track != "" && s.ready == "true" {
			tracks = []string{s.track}
		}
		assertApplication(t, a, applicationAssertion{
			namespace: abnapp.GetNamespaceFromKey(s.application),
			name:      abnapp.GetNameFromKey(s.application),
			tracks:    tracks,
			versions:  []string{s.version},
		})
	}
}

func TestDelete(t *testing.T) {
	testcases := map[string]scenario{
		"no application":      {iter8: "true", application: "", version: "", track: "", ready: ""},
		"no version":          {iter8: "true", application: "ns/app", version: "", track: "", ready: ""},
		"w/o track ready":     {iter8: "true", application: "ns/app", version: "version", track: "", ready: "true"},
		"w/o track not ready": {iter8: "true", application: "ns/app", version: "version", track: "", ready: "false"},
		"w/ track ready":      {iter8: "true", application: "ns/app", version: "version", track: "track", ready: "true"},
		"w/ track not ready":  {iter8: "true", application: "ns/app", version: "version", track: "track", ready: ""},
	}
	for label, s := range testcases {
		t.Run(label, func(t *testing.T) {
			setup()
			addObject(wo("true", "ns/app", "version", "track", "true"))
			testDelete(t, s)
		})
	}
}

func testDelete(t *testing.T, s scenario) {
	deleteObject(wo(s.iter8, s.application, s.version, s.track, s.ready))

	abnapp.NumApplications(t, 1)
	a, _ := abnapp.Applications.Get("ns/app", true)
	assert.NotNil(t, a)

	if s.iter8 != "true" ||
		s.application == "" ||
		s.version == "" {
		assertApplication(t, a, applicationAssertion{
			namespace: "ns",
			name:      "app",
			tracks:    []string{"track"},
			versions:  []string{"version"},
		})
	} else {
		assertApplication(t, a, applicationAssertion{
			namespace: "ns",
			name:      "app",
			tracks:    []string{},
			versions:  []string{s.version},
		})
	}
}

func setup() {
	abnapp.Applications.Clear()
	abnapp.Applications.SetReaderWriter(&abnapp.ApplicationReaderWriter{Client: k8sdriver.NewFakeKubeDriver(cli.New()).Clientset})
}

func wo(iter8, name, version, track, ready string) WatchedObject {
	labels := map[string]string{}
	if name != "" {
		labels[NAME_LABEL] = abnapp.GetNameFromKey(name)
	}
	if version != "" {
		labels[VERSION_LABEL] = version
	}
	annotations := map[string]string{}
	if track != "" {
		annotations[TRACK_ANNOTATION] = track
	}
	if ready != "" {
		annotations[READY_ANNOTATION] = ready
	}
	if iter8 != "" {
		annotations[ITER8_ANNOTATION] = iter8
	}

	o := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        abnapp.GetNameFromKey(name),
			Namespace:   abnapp.GetNamespaceFromKey(name),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{},
	}
	obj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&o)
	wo := WatchedObject{Obj: &unstructured.Unstructured{Object: obj}}

	return wo
}
