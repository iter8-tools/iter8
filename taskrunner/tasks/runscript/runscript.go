package runscript

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

const (
	ScratchEnv string = "SCRATCH_DIR=/scratch"
)

// Inputs for the run task may contain a secret reference
type Inputs struct {
	Secret          *string `json:"secret" yaml:"secret"`
	interpolatedRun string
}

// Task encapsulates a command that can be executed.
type Task struct {
	core.TaskMeta `json:",inline" yaml:",inline"`
	With          Inputs `json:"with" yaml:"with"`
}

// Make converts an run spec into a run.
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if !core.IsARun(t) {
		return nil, fmt.Errorf("invalid run spec")
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

// EnhancedExperiment supports enhanced interpolation behaviors
type EnhancedExperiment struct {
	*core.Experiment
	sec *corev1.Secret
}

// Secret returns a value (of type string) for a key.
// This uses the secret embedded in the enhanced experiment.
// If the key is absent, an error will be returned.
// If the value is not a string, an error will be returned.
func (ee *EnhancedExperiment) Secret(key string) (string, error) {
	if ee.sec == nil {
		return "", errors.New("no secret specified")
	}
	val, ok := ee.sec.Data[key]
	if !ok {
		return "", fmt.Errorf("specified key %s seems absent in secret", key)
	}
	return string(val), nil
}

// Interpolate the script.
func (t *Task) Interpolate(ctx context.Context) error {
	exp, err := core.GetExperimentFromContext(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	ee := EnhancedExperiment{Experiment: exp}

	log.Trace("got past enhanced experiment")

	if t.With.Secret != nil {
		secret, err := core.GetSecret(*t.With.Secret)
		if err != nil {
			return err
		}
		ee.sec = secret
	}

	log.Trace("got past secret")

	var templ *template.Template
	if templ, err = template.New("templated script").Parse(*t.TaskMeta.Run); err == nil {
		buf := bytes.Buffer{}
		if err = templ.Execute(&buf, &ee); err == nil {
			t.With.interpolatedRun = buf.String()
			log.Trace("interpolated with enhanced experiment")
			log.Trace(t.With.interpolatedRun)
			return nil
		}
		log.Error("template execution error: ", err)
		return errors.New("cannot interpolate string due to template execution error")
	}
	log.Error("template creation error: ", err)
	return errors.New("cannot interpolate string due to template creation error")
}

// get the command
func (t *Task) getCommand(ctx context.Context) (*exec.Cmd, error) {
	err := t.Interpolate(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	cmd := exec.Command("/bin/bash", "-c", t.With.interpolatedRun)
	// append the scratch environment variable
	cmd.Env = append(os.Environ(), ScratchEnv)
	return cmd, nil
}

// Run the command.
func (t *Task) Run(ctx context.Context) error {
	cmd, err := t.getCommand(ctx)
	if err != nil {
		return err
	}
	out, err := cmd.CombinedOutput()
	log.Trace("Running task: " + cmd.String())
	log.Trace("Got combined output: " + string(out))
	return err
}
