---
template: main.html
---

# Your First Experiment

!!! tip "Load Test an HTTP Service"
    Get started with your first [Iter8 experiment](concepts.md#what-is-an-iter8-experiment) by load testing an HTTP service. 
    
***

## 1. Install Iter8
=== "Brew"
    Install the latest stable release of the Iter8 CLI using `brew` as follows.

    ```shell
    brew tap iter8-tools/iter8
    brew install iter8
    ```
    
=== "Binaries"
    Pre-compiled Iter8 binaries for many platforms are available [here](https://github.com/iter8-tools/iter8/releases). Uncompress the iter8-X-Y.tar.gz archive for your platform, and move the `iter8` binary to any folder in your PATH.

=== "Source"
    Build Iter8 from source as follows. Go `1.17+` is a pre-requisite.
    ```shell
    # you can replace master with a specific tag such as v0.8.29
    export REF=master
    https://github.com/iter8-tools/iter8.git?ref=$REF
    cd iter8
    make install
    ```

=== "Go 1.17+"
    Install the latest stable release of the Iter8 CLI using `go 1.17+` as follows.

    ```shell
    go install github.com/iter8-tools/iter8@latest
    ```
    You can now run `iter8` (from your gopath bin/ directory)

## 2. Download experiment chart
Download the `load-test-http` [experiment chart](concepts.md#experiment-chart) from [Iter8 hub](concepts.md#iter8-hub) as follows.

```shell
iter8 hub -e load-test-http
cd load-test-http
```

## 3. Run experiment
We will load test and validate the HTTP service whose URL (`url`) is https://example.com. For validation of SLOs, we will specify that the error rate (`SLOs.error-rate`) must be 0, the mean latency (`SLOs.latency-mean`) must be under 50 msec, the 90th percentile latency (`SLOs.latency-p90`) must be under 100 msec, and the 97.5th percentile latency (`SLOs.latency-p'97\.5'`) must be under 200 msec. 

The `iter8 run` command combines an experiment chart with the supplied values to generate the `experiment.yaml` file, runs the experiment, and writes results into the `result.yaml` file. Run the experiment as follows.

```shell
iter8 run --set url=https://httpbin.org/get
```


??? note "Sample output from `iter8 run`"

    ```shell
    INFO[2021-12-14 10:23:26] starting experiment run                      
    INFO[2021-12-14 10:23:26] task 1: gen-load-and-collect-metrics-http : started 
    INFO[2021-12-14 10:23:39] task 1: gen-load-and-collect-metrics-http : completed 
    INFO[2021-12-14 10:23:39] task 2: assess-app-versions : started        
    INFO[2021-12-14 10:23:39] task 2: assess-app-versions : completed      
    INFO[2021-12-14 10:23:39] experiment completed successfully    
    ```

??? note "Iter8 and Helm"
    If you are familiar with [Helm](https://helm.sh), you probably noticed that the `load-test-http` folder resembles a Helm chart. This is because, Iter8 experiment charts *are* Helm charts under the covers. The [`iter8 run` command](../user-guide/commands/iter8_run.md) used above combines the experiment chart with values to generate the `experiments.yaml` file, much like how Helm charts can be combined with values to produce Kubernetes manifests.

## 4. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

The `iter8 assert` subcommand asserts if experiment result satisfies the specified conditions. 
If assert conditions are satisfied, it exits with code `0`, and exits with code `1` otherwise. Assertions are especially useful within CI/CD/GitOps pipelines.

??? note "Sample output from `iter8 assert`"

    ```shell
    INFO[2021-11-10 09:33:12] experiment completed
    INFO[2021-11-10 09:33:12] experiment has no failure                    
    INFO[2021-11-10 09:33:12] SLOs are satisfied                           
    INFO[2021-11-10 09:33:12] all conditions were satisfied
    ```

## 5. View report
View a report of the experiment in HTML or text formats as follows.

=== "HTML"
    ```shell
    iter8 report -o html > report.html
    # open report.html with a browser. In MacOS, you can use the command:
    # open report.html
    ```

    ??? note "The HTML report looks like this"
        ![HTML report](images/report.html.png)

=== "Text"
    ```shell
    iter8 report
    ```

    ??? note "The text report looks like this"
        ```shell
        Experiment summary:
        *******************

          Experiment completed: true
          No task failures: true
          Total number of tasks: 1
          Number of completed tasks: 1

        Latest observed values for metrics:
        ***********************************

          Metric                              |value
          -------                             |-----
          built-in/http-error-count           |0.00
          built-in/http-error-rate            |0.00
          built-in/http-latency-max (msec)    |203.78
          built-in/http-latency-mean (msec)   |17.00
          built-in/http-latency-min (msec)    |4.20
          built-in/http-latency-p50 (msec)    |10.67
          built-in/http-latency-p75 (msec)    |12.33
          built-in/http-latency-p90 (msec)    |14.00
          built-in/http-latency-p95 (msec)    |15.67
          built-in/http-latency-p99 (msec)    |202.84
          built-in/http-latency-p99.9 (msec)  |203.69
          built-in/http-latency-stddev (msec) |37.94
          built-in/http-request-count         |100.00
        ```

Congratulations! :tada: You completed your first Iter8 experiment.

***

???+ tip "Next steps"

    1. Learn more about [load testing HTTP services with service-level objectives (SLOs)](../tutorials/load-test-http/usage.md).
    2. Learn more about [load testing gRPC services with service-level objectives (SLOs)](../tutorials/load-test-grpc/usage.md).