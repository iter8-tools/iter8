package base

import (
	"errors"
	"strings"

	log "github.com/iter8-tools/iter8/base/log"
)

const (
	CollectABNMetrics = "abnmetrics"
	abnMetricProvider = "abn"
)

type ABNMetricsInputs struct {
	// Application is name of application to evaluate
	Application string `json:"application" yaml:"application"`
	// // Tracks are logical version names that should be evaluated
	// Tracks []string `json:"tracks" yaml:"tracks"`
}

type collectABNMetricsTask struct {
	TaskMeta
	With ABNMetricsInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the task
func (t *collectABNMetricsTask) initializeDefaults() {
	kd.InitKube()
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

	// // initialize insights in Result with number of tracks
	// err = exp.Result.initInsightsWithNumVersions(len(t.With.Tracks))
	// if err != nil {
	// 	return err
	// }

	//////////////////////////////////////////////////////////////////////////
	ms, err := NewMetricStoreSecret(t.With.Application, kd)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Warn("unable to read metrics")
	}

	// expect an error since we are not specifying the version or metric
	// but should still get the full appData object (or an empty one if none exists)
	c, err := ms.Read("", "")
	// if there are no metrics, we want to fail
	if err != nil {
		if strings.Contains(err.Error(), "no secret for application") ||
			strings.Contains(err.Error(), "expected key not found in secret") ||
			strings.Contains(err.Error(), "unable to unmarshal appData from secret") {
			return errors.New("unable to read metrics: " + err.Error())
		}
	}
	// for versionIndex, track := range t.With.Tracks {
	// 	for metricName, metricData := range c.appData[track].Metrics {
	// 		in.updateMetric(
	// 			abnMetricProvider+"/"+metricName,
	// 			MetricMeta{
	// 				Description: "summary metric",
	// 				Type:        SummaryMetricType,
	// 			},
	// 			versionIndex,
	// 			metricData,
	// 		)
	// 	}
	// }

	// count number of tracks
	numTracks := 0
	for _, versionData := range c.appData {
		lastEvent := versionData.History[len(versionData.History)-1]
		if lastEvent.Event == VersionMapTrackEvent {
			numTracks++
		}
	}
	if numTracks == 0 {
		log.Logger.Warnf("no tracks detected in application %s", ms.App)
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
	for versionName, versionData := range c.appData {
		lastEvent := versionData.History[len(versionData.History)-1]
		if lastEvent.Event == VersionMapTrackEvent {
			log.Logger.Tracef("version %s is mapped to track %s; using index %d", versionName, lastEvent.Track, versionIndex)
			for metricName, metricData := range versionData.Metrics {
				log.Logger.Tracef("   updating metric %s with data %+v", metricName, metricData)
				in.updateMetric(
					abnMetricProvider+"/"+metricName,
					MetricMeta{
						Description: "summary metric",
						Type:        SummaryMetricType,
					},
					versionIndex,
					metricData,
				)
			}
			versionIndex++
		}
	}

	return nil
}
