package readiness

import (
	"encoding/json"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestInvalidObjName(t *testing.T) {
	initDelay, _ := json.Marshal(5)
	numRetries, _ := json.Marshal(3)
	intervalSeconds, _ := json.Marshal(5)
	objRefs, _ := json.Marshal([]ObjRef{
		{
			Kind:      "deploy",
			Namespace: core.StringPointer("default"),
			Name:      "hello world",
			WaitFor:   core.StringPointer("condition=available"),
		},
	})
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"initialDelaySeconds": {Raw: initDelay},
			"numRetries":          {Raw: numRetries},
			"intervalSeconds":     {Raw: intervalSeconds},
			"objRefs":             {Raw: objRefs},
		},
	})
	assert.Error(t, err)
}

func TestMakeReadinessTask(t *testing.T) {
	initDelay, _ := json.Marshal(5)
	numRetries, _ := json.Marshal(3)
	intervalSeconds, _ := json.Marshal(5)
	objRefs, _ := json.Marshal([]ObjRef{
		{
			Kind:      "deploy",
			Namespace: core.StringPointer("default"),
			Name:      "hello",
			WaitFor:   core.StringPointer("condition=available"),
		},
		{
			Kind:      "deploy",
			Namespace: core.StringPointer("default"),
			Name:      "hello-candidate",
			WaitFor:   core.StringPointer("condition=available"),
		},
	})
	task, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer(TaskName),
		With: map[string]apiextensionsv1.JSON{
			"initialDelaySeconds": {Raw: initDelay},
			"numRetries":          {Raw: numRetries},
			"intervalSeconds":     {Raw: intervalSeconds},
			"objRefs":             {Raw: objRefs},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), *task.(*ReadinessTask).With.InitialDelaySeconds)
	assert.Equal(t, int32(3), *task.(*ReadinessTask).With.NumRetries)
	assert.Equal(t, int32(5), *task.(*ReadinessTask).With.IntervalSeconds)
	assert.Equal(t, 2, len(task.(*ReadinessTask).With.ObjRefs))
}
