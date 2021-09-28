package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/etc3/controllers"
	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	// TaskName is the name of the task this file implements
	TaskName string = "metrics/collect"

	// DefaultQPS is the default value of QPS (queries per sec) in the collect task
	DefaultQPS float32 = 8

	// DefaultNumQueries is the default value of the number of queries used by the collect task
	DefaultNumQueries uint32 = 100

	// DefaultConnections is the default value of the number of connections
	DefaultConnections uint32 = 4

	// fortioFolder is where fortio data is held
	fortioFolder string = "/tmp"

	// fortioOutputFile is where fortio data is held within the fortioFolder
	fortioOutputFile string = "output.json"

	// fortioPayloadFile is where fortio payload data is held within the fortioFolder
	fortioPayloadFile string = "payload.out"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

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
	// if LoadOnly is true, send requests without collecting metrics; optional; default false
	LoadOnly *bool `json:"loadOnly,omitempty" yaml:"loadOnly,omitempty"` // list of versions
	// string to be sent during queries as payload; optional
	PayloadStr *string `json:"payloadStr,omitempty" yaml:"payloadStr,omitempty"`
	// URL whose content will be sent as payload during queries; optional
	// if both payloadURL and payloadStr are specified, the URL takes precedence
	PayloadURL *string `json:"payloadURL,omitempty" yaml:"payloadURL,omitempty"`
	// valid HTTP content type string; specifying this switches the request from GET to POST
	ContentType *string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	// information about versions
	Versions []Version `json:"versions" yaml:"versions"`
}

// CollectTask enables collection of Iter8's built-in metrics.
type CollectTask struct {
	core.TaskMeta
	With CollectInputs `json:"with" yaml:"with"`
}

