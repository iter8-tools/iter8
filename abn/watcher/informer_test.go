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
		ready        string
	}{
		"iter8 not set":        {iter8related: "", namespace: "namespace", application: "name", version: "version", track: "track", ready: "true"},
		"iter8 not true":       {iter8related: "false", namespace: "namespace", application: "name", version: "version", track: "track", ready: "true"},
		"no application":       {iter8related: "true", namespace: "namespace", application: "", version: "version", track: "track", ready: "true"},
		"no version":           {iter8related: "true", namespace: "namespace", application: "name", version: "", track: "track", ready: "true"},
		"w/o track, ready":     {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "", ready: "ready"},
		"w/o track, not ready": {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "", ready: "false"},
		"w/ track, ready":      {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track", ready: "true"},
		"w/ track, not ready":  {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track", ready: ""},
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

			createObject(t, gvr, s.iter8related, s.namespace, s.application, s.version, s.track, s.ready)

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
					if s.track != "" && s.ready == "true" {
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
		ready        string
	}{
		"no application":       {iter8related: "true", namespace: "namespace", application: "", version: "version", track: "track", ready: "true"},
		"no version":           {iter8related: "true", namespace: "namespace", application: "name", version: "", track: "track", ready: "true"},
		"w/o track, ready":     {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "", ready: "ready"},
		"w/o track, not ready": {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "", ready: "false"},
		"w/ track, ready":      {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track", ready: "true"},
		"w/ track, not ready":  {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track", ready: ""},
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
			existingObj := createObject(t, gvr, "true", "namespace", "name", "version", "", "")
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

			//update object
			if s.iter8related == "" {
				delete((existingObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{}), ITER8_LABEL)
			} else {
				(existingObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[ITER8_LABEL] = s.iter8related
			}
			if s.iter8related == "" {
				delete((existingObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{}), NAME_LABEL)
			} else {
				(existingObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[NAME_LABEL] = s.application
			}
			if s.iter8related == "" {
				delete((existingObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{}), VERSION_LABEL)
			} else {
				(existingObj.Object["metadata"].(map[string]interface{}))["labels"].(map[string]interface{})[VERSION_LABEL] = s.version
			}

			if s.track == "" {
				delete((existingObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{}), TRACK_ANNOTATION)
			} else {
				(existingObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{})[TRACK_ANNOTATION] = s.track
			}
			if s.ready == "" {
				delete((existingObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{}), READY_ANNOTATION)
			} else {
				(existingObj.Object["metadata"].(map[string]interface{}))["annotations"].(map[string]interface{})[READY_ANNOTATION] = s.ready
			}

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
					if s.track != "" && s.ready == "true" {
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

func TestDelete(t *testing.T) {
	scenarios := map[string]struct {
		iter8related string
		namespace    string
		application  string
		version      string
		track        string
		ready        string
	}{
		"w/o track, ready":     {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "", ready: "ready"},
		"w/o track, not ready": {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "", ready: "false"},
		"w/ track, ready":      {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track", ready: "true"},
		"w/ track, not ready":  {iter8related: "true", namespace: "namespace", application: "name", version: "version", track: "track", ready: ""},
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
			createObject(t, gvr, "true", "namespace", "name", "version", "track", "true")
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

func createObject(t *testing.T, gvr schema.GroupVersionResource, iter8related, namespace, application, version, track string, ready string) *unstructured.Unstructured {
	// create object; no track defined
	createdObj, err := k8sclient.Client.Dynamic().
		Resource(gvr).Namespace(namespace).
		Create(
			context.TODO(),
			newUnstructuredDeployment(iter8related, namespace, application, version, track, ready),
			metav1.CreateOptions{},
		)
	assert.NoError(t, err)
	assert.NotNil(t, createdObj)
	return createdObj
}

func newUnstructuredDeployment(iter8related, namespace, application, version, track string, ready string) *unstructured.Unstructured {
	labels := map[string]interface{}{}
	if application != "" {
		labels[NAME_LABEL] = application
	}
	if version != "" {
		labels[VERSION_LABEL] = version
	}
	if iter8related != "" {
		labels[ITER8_LABEL] = iter8related
	}

	annotations := map[string]interface{}{}
	if ready != "" {
		annotations[READY_ANNOTATION] = ready
	}
	if track != "" {
		annotations[TRACK_ANNOTATION] = track
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace":   namespace,
				"name":        application,
				"labels":      labels,
				"annotations": annotations,
			},
			"spec": application,
		},
	}
}
