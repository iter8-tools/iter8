---
template: main.html
---

# Your First Experiment

!!! tip "Load test https://example.com and validate SLOs" 
    Use an [Iter8 experiment](concepts.md#what-is-an-iter8-experiment) to load test https://example.com and validate latency and error-related service level objectives (SLOs).

## 1. Install Iter8
Install Iter8 using [Go 1.16+](https://golang.org/) as follows.
```shell
go install github.com/iter8-tools/iter8@latest
# you can now run iter8 (from your gopath bin/ directory)
```

## 2. Download experiment
Download the `load-test` experiment folder from [Iter8 hub](../user-guide/topics/iter8hub.md) as follows.

```shell
iter8 hub -e load-test
```

## 3. Run experiment
[Iter8 experiments](concepts.md#what-is-an-iter8-experiment) are specified using the `experiment.yaml` file. The `iter8 run` command reads this file, runs the specified experiment, and writes the results of the experiment into the `result.yaml` file.

Run the experiment you downloaded above as follows.

```shell
cd load-test
iter8 run
```

??? note "Look inside experiment.yaml"
    This experiment contains the [`gen-load-and-collect-metrics` task](../user-guide/tasks/collect.md) for generating load and collecting metrics, and the [`assess-app-versions` task](../user-guide/tasks/assess.md) for validating SLOs.

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

??? note "Sample output from `iter8 run`"

    ```shell
    INFO[2021-12-14 10:23:26] starting experiment run                      
    INFO[2021-12-14 10:23:26] task 1: gen-load-and-collect-metrics : started 
    INFO[2021-12-14 10:23:39] task 1: gen-load-and-collect-metrics : completed 
    INFO[2021-12-14 10:23:39] task 2: assess-app-versions : started        
    INFO[2021-12-14 10:23:39] task 2: assess-app-versions : completed      
    INFO[2021-12-14 10:23:39] experiment completed successfully    
    ```

## 4. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

??? note "Sample output from `iter8 assert`"

    ```shell
    INFO[2021-11-10 09:33:12] experiment completed
    INFO[2021-11-10 09:33:12] experiment has no failure                    
    INFO[2021-11-10 09:33:12] SLOs are satisfied                           
    INFO[2021-11-10 09:33:12] all conditions were satisfied
    ```

## 5. Generate report
Generate a report of the experiment in HTML or text formats as follows.

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
           Number of completed tasks |4
        -----------------------------|-----



        -----------------------------|-----
                                 SLOs|
        -----------------------------|-----
                  built-in/error-rate|true
        -----------------------------|-----
                built-in/p95.0 (msec)|true
        -----------------------------|-----


          -----------------------------|-----
                                Metrics|
          -----------------------------|-----
                   built-in/error-count|0
          -----------------------------|-----
                    built-in/error-rate|0
          -----------------------------|-----
                built-in/latency (msec)|3
          -----------------------------|-----
            built-in/max-latency (msec)|213.21
          -----------------------------|-----
           built-in/mean-latency (msec)|17.47
          -----------------------------|-----
            built-in/min-latency (msec)|4.30
          -----------------------------|-----
                  built-in/p50.0 (msec)|10.80
          -----------------------------|-----
                  built-in/p75.0 (msec)|12.40
          -----------------------------|-----
                  built-in/p90.0 (msec)|13.60
          -----------------------------|-----
                  built-in/p95.0 (msec)|14
          -----------------------------|-----
                  built-in/p99.0 (msec)|209.91
          -----------------------------|-----
                  built-in/p99.9 (msec)|212.88
          -----------------------------|-----
                 built-in/request-count|100
          -----------------------------|-----
         built-in/stddev-latency (msec)|39.90
          -----------------------------|-----
        ```

Congratulations! :tada: You completed your first Iter8 experiment.
