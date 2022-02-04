---
template: main.html
---

# Data and Metadata

!!! tip "Send call data and metadata"
    gRPC calls may include data serialized as [Protocol Buffer messages](https://grpc.io/docs/what-is-grpc/introduction/#working-with-protocol-buffers), and may also include [metadata](https://grpc.io/docs/what-is-grpc/core-concepts/#metadata) which is information about a particular call. 
    
    This tutorial shows how to send data and metadata as part of gRPC load tests.

***

Follow the [introductory tutorial for load testing with SLO validation for gRPC services](unary.md). Modify the `iter8 run` command using one of the following variations.

### Data

The [protobuf file specifying the gRPC service](https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto) defines the following `HelloRequest` message format:
```protobuf
message HelloRequest {
  string name = 1;
}
```

Suppose you want include the following `HelloRequest` message with every call.
```yaml
name: frodo
```

You can do so by running the Iter8 experiment as follows.
```shell
iter8 run --set-string host="127.0.0.1:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set data.name="frodo" \
          --set SLOs.error-rate=0 \
          --set SLOs.latency/mean=50 \
          --set SLOs.latency/p90=100 \
          --set SLOs.latency/p'97\.5'=200
```

??? note "Nested data"
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

### Data URL
Suppose that call data you want to send is contained in a JSON file and hosted at the url https://location.of/data.json. Iter8 can fetch this JSON file and use the data contained in it during the gRPC load test, when you run the experiment as follows.

```shell
iter8 run --set-string host="127.0.0.1:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set dataURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set SLOs.error-rate=0 \
          --set SLOs.latency/mean=50 \
          --set SLOs.latency/p90=100 \
          --set SLOs.latency/p'97\.5'=200
```

### Binary data URL
Suppose that call data you want to send is contained in a JSON file and hosted at the url https://location.of/data.json. Iter8 can fetch this JSON file and use the data contained in it during the gRPC load test, when you run the experiment as follows.

```shell
iter8 run --set-string host="127.0.0.1:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set dataURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set SLOs.error-rate=0 \
          --set SLOs.latency/mean=50 \
          --set SLOs.latency/p90=100 \
          --set SLOs.latency/p'97\.5'=200
```


#### data vs dataURL vs binaryDataURL
If the call data is shallow and has only a few fields, setting the `data` value directly is the easier approach. If it is deep with many fields, storing this as a JSON or binary file, and providing the `dataURL` of `binaryDataURL` location is the easier approach. ... takes precedence over ...

### Metadata
### Metadata URL

#### metadata vs metadataURL
If the metadata is shallow and has only a few fields, setting the `metadata` value directly is the easier approach. If it is deep with many fields, storing this as a JSON file, and providing the `metadataURL` location is the easier approach. ... takes precedence over ...
