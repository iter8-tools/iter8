---
template: main.html
---

# K8s App/serverless/ML Frameworks

This tutorial provides examples of using the `load-test-http` experiment chart with various Kubernetes app/serverless/ML frameworks. Refer to [`load-test-http` usage](usage.md) to learn more about this chart.

!!! tip "Dear Iter8 community" 

    These examples are maintained by members of the Iter8 community, and may become outdated. If you find that something is not working, lend a helping hand and fix it in a PR. More examples are always welcome.

## Knative

???+ note "Before you begin"
    1. [Install Iter8](../../getting-started/install.md).
    2. [Install Knative and deploy your first Knative Service](https://knative.dev/docs/getting-started/first-service/). As noted at the end of the Knative tutorial, when you curl the Knative service,
    ```shell
    curl http://hello.default.127.0.0.1.sslip.io
    ```
    you should see the expected output as follows.
    ```
    Hello World!
    ```
    3. Download experiment chart.
    ```shell
    iter8 hub -e load-test-http
    cd load-test-http
    ```

### 1. Run experiment
```shell
iter8 run --set url=http://hello.default.127.0.0.1.sslip.io \
          --set SLOs.error-rate=0 \
          --set SLOs.latency-mean=50 \
          --set SLOs.latency-p90=100 \
          --set SLOs.latency-p'97\.5'=200
```

### 2. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

### 3. View report
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
