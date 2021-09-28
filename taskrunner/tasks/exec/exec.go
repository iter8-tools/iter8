package exec

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
	// TaskName is the name of the task this file implements
	TaskName string = "common/exec"
)

// Inputs contain the name and arguments of the command to be executed.
type Inputs struct {
	Cmd                  string        `json:"cmd" yaml:"cmd"`
	Args                 []interface{} `json:"args,omitempty" yaml:"args,omitempty"`
	DisableInterpolation bool          `json:"disableInterpolation,omitempty" yaml:"disableInterpolation,omitempty"`
}

// Task encapsulates a command that can be executed.
type Task struct {
	core.TaskMeta `json:",inline" yaml:",inline"`
	With          Inputs `json:"with" yaml:"with"`
}

// Run the command.
func (t *Task) Run(ctx context.Context) error {
	exp, err := core.GetExperimentFromContext(ctx)
	if err == nil {
		inputArgs := make([]string, len(t.With.Args))
		for i := 0; i < len(inputArgs); i++ {
			inputArgs[i] = fmt.Sprint(t.With.Args[i])
		}
		log.Trace(inputArgs)
		var args []string
		if t.With.DisableInterpolation {
			args = inputArgs
		} else {
			args, err = exp.Interpolate(inputArgs)
		}
		if err == nil {
			log.Trace("interpolated args: ", args)
			cmd := exec.Command(t.With.Cmd, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Info("Running task: " + cmd.String())
			log.Trace(args)
			err = cmd.Run()
		}
	}
	if err != nil {
		log.Error(err)
	}
	return err
}

// Make converts an exec task spec into an exec task.
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if *t.Task != TaskName {
		return nil, fmt.Errorf("library and task need to be '%s'", TaskName)
	}
	var err error
	var jsonBytes []byte
	var et Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to ExecTask
	if err == nil {
		et = Task{}
		err = json.Unmarshal(jsonBytes, &et)
	}
	return &et, err
}