// Make constructs a CollectTask out of a collect task spec
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if *t.Task != TaskName {
		return nil, errors.New("task need to be " + TaskName)
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
		if ct.With.Versions == nil {
			return nil, errors.New("collect task with nil versions")
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
	if t.With.LoadOnly == nil {
		t.With.LoadOnly = core.BoolPointer(false)
	}
	if t.With.QPS == nil {
		t.With.QPS = core.Float32Pointer(DefaultQPS)
	}
	if t.With.Connections == nil {
		t.With.Connections = core.UInt32Pointer(DefaultConnections)
	}
}

////
/////////////
////

// DurationSample is a Fortio duration sample
type DurationSample struct {
	Start float64
	End   float64
	Count int
}

// DurationHist is the Fortio duration histogram
type DurationHist struct {
	Count int
	Max   float64
	Sum   float64
	Data  []DurationSample
}

// Result is the result of a single Fortio run; it contains the result for a single version
type Result struct {
	DurationHistogram DurationHist
	RetCodes          map[string]int
}

// aggregate existing results, with a new result for a specific version
func aggregate(oldResults map[string]*Result, version string, newResult *Result) map[string]*Result {
	// there are no existing results...
	if oldResults == nil {
		m := make(map[string]*Result)
		m[version] = newResult
		return m
	}
	if updatedResult, ok := oldResults[version]; ok {
		// there are existing results for the input version
		// aggregate count, max and sum
		updatedResult.DurationHistogram.Count += newResult.DurationHistogram.Count
		updatedResult.DurationHistogram.Max = math.Max(oldResults[version].DurationHistogram.Max, newResult.DurationHistogram.Max)
		updatedResult.DurationHistogram.Sum = oldResults[version].DurationHistogram.Sum + newResult.DurationHistogram.Sum

		// aggregation duration histogram data
		updatedResult.DurationHistogram.Data = append(updatedResult.DurationHistogram.Data, newResult.DurationHistogram.Data...)

		// aggregate return code counts
		if updatedResult.RetCodes == nil {
			updatedResult.RetCodes = newResult.RetCodes
		} else {
			for key := range newResult.RetCodes {
				oldResults[version].RetCodes[key] += newResult.RetCodes[key]
			}
		}
	} else {
		// there are no existing results for the input version
		oldResults[version] = newResult
	}
	// this is efficient because oldResults is a map with pointer values
	// no deep copies of structs
	return oldResults
}

// getResultFromFile reads the contents from a Fortio output file and returns it as a Result
func getResultFromFile(fortioOutputFile string) (*Result, error) {
	// open JSON file
	jsonFile, err := os.Open(fortioOutputFile)
	// if os.Open returns an error, handle it
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// defer the closing of jsonFile so that we can parse it below
	defer jsonFile.Close()

	// read jsonFile as a byte array.
	bytes, err := ioutil.ReadAll(jsonFile)
	// if ioutil.ReadAll returns an error, handle it
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// unmarshal the result and return
	var res Result
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &res, nil
}

// payloadFile downloads payload from a URL into a temp file, and returns its name
func payloadFile(url string) (string, error) {
	content, err := core.GetPayloadBytes(url)
	if err != nil {
		log.Error("Error while getting payload bytes: ", err)
		return "", err
	}

	tmpfile, err := ioutil.TempFile(fortioFolder, fortioPayloadFile)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		log.Fatal(err)
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
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
	for header, value := range t.With.Versions[j].Headers {
		args = append(args, "-H", fmt.Sprintf("\"%v: %v\"", header, value))
	}

	// append output file
	args = append(args, "-json", filepath.Join(fortioFolder, fortioOutputFile))

	// append URL to be queried by Fortio
	args = append(args, t.With.Versions[j].URL)

	return args, nil
}

// resultForVersion collects Fortio result for a given version
func (t *CollectTask) resultForVersion(j int) (*Result, error) {
	// the main idea is to run Fortio shell command with proper args
	// collect Fortio output as a file
	// and extract the result from the file, and return the result

	var execOut bytes.Buffer

	// get fortio args
	args, err := t.getFortioArgs(j)
	if err != nil {
		return nil, err
	}

	// setup Fortio command
	cmd := exec.Command("fortio", args...)
	cmd.Stdout = &execOut
	cmd.Stderr = os.Stderr
	log.Trace("Invoking: " + cmd.String())

	// execute Fortio command
	err = cmd.Run()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// extract result from Fortio output file
	ifr, err := getResultFromFile(filepath.Join(fortioFolder, fortioOutputFile))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return ifr, err
}

// Run executes the metrics/collect task
// ToDo: error handling
func (t *CollectTask) Run(ctx context.Context) error {
	log.Trace("collect task run started...")
	t.InitializeDefaults()

	// fortioData will be used if this not a loadOnly task
	fortioData := make(map[string]*Result)

	// get experiment
	exp, err := core.GetExperimentFromContext(ctx)
	if err != nil {
		return err
	}

	// if this task is **not** loadOnly
	if !*t.With.LoadOnly {
		// bootstrap AggregatedBuiltinHists with data already present in experiment status
		if exp.Status.Analysis != nil && exp.Status.Analysis.AggregatedBuiltinHists != nil {
			jsonBytes, _ := json.Marshal(exp.Status.Analysis.AggregatedBuiltinHists.Data)
			json.Unmarshal(jsonBytes, &fortioData)
		}
	}

	// run fortio queries for each version sequentially
	for j := range t.With.Versions {
		data, err := t.resultForVersion(j)
		if err == nil {
			// if this task is **not** loadOnly
			if !*t.With.LoadOnly {
				// Update fortioData in a threadsafe manner
				fortioData = aggregate(fortioData, t.With.Versions[j].Name, data)
			}
		} else {
			return err
		}
	}

	// if this task is **not** loadOnly
	if !*t.With.LoadOnly {
		// Update experiment status with results
		// update to experiment status will result in reconcile request to etc3
		// unless the task runner job executing this action is completed, this request will not have have an immediate effect in the experiment reconcilation process

		bytes1, _ := json.Marshal(fortioData)

		exp.SetAggregatedBuiltinHists(v1.JSON{Raw: bytes1})

		core.UpdateInClusterExperimentStatus(exp)

		var prettyBody bytes.Buffer
		bytes2, _ := json.Marshal(exp)

		json.Indent(&prettyBody, bytes2, "", "  ")
		log.Trace(prettyBody.String())
	}

	// Iter8Log
	if err == nil {
		// get action from context
		a, err := core.GetActionStringFromContext(ctx)
		if err != nil {
			return err
		}

		il := controllers.Iter8Log{
			IsIter8Log:          true,
			ExperimentName:      exp.Name,
			ExperimentNamespace: exp.Namespace,
			Source:              controllers.Iter8LogSourceTR,
			Priority:            controllers.Iter8LogPriorityLow,
			Message:             "metrics collection completed for all versions",
			Precedence:          core.GetIter8LogPrecedence(exp, a),
		}
		fmt.Println(il.JSON())
	}

	return err
}
