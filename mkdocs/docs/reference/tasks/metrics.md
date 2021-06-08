---
template: main.html
---

# Metrics Tasks

## `metrics/collect`

### Overview

The `metrics/collect` takes enables collection of [builtin metrics](). It generates a stream of HTTP requests to one or more app/ML model versions, and collects latency/error metrics.

### Example

The following loop action contains a `metrics/collect` task which is executed during every loop of the experiment. The task sends a stream of HTTP requests to the version named `sample-app-v1` at the specified URL and collects builtin latency/error metrics for this version.

```yaml
loop:
- task: metrics/collect
  with:
    versions: 
    - name: sample-app-v1
      url: http://sample-app.default.svc.cluster.local
```

### Inputs

<!-- const (
	// CollectTaskName is the name of the task this file implements
	CollectTaskName string = "collect"

	// DefaultQPS is the default value of QPS (queries per sec) in collect task inputs
	DefaultQPS float32 = 8

	// DefaultTime is the default value of time (duration of queries) in collect task inputs
	DefaultTime string = "5s"
)

// Version contains header and url information needed to send requests to each version.
type Version struct {
	// name of the version
	// version names must be unique and must match one of the version names in the
	// VersionInfo field of the experiment
	Name string `json:"name" yaml:"name"`
	// how many queries per second will be sent to this version; optional; default 8
	QPS *float32 `json:"qps,omitempty" yaml:"qps,omitempty"`
	// HTTP headers to use in the query for this version; optional
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// URL to use for querying this version
	URL string `json:"url" yaml:"url"`
}

// CollectInputs contain the inputs to the metrics collection task to be executed.
type CollectInputs struct {
	// how long to run the metrics collector; optional; default 5s
	Time *string `json:"time,omitempty" yaml:"time,omitempty"`
	// list of versions
	Versions []Version `json:"versions" yaml:"versions"`
	// URL of the JSON file to send during the query; optional
	PayloadURL *string `json:"payloadURL,omitempty" yaml:"payloadURL,omitempty"`
	// if LoadOnly is set to true, this task will send requests without collecting metrics; optional
	LoadOnly *bool `json:"loadOnly,omitempty" yaml:"loadOnly,omitempty"`	
} -->

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| time | string | Duration of the `metrics/collect` task run. Specified in the [Go duration string format](https://golang.org/pkg/time/#ParseDuration). Default value is `5s`. | No |
| payloadURL | string | URL of JSON-encoded data. If this field is specified, the metrics collector will send HTTP POST requests to versions, and the POST requests will contain this JSON data as payload. | No |
| versions | [][Version](#version) | A non-empty list of versions. | Yes |
| loadOnly | bool | If set to true, this task will send requests without collecting metrics. Default value is `false`. | No |

#### Version
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| name | string | Name of the version. Version names must be unique and must match one of the version names in the VersionInfo field of the experiment. | Yes |
| qps | float | How many queries per second will be sent to this version. Default is 8.0. | No |
| headers | map[string]string | HTTP headers to be used in requests sent to this version. | No |
| url | string | HTTP URL of this version. | Yes |


### Result

This task will run for the specified duration (`time`), send requests to each version (`versions`) at the specified rate (`qps`), and will collect [built-in metrics]() for each version. Builtin metric values are stored in the metrics field of the experiment status in the same manner as custom metric values.

The task may result in an error, for instance, if one or more required fields are missing or if URLs are mis-specified. In this case, the experiment to which it belongs will fail.

### Start vs loop actions
If this task is embedded in start actions, it will run once at the beginning of the experiment.

If this task is embedded in loop actions, it will run in each loop of the experiment. The results from each run will be aggregated.

### Load generation without metrics collection
You can use this task to send HTTP GET and POST requests to app/ML model versions without collecting metrics by setting the [`loadOnly` input](#inputs) to `true`.