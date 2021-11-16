package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"text/template"

	"github.com/Masterminds/sprig"
	log "github.com/iter8-tools/iter8/base/log"
)

const (
	RunTaskName        = "run"
	scratchEnv  string = "SCRATCH_DIR=/scratch"
)

// runInputs contains inputs for the run task
type runInputs struct {
	Template bool `json:"template" yaml:"template"`
}

// runTask enables running a shell script
type runTask struct {
	taskMeta
	With runInputs `json:"with" yaml:"with"`
}

// MakeRun constructs a RunTask out of a run task spec
func MakeRun(t *TaskSpec) (Task, error) {
	if t.Run == nil {
		return nil, errors.New("task need to have a run command")
	}
	var err error
	var jsonBytes []byte
	var bt Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to RunTask
	if err == nil {
		rt := &runTask{}
		err = json.Unmarshal(jsonBytes, &rt)
		bt = rt
	}
	return bt, err
}

// interpolate the script.
func (t *runTask) interpolate(exp *Experiment) (string, error) {
	// ensure it is a valid template
	tmpl, err := template.New("tpl").Funcs(sprig.TxtFuncMap()).Option("missingkey=error").Parse(*t.taskMeta.Run)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse templated run command")
		return "", err
	}

	// execute template
	var b bytes.Buffer
	err = tmpl.Execute(&b, exp)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to execute command template")
		return "", err
	}

	// print output
	return b.String(), nil

}

// get the command
func (t *runTask) getCommand(exp *Experiment) (*exec.Cmd, error) {
	var cmdStr string
	var err error
	if t.With.Template {
		cmdStr, err = t.interpolate(exp)
	} else {
		cmdStr = *t.taskMeta.Run
	}
	if err != nil {
		return nil, err
	}

	// create command to be executed
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	// append the scratch environment variable
	cmd.Env = append(os.Environ(), scratchEnv)
	return cmd, nil
}

// GetName returns the name of the assess task
func (t *runTask) GetName() string {
	return RunTaskName
}

// Run the command.
func (t *runTask) Run(exp *Experiment) error {
	cmd, err := t.getCommand(exp)
	if err != nil {
		return err
	}
	log.Logger.WithStackTrace(cmd.String()).Trace("running command")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("combined execution failed")
		log.Logger.WithStackTrace(string(out)).Error("combined output from command")
		return err
	}
	log.Logger.WithStackTrace(string(out)).Trace("combined output from command")
	return nil
}
