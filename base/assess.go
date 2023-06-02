package base

import (
	"errors"

	"github.com/iter8-tools/iter8/base/log"
)

// assessInputs contain the inputs to the assess-app-versions task to be executed.
type assessInputs struct {
	// Rewards are the reward metrics
	Rewards *Rewards `json:"rewards,omitempty" yaml:"rewards,omitempty"`

	// SLOs are the SLO limits
	SLOs *SLOLimits `json:"SLOs,omitempty" yaml:"SLOs,omitempty"`
}

// assessTask enables assessment of versions
type assessTask struct {
	// TaskMeta has fields common to all tasks
	TaskMeta
	// With contains the inputs to this task
	With assessInputs `json:"with" yaml:"with"`
}

const (
	// AssessTaskName is the name of the task this file implements
	AssessTaskName = "assess"
)

// initializeDefaults sets default values for task inputs
func (t *assessTask) initializeDefaults() {}

// validateInputs for this task
func (t *assessTask) validateInputs() error {
	return nil
}

// Run executes the assess-app-versions task
func (t *assessTask) run(exp *Experiment) error {
	err := t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	if exp.Result.Insights == nil {
		log.Logger.Error("uninitialized insights within experiment")
		return errors.New("uninitialized insights within experiment")
	}
	if t.With.SLOs == nil ||
		exp.Result.Insights.NumVersions == 0 {
		// do nothing for now
		// todo: fix when rewards are introduced

		log.Logger.Warn("nothing to do; returning")
		return nil
	}

	// set rewards (if needed)
	err = exp.Result.Insights.setRewards(t.With.Rewards)
	if err != nil {
		return err
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
	if t.With.SLOs != nil {
		exp.Result.Insights.SLOsSatisfied = &SLOResults{
			Upper: evaluateSLOs(exp, t.With.SLOs.Upper, true),
			Lower: evaluateSLOs(exp, t.With.SLOs.Lower, false),
		}
	}

	// set RewardsWinners
	if t.With.Rewards != nil {
		exp.Result.Insights.RewardsWinners = &RewardsWinners{
			Max: evaluateRewards(exp, t.With.Rewards.Max, true),
			Min: evaluateRewards(exp, t.With.Rewards.Min, false),
		}
	}

	return err
}

func evaluateRewards(exp *Experiment, rewards []string, max bool) []int {
	winners := make([]int, len(rewards))
	for i := 0; i < len(rewards); i++ {
		for j := 0; j < exp.Result.Insights.NumVersions; j++ {
			winners[i] = identifyWinner(exp, rewards[i], max)
		}
	}
	return winners
}

func identifyWinner(e *Experiment, reward string, max bool) int {
	currentWinner := -1
	var currentWinningValue *float64

	for j := 0; j < e.Result.Insights.NumVersions; j++ {
		val := e.Result.Insights.ScalarMetricValue(j, reward)
		if val == nil {
			log.Logger.Warnf("unable to find value for version %v and metric %s", j, reward)
			continue
		}
		if currentWinningValue == nil || (max && *val > *currentWinningValue) || (!max && *val < *currentWinningValue) {
			currentWinningValue = val
			currentWinner = j
		}
	}

	return currentWinner
}

// evaluate SLOs and output the boolean SLO X version matrix
func evaluateSLOs(exp *Experiment, slos []SLO, upper bool) [][]bool {
	slosSatisfied := make([][]bool, len(slos))
	for i := 0; i < len(slos); i++ {
		slosSatisfied[i] = make([]bool, exp.Result.Insights.NumVersions)
		for j := 0; j < exp.Result.Insights.NumVersions; j++ {
			slosSatisfied[i][j] = sloSatisfied(exp, slos, i, j, upper)
		}
	}
	return slosSatisfied
}

// sloSatisfied returns true if SLO i satisfied by version j
func sloSatisfied(e *Experiment, slos []SLO, i int, j int, upper bool) bool {
	val := e.Result.Insights.ScalarMetricValue(j, slos[i].Metric)
	// check if metric is available
	if val == nil {
		log.Logger.Warnf("unable to find value for version %v and metric %s", j, slos[i].Metric)
		return false
	}

	if upper {
		// check upper limit
		if *val > slos[i].Limit {
			return false
		}
	} else {
		// check lower limit
		if *val < slos[i].Limit {
			return false
		}
	}

	return true
}
