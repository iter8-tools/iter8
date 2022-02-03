# Load test gRPC services with SLO validation

Iter8 experiment chart that enables load testing and validation of latency and error-related service level objectives (SLOs) for gRPC services.
***

## Examples

The following `iter8 run` command will load test and validate the gRPC sample service with host `127.0.0.1:50051`, fully-qualified method name `helloworld.Greeter.SayHello`, and defined by the Protocol Buffer file located at the `protoURL`. The gRPC requests made by the Iter8 experiment will include `{'name': 'frodo'}` as the data, serialized in the protobuf format. The command also specifies that the error rate must be 0, the mean latency must be under 50 msec, the 90th percentile latency must be under 100 msec, and the 97.5th percentile latency must be under 200 msec.

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
