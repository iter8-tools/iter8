package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fortio.org/fortio/fhttp"
	"github.com/iter8-tools/iter8/core"
)

const (
	// TaskName is the name of the task this file implements
	CollectTaskName string = "collect-fortio-metrics"

	// DefaultQPS is the default value of QPS (queries per sec) in the collect task
	DefaultQPS float32 = 8

	// DefaultNumQueries is the default value of the number of queries used by the collect task
	DefaultNumQueries uint32 = 100

	// DefaultConnections is the default value of the number of connections
	DefaultConnections uint32 = 4

	// fortioOutputFile is where fortio data is held within the fortioFolder
	fortioOutputFile string = "output.json"

	// fortioPayloadFile is where fortio payload data is held within the fortioFolder
	fortioPayloadFile string = "payload.out"
)

var (
	// DefaultErrorRanges is the default value of the error ranges
	DefaultErrorRanges = []ErrorRange{{Lower: core.IntPointer(500)}}

	// fortioFolder is where fortio data is held
	fortioFolder = "/tmp"
)

// Version contains header and url information needed to send requests to each version.
type Version struct {
	// name of the version
	// version names must be unique and must match one of the version names in the
	// VersionInfo field of the experiment
	Name string `json:"name" yaml:"name"`
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url"`
}

// HTTP status code within this range is considered an error
type ErrorRange struct {
	Lower *int `json:"lower,omitempty" yaml:"lower,omitempty"`
	Upper *int `json:"upper,omitempty" yaml:"upper,omitempty"`
}

// CollectInputs contain the inputs to the metrics collection task to be executed.
type CollectInputs struct {
	// how many queries will be sent for each version; optional; default 100
	NumQueries *uint32 `json:"numQueries,omitempty" yaml:"numQueries,omitempty"`
	// how long to run the metrics collector; optional;
	// if both time and numQueries are specified, numQueries takes precedence
	Time *string `json:"time,omitempty" yaml:"time,omitempty"`
	// how many queries per second will be sent; optional; default 8
	QPS *float32 `json:"qps,omitempty" yaml:"qps,omitempty"`
	// how many parallel connections will be used; optional; default 4
	Connections *uint32 `json:"connections,omitempty" yaml:"connections,omitempty"`
	// string to be sent during queries as payload; optional
	PayloadStr *string `json:"payloadStr,omitempty" yaml:"payloadStr,omitempty"`
	// URL whose content will be sent as payload during queries; optional
	// if both payloadURL and payloadStr are specified, the URL takes precedence
	PayloadURL *string `json:"payloadURL,omitempty" yaml:"payloadURL,omitempty"`
	// valid HTTP content type string; specifying this switches the request from GET to POST
	ContentType *string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	// ranges of HTTP status codes that are considered as errors
	ErrorRanges []ErrorRange `json:"errorRanges,omitempty" yaml:"errorRanges,omitempty"`
	// information about versions
	VersionInfo []*Version `json:"versionInfo" yaml:"versionInfo"`
}

// ErrorCode checks if a given code is an error code
func (t *CollectTask) ErrorCode(code int) bool {
	for _, lims := range t.With.ErrorRanges {
		// if no lower limit (check upper)
		if lims.Lower == nil && code <= *lims.Upper {
			return true
		}
		// if no upper limit (check lower)
		if lims.Upper == nil && code >= *lims.Lower {
			return true
		}
		// if both limits are present (check both)
		if lims.Upper != nil && lims.Lower != nil && code <= *lims.Upper && code >= *lims.Lower {
			return true
		}
	}
	return false
}

// CollectTask enables collection of Iter8's built-in metrics.
type CollectTask struct {
	core.TaskMeta
	With CollectInputs `json:"with" yaml:"with"`
}

// MakeCollect constructs a CollectTask out of a collect task spec
func MakeCollect(t *core.TaskSpec) (core.Task, error) {
	if *t.Task != CollectTaskName {
		return nil, errors.New("task need to be " + CollectTaskName)
	}
	var err error
	var jsonBytes []byte
	var bt core.Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to CollectTask
	if err == nil {
		ct := &CollectTask{}
		err = json.Unmarshal(jsonBytes, &ct)
		if ct.With.VersionInfo == nil {
			return nil, errors.New("collect task with nil versionInfo")
		}
		bt = ct
	}
	return bt, err
}

// InitializeDefaults sets default values for the collect task
func (t *CollectTask) InitializeDefaults() {
	if t.With.NumQueries == nil && t.With.Time == nil {
		t.With.NumQueries = core.UInt32Pointer(DefaultNumQueries)
	}
	if t.With.QPS == nil {
		t.With.QPS = core.Float32Pointer(DefaultQPS)
	}
	if t.With.Connections == nil {
		t.With.Connections = core.UInt32Pointer(DefaultConnections)
	}
	if t.With.ErrorRanges == nil {
		t.With.ErrorRanges = DefaultErrorRanges
	}
}

