package base

import (
	abnapp "github.com/iter8-tools/iter8/abn/application"
	k8sdriver "github.com/iter8-tools/iter8/base/k8sdriver"
	log "github.com/iter8-tools/iter8/base/log"
)

const (
	CollectABNMetrics = "abnmetrics"
	abnMetricProvider = "abn"
)

type ABNMetricsInputs struct {
	// Application is name of application to evaluate
	Application string `json:"application" yaml:"application"`
}

type collectABNMetricsTask struct {
	TaskMeta
	With ABNMetricsInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the task
func (t *collectABNMetricsTask) initializeDefaults() {
	k8sdriver.Driver.InitKube()
}

// validate task inputs
func (t *collectABNMetricsTask) validateInputs() error {
	return nil
}

// run exeuctes this task
func (t *collectABNMetricsTask) run(exp *Experiment) error {
	var err error

	// validate inputs
	err = t.validateInputs()
	if err != nil {
		return err
	}

	// initialize defaults
	t.initializeDefaults()

	rw := abnapp.ApplicationReaderWriter{Client: k8sdriver.Driver.Clientset}
	a, _ := rw.Read(t.With.Application)

	// count number of tracks
	numTracks := len(a.Tracks)
	if numTracks == 0 {
		log.Logger.Warn("no tracks detected in application")
		return nil
	}

	// initialize insights in Result with number of tracks
	err = exp.Result.initInsightsWithNumVersions(numTracks)
	if err != nil {
		return err
	}
	log.Logger.Tracef("intialized insights with %d versions", numTracks)

	in := exp.Result.Insights

	// add metrics for tracks
	versionIndex := 0
	for version, v := range a.Versions {
		t := v.GetTrack()
		if t != nil {
			log.Logger.Tracef("version %s is mapped to track %s; using index %d", version, *t, versionIndex)
			for metric, m := range v.Metrics {
				log.Logger.Tracef("   updating metric %s with data %+v", metric, m)
				in.updateMetric(
					abnMetricProvider+"/"+metric,
					MetricMeta{
						Description: "summary metric",
						Type:        SummaryMetricType,
					},
					versionIndex,
					m,
				)
			}
			versionIndex++
		}
	}

	return nil
}
