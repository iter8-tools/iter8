# Iter8

<img alt="Iter8" src="images/iter8.png" align="center">

***

[![Iter8 release (latest SemVer)](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
[![Test Status](https://github.com/iter8-tools/iter8/workflows/tests/badge.svg)](https://github.com/iter8-tools/iter8/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/iter8-tools/iter8/branch/master/graph/badge.svg)](https://codecov.io/gh/iter8-tools/iter8)
[![Slack channel](https://img.shields.io/badge/Slack-Join-purple)](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
[![Community meetups](https://img.shields.io/badge/meet-Iter8%20community%20meetups-brightgreen)](https://iter8.tools/0.7/getting-started/help/#iter8-community-meetings)

## 1. Install Iter8
```shell
brew tap iter8-tools/iter8
brew install iter8
```

You can also install Iter8 using [pre-compiled binaries](https://iter8.tools/latest/getting-started/install/) or [`go 1.16+`](https://iter8.tools/latest/getting-started/install/).

## 2. Your first experiment
Load test an HTTP service and validate its latency and error-related service level objectives (SLOs).

```shell
iter8 launch load-test-http --set url=https://example.com \
                            --set numRequests=200 \
                            --set rps=10.0 \
                            --set SLOs.http-error-rate=0 \
                            --set SLOs.http-latency/mean=30 \
                            --set SLOs.http-latency/p95=100
```

The `iter8 launch` command shown above does the following.
1.  Create a local folder called `load-test` containing the chart.
2.  Generate an Iter8 experiment spec in a file named `experiment.yaml`, by combining the chart with the supplied values (`--set`).
3.  Run the load test experiment, and output results to a file named `result.yaml`.

## 3. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

## 4. View report
View a report of the experiment in text or HTML formats.

### Text Report
```shell
iter8 report
```

A sample text report is as follows.

```shell

Experiment summary:
*******************

  Experiment completed: true
  No failed tasks: true
  Total number of tasks: 2
  Number of completed tasks: 2

Whether or not service level objectives (SLOs) are satisfied:
*************************************************************

  SLO Conditions                           |Satisfied
  --------------                           |---------
  built-in/http-latency-mean (msec) <= 100 |true
  built-in/http-latency-p95 (msec) <= 150  |true
  

Latest observed values for metrics:
***********************************

  Metric                              |value
  -------                             |-----
  built-in/http-error-count           |0.00
  built-in/http-error-rate            |0.00
  built-in/http-latency-max (msec)    |186.56
  built-in/http-latency-mean (msec)   |11.25
  built-in/http-latency-min (msec)    |4.05
  built-in/http-latency-p50 (msec)    |6.59
  built-in/http-latency-p75 (msec)    |7.79
  built-in/http-latency-p90 (msec)    |9.23
  built-in/http-latency-p95 (msec)    |10.00
  built-in/http-latency-p99 (msec)    |183.94
  built-in/http-latency-p99.9 (msec)  |186.30
  built-in/http-latency-stddev (msec) |28.01
  built-in/http-request-count         |200.00
```

### HTML Report
```shell
iter8 report -o html > report.html
```

You can view `report.html` in any browser. A sample report is [here](mkdocs/docs/getting-started/images/report.html.png).


## More Examples

## Documentation
Iter8 documentation is available at https://iter8.tools.

## Contributing
See [here](https://iter8.tools/0.8/contributing/) for information about ways to contribute, Iter8 community meetings, finding an issue, asking for help, pull-request lifecycle, and more.
