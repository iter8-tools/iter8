package base

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/Masterminds/sprig"
	log "github.com/iter8-tools/iter8/base/log"
)

const (
	// RunTaskName is the name of the run task which performs running of a shell script.
	RunTaskName = "run"
)

var (
	tempDirEnv string = fmt.Sprintf("TEMP_DIR=%v", os.TempDir())
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

// initializeDefaults sets default values for task inputs
func (t *runTask) initializeDefaults() {}

//validateInputs for this task
func (t *runTask) validateInputs() error {
	return nil
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
	// append the environment variable for temp dir
	cmd.Env = append(os.Environ(), tempDirEnv)
	return cmd, nil
}

// Run the command.
func (t *runTask) Run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	cmd, err := t.getCommand(exp)
	if err != nil {
		return err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("combined execution failed")
		log.Logger.WithStackTrace(string(out)).Error("combined output from command")
		return err
	}
	log.Logger.WithStackTrace(string(out)).Trace("combined output from command")
	return nil
}