// getResultFromFile reads the contents from a Fortio output file and returns it as a fortio result
func getResultFromFile(fortioOutputFile string) (*fhttp.HTTPRunnerResults, error) {
	// open JSON file
	jsonFile, err := os.Open(fortioOutputFile)
	// if os.Open returns an error, handle it
	if err != nil {
		return nil, err
	}

	// defer the closing of jsonFile so that we can parse it below
	defer jsonFile.Close()

	// read jsonFile as a byte array.
	bytes, err := ioutil.ReadAll(jsonFile)
	// if ioutil.ReadAll returns an error, handle it
	if err != nil {
		return nil, err
	}

	// unmarshal the result and return
	var res fhttp.HTTPRunnerResults
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// payloadFile downloads payload from a URL into a temp file, and returns its name
func payloadFile(url string) (string, error) {
	content, err := GetPayloadBytes(url)
	if err != nil {
		return "", err
	}

	tmpfile, err := ioutil.TempFile(fortioFolder, fortioPayloadFile)
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	return tmpfile.Name(), nil
}

// getFortioArgs constructs args for the Fortio command using the collect task spec for the j^th version
func (t *CollectTask) getFortioArgs(j int) ([]string, error) {
	// append Fortio load subcommand
	args := []string{"load"}

	// append numQueries or time
	if t.With.NumQueries != nil {
		args = append(args, "-n", fmt.Sprintf("%v", *t.With.NumQueries))
	} else {
		args = append(args, "-t", *t.With.Time)
	}

	// append qps
	args = append(args, "-qps", fmt.Sprintf("%f", *t.With.QPS))

	// append connections
	args = append(args, "-c", fmt.Sprintf("%v", *t.With.Connections))

	// append payload file if URL is specified
	if t.With.PayloadURL != nil {
		pf, err := payloadFile(*t.With.PayloadURL)
		if err != nil {
			return nil, err
		}
		args = append(args, "-payload-file", pf)
	} else if t.With.PayloadStr != nil {
		// append double quoted payload string if specified
		args = append(args, "-payload", fmt.Sprintf("%q", *t.With.PayloadStr))
	}

	// append content type
	if t.With.ContentType != nil {
		args = append(args, "-content-type", fmt.Sprintf("%q", *t.With.ContentType))
	}

	// append headers
	for header, value := range t.With.VersionInfo[j].Headers {
		args = append(args, "-H", fmt.Sprintf("\"%v: %v\"", header, value))
	}

	// append output file
	args = append(args, "-json", filepath.Join(fortioFolder, fortioOutputFile))

	// append URL to be queried by Fortio
	args = append(args, t.With.VersionInfo[j].URL)

	return args, nil
}

// resultForVersion collects Fortio result for a given version
func (t *CollectTask) resultForVersion(j int) (*fhttp.HTTPRunnerResults, error) {
	// the main idea is to run Fortio shell command with proper args
	// collect Fortio output as a file
	// and extract the result from the file, and return the result

	// get fortio args
	args, err := t.getFortioArgs(j)
	if err != nil {
		return nil, err
	}

	// setup Fortio command
	cmd := exec.Command("fortio", args...)

	// execute Fortio command
	stdoutBytes, err := cmd.CombinedOutput()
	if err != nil {
		core.Logger.WithStackTrace(err.Error()).Error("fortio execution error")
		core.Logger.WithStackTrace(string(stdoutBytes)).Error("output from fortio")
		return nil, err
	} else {
		core.Logger.WithStackTrace(string(stdoutBytes)).Trace("output from fortio")
	}

	// extract result from Fortio output file
	ifr, err := getResultFromFile(filepath.Join(fortioFolder, fortioOutputFile))
	if err != nil {
		return nil, err
	}

	return ifr, err
}

// Run executes the metrics/collect task
func (t *CollectTask) Run(exp *core.Experiment) error {
	var err error
	t.InitializeDefaults()

	fm := make([]*fhttp.HTTPRunnerResults, len(exp.Spec.Versions))

	// run fortio queries for each version sequentially
	for j := range t.With.VersionInfo {
		var data *fhttp.HTTPRunnerResults
		var err error
		if t.With.VersionInfo[j] == nil {
			data = nil
		} else {
			data, err = t.resultForVersion(j)
			if err == nil {
				fm[j] = data
			} else {
				return err
			}
		}
	}
	err = exp.SetFortioMetrics(fm)
	if err != nil {
		return err
	}

	// set metrics for each version for which fortio metrics are available
	for i := range exp.Spec.Versions {
		if exp.Result.Analysis.FortioMetrics[i] != nil {
			for _, m := range core.IFBackend.Metrics {
				fqName := core.IFBackend.Name + "/" + m.Name
				switch m.Name {
				case "request-count":
					err = exp.UpdateMetricForVersion(fqName, i, float64(exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Count))
					if err != nil {
						return err
					}
				case "error-count": // and error rate
					// compute val
					val := float64(0)
					for code, count := range exp.Result.Analysis.FortioMetrics[i].RetCodes {
						if t.ErrorCode(code) {
							val += float64(count)
						}
					}
					err = exp.UpdateMetricForVersion(fqName, i, val)
					if err != nil {
						return err
					}

					// error-rate
					rc := float64(exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Count)
					if rc != 0 {
						err = exp.UpdateMetricForVersion(core.IFBackend.Name+"/"+"error-rate", i, val/rc)
						if err != nil {
							return err
						}
					}

				case "mean-latency":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Avg)
					if err != nil {
						return err
					}

				case "min-latency":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Min)
					if err != nil {
						return err
					}

				case "max-latency":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Max)
					if err != nil {
						return err
					}

				case "stddev-latency":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.StdDev)
					if err != nil {
						return err
					}

				case "p50":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Percentiles[0].Value)
					if err != nil {
						return err
					}

				case "p75":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Percentiles[1].Value)
					if err != nil {
						return err
					}

				case "p90":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Percentiles[2].Value)
					if err != nil {
						return err
					}

				case "p99":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Percentiles[3].Value)
					if err != nil {
						return err
					}

				case "p99.9":
					err = exp.UpdateMetricForVersion(fqName, i, exp.Result.Analysis.FortioMetrics[i].DurationHistogram.Percentiles[4].Value)
					if err != nil {
						return err
					}

				}
			}
		}
	}
	return err
}

// GetPayloadBytes downloads payload from URL and returns a byte slice
func GetPayloadBytes(url string) ([]byte, error) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil || r.StatusCode >= 400 {
		return nil, errors.New("error while fetching payload")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, err
}
