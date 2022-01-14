---
template: main.html
---

# Load test a Knative HTTP service

!!! tip "Load test a Knative HTTP Service with a GET endpoint and validate SLOs"
    Use an [Iter8 experiment](../../../../getting-started/concepts.md#what-is-an-iter8-experiment) to load test a [Knative](https://knative.dev/) HTTP service and validate latency and error-related [service level objectives (SLOs)](../../../../user-guide/topics/slos.md).

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
    3. This tutorial combines [this tutorial](../../../../getting-started/your-first-experiment.md) and [this tutorial](../../requests.md) in the context of Knative. Please refer to these tutorials for more details.

## 1. Download experiment chart
Download the `load-test` [experiment chart](../../../../getting-started/concepts.md#experiment-chart) from [Iter8 hub](../../../../user-guide/topics/iter8hub.md) as follows.

```shell
iter8 hub -e load-test
cd load-test
```

## 2. Run experiment
The `iter8 run` command generates the `experiment.yaml` file from an experiment chart, runs the experiment, and writes the results of the experiment into the `result.yaml` file. Run the load test experiment as follows.

```shell
iter8 run --set url=http://hello.default.127.0.0.1.sslip.io
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

## 5. Control request generation
While running a load test, you can set the total number of requests/the duration of the load test, the number of requests sent per second, and the number of parallel connections used to send requests. This provides fine-grained control over the request generation process.

Re-run your experiment by setting the following parameters of the load test experiment.

```shell
iter8 run --set url=http://hello.default.127.0.0.1.sslip.io \
          --set numQueries=200 \
          --set qps=10 \
          --set connections= 5
```

Assert outcomes and view reports as described above.

Congratulations! :tada: You completed your Iter8 experiment with Knative.
