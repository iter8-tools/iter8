package core

// import (
// 	"encoding/json"
// 	"testing"

// 	"github.com/iter8-tools/etc3/api/v2alpha2"
// 	"github.com/iter8-tools/etc3/taskrunner/core"
// 	"github.com/stretchr/testify/assert"
// 	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
// )

// func TestMakeFakeMetricsTask(t *testing.T) {
// 	_, err := Make(&v2alpha2.TaskSpec{
// 		Task: core.StringPointer("fake/fake"),
// 	})
// 	assert.Error(t, err)
// }

// func TestMakeMetricsTask(t *testing.T) {
// 	vers, _ := json.Marshal([]Version{
// 		{
// 			Name: "test",
// 			URL:  "https://iter8.tools",
// 		},
// 	})
// 	task, err := Make(&v2alpha2.TaskSpec{
// 		Task: core.StringPointer("metrics/collect"),
// 		With: map[string]v1.JSON{
// 			"versions": {Raw: vers},
// 		},
// 	})
// 	assert.NotEmpty(t, task)
// 	assert.NoError(t, err)

// 	task, err = Make(&v2alpha2.TaskSpec{
// 		Task: core.StringPointer("metrics/collect"),
// 		With: map[string]v1.JSON{
// 			"versionables": {Raw: vers},
// 		},
// 	})
// 	assert.Empty(t, task)
// 	assert.Error(t, err)

// 	task, err = Make(&v2alpha2.TaskSpec{
// 		Task: core.StringPointer("metrics/collect-it"),
// 	})
// 	assert.Nil(t, task)
// 	assert.Error(t, err)
// }
