package base

import (
	"fmt"
	"os"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/imdario/mergo"
	log "github.com/iter8-tools/iter8/base/log"
	gd "github.com/mcuadros/go-defaults"
)

const (
	// CollectGRPCTaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectGRPCTaskName = "grpc"
	// gRPC metric prefix
	gRPCMetricPrefix = "grpc"
	// gRPCRequestCountMetricName is name of the gRPC request count metric
	gRPCRequestCountMetricName = "request-count"
	// gRPCErrorCountMetricName is name of the gRPC error count metric
	gRPCErrorCountMetricName = "error-count"
	// gRPCErrorRateMetricName is name of the gRPC error rate metric
	gRPCErrorRateMetricName = "error-rate"
	// gRPCLatencySampleMetricName is name of the gRPC latency sample metric
	gRPCLatencySampleMetricName = "latency"
	// countErrorsDefault is the default value which indicates if errors are counted
	countErrorsDefault = true
	// insucureDefault is the default value which indicates that plaintext and insecure connection should be used
	insecureDefault = true
)

// collectHTTPInputs contain the inputs to the metrics collection task to be executed.
type collectGRPCInputs struct {
	runner.Config

	// Warmup indicates if task execution is for warmup purposes; if so the results will be ignored
	Warmup *bool `json:"warmup,omitempty" yaml:"warmup,omitempty"`

	// Endpoints is used to define multiple endpoints to test
	Endpoints map[string]runner.Config `json:"endpoints" yaml:"endpoints"`

	// Determines if Grafana dashboard should be created
	// dasboard vs report/assess tasks
	// TODO: remove
	Grafana bool `json:"grafana" yaml:"grafana"`
}

// collectGRPCTask enables load testing of gRPC services.
type collectGRPCTask struct {
	// TaskMeta has fields common to all tasks
	TaskMeta

	// With contains the inputs to this task
	With collectGRPCInputs `json:"with" yaml:"with"`
}

// GHZResult is the raw data sent to the metrics server
// This data will be transformed into httpDashboard when getGHZGrafana is called
type GHZResult struct {
	// key is the endpoint
	EndpointResults map[string]*runner.Report

	Summary Insights
}

// initializeDefaults sets default values for the collect task
func (t *collectGRPCTask) initializeDefaults() {
	// set defaults
	gd.SetDefaults(&t.With)
	// if dial timeout is zero, then set a default...
	if t.With.DialTimeout == 0 {
		td, _ := time.ParseDuration("10s")
		t.With.DialTimeout = runner.Duration(td)
	}
	// always count errors
	t.With.CountErrors = countErrorsDefault
	// todo: document how to use security credentials
	// remove this default altogether after enabling secure
	t.With.Insecure = insecureDefault
}

// validate task inputs
func (t *collectGRPCTask) validateInputs() error {
	return nil
}

// resultForVersion collects gRPC test result for a given version
func (t *collectGRPCTask) resultForVersion() (map[string]*runner.Report, error) {
	// the main idea is to run ghz with proper options

	var err error
	results := map[string]*runner.Report{}

	if len(t.With.Endpoints) > 0 {
		log.Logger.Trace("multiple endpoints")
		for endpointID, endpoint := range t.With.Endpoints {
			endpoint := endpoint // prevent implicit memory aliasing
			log.Logger.Trace(fmt.Sprintf("endpoint: %s", endpointID))

			// default from baseline
			call := t.With.Call
			if endpoint.Call != "" {
				call = endpoint.Call
			}

			host := t.With.Host
			if endpoint.Host != "" {
				host = endpoint.Host
			}

			// merge endpoint options with baseline options
			if err := mergo.Merge(&endpoint, t.With.Config); err != nil {
				log.Logger.Error(fmt.Sprintf("could not merge ghz options for endpoint \"%s\"", endpointID))
				return nil, err
			}
			eOpts := runner.WithConfig(&endpoint) // endpoint options

			log.Logger.Trace("run ghz gRPC test")
			igr, err := runner.Run(call, host, eOpts)
			if err != nil {
				log.Logger.WithStackTrace(err.Error()).Error(err)
				continue
			}

			resultsKey := gRPCMetricPrefix + "-" + endpointID
			if t.With.Grafana {
				resultsKey = endpoint.Call
			}
			results[resultsKey] = igr
		}
	} else {
		// TODO: supply all the allowed options
		opts := runner.WithConfig(&t.With.Config)

		log.Logger.Trace("run ghz gRPC test")
		igr, err := runner.Run(t.With.Call, t.With.Host, opts)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error(err)
			return results, err
		}

		resultsKey := gRPCMetricPrefix
		if t.With.Grafana {
			resultsKey = t.With.Call
		}
		results[resultsKey] = igr
	}

	return results, err
}

