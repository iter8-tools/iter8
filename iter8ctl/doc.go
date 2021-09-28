// Package iter8ctl provides iter8's command line utility for service operators to understand and diagnose their iter8 experiments.
//
// Installation
//
// The following command installs `iter8ctl` under the `/usr/local/bin` directory. To install under a different directory, change the value of `GOBIN` below.
//  GOBIN=/usr/local/bin go install github.com/iter8-tools/iter8ctl
//
// Usage Example 1
//
// Describe an iter8 Experiment resource object present in your Kubernetes cluster.
//  kubectl get experiment sklearn-iris-experiment-1 -n kfserving-test -o yaml > experiment.yaml
//  iter8ctl describe -f experiment.yaml
//
// Usage Example 2
//
// Supply experiment YAML using console input.
//  kubectl get experiment sklearn-iris-experiment-1 -n kfserving-test -o yaml > experiment.yaml
//  iter8ctl describe -f experiment.yaml
//
// Usage Example 3
//
// Periodically fetch an iter8 Experiment resource object present in your Kubernetes cluster and describe it. You can change the frequency by adjusting the sleep interval below.
//  kubectl get experiment sklearn-iris-experiment-1 -n kfserving-test -o yaml > experiment.yaml
//  iter8ctl describe -f experiment.yaml
//
// Sample output
//
// The following is the output of executing `iter8ctl describe -f testdata/experiment8.yaml`; the `testdata` folder is part of the `iter8ctl` GitHub repo and contains sample experiments used in tests.
//  ******
//  Experiment name: sklearn-iris-experiment-1
//  Experiment namespace: kfserving-test
//  Experiment target: kfserving-test/sklearn-iris
//
//  ******
//  Number of completed iterations: 10
//
//  ******
//  Winning version: canary
//
//  ******
//  Objectives
//  +----------------------+---------+--------+
//  |      OBJECTIVE       | DEFAULT | CANARY |
//  +----------------------+---------+--------+
//  | mean-latency <= 1000 | true    | true   |
//  +----------------------+---------+--------+
//  | error-rate <= 0.010  | true    | true   |
//  +----------------------+---------+--------+
//  ******
//  Metrics
//  +--------------------------------+---------------+---------------+
//  |             METRIC             |    DEFAULT    |    CANARY     |
//  +--------------------------------+---------------+---------------+
//  | 95th-percentile-tail-latency   | 330.681818182 | 310.319302313 |
//  | (milliseconds)                 |               |               |
//  +--------------------------------+---------------+---------------+
//  | mean-latency (milliseconds)    | 228.419047620 | 229.001070304 |
//  +--------------------------------+---------------+---------------+
//  | error-rate                     |             0 |             0 |
//  +--------------------------------+---------------+---------------+
//  | request-count                  | 117.444444445 |  57.714400001 |
//  +--------------------------------+---------------+---------------+
//
// Removal
//
// Remove `iter8ctl` as follows.
//  rm <GOBIN-value-used-during-install>/iter8ctl
//
package main
