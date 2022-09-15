package watcher

import (
	"context"
	"strings"
	"testing"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/abn/k8sclient"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAdd(t *testing.T) {
	scenarios := map[string]struct {
		iter8related string
		namespace    string
		application  string
		version      string
		track        string
	}{
		"iter8 not set":  {iter8related: "", namespace: "namespace", application: "name", version: "version", track: "track"},
		"iter8 not true": {iter8related: "false", namespace: "namespace", application: "name", version: "version", track: "track"},
		"no application": {iter8related: "true", namespace: "namespace", application: "", version: "version", track: "track"},
		"no version":     {iter8related: "true", namespace: "namespace", application: "name", version: "", track: "track"},
		"no track":       {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: ""},
		"all":            {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track"},
	}

	for label, s := range scenarios {
		t.Run(label, func(t *testing.T) {
			gvr := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}

			done := make(chan struct{})
			defer close(done)
			setup(t, gvr, s.namespace, done)

			createObject(t, gvr, s.iter8related, s.namespace, s.application, s.version, s.track)

			assert.Eventually(
				t,
				func() bool {
					r := true
					// if any preconditions are not met, then no application added
					if strings.ToLower(s.iter8related) != "true" ||
						s.application == "" ||
						s.version == "" {
						r = r && abnapp.NumApplications(t, 0)
						return r
					}

					// otherwise application created

					a, err := abnapp.Applications.Get(s.namespace + "/" + s.application)
					r = r && assert.NoError(t, err)
					r = r && assert.NotNil(t, a)

					tracks := []string{}
					if s.track != "" {
						tracks = []string{s.track}
					}
					r = r && assertApplication(t, a, applicationAssertion{
						namespace: s.namespace,
						name:      s.application,
						tracks:    tracks,
						versions:  []string{s.version},
					})
					return r
				},
				10*time.Second,
				100*time.Millisecond,
			)
		})
	}
}

func TestUpdate(t *testing.T) {
	scenarios := map[string]struct {
		iter8related string
		namespace    string
		application  string
		version      string
		track        string
	}{
		"no application": {iter8related: "true", namespace: "namespace", application: "", version: "version", track: "track"},
		"no version":     {iter8related: "true", namespace: "namespace", application: "name", version: "", track: "track"},
		"no track":       {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: ""},
		"all":            {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track"},
	}

	for label, s := range scenarios {
		t.Run(label, func(t *testing.T) {
			gvr := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}

			done := make(chan struct{})
			defer close(done)
			setup(t, gvr, s.namespace, done)
			existingObj := createObject(t, gvr, "true", "namespace", "name", "version", "")
			// verify that existing object is as expected
			assert.Eventually(
				t,
				func() bool {
					r := true
					a, err := abnapp.Applications.Get("namespace/name")
					r = r && assert.NoError(t, err)
					r = r && assert.NotNil(t, a)
					r = r && assertApplication(t, a, applicationAssertion{
						namespace: "namespace",
						name:      "name",
						tracks:    []string{},
						versions:  []string{"version"},
					})
					return r
				},
				10*time.Second,
				100*time.Millisecond,
			)

			//update object labels
			updateLabel(existingObj, iter8Label, s.iter8related)
			updateLabel(existingObj, nameLabel, s.application)
			updateLabel(existingObj, versionLabel, s.version)
			updateLabel(existingObj, trackLabel, s.track)

			updatedObj, err := k8sclient.Client.Dynamic().
				Resource(gvr).Namespace(s.namespace).
				Update(
					context.TODO(),
					existingObj,
					metav1.UpdateOptions{},
				)
			assert.NoError(t, err)
			assert.NotNil(t, updatedObj)

			assert.Eventually(
				t,
				func() bool {
					r := true
					// if any preconditions are not met, then no change was made to applications
					if strings.ToLower(s.iter8related) != "true" ||
						s.application == "" ||
						s.version == "" {
						// abnapp.NumApplications(t, 1)
						a, err := abnapp.Applications.Get("namespace/name")
						r = r && assert.NoError(t, err)
						r = r && assert.NotNil(t, a)
						r = r && assertApplication(t, a, applicationAssertion{
							namespace: "namespace",
							name:      "name",
							tracks:    []string{},
							versions:  []string{"version"},
						})
						return r
					}

					// otherwise application was possibily modified in some way
					// at least this application exists since it was preexisting
					a, err := abnapp.Applications.Get(s.namespace + "/" + s.application)
					r = r && assert.NoError(t, err)
					r = r && assert.NotNil(t, a)

					tracks := []string{}
					if s.track != "" {
						tracks = []string{s.track}
					}
					r = r && assertApplication(t, a, applicationAssertion{
						namespace: s.namespace,
						name:      s.application,
						tracks:    tracks,
						versions:  []string{s.version},
					})
					return r
				},
				10*time.Second,
				100*time.Millisecond,
			)
		})
	}
}

