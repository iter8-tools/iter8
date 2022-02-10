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
We will load test the HTTP service whose URL (`url`) is https://httpbin.org/get. 

The `iter8 run` command combines an experiment chart with values, generates the `experiment.yaml` file, runs the experiment, and writes results into the `result.yaml` file. Run the experiment as follows.

```shell
iter8 run --set url=https://httpbin.org/get
```

## 4. View report
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