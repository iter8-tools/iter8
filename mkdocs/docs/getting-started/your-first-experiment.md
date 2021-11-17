---
template: main.html
---

# Your First Experiment

!!! tip "Load test https://example.com"
    Use an Iter8 experiment to load test https://example.com and validate error and latency related service level objectives (SLOs).

## 1. [Install Iter8](install.md)

## 2. Download experiment
[Iter8 Hub](../README.md) enables users to share, find, and download Iter8 experiment samples. Download the `load-test` experiment sample from the hub as follows.

```shell
iter8 hub -e load-test
```

## 3. Launch experiment
```shell
cd load-test
iter8 run
```

The above command reads in the experiment specified in the `experiment.yaml` file, runs the experiment, and writes the results of the experiment into the `results.yaml` file.

??? note "Look inside experiment.yaml"
    ```yaml
    # task 1: generate HTTP requests for https://example.com and
    # collect Iter8's built-in latency and error related metrics
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
The experiment should complete in a few seconds. Upon completion, assert that the experiment completed without any failures, and SLOs are satisfied, as follows.

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

```shell
iter8 gen 
```

??? note "Look inside a sample report"
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
         built-in/max-latency|200.45 (msec)
    -----------------------------|-----
        built-in/mean-latency|18.77 (msec)
    -----------------------------|-----
         built-in/min-latency|4.76 (msec)
    -----------------------------|-----
               built-in/p50.0|11.74 (msec)
    -----------------------------|-----
               built-in/p75.0|13.29 (msec)
    -----------------------------|-----
               built-in/p90.0|15.60 (msec)
    -----------------------------|-----
               built-in/p95.0|25 (msec)
    -----------------------------|-----
               built-in/p99.0|200.34 (msec)
    -----------------------------|-----
               built-in/p99.9|200.44 (msec)
    -----------------------------|-----
       built-in/request-count|100
    -----------------------------|-----
      built-in/stddev-latency|37.23 (msec)
    -----------------------------|-----
    ```

Congratulations :tada: You completed your first Iter8 experiment.
