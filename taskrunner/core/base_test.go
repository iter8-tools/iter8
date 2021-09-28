package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	log = GetLogger()
}

func TestWithoutExperiment(t *testing.T) {
	tags := GetDefaultTags(context.Background())
	assert.Empty(t, tags.M)
}

func TestWithExperiment(t *testing.T) {
	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), ContextKey("experiment"), exp)
	tags := GetDefaultTags(ctx)

	testStr := []string{
		"{{.this.apiVersion}}",
		"{{.this.metadata.name}}",
		"{{.this.spec.duration.intervalSeconds}}",
		"{{(index .this.spec.versionInfo.baseline.variables 0).value}}",
		"{{.this.status.versionRecommendedForPromotion}}",
	}
	expectedOut := []string{
		"iter8.tools/v2alpha2",
		"sklearn-iris-experiment-1",
		"15",
		"revision1",
		"default",
	}

	for i, in := range testStr {
		out, err := tags.Interpolate(&in)
		assert.NoError(t, err)
		assert.Equal(t, expectedOut[i], out)
	}
}

type testTask struct {
	TaskMeta
}

func (t *testTask) Run(ctx context.Context) error {
	return nil
}

// multiple tasks successfully execute
func TestActionRun(t *testing.T) {
	action := Action{}
	t1 := testTask{}
	t2 := testTask{}
	action = append(action, &t1)
	action = append(action, &t2)

	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment10.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), ContextKey("experiment"), exp)

	a := &action
	err = a.Run(ctx)
	assert.NoError(t, err)
}

type badTestTask struct {
	TaskMeta
}

func (t *badTestTask) Run(ctx context.Context) error {
	return errors.New("shouldn't have run")
}

func TestBadActionRun(t *testing.T) {
	action := Action{}
	t1 := badTestTask{
		TaskMeta: TaskMeta{
			Task: StringPointer("hello/world"),
			If:   StringPointer("WinnerFound()"),
		},
	}
	action = append(action, &t1)

	exp, err := (&Builder{}).FromFile(CompletePath("../", "testdata/experiment10.yaml")).Build()
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), ContextKey("experiment"), exp)

	a := &action
	err = a.Run(ctx)
	assert.NoError(t, err)

	t2 := badTestTask{}
	action = append(action, &t2)

	err = a.Run(ctx)
	assert.Error(t, err)

	t2.If = StringPointer("AnotherOneBytesTheDust()")
	err = a.Run(ctx)
	assert.Error(t, err)
}
