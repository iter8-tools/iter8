package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
)

const (
	// TaskName is the name of the HTTP request task
	TaskName string = "notification/http"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

// Inputs contain the name and arguments of the task.
type Inputs struct {
	URL           string                `json:"URL" yaml:"URL"`
	Method        *v2alpha2.MethodType  `json:"method,omitempty" yaml:"method,omitempty"`
	AuthType      *v2alpha2.AuthType    `json:"authType,omitempty" yaml:"authType,omitempty"`
	Secret        *string               `json:"secret,omitempty" yaml:"secret,omitempty"`
	Headers       []v2alpha2.NamedValue `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body          *string               `json:"body,omitempty" yaml:"body,omitempty"`
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

func (t *Task) prepareRequest(ctx context.Context) (*http.Request, error) {
	tags := core.NewTags()
	exp, err := core.GetExperimentFromContext(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	obj, err := exp.ToMap()
	if err != nil {
		// error already logged by ToMap()
		// don't log it again
		return nil, err
	}

	// prepare for interpolation; add experiment as tag
	// Note that if versionRecommendedForPromotion is not set or there is no version corresponding to it,
	// then some placeholders may not be replaced
	tags = tags.
		With("this", obj).
		WithRecommendedVersionForPromotion(&exp.Experiment, t.With.VersionInfo)

	// log tags now before secret is added; we don't log the secret
	log.Trace("tags without secrets: ", tags)

	secretName := t.With.Secret
	if secretName != nil {
		secret, err := core.GetSecret(*secretName)
		if err == nil {
			tags = tags.WithSecret("secret", secret)
		}
	}
	log.Trace("tags with secrets: ", tags)

	defaultMethod := v2alpha2.POSTMethodType
	method := t.With.Method
	if method == nil {
		method = &defaultMethod
	}
	log.Trace("method: ", *method)

	body := t.With.Body
	if body != nil {
		if interpolated, err := tags.Interpolate(body); err == nil {
			body = &interpolated
		}
	} else {
		// body should be defaulted
		b, err := defaultBody(exp.Experiment)
		if err != nil {
			return nil, err
		}
		body = &b
	}
	log.Trace("body:", *body)

	defaultAuthType := v2alpha2.AuthType("None")
	authType := t.With.AuthType
	if authType == nil {
		authType = &defaultAuthType
	}
	log.Trace("authType: ", *authType)

	req, err := http.NewRequest(string(*method), t.With.URL, strings.NewReader(*body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")
	for _, h := range t.With.Headers {
		hValue, err := tags.Interpolate(&h.Value)
		if err != nil {
			log.Warn("Unable to interpolate header "+h.Name, err)
		} else {
			req.Header.Set(h.Name, hValue)
		}
	}

	if *authType == v2alpha2.BasicAuthType {
		usernameTemplate := "{{ .secret.username }}"
		passwordTemplate := "{{ .secret.password }}"
		username, _ := tags.Interpolate(&usernameTemplate)
		password, _ := tags.Interpolate(&passwordTemplate)
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	} else if *authType == v2alpha2.BearerAuthType {
		tokenTemplate := "{{ .secret.token }}"
		token, _ := tags.Interpolate(&tokenTemplate)
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, err
}

// helper types for creating a default body
type defaultbody struct {
	Summary    experimentsummary   `json:"summary" yaml:"summary"`
	Experiment v2alpha2.Experiment `json:"experiment" yaml:"experiment"`
}

type experimentsummary struct {
	WinnerFound                    bool                  `json:"winnerFound" yaml:"winnerFound"`
	Winner                         *string               `json:"winner,omitempty" yaml:"winner,omitempty"`
	VersionRecommendedForPromotion *string               `json:"versionRecommendedForPromotion,omitempty" yaml:"versionRecommendedForPromotion,omitempty"`
	LastRecommendedWeights         []v2alpha2.WeightData `json:"lastRecommendedWeights,omitempty" yaml:"lastRecommendedWeights,omitempty"`
}

func defaultBody(experiment v2alpha2.Experiment) (string, error) {
	defaultBody := defaultbody{
		Summary: experimentsummary{
			WinnerFound: false,
		},
		Experiment: experiment,
	}

	// WinnerFound, Winner
	if experiment.Status.Analysis != nil &&
		experiment.Status.Analysis.WinnerAssessment != nil {
		defaultBody.Summary.WinnerFound = experiment.Status.Analysis.WinnerAssessment.Data.WinnerFound
		if experiment.Status.Analysis.WinnerAssessment.Data.Winner != nil {
			defaultBody.Summary.Winner = experiment.Status.Analysis.WinnerAssessment.Data.Winner
		}
	}

	// VersionRecommendedForPromotion
	if experiment.Status.VersionRecommendedForPromotion != nil {
		defaultBody.Summary.VersionRecommendedForPromotion = experiment.Status.VersionRecommendedForPromotion
	}

	// LastRecommendedWeights
	if experiment.Status.Analysis != nil && experiment.Status.Analysis.Weights != nil {
		defaultBody.Summary.LastRecommendedWeights = make([]v2alpha2.WeightData, len(experiment.Status.Analysis.Weights.Data))
		for i, w := range experiment.Status.Analysis.Weights.Data {
			defaultBody.Summary.LastRecommendedWeights[i] = v2alpha2.WeightData{Name: w.Name, Value: w.Value}
		}
	}

	b, err := json.Marshal(defaultBody)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Run the command.
func (t *Task) internalRun(ctx context.Context) error {
	req, err := t.prepareRequest(ctx)

	if err != nil {
		return err
	}

	// send request
	var httpClient = &http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("RESPONSE STATUS: " + resp.Status)
	if resp.StatusCode >= 400 {

		err = errors.New(resp.Status)
		log.Error(err)
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	log.Info(buf.String())

	return nil
}

// Run the task. Ignores failures unless the task indicates ignoreFailures: false
func (t *Task) Run(ctx context.Context) error {
	err := t.internalRun(ctx)
	if t.With.IgnoreFailure != nil && !*t.With.IgnoreFailure {
		return err
	}
	return nil
}
