---
template: main.html
---

# Your First Experiment

!!! tip "Load testing with SLOs"
    Launch an Iter8 experiment that performs a load test with SLO validation. Iter8 will:

      1. Generate HTTP GET requests for the app (hosted at https://example.com). 
      2. Collect built-in latency and error related metrics from the responses.
      3. Validate if the app satisfies service level objectives (SLOs) specified in the experiment.

???+ warning "Setup"
    1. **Install the `iter8` command line utility**
    ```shell
    GOBIN=/usr/local/bin/ go get github.com/iter8-tools/iter8
    ```
    `GOBIN` can be any folder in your `PATH` environment variable.

    2. **Clone the Iter8 repo**

        Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Clone your fork, and set the `ITER8` environment variable as follows.
        ```shell
        export USERNAME=<your GitHub username>
        ```
        ```shell
        git clone git@github.com:$USERNAME/iter8.git
        cd iter8
        export ITER8=$(pwd)
        ```

## 1. Launch experiment
Launch an Iter8 experiment for load testing with SLO validation as follows.
```shell
cd $ITER8/mkdocs/docs/experimenthub/load-test
iter8 run
```

The above command runs the [Iter8 experiment](../../../getting-started/what-is-iter8.md#what-is-an-iter8-experiment) defined in the file `$ITER8/mkdocs/docs/experimenthub/load-test/experiment.yaml`. 

This experiment generates requests, collects latency and error rate metrics for the app, and verifies that the app satisfies mean latency (50 msec), error rate (0.0), 95th percentile tail latency (100 msec) SLOs.

## 2. Observe experiment

### 2.a) Assert outcomes
Assert that the experiment completed and the app satisfies SLOs.
```shell
iter8 assert -c completed -c satisfiesSLOs --timeout=30s
```

### 2.b) Report results
Report the results of the experiment.

```shell
iter8 report
```

???+ info "Sample report"
    ```shell
    ****** Overview ******
    Experiment name: slo-validation-1dhq3
    Experiment namespace: default
    Target: app
    Testing pattern: Conformance
    Deployment pattern: Progressive

    ****** Progress Summary ******
    Experiment stage: Completed
    Number of completed iterations: 1

    ****** Winner Assessment ******
    > If the version being validated; i.e., the baseline version, satisfies the experiment objectives, it is the winner.
    > Otherwise, there is no winner.
    Winning version: my-app

    ****** Objective Assessment ******
    > Whether objectives specified in the experiment are satisfied by versions.
    > This assessment is based on last known metric values for each version.
    +--------------------------------------+------------+--------+
    |                METRIC                | CONDITION  | MY-APP |
    +--------------------------------------+------------+--------+
    | iter8-system/mean-latency            | <= 50.000  | true   |
    +--------------------------------------+------------+--------+
    | iter8-system/error-rate              | <= 0.000   | true   |
    +--------------------------------------+------------+--------+
    | iter8-system/latency-95th-percentile | <= 100.000 | true   |
    +--------------------------------------+------------+--------+

    ****** Metrics Assessment ******
    > Last known metric values for each version.
    +--------------------------------------+--------+
    |                METRIC                | MY-APP |
    +--------------------------------------+--------+
    | iter8-system/mean-latency            |  1.285 |
    +--------------------------------------+--------+
    | iter8-system/error-rate              |  0.000 |
    +--------------------------------------+--------+
    | iter8-system/latency-95th-percentile |  2.208 |
    +--------------------------------------+--------+
    | iter8-system/request-count           | 40.000 |
    +--------------------------------------+--------+
    | iter8-system/error-count             |  0.000 |
    +--------------------------------------+--------+
    ```
