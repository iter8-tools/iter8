package base

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/iter8-tools/iter8/base/log"
)

const (
	// RunTaskName is the name of the run task which performs running of a shell script
	RunTaskName = "run"
)

var (
	// tempDirEnv is a temporary directory
	tempDirEnv string = fmt.Sprintf("TEMP_DIR=%v", os.TempDir())
)

// runTask enables running a shell script
type runTask struct {
	// TaskMeta has fields common to all tasks
	TaskMeta
}

// initializeDefaults sets default values for task inputs
func (t *runTask) initializeDefaults() {}

// validateInputs for this task
func (t *runTask) validateInputs() error {
	return nil
}

// getCommand gets the executable command
func (t *runTask) getCommand() *exec.Cmd {
	cmdStr := *t.TaskMeta.Run
	// create command to be executed
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	// append the environment variable for temp dir
	cmd.Env = append(os.Environ(), tempDirEnv)
	return cmd
}

// run the command
func (t *runTask) run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	cmd := t.getCommand()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("combined execution failed")
		log.Logger.WithStackTrace(string(out)).Error("combined output from command")
		return err
	}
	log.Logger.WithStackTrace(string(out)).Trace("combined output from command")
	return nil
}