// latencySample extracts a latency sample from ghz result details
func latencySample(rd []runner.ResultDetail) []float64 {
	f := make([]float64, len(rd))
	for i := 0; i < len(rd); i++ {
		f[i] = float64(rd[i].Latency.Milliseconds())
	}
	return f
}

// Run executes this task
func (t *collectGRPCTask) run(exp *Experiment) error {
	// 1. initialize defaults
	var err error

	err = t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	// 2. collect raw results from ghz

	// run ghz test
	// collect ghz report
	// ghz reports will be further processed to populate metrics
	data, err := t.resultForVersion()
	if err != nil {
		return err
	}

	// ignore results if warmup
	if t.With.Warmup != nil && *t.With.Warmup {
		log.Logger.Debug("warmup: ignoring results")
		return nil
	}

	// 3. Init insights with num versions: always 1 in this task
	if err = exp.Result.initInsightsWithNumVersions(1); err != nil {
		return err
	}
	in := exp.Result.Insights

	if t.With.Grafana {
		// push data to metrics service
		ghzResult := GHZResult{
			EndpointResults: data,
			Summary:         *exp.Result.Insights,
		}

		// get URL of metrics server from environment variable
		metricsServerURL, ok := os.LookupEnv(MetricsServerURL)
		if !ok {
			errorMessage := "could not look up METRICS_SERVER_URL environment variable"
			log.Logger.Error(errorMessage)
			return fmt.Errorf(errorMessage)
		}

		if err = putPerformanceResultToMetricsService(metricsServerURL, exp.Metadata.Namespace, exp.Metadata.Name, ghzResult); err != nil {
			return err
		}
	} else {
		// 4. Populate all metrics collected by this task
		for provider, data := range data {
			// populate grpc request count
			// todo: this logic breaks for looped experiments. Fix when we get to loops.
			m := provider + "/" + gRPCRequestCountMetricName
			mm := MetricMeta{
				Description: "number of gRPC requests sent",
				Type:        CounterMetricType,
			}
			if err = in.updateMetric(m, mm, 0, float64(data.Count)); err != nil {
				return err
			}

			// populate error count & rate
			ec := float64(0)
			for _, count := range data.ErrorDist {
				ec += float64(count)
			}

			// populate count
			// todo: This logic breaks for looped experiments. Fix when we get to loops.
			m = provider + "/" + gRPCErrorCountMetricName
			mm = MetricMeta{
				Description: "number of responses that were errors",
				Type:        CounterMetricType,
			}
			if err = in.updateMetric(m, mm, 0, ec); err != nil {
				return err
			}

			// populate rate
			// todo: This logic breaks for looped experiments. Fix when we get to loops.
			m = provider + "/" + gRPCErrorRateMetricName
			rc := float64(data.Count)
			if rc != 0 {
				mm = MetricMeta{
					Description: "fraction of responses that were errors",
					Type:        GaugeMetricType,
				}
				if err = in.updateMetric(m, mm, 0, ec/rc); err != nil {
					return err
				}
			}

			// populate latency sample
			m = provider + "/" + gRPCLatencySampleMetricName
			mm = MetricMeta{
				Description: "gRPC Latency Sample",
				Type:        SampleMetricType,
				Units:       StringPointer("msec"),
			}
			lh := latencySample(data.Details)
			if err = in.updateMetric(m, mm, 0, lh); err != nil {
				return err
			}
		}
	}

	return nil
}
