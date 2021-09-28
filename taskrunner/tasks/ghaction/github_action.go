package ghaction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/iter8-tools/etc3/taskrunner/tasks/http"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

const (
	// TaskName is the name of the GitHub action request task
	TaskName string = "notification/github-workflow"
	// DefaultRef is default github reference (branch)
	DefaultRef string = "master"
)

// Inputs contain the name and arguments of the task.
type Inputs struct {
	Repository    string                `json:"repository" yaml:"repository"`
	Workflow      string                `json:"workflow" yaml:"workflow"`
	Secret        string                `json:"secret" yaml:"secret"`
	Ref           *string               `json:"ref,omitempty" yaml:"ref,omitempty"`
	WFInputs      []v2alpha2.NamedValue `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	VersionInfo   []core.VersionInfo    `json:"versionInfo,omitempty" yaml:"versionInfo,omitempty"`
	IgnoreFailure *bool                 `json:"ignoreFailure,omitempty" yaml:"ignoreFailure,omitempty"`
}

// Task encapsulates the task.
type Task struct {
	core.TaskMeta `json:",inline" yaml:",inline"`
	With          Inputs `json:"with" yaml:"with"`
}

// Make converts an spec to a task.
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if *t.Task != TaskName {
		return nil, fmt.Errorf("library and task need to be '%s'", TaskName)
	}
	var jsonBytes []byte
	var task Task
	// convert t to jsonBytes
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	// convert jsonString to task
	task = Task{}
	err = json.Unmarshal(jsonBytes, &task)
	return &task, err
}

// ToHTTPTask converts a Task to an http Task
func (t *Task) ToHTTPTask() *http.Task {
	authType := v2alpha2.BearerAuthType
	authtype := &authType
	secret := &t.With.Secret

	ref := DefaultRef
	if t.With.Ref != nil {
		ref = *t.With.Ref
	}

	// compose body of POST request
	body := ""
	body += "{"
	body += "\"ref\": \"" + ref + "\","
	body += "\"inputs\": {"
	numWFInputs := len(t.With.WFInputs)
	for i := 0; i < numWFInputs; i++ {
		body += "\"" + t.With.WFInputs[i].Name + "\": \"" + t.With.WFInputs[i].Value + "\""
		if i+1 < numWFInputs {
			body += ","
		}
	}
	body += "}"
	body += "}"

	tSpec := &http.Task{
		TaskMeta: core.TaskMeta{
			Task: core.StringPointer(TaskName),
		},
		With: http.Inputs{
			URL:      "https://api.github.com/repos/" + t.With.Repository + "/actions/workflows/" + t.With.Workflow + "/dispatches",
			AuthType: authtype,
			Secret:   secret,
			Headers: []v2alpha2.NamedValue{{
				Name:  "Accept",
				Value: "application/vnd.github.v3+json",
			}},
			Body:          &body,
			IgnoreFailure: t.With.IgnoreFailure,
		},
	}

	if t.With.IgnoreFailure != nil {
		tSpec.With.IgnoreFailure = t.With.IgnoreFailure
	}

	log.Info("Dispatching GitHub workflow: ", tSpec.With.URL)
	log.Info(*tSpec.With.Body)

	return tSpec
}

// Run the task. Ignores failures unless the task indicates ignoreFailures: false
func (t *Task) Run(ctx context.Context) error {
	return t.ToHTTPTask().Run(ctx)
}
