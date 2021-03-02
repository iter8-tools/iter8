---
template: overrides/main.html
hide:
- toc
---

# Realtime Observability

!!! abstract ""
    The  **iter8ctl** CLI enables you to observe an experiment in realtime. Use iter8ctl to observe metric values for each version, whether or not versions satisfy objectives, and the `winner`.


??? example "Sample output from iter8ctl"
    ```shell
    ****** Overview ******
    Experiment name: quickstart-exp
    Experiment namespace: default
    Target: default/sample-app
    Testing pattern: Canary
    Deployment pattern: Progressive

    ****** Progress Summary ******
    Experiment stage: Running
    Number of completed iterations: 3

    ****** Winner Assessment ******
    App versions in this experiment: [current candidate]
    Winning version: candidate
    Recommended baseline: candidate

    ****** Objective Assessment ******
    +--------------------------------+---------+-----------+
    |           OBJECTIVE            | CURRENT | CANDIDATE |
    +--------------------------------+---------+-----------+
    | mean-latency <= 50.000         | true    | true      |
    +--------------------------------+---------+-----------+
    | 95th-percentile-tail-latency   | true    | true      |
    | <= 100.000                     |         |           |
    +--------------------------------+---------+-----------+
    | error-rate <= 0.010            | true    | true      |
    +--------------------------------+---------+-----------+

    ****** Metrics Assessment ******
    +--------------------------------+---------+-----------+
    |             METRIC             | CURRENT | CANDIDATE |
    +--------------------------------+---------+-----------+
    | request-count                  | 429.334 |    16.841 |
    +--------------------------------+---------+-----------+
    | mean-latency (milliseconds)    |   0.522 |     0.712 |
    +--------------------------------+---------+-----------+
    | 95th-percentile-tail-latency   |   4.835 |     4.750 |
    | (milliseconds)                 |         |           |
    +--------------------------------+---------+-----------+
    | error-rate                     |   0.000 |     0.000 |
    +--------------------------------+---------+-----------+
    ```    

See [here](/getting-started/install/#step-4-install-iter8ctl) for instructions on installing iter8ctl. See [here](/getting-started/quick-start/with-knative/#7-observe-experiment) for an example of using iter8ctl to observe an experiment in realtime.
