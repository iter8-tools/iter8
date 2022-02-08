---
template: main.html
---

# Load Test gRPC Services with SLOs

!!! tip "Overview"
    Iter8's `load-test-grpc` experiment chart can be used to generate call requests for gRPC services, collect built-in latency and error-related metrics, and validate service-level objectives (SLOs).

    <p align='center'>
      <img alt-text="load-test-http" src="../images/grpc-overview.png" width="75%" />
    </p>

    **Use-cases:** Rapid testing, validation, safe rollouts, and continuous delivery (CD) of gRPC services are the motivating use-cases for this experiment type. If the gRPC service satisfies the SLOs, it may be safely rolled out, for example, from a test environment to a production environment.  

***

[Your first experiment](../../getting-started/your-first-experiment.md) provides a basic example of using the `load-test-http` experiment chart. This tutorial provides additional examples.

???+ warning "Before you try these examples"
    1. [Install Iter8](../../getting-started/install.md).
    2. Choose a language, and follow the linked instructions to run the gRPC sample app.

        !!! warning "Update step is not needed" 
            The linked instructions show how to update the app, and re-run the updated app. For the purpose of this tutorial, there is **no need** to update and re-run. Running the basic service is sufficient.

        === "C#"
            [Run the C# gRPC app](https://grpc.io/docs/languages/csharp/quickstart/#run-a-grpc-application).

        === "C++"
            [Run the C++ gRPC app](https://grpc.io/docs/languages/cpp/quickstart/#try-it).

        === "Dart"
            [Run the Dart gRPC app](https://grpc.io/docs/languages/dart/quickstart/#run-the-example).

        === "Go"
            [Run the Go gRPC app](https://grpc.io/docs/languages/go/quickstart/#run-the-example).

        === "Java"
            [Run the Java gRPC app](https://grpc.io/docs/languages/java/quickstart/#run-the-example).

        === "Kotlin"
            [Run the Kotlin gRPC app](https://grpc.io/docs/languages/kotlin/quickstart/#run-the-example).

        === "Node"
            [Run the Node gRPC app](https://grpc.io/docs/languages/node/quickstart/#run-a-grpc-application).

        === "Objective-C"
            [Run the Objective-C gRPC app](https://grpc.io/docs/languages/objective-c/quickstart/#run-the-server).

        === "PHP"
            [Run the PHP gRPC app](https://grpc.io/docs/languages/php/quickstart/#run-the-example).

        === "Python"
            [Run the Python gRPC app](https://grpc.io/docs/languages/python/quickstart/#run-a-grpc-application).

        === "Ruby"
            [Run the Ruby gRPC app](https://grpc.io/docs/languages/ruby/quickstart/#run-a-grpc-application).

    3. Download experiment chart.
    ```shell
    iter8 hub -e load-test-grpc
    cd load-test-grpc
    ```

***

## Basic example
Load test the gRPC sample service with `host` value `127.0.0.1:50051`, fully-qualified method name (`call`) `helloworld.Greeter.SayHello`, and defined by the Protocol Buffer file located at the `protoURL`.

```shell
iter8 run --set-string host="127.0.0.1:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto"
```

View a report of this experiment as described in the [quick start tutorial](../../getting-started/your-first-experiment.md).

## Load profile
Control the characteristics of the load generated by the `load-test-grpc` experiment by  setting the number of requests (`total`)/duration (`duration`), the number of requests per second (`rps`), number of connections to use (`connections`), and the number of concurrent request workers to use in each connection (`concurrency`).


=== "Number of requests"
    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set data.name="frodo" \
              --set total=500 \
              --set rps=25 \
              --set concurrency=50 \
              --set connections=10
    ```

=== "Duration"
    The duration value may be any [Go duration string](https://pkg.go.dev/maze.io/x/duration#ParseDuration).

    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set data.name="frodo" \
              --set duration="20s" \
              --set rps=25 \
              --set concurrency=50 \
              --set connections=10
    ```

When you set `total` and `qps`, the duration of the load test is automatically determined. Similarly, when you set `duration` and `qps`, the number of requests is automatically determined. If you set both `total` and `duration`, the former will be ignored.

***

## Call data
gRPC calls may include data serialized as [Protocol Buffer messages](https://grpc.io/docs/what-is-grpc/introduction/#working-with-protocol-buffers). Supply them as values, or by pointing to JSON or binary files containing the data.

=== "Data"
    The [protobuf file specifying the gRPC service](https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto) used in this tutorial defines the following `HelloRequest` message format:
    ```protobuf
    message HelloRequest {
      string name = 1;
    }
    ```

    Suppose you want include the following `HelloRequest` message with every call.
    ```yaml
    name: frodo
    ```

    To do so, run the Iter8 experiment as follows.
    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set data.name="frodo"
    ```

    ???+ note "Nested data"
        Call data may be nested. For example, consider the data:
        ```yaml
        name: frodo
        realm:
          planet: earth
          location: middle
        ```
        You can set the above data in `iter8 run` as follows:
        ```shell
        --set data.name="frodo" \
        --set data.realm.planet="earth" \
        --set data.realm.location="middle" 
        ```

=== "Data URL"
    Suppose the call data you want to send is contained in a JSON file and hosted at the url https://location.of/data.json. Iter8 can fetch this JSON file and use the data contained in it during the gRPC load test. To do so, run the experiment as follows.

    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set dataURL="https://location.of/data.json"
    ```

=== "Binary data URL"
    Suppose that call data you want to send is contained in a binary file as a serialized binary message or multiple count-prefixed messages, and hosted at the url https://location.of/data.bin. Iter8 can fetch this binary file and use the data contained in it during the gRPC load test. To do so, run the experiment as follows.

    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set binaryDataURL="https://location.of/data.bin"
    ```

### data vs dataURL vs binaryDataURL
If the call data is shallow and has only a few fields, setting the `data` value directly is the easiest of the three approaches. If it is deeply nested, or contains many fields, storing the data as a JSON or binary file, and providing the `dataURL` or `binaryDataURL` value might be the easier approach. When more than one of these options are specified, the `data` field takes precedence over the `dataURL` field which in turn takes precedence over the `binaryDataURL` field.

***

## Call metadata
gRPC calls may include [metadata](https://grpc.io/docs/what-is-grpc/core-concepts/#metadata) which is information about a particular call. Supply them as values, or by pointing to a JSON file containing the metadata.

=== "Metadata"
    You can supply metadata of type `map[string]string` (i.e., a map whose keys and values are strings) in the `gRPC` load test. Suppose you want to use the following metadata.
    ```yaml
    darth: vader
    lord: sauron
    volde: mort
    ```

    To do so, run the Iter8 experiment as follows.
    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set metadata.darth="vader" \
              --set metadata.lord="sauron" \
              --set metadata.volde="mort"
    ```

=== "Metadata URL"
    Suppose the call metadata you want to send is contained in a JSON file and hosted at the url https://location.of/metadata.json. Iter8 can fetch this JSON file and use its contents as the metadata during the gRPC load test. To do so, run the experiment as follows.

    ```shell
    iter8 run --set-string host="127.0.0.1:50051" \
              --set-string call="helloworld.Greeter.SayHello" \
              --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
              --set metadataURL="https://location.of/metadata.json"
    ```

### metadata vs metadataURL
If the call metadata is shallow and has only a few fields, setting the `metadata` value directly is the easier approach. If it is deeply nested, or contains many fields, storing the data as a JSON binary file, and providing the `metadataURL` value might be the easier approach. When both these options are specified, the `metadata` field takes precedence over the `metadataURL` field.