func updateLabel(obj *unstructured.Unstructured, key string, value string) {
	if value == "" {
		delete((obj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{}), key)
	} else {
		(obj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[key] = value
	}

}

func TestDelete(t *testing.T) {
	scenarios := map[string]struct {
		iter8related string
		namespace    string
		application  string
		version      string
		track        string
	}{
		"no track, ready": {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: ""},
		"track, ready":    {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track"},
	}

	for label, s := range scenarios {
		t.Run(label, func(t *testing.T) {
			gvr := schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}

			done := make(chan struct{})
			defer close(done)
			setup(t, gvr, s.namespace, done)

			// add existing object (to delete)
			createObject(t, gvr, "true", "namespace", "name", "version", "track")
			// verify that existing object is as expected
			assert.Eventually(
				t,
				func() bool {
					r := true
					a, err := abnapp.Applications.Get("namespace/name")
					r = r && assert.NoError(t, err)
					r = r && assert.NotNil(t, a)
					r = r && assertApplication(t, a, applicationAssertion{
						namespace: "namespace",
						name:      "name",
						tracks:    []string{"track"},
						versions:  []string{"version"},
					})
					return r
				},
				10*time.Second,
				100*time.Millisecond,
			)

			// delete object --> no track anymore
			err := k8sclient.Client.Dynamic().
				Resource(gvr).Namespace(s.namespace).
				Delete(
					context.TODO(),
					s.application,
					metav1.DeleteOptions{},
				)
			assert.NoError(t, err)

			assert.Eventually(
				t,
				func() bool {
					r := true
					// if any preconditions are not met, then no change was made to applications
					if strings.ToLower(s.iter8related) != "true" ||
						s.application == "" ||
						s.version == "" {
						a, err := abnapp.Applications.Get("namespace/name")
						r = r && assert.NoError(t, err)
						r = r && assert.NotNil(t, a)
						r = r && assertApplication(t, a, applicationAssertion{
							namespace: "namespace",
							name:      "name",
							tracks:    []string{"track"},
							versions:  []string{"version"},
						})
						return r
					}

					// otherwise application was possibily modified in some way
					// at least this application exists since it was preexisting
					a, err := abnapp.Applications.Get(s.namespace + "/" + s.application)
					r = r && assert.NoError(t, err)
					r = r && assert.NotNil(t, a)

					r = r && assertApplication(t, a, applicationAssertion{
						namespace: s.namespace,
						name:      s.application,
						tracks:    []string{},
						versions:  []string{s.version},
					})

					return r
				},
				10*time.Second,
				100*time.Millisecond,
			)
		})
	}
}

func setup(t *testing.T, gvr schema.GroupVersionResource, namespace string, done chan struct{}) {
	// ensure global record of applications is empty
	abnapp.Applications.Clear()

	// define watcher
	k8sclient.Client = *k8sclient.NewFakeKubeClient(cli.New())
	w := NewIter8Watcher(
		[]schema.GroupVersionResource{gvr},
		[]string{namespace},
	)
	assert.NotNil(t, w)

	// start informers
	w.Start(done)
}

func createObject(t *testing.T, gvr schema.GroupVersionResource, iter8related, namespace, application, version, track string) *unstructured.Unstructured {
	// create object; no track defined
	createdObj, err := k8sclient.Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(iter8related, namespace, application, version, track),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)
	return createdObj
}

func newUnstructuredDeployment(iter8related, namespace, application, version, track string) *unstructured.Unstructured {
	labels := map[string]interface{}{}
	if application != "" {
		labels[nameLabel] = application
	}
	if version != "" {
		labels[versionLabel] = version
	}
	if iter8related != "" {
		labels[iter8Label] = iter8related
	}

	if track != "" {
		labels[trackLabel] = track
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      application,
				"labels":    labels,
			},
			"spec": application,
		},
	}
}
