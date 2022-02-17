package base

import (
	"encoding/json"
	"errors"

	"github.com/bojand/ghz/runner"
	log "github.com/iter8-tools/iter8/base/log"
	gd "github.com/mcuadros/go-defaults"
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
	// ProtoURL	URL pointing to the Protocol Buffer file.
	ProtoURL *string `json:"protoURL,omitempty" yaml:"protoURL,omitempty"`
	// ProtosetURL	URL pointing to the protoset file.
	ProtosetURL *string `json:"protosetURL,omitempty" yaml:"protosetURL,omitempty"`
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
	// CollectGRPCTaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectGRPCTaskName = "gen-load-and-collect-metrics-grpc"
	// protoFileName is the proto buf file
	protoFileName = "ghz.proto"
	// protosetFileName is the protoset file
	protosetFileName = "ghz.protoset"
	// callDataJSONFileName is the JSON call data file
	callDataJSONFileName = "ghz-call-data.json"
	// callDataBinaryFileName is the binary call data file
	callDataBinaryFileName = "ghz-call-data.bin"
	// callMetadataJSONFileName is the JSON call metadata file
	callMetadataJSONFileName = "ghz-call-metadata.json"
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

// collectGRPCTask enables load testing of gRPC services.
type collectGRPCTask struct {
	TaskMeta
	With collectGRPCInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the collect task
func (t *collectGRPCTask) initializeDefaults() {
	// set defaults
	gd.SetDefaults(&t.With.Config)
	// always count errors
	t.With.Config.CountErrors = countErrorsDefault
	// todo: document how to use security credentials
	// remove this default altogether after enabling secure
	t.With.Config.Insecure = insecureDefault
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
		err := getFileFromURL(*t.With.ProtoURL, protoFileName)
		if err != nil {
			return nil, err
		}
		ghzc.Proto = protoFileName
	}
	// get protoset file
	if t.With.ProtosetURL != nil {
		err := getFileFromURL(*t.With.ProtosetURL, protosetFileName)
		if err != nil {
			return nil, err
		}
		ghzc.Protoset = protosetFileName
	}
	// get JSON call data file
	if t.With.DataURL != nil {
		err := getFileFromURL(*t.With.ProtoURL, callDataJSONFileName)
		if err != nil {
			return nil, err
		}
		ghzc.DataPath = callDataJSONFileName
	}
	// get binary data file
	if t.With.BinaryDataURL != nil {
		err := getFileFromURL(*t.With.ProtoURL, callDataBinaryFileName)
		if err != nil {
			return nil, err
		}
		ghzc.BinDataPath = callDataBinaryFileName
	}
	// get metadata file
	if t.With.MetadataURL != nil {
		err := getFileFromURL(*t.With.ProtoURL, callMetadataJSONFileName)
		if err != nil {
			return nil, err
		}
		ghzc.MetadataPath = callMetadataJSONFileName
	}

	return &ghzc, nil
}

// resultForVersion collects gRPC test result for a given version
func (t *collectGRPCTask) resultForVersion(j int) (*runner.Report, error) {
	// the main idea is to run ghz with proper options
	ghzc, err := t.getGhzConfig(j)
	if err != nil {
		return nil, err
	}
	ghzcBytes, _ := json.MarshalIndent(ghzc, "", "	")
	log.Logger.WithStackTrace(string(ghzcBytes)).Trace("runner config")

	opts := runner.WithConfig(ghzc)

	// todo: supply all the allowed options
	igr, err := runner.Run(t.With.VersionInfo[j].Call, t.With.VersionInfo[j].Host, opts)
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
			// todo: this logic breaks for looped experiments. Fix when we get to loops.
			m := gRPCMetricPrefix + "/" + gRPCRequestCountMetricName
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
			// todo: This logic breaks for looped experiments. Fix when we get to loops.
			m = gRPCMetricPrefix + "/" + gRPCErrorCountMetricName
			mm = MetricMeta{
				Description: "number of responses that were errors",
				Type:        CounterMetricType,
			}
			in.updateMetric(m, mm, i, ec)

			// populate rate
			// todo: This logic breaks for looped experiments. Fix when we get to loops.
			m = gRPCMetricPrefix + "/" + gRPCErrorRateMetricName
			rc := float64(gr[i].Count)
			if rc != 0 {
				mm = MetricMeta{
					Description: "fraction of responses that were errors",
					Type:        GaugeMetricType,
				}
				in.updateMetric(m, mm, i, ec/rc)
			}

			// populate latency sample
			m = gRPCMetricPrefix + "/" + gRPCLatencySampleMetricName
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
