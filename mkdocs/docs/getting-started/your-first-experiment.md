---
template: main.html
---

# Your First Experiment

!!! tip "Load test https://example.com"
    Use an [Iter8 experiment](concepts.md#what-is-an-iter8-experiment) to load test https://example.com and validate latency and error-related service level objectives (SLOs).

## 1. [Install Iter8](install.md)

## 2. Download experiment
Download the `load-test` experiment folder from the [Iter8 hub](../user-guide/topics/iter8hub.md) as follows.

```shell
iter8 hub -e load-test
```

## 3. Run experiment
An [Iter8 experiment](concepts.md#what-is-an-iter8-experiment) is a sequence of tasks that produce metrics-driven insights for your app/ML model, validates it, and optionally performs a rollout of your app/ML model. 

Experiments are specified declaratively using the `experiment.yaml` file. The `iter8 run` command reads this file, runs the experiment, and writes the result of the experiment run into the `result.yaml` file. Run the `load-test` experiment as follows.

```shell
cd load-test
iter8 run
```

This experiment uses the [`gen-load-and-collect-metrics` task](../user-guide/tasks/collect.md) for generating load and collecting metrics, and the [`assess-app-versions` task](../user-guide/tasks/assess.md) for validating SLOs.

??? note "Look inside experiment.yaml"
    ```yaml
    # task 1: generate HTTP requests for https://example.com and
    # collect Iter8's built-in latency and error-related metrics
    - task: gen-load-and-collect-metrics
      with:
        versionInfo:
        - url: https://example.com
    # task 2: validate if the app (hosted at https://example.com) satisfies 
    # service level objectives (SLOs)
    # this task uses the built-in metrics collected by task 1 for validation
    - task: assess-app-versions
      with:
        SLOs:
          # error rate must be 0
        - metric: built-in/error-rate
          upperLimit: 0
          # 95th percentile latency must be under 100 msec
        - metric: built-in/p95.0
          upperLimit: 100
    ```

## 4. Assert outcomes
The `load-test` experiment you ran above should complete in a few seconds. Upon completion, assert that the experiment completed without any failures and SLOs are satisfied, as follows.

```shell
iter8 assert -c completed -c nofailure -c slos
```

??? note "Look inside sample output of assert"

    ```shell
    INFO[2021-11-10 09:33:12] experiment completed
    INFO[2021-11-10 09:33:12] experiment has no failure                    
    INFO[2021-11-10 09:33:12] SLOs are satisfied                           
    INFO[2021-11-10 09:33:12] all conditions were satisfied
    ```

## 5. Generate report
Generate a report of the experiment including a summary of the experiment, SLOs, and metrics.

=== "HTML"
    ```shell
    iter8 report -o html > report.html
    # open report.html with a browser. In MacOS, you can use the command:
    # open report.html
    ```

    ???+ note "The HTML report looks as follows"
        ![HTML report](images/report.html.png)

=== "Text"
    ```shell
    iter8 report -o text
    ```

    ???+ note "The text report looks as follows."

        ```
        -----------------------------|-----
                   Experiment summary|
        -----------------------------|-----
                Experiment completed |true
        -----------------------------|-----
                   Experiment failed |false
        -----------------------------|-----
           Number of completed tasks |2
        -----------------------------|-----



        -----------------------------|-----
                                 SLOs|
        -----------------------------|-----
             built-in/error-rate <= 0|true
        -----------------------------|-----
                built-in/p95.0 <= 100|true
        -----------------------------|-----


        -----------------------------|-----
                              Metrics|
        -----------------------------|-----
                 built-in/error-count|0
        -----------------------------|-----
                  built-in/error-rate|0
        -----------------------------|-----
                 built-in/max-latency|201.75 (msec)
        -----------------------------|-----
                built-in/mean-latency|17.02 (msec)
        -----------------------------|-----
                 built-in/min-latency|3.80 (msec)
        -----------------------------|-----
                       built-in/p50.0|10.75 (msec)
        -----------------------------|-----
                       built-in/p75.0|12.12 (msec)
        -----------------------------|-----
                       built-in/p90.0|13.88 (msec)
        -----------------------------|-----
                       built-in/p95.0|15.60 (msec)
        -----------------------------|-----
                       built-in/p99.0|201.31 (msec)
        -----------------------------|-----
                       built-in/p99.9|201.71 (msec)
        -----------------------------|-----
               built-in/request-count|100
        -----------------------------|-----
              built-in/stddev-latency|37.81 (msec)
        -----------------------------|-----
        ```

Congratulations! :tada: You completed your first Iter8 experiment.

???+ tip "Customize"
    1.  To load test and validate SLOs for your service, change the URL in `experiment.yaml` to that of your service.
    2.  The [`gen-load-and-collect-metrics` task](../user-guide/tasks/collect.md) used in the experiment can be customized with various inputs including the number of queries sent to the URL, number of queries sent per second, number of parallel connections used, and the payload to be used as part of the queries.
    3.  The SLOs specified as part of the [`assess-app-versions` task](../user-guide/tasks/assess.md#illustrative-example) can be customized, both in terms of the [built-in metrics](../user-guide/tasks/collect.md#built-in-metrics) used and their limits.
