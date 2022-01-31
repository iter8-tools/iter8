---
template: main.html
---

# Load test a Knative HTTP service

!!! tip "Load test a Knative HTTP Service and validate SLOs"
    Use an [Iter8 experiment](../../../../getting-started/concepts.md#what-is-an-iter8-experiment) to load test a [Knative](https://knative.dev/) HTTP service and validate latency and error-related service level objectives (SLOs).

???+ note "Before you begin"
    1. [Install Iter8](../../../../getting-started/install.md).
    2. [Install Knative and deploy your first Knative Service](https://knative.dev/docs/getting-started/first-service/). As noted at the end of the Knative tutorial, when you curl the Knative service,
    ```shell
    curl http://hello.default.127.0.0.1.sslip.io
    ```
    you should see the expected output as follows.
    ```
    Hello World!
    ```
    3. Complete the [Iter8 quick start tutorial](../../../../getting-started/your-first-experiment.md).


## 1. Download experiment chart
Download the `load-test-http` [experiment chart](../../../../getting-started/concepts.md#experiment-chart) from [Iter8 hub](../../../../user-guide/topics/iter8hub.md) as follows.

```shell
iter8 hub -e load-test-http
cd load-test-http
```

## 2. Run experiment
The `iter8 run` command combines an experiment chart with the supplied values to generate the `experiment.yaml` file, runs the experiment, and writes results into the `result.yaml` file.

```shell
iter8 run --set url=http://hello.default.127.0.0.1.sslip.io \
          --set SLOs.error-rate=0 \
          --set SLOs.latency-mean=50 \
          --set SLOs.latency-p90=100 \
          --set SLOs.latency-p'97\.5'=200
```

## 3. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

## 4. View report
View a report of the experiment in HTML or text formats as follows.

=== "HTML"
    ```shell
    iter8 report -o html > report.html
    # open report.html with a browser. In MacOS, you can use the command:
    # open report.html
    ```

=== "Text"
    ```shell
    iter8 report -o text
    ```

Congratulations! :tada: You completed your Iter8-Knative experiment.

***

???+ tip "Useful variations of this experiment"

    1. [Control the load characteristics during the HTTP load test experiment](../../loadcharacteristics.md) by setting the number of queries/duration, the number of queries sent per second, and the number of parallel connections used to send requests.

    2. HTTP services with POST endpoints may accept payloads. [Send various types of content as payload](../../payload.md) during the load test.

    3. [Learn more about the built-in metrics that are collected and the SLOs that are validated during the load test](../../metricsandslos.md).
    
    4. The `values.yaml` file in the experiment chart folder documents all the values that can be supplied during the experiment.

