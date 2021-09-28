package bash

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

const (
	// TaskName is the name of the bash task
	TaskName string = "common/bash"
)

// Inputs contain the name and arguments of the command to be executed.
type Inputs struct {
	VersionInfo []core.VersionInfo `json:"versionInfo,omitempty" yaml:"versionInfo,omitempty"`
	Script      string             `json:"script" yaml:"script"`
}

// Task encapsulates a command that can be executed.
type Task struct {
	core.TaskMeta `json:",inline" yaml:",inline"`
	With          Inputs `json:"with" yaml:"with"`
}

// Make converts an bash task spec into an bash task.
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if *t.Task != TaskName {
		return nil, fmt.Errorf("task need to be '%s'", TaskName)
	}
	var jsonBytes []byte
	var task Task
	// convert t to jsonBytes
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	// convert jsonString to ExecTask
	task = Task{}
	err = json.Unmarshal(jsonBytes, &task)
	return &task, err
}

// Run the command.
func (t *Task) Run(ctx context.Context) error {
	exp, err := core.GetExperimentFromContext(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Trace("experiment", exp)

	obj, err := exp.ToMap()
	if err != nil {
		// error already logged by ToMap()
		// don't log it again
		return err
	}

	// prepare for interpolation; add experiment as tag
	// Note that if versionRecommendedForPromotion is not set or there is no version corresponding to it,
	// then some placeholders may not be replaced
	tags := core.NewTags().
		With("this", obj).
		WithRecommendedVersionForPromotion(&exp.Experiment, t.With.VersionInfo)

	// interpolate - replaces placeholders in the script with values
	script, _ := tags.Interpolate(&t.With.Script)

	log.Trace(script)
	args := []string{"-c", script}

	cmd := exec.Command("/bin/bash", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running task: " + cmd.String())
	log.Trace(args)
	err = cmd.Run()

	return err
}
