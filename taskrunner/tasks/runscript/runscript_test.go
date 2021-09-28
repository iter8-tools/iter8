package runscript

import (
	"context"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
)

func TestMakeFakeRun(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeRun(t *testing.T) {
	task, err := Make(&v2alpha2.TaskSpec{
		Run: core.StringPointer("echo hello"),
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
}

func TestRunOne(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)

	task, err := Make(&actionSpec[0])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	err = task.Run(ctx)
	assert.NoError(t, err)
}

func TestRunTwo(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)

	task, err := Make(&actionSpec[1])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	err = task.Run(ctx)
	assert.NoError(t, err)

}

func TestRunThree(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)

	task, err := Make(&actionSpec[2])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	err = task.Run(ctx)
	assert.NoError(t, err)

}

func TestRunFour(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)

	task, err := Make(&actionSpec[3])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	err = task.Run(ctx)
	assert.Error(t, err)
}

func TestRunFive(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)

	task, err := Make(&actionSpec[4])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	err = task.Run(ctx)
	log.Error(err)
	assert.Error(t, err)
}

func TestRunEnv(t *testing.T) {

	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)

	task, err := Make(&v2alpha2.TaskSpec{
		Run: core.StringPointer("echo $SCRATCH_DIR"),
	})
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	cmd, err := task.(*Task).getCommand(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "/bin/bash -c echo $SCRATCH_DIR", cmd.String())

	out, err := cmd.CombinedOutput()
	assert.Equal(t, "/scratch\n", string(out))
	assert.NoError(t, err)
}
