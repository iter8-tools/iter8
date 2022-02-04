---
template: main.html
---

# Load Characteristics

!!! tip "Control the load characteristics during the gRPC load test experiment"
    Control the load characteristics during the gRPC load test experiment by setting the number of requests/duration, the number of requests per second, number of connections to use, and the number of concurrent request workers to use in each connection.

***

Follow the [introductory tutorial for load testing with SLO validation for gRPC services](unary.md). In the step where you run the experiment, replace the `iter8 run` command with either of the following commands.

### Number of requests
Set the total number of requests sent during the load-test (`total`) to 500, the number of requests per second (`rps`) to 25, the number of parallel connections (`connections`) used to 10, and the number of concurrent requests within each connection (`concurrency`) to 50, as follows.

```shell
iter8 run --set-string host="127.0.0.1:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set data.name="frodo" \
          --set SLOs.error-rate=0 \
          --set SLOs.latency/mean=50 \
          --set SLOs.latency/p90=100 \
          --set SLOs.latency/p'97\.5'=200 \
          --set total=500 \
          --set rps=25 \
          --set concurrency=50 \
          --set connections=10
```

### Duration
Set the total duration of the load-test (`loadMaxDuration`) to 20 seconds, the number of requests per second (`rps`) to 25, the number of parallel connections (`connections`) used to 10, and the number of concurrent requests within each connection (`concurrency`) to 50, as follows. The duration value may be any [Go duration string](https://pkg.go.dev/maze.io/x/duration#ParseDuration).

```shell
iter8 run --set-string host="127.0.0.1:50051" \
          --set-string call="helloworld.Greeter.SayHello" \
          --set-string protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto" \
          --set data.name="frodo" \
          --set SLOs.error-rate=0 \
          --set SLOs.latency/mean=50 \
          --set SLOs.latency/p90=100 \
          --set SLOs.latency/p'97\.5'=200 \
          --set loadMaxDuration="20s" \
          --set rps=25 \
          --set concurrency=50 \
          --set connections=10
```

***

When you set `total` and `qps`, the duration of the load test is automatically determined. Similarly, when you set `loadMaxDuration` and `qps`, the number of requests is automatically determined. If you set both `total` and `loadMaxDuration`, the former will be ignored.

