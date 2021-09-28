package bash

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeFakeTask(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeBashTask(t *testing.T) {
	script, _ := json.Marshal("echo hello")
	task, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"script": {Raw: script},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "echo hello", task.(*Task).With.Script)
}

func TestBashRun(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "bashexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)
	// action, err := GetAction(exp, actionSpec)
	action, err := Make(&actionSpec[0])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err = action.Run(ctx)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(buf.String(), "\necho \"v1\"\n"))
}
