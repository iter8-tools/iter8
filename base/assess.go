// Package base provides the core definitions and primitives for Iter8 experiment and experimeent tasks.
package base

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
)

// assessInputs contain the inputs to the assess-app-versions task to be executed.
type assessInputs struct {
	// SLOs is a list of service level objectives
	SLOs []SLO `json:"SLOs,omitempty" yaml:"SLOs,omitempty"`
}

// assessTask enables assessment of versions
type assessTask struct {
	taskMeta
	// With contains the inputs for the assessTask
	With assessInputs `json:"with" yaml:"with"`
}

const (
	// AssessTaskName is the name of the task this file implements
	AssessTaskName = "assess-app-versions"
)

// initializeDefaults sets default values for task inputs
func (t *assessTask) initializeDefaults() {}

//validateInputs for this task
func (t *assessTask) validateInputs() error {
	return nil
}

// Run executes the assess-app-versions task
func (t *assessTask) Run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	if exp.Result.Insights == nil {
		log.Logger.Error("uninitialized insights within experiment")
		return errors.New("uninitialized insights within experiment")
	}
	if len(t.With.SLOs) == 0 ||
		exp.Result.Insights.NumVersions == 0 {
		// do nothing for now
		// todo: fix when rewards are introduced

		log.Logger.Warn("nothing to do; returning")
		return nil
	}

	// set SLOs (if needed)
	err = exp.Result.Insights.setSLOs(t.With.SLOs)
	if err != nil {
		return err
	}

	// set initialize SLOsSatisfied (if needed)
	err = exp.initializeSLOsSatisfied()
	if err != nil {
		return err
	}

	// set SLOsSatisfied
	exp.Result.Insights.SLOsSatisfied = evaluateSLOs(exp, t.With.SLOs)

	return err
}

// evaluate SLOs
func evaluateSLOs(exp *Experiment, slos []SLO) [][]bool {
	slosSatisfied := make([][]bool, len(slos))
	for i := 0; i < len(slos); i++ {
		slosSatisfied[i] = make([]bool, exp.Result.Insights.NumVersions)
		for j := 0; j < exp.Result.Insights.NumVersions; j++ {
			slosSatisfied[i][j] = sloSatisfied(exp, slos, i, j)
		}
	}
	return slosSatisfied
}

// return true if SLO i satisfied by version j
func sloSatisfied(e *Experiment, slos []SLO, i int, j int) bool {
	val := e.Result.Insights.getMetricValue(j, slos[i].Metric)
	// check if metric is available
	if val == nil {
		log.Logger.Warnf("unable to find value for version %v and metric %s", j, slos[i].Metric)
		return false
	}
	// check lower limit
	if slos[i].LowerLimit != nil {
		if *val < *slos[i].LowerLimit {
			return false
		}
	}
	// check upper limit
	if slos[i].UpperLimit != nil {
		if *val > *slos[i].UpperLimit {
			return false
		}
	}
	return true
}
