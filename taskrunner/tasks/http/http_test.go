package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeFakeNotificationTask(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeFakeHTTPTask(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeHttpTask(t *testing.T) {
	url, _ := json.Marshal("http://postman-echo.com/post")
	body, _ := json.Marshal("{\"hello\":\"world\"}")
	headers, _ := json.Marshal([]v2alpha2.NamedValue{{
		Name:  "x-foo",
		Value: "bar",
	}, {
		Name:  "Authentication",
		Value: "Basic: dXNlcm5hbWU6cGFzc3dvcmQK",
	}})
	task, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"URL":     {Raw: url},
			"body":    {Raw: body},
			"headers": {Raw: headers},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, task)
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	req, err := task.(*Task).prepareRequest(ctx)
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	assert.Equal(t, "http://postman-echo.com/post", req.URL.String())
	assert.Equal(t, "bar", req.Header.Get("x-foo"))

	err = task.Run(ctx)
	assert.NoError(t, err)
}

func TestMakeHttpTaskDefaults(t *testing.T) {
	url, _ := json.Marshal("http://target")
	task, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"URL": {Raw: url},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, task)

	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	req, err := task.(*Task).prepareRequest(ctx)
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, 1, len(req.Header))
	assert.Equal(t, "application/json", req.Header.Get("Content-type"))

	data, err := ioutil.ReadAll(req.Body)
	assert.NoError(t, err)

	expectedBody := `{"summary":{"winnerFound":false,"versionRecommendedForPromotion":"default"},"experiment":{"kind":"Experiment","apiVersion":"iter8.tools/v2alpha2","metadata":{"name":"sklearn-iris-experiment-1","namespace":"default","selfLink":"/apis/iter8.tools/v2alpha2/namespaces/default/experiments/sklearn-iris-experiment-1","uid":"b99489b6-a1b4-420f-9615-165d6ff88293","generation":2,"creationTimestamp":"2020-12-27T21:55:48Z","annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"iter8.tools/v2alpha2\",\"kind\":\"Experiment\",\"metadata\":{\"annotations\":{},\"name\":\"sklearn-iris-experiment-1\",\"namespace\":\"default\"},\"spec\":{\"criteria\":{\"indicators\":[\"95th-percentile-tail-latency\"],\"objectives\":[{\"metric\":\"mean-latency\",\"upperLimit\":1000},{\"metric\":\"error-rate\",\"upperLimit\":\"0.01\"}]},\"duration\":{\"intervalSeconds\":15,\"iterationsPerLoop\":10},\"strategy\":{\"type\":\"Canary\"},\"target\":\"default/sklearn-iris\"}}\n"}},"spec":{"target":"default/sklearn-iris","versionInfo":{"baseline":{"name":"default","variables":[{"name":"revision","value":"revision1"}]},"candidates":[{"name":"canary","variables":[{"name":"revision","value":"revision2"}],"weightObjRef":{"kind":"InferenceService","namespace":"default","name":"sklearn-iris","apiVersion":"serving.kubeflow.org/v1alpha2","fieldPath":".spec.canaryTrafficPercent"}}]},"strategy":{"testingPattern":"Canary","deploymentPattern":"Progressive","actions":{"finish":[{"task":"common/exec","with":{"args":["build","."],"cmd":"kustomize"}}],"start":[{"task":"common/exec","with":{"args":["hello-world","hello {{ revision }} world","hello {{ omg }} world"],"cmd":"echo"}},{"task":"common/exec","with":{"args":["v1","v2",20,40.5],"cmd":"helm"}}]},"weights":{"maxCandidateWeight":100,"maxCandidateWeightIncrement":10}},"criteria":{"requestCount":"request-count","indicators":["95th-percentile-tail-latency"],"objectives":[{"metric":"mean-latency","upperLimit":"1k"},{"metric":"error-rate","upperLimit":"10m"}],"strength":null},"duration":{"intervalSeconds":15,"iterationsPerLoop":10}},"status":{"conditions":[{"type":"Completed","status":"False","lastTransitionTime":"2020-12-27T21:55:49Z","reason":"StartHandlerLaunched","message":"Start handler 'start' launched"},{"type":"Failed","status":"False","lastTransitionTime":"2020-12-27T21:55:48Z"}],"initTime":"2020-12-27T21:55:48Z","lastUpdateTime":"2020-12-27T21:55:48Z","completedIterations":0,"versionRecommendedForPromotion":"default","message":"StartHandlerLaunched: Start handler 'start' launched"}}}`
	assert.Equal(t, expectedBody, string(data))
}
