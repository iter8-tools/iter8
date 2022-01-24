package base

import (
	"errors"

	"github.com/bojand/ghz/runner"
	log "github.com/iter8-tools/iter8/base/log"
)

/*
Credit: The inputs for this task, including their JSON/YAML tags and go doc comments are derived from https://github.com/bojand/ghz.
*/

// versionGRPC contains call and host information needed to send gRPC requests to each version.
// this struct holds inputs that is specific to each version
type versionGRPC struct {
	// Call A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.
	Call string `json:"call" yaml:"call"`
	// Host gRPC host in the hostname[:port] format.
	Host string `json:"host" yaml:"host"`
}

// collectGRPCInputs holds all the inputs for this task
type collectGRPCInputs struct {
	// even though runner.Config is embedded,
	// many of the fields in runner.Config will not be supported.
	// Refer to iter8.tools documentation for supported fields.
	runner.Config
	// additional settings
	// ProtoURL	URL pointing to the proto buf.
	ProtoURL *string `json:"protoURL,omitempty" yaml:"protoURL,omitempty"`
	// data and metadata settings
	// DataURL	URL pointing to JSON data to be sent as part of the call.
	// This takes precedence over Data.
	DataURL *string `json:"JSONDataURL,omitempty" yaml:"JSONDataURL,omitempty"`
	// BinaryDataURL	URL pointing to call data as serialized binary message or multiple count-prefixed messages.
	BinaryDataURL *string `json:"BinaryDataURL,omitempty" yaml:"BinaryDataURL,omitempty"`
	// MetadataURL	URL pointing to call metadata in JSON format.
	// This takes precedence over Metadata.
	MetadataURL *string `json:"MetadataURL,omitempty" yaml:"MetadataURL,omitempty"`

	// VersionInfo is a non-empty list of versionGRPC values.
	VersionInfo []*versionGRPC `json:"versionInfo" yaml:"versionInfo"`
}

const (
	// CollectGPRCTaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectGPRCTaskName = "gen-load-and-collect-metrics-grpc"
	// protoFileName is the proto buf file
	protoFileName = "ghz.proto"
	// callDataJSONFileName is the JSON call data file
	callDataJSONFileName = "ghz-call-data.json"
	// callDataBinaryFileName is the binary call data file
	callDataBinaryFileName = "ghz-call-data.bin"
	// callMetadataJSONFileName is the JSON call metadata file
	callMetadataJSONFileName = "ghz-call-metadata.json"
	// gRPCRequestCountMetricName is name of the gRPC request count metric
	gRPCRequestCountMetricName = "grpc-request-count"
	// gRPCErrorCountMetricName is name of the gRPC error count metric
	gRPCErrorCountMetricName = "grpc-error-count"
	// gRPCErrorRateMetricName is name of the gRPC error rate metric
	gRPCErrorRateMetricName = "grpc-error-rate"
	// gRPCLatencySampleMetricName is name of the gRPC latency sample metric
	gRPCLatencySampleMetricName = "grpc-latency-sample"
	// countErrorsDefault is the default value which indicates if errors are counted
	countErrorsDefault = true
)

// collectGRPCTask enables load testing of gRPC services.
type collectGRPCTask struct {
	taskMeta
	With collectGRPCInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectGRPCTask) initializeDefaults() {
	// always count errors
	t.With.Config.CountErrors = countErrorsDefault
}

// validate task inputs
func (t *collectGRPCTask) validateInputs() error {
	return nil
}

// getGhzConfig constructs ghz's runner.Config for version j based on task inputs
func (t *collectGRPCTask) getGhzConfig(j int) (*runner.Config, error) {
	ghzc := t.With.Config
	ghzc.Call = t.With.VersionInfo[j].Call
	ghzc.Host = t.With.VersionInfo[j].Host

	// get proto file
	if t.With.ProtoURL != nil {
		err := GetFileFromURL(*t.With.ProtoURL, protoFileName)
		if err != nil {
			return nil, err
		}
		ghzc.Proto = protoFileName
	}
	// get JSON call data file
	if t.With.DataURL != nil {
		err := GetFileFromURL(*t.With.ProtoURL, callDataJSONFileName)
		if err != nil {
			return nil, err
		}
		ghzc.DataPath = callDataJSONFileName
	}
	// get binary data file
	if t.With.BinaryDataURL != nil {
		err := GetFileFromURL(*t.With.ProtoURL, callDataBinaryFileName)
		if err != nil {
			return nil, err
		}
		ghzc.BinDataPath = callDataBinaryFileName
	}
	// get metadata file
	if t.With.MetadataURL != nil {
		err := GetFileFromURL(*t.With.ProtoURL, callMetadataJSONFileName)
		if err != nil {
			return nil, err
		}
		ghzc.MetadataPath = callMetadataJSONFileName
	}

	return &ghzc, nil
}

