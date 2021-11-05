---
template: main.html
---

# Your First Experiment

!!! tip "Load test https://example.com"
    Use an Iter8 experiment to load test https://example.com and validate error and latency related service level objectives (SLOs).

## 1. Install Iter8
Install the Iter8 command line utility using one of [these methods](../../getting-started/install.md).

## 2. Download experiment
[Iter8hub](iter8hub.md) provides useful samples for accelerating the creation of experiments. Download the `load-test` experiment sample from Iter8hub as follows.

```shell
iter8 hub -e load-test
```

## 3. Launch experiment
```shell
cd load-test
iter8 run
```

The above command reads in the `experiment.yaml` file, executes the specified tasks, and writes the results into the `results.yaml` file. The contents of the `experiment.yaml` file is as follows.

```yaml
# Task 1: generate HTTP requests for https://example.com
# collect Iter8's built-in latency and error related metrics
- task: collect-fortio-metrics
  with:
    versionInfo:
    - url: https://example.com
# Task 2: validate service level objectives for https://example.com 
# using the metrics collected in the above task
- task: assess-versions
  with:
    criteria:
      objectives:
        # error rate must be 0
      - metric: iter8-fortio/error-rate
        upperLimit: 0
        # 95th percentile latency must be under 100 msec
      - metric: iter8-fortio/p95.0
        upperLimit: 100
```

## 4. Assert outcomes
The above experiment must complete in a few seconds. Upon completion assert that all the SLOs are satisfied as follows.

```shell
iter8 assert -c valid=v0
```

This experiment involves only a single version of an app which serves the https://example.com URL. Iter8 names this version `v0`. The above command asserts that `v0` satisfies the `error-rate` and `p95.0` SLOs specified in the experiment.

## 5. Generate report
Generate a report of the experiment, including metrics and objectives.

```shell
iter8ctl gen 
```

A sample report is as follows.
```
--------------------------|-----
        Experiment summary|
--------------------------|-----
                     Name |my-exp
             App versions |my-app
                   Winner |my-app
          Testing pattern |SLOValidation
     Experiment completed |true
        Experiment failed |false
Number of completed tasks |2
```