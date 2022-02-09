---
template: main.html
---

# App/serverless/ML Frameworks

This tutorial provides examples of using the `load-test-grpc` experiment chart with various Kubernetes app/serverless/ML frameworks. Refer to [`load-test-grpc` usage](usage.md) to learn more about this chart.

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
    3. Update the Knative service deployed above to a gRPC service as follows.
    ```shell
    kn service update hello \
    --image docker.io/grpc/java-example-hostname:latest \
    --port 50051 \
    --revision-name=grpc
    ```
    4. Download experiment chart.
    ```shell
    iter8 hub -e load-test-grpc
    cd load-test-grpc
    ```

### 1. Run experiment
```shell
iter8 run --set-string host="hello.default.127.0.0.1.sslip.io:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-java/master/examples/example-hostname/src/main/proto/helloworld/helloworld.proto"
          --set data.name="frodo" \
          --set SLOs.error-rate=0 \
          --set SLOs.latency/mean=400 \
          --set SLOs.latency/p90=500 \
          --set SLOs.latency/p'97\.5'=600
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