// getGhzOption constructs ghz's runner.Option based on task inputs
func (t *collectGRPCTask) getGhzOption(j int) (runner.Option, error) {
	ghzc, err := t.getGhzConfig(j)
	if err != nil {
		return nil, err
	}
	return runner.WithConfig(ghzc), nil
}

// resultForVersion collects gRPC test result for a given version
func (t *collectGRPCTask) resultForVersion(j int) (*runner.Report, error) {
	// the main idea is to run ghz with proper options

	ghzo, err := t.getGhzOption(j)
	if err != nil {
		return nil, err
	}
	log.Logger.Trace("got ghz options")
	igr, err := runner.Run(t.With.VersionInfo[j].Call, t.With.VersionInfo[j].Host, ghzo)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("ghz failed")
		if igr == nil {
			log.Logger.Error("failed to get results since ghz run was aborted")
		}
	}
	log.Logger.Trace("ran ghz gRPC test")
	return igr, err
}

// latencySample extracts a latency sample from ghz result details
func latencySample(rd []runner.ResultDetail) []float64 {
	f := make([]float64, len(rd))
	for i := 0; i < len(rd); i++ {
		f[i] = float64(rd[i].Latency)
	}
	return f
}

// Run executes this task
func (t *collectGRPCTask) Run(exp *Experiment) error {
	// 1. initialize defaults
	var err error

	err = t.validateInputs()
	if err != nil {
		return err
	}

	t.initializeDefaults()

	// 2. validate input (without going crazy in terms of rules)
	// idea is to check for the most obvious errors here
	// validation also happens at the chart/values level
	if len(t.With.VersionInfo) == 0 {
		log.Logger.Error("collect task must specify info for at least one version")
		return errors.New("collect task must specify info for at least one version")
	}

	// 3. collect raw results from ghz for each version

	// initialize a slice to hold reports from the ghz tests
	gr := make([]*runner.Report, len(t.With.VersionInfo))

	// run ghz test for each version sequentially
	// collect ghz report for each version
	// ghz reports will be further processed to populate metrics
	for j := range t.With.VersionInfo {
		log.Logger.Trace("initiating ghz for version ", j)
		var data *runner.Report
		var err error
		if t.With.VersionInfo[j] == nil { // nothing to do for this version
			data = nil
		} else {
			// this is where we get the raw ghz data for version j
			data, err = t.resultForVersion(j)
			if err == nil {
				gr[j] = data
			} else {
				return err
			}
		}
	}

	// 4. The inputs for this task determine the number of versions participating in the experiment.
	// Hence, init insights with num versions
	err = exp.Result.initInsightsWithNumVersions(len(t.With.VersionInfo))
	if err != nil {
		return err
	}
	in := exp.Result.Insights

	// 5. Populate all metrics collected by this task
	for i := range t.With.VersionInfo { // for each version
		if gr[i] != nil { // assuming there is some raw ghz result to process for this version
			// populate grpc request count
			// ToDo: This logic breaks for looped experiments. Fix when we get to loops.
			m := iter8BuiltInPrefix + "/" + gRPCRequestCountMetricName
			mm := MetricMeta{
				Description: "number of gRPC requests sent",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, float64(gr[i].Count))

			// populate error count & rate
			ec := float64(0)
			for _, count := range gr[i].ErrorDist {
				ec += float64(count)
			}

			// populate count
			// ToDo: This logic breaks for looped experiments. Fix when we get to loops.
			m = iter8BuiltInPrefix + "/" + gRPCErrorCountMetricName
			mm = MetricMeta{
				Description: "number of responses that were errors",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, ec)

			// populate rate
			// ToDo: This logic breaks for looped experiments. Fix when we get to loops.
			m = iter8BuiltInPrefix + "/" + gRPCErrorRateMetricName
			rc := float64(gr[i].Count)
			if rc != 0 {
				mm = MetricMeta{
					Description: "fraction of responses that were errors",
					Type:        GaugeMetricType,
				}
				in.updateMetric(m, mm, i, ec/rc)
			}

			// populate latency sample
			m = iter8BuiltInPrefix + "/" + gRPCLatencySampleMetricName
			mm = MetricMeta{
				Description: "gRPC Latency Sample",
				Type:        SampleMetricType,
				Units:       StringPointer("msec"),
			}
			lh := latencySample(gr[i].Details)
			in.updateMetric(m, mm, i, lh)
		}
	}
	return nil
}
