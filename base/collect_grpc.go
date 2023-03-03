package base

import (
	"errors"
	"time"

	"github.com/bojand/ghz/runner"
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

type collectGRPCInputs struct {
	runner.Config
	// Warmup indicates if task execution is for warmup purposes; if so the results will be ignored
	Warmup *bool `json:"warmup,omitempty" yaml:"warmup,omitempty"`
}

// collectGRPCTask enables load testing of gRPC services.
type collectGRPCTask struct {
	// TaskMeta has fields common to all tasks
	TaskMeta
	// With contains the inputs to this task
	With collectGRPCInputs `json:"with" yaml:"with"`
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
func (t *collectGRPCTask) resultForVersion() (*runner.Report, error) {
	// the main idea is to run ghz with proper options

	opts := runner.WithConfig(&t.With.Config)

	// todo: supply all the allowed options
	igr, err := runner.Run(t.With.Call, t.With.Host, opts)
	if err != nil {
		e := errors.New("ghz run failed")
		log.Logger.WithStackTrace(err.Error()).Error(e)
		if igr == nil {
			e = errors.New("failed to get results since ghz run was aborted")
			log.Logger.Error(e)
		}
		return nil, e
	}
	log.Logger.Trace("ran ghz gRPC test")
	log.Logger.Trace(igr.ErrorDist)
	return igr, err
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

	// 4. Populate all metrics collected by this task
	if data != nil { // assuming there is some raw ghz result to process
		// populate grpc request count
		// todo: this logic breaks for looped experiments. Fix when we get to loops.
		m := gRPCMetricPrefix + "/" + gRPCRequestCountMetricName
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
		m = gRPCMetricPrefix + "/" + gRPCErrorCountMetricName
		mm = MetricMeta{
			Description: "number of responses that were errors",
			Type:        CounterMetricType,
		}
		if err = in.updateMetric(m, mm, 0, ec); err != nil {
			return err
		}

		// populate rate
		// todo: This logic breaks for looped experiments. Fix when we get to loops.
		m = gRPCMetricPrefix + "/" + gRPCErrorRateMetricName
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
		m = gRPCMetricPrefix + "/" + gRPCLatencySampleMetricName
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
	return nil
}
