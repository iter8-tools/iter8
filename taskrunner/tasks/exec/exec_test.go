package exec

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeTask(t *testing.T) {
	b, _ := json.Marshal("echo")
	a, _ := json.Marshal([]string{"hello", "people", "of", "earth"})
	task, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("common/exec"),
		With: map[string]apiextensionsv1.JSON{
			"cmd":  {Raw: b},
			"args": {Raw: a},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "earth", task.(*Task).With.Args[3])
	log.Trace(task.(*Task).With.Args)

	exp, _ := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment10.yaml")).Build()
	task.Run(context.WithValue(context.Background(), core.ContextKey("experiment"), exp))

	task, err = Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("common/run"),
		With: map[string]apiextensionsv1.JSON{
			"cmd": {Raw: b},
		},
	})
	assert.Nil(t, task)
	assert.Error(t, err)
}

func TestExecTaskNoInterpolation(t *testing.T) {
	b, _ := json.Marshal("echo")
	a, _ := json.Marshal([]string{"hello", "{{ omg }}", "world"})
	c, _ := json.Marshal(true)
	task, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("common/exec"),
		With: map[string]apiextensionsv1.JSON{
			"cmd":                  {Raw: b},
			"args":                 {Raw: a},
			"disableInterpolation": {Raw: c},
		},
	})

	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "world", task.(*Task).With.Args[2])
	log.Trace(task.(*Task).With.Args)

	exp, _ := (&core.Builder{}).FromFile(core.CompletePath("../../", "testdata/experiment10.yaml")).Build()
	task.Run(context.WithValue(context.Background(), core.ContextKey("experiment"), exp))
}
