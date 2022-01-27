# Load Test with SLOs

Use this Iter8 experiment chart to load test an HTTP service and validate latency and error-related service level objectives (SLOs).

***

## Examples

The following `iter8 run` command will load test and validate the HTTP service whose URL is https://example.com. The command specifies that error rate must be 0, the mean latency must be under 50 msec, the 90th percentile latency must be under 100 msec, and the 97.5th percentile latency must be under 200 msec.

```shell
iter8 run --set url=https://example.com \
          --set SLOs.error-rate=0 \
          --set SLOs.mean-latency=50 \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200
```

***

Set the number of requests sent during the load-test to 200, the number of requests per second to 10, and the number of parallel connections used to send the requests to 5, as follows.

```shell
iter8 run --set url=https://example.com \
          --set SLOs.error-rate=0 \
          --set SLOs.mean-latency=50 \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200 \
          --set numQueries=200 \
          --set qps=10 \
          --set connections=5
```

***

Supply a string as payload. This sets the content type to application/octet-stream.

```shell
iter8 run --set url=http://127.0.0.1/post \
          --set payloadStr="abc123" \
          --set SLOs.error-rate=0 \
          --set SLOs.mean-latency=50 \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200
```

***

Fetch JSON content from a payload URL. Use this JSON as payload and explicitly set the content type to application/json.

```shell
iter8 run --set url=http://127.0.0.1/post \
          --set payloadURL=https://httpbin.org/stream/1 \
          --set contentType="application/json" \
          --set SLOs.error-rate=0 \
          --set SLOs.mean-latency=50 \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200
```

***

By default, the load test experiment collects the following built-in metrics: `error-count`, `error-rate`, `mean-latency`, and latency percentiles in the list `[50.0, 75.0, 90.0, 95.0, 99.0, 99.9]`. In addition, any other latency percentiles that are specified as part of SLOs are also collected. 

Consider the following command.
```shell
iter8 run --set url=https://example.com \
          --set SLOs.error-rate=0 \
          --set SLOs.mean-latency=50 \
          --set 'percentiles={25.0, 75.0}' \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200
```

The above command ensures the following.

1. The following latency percentiles are collected and reported: `[25.0, 50.0, 75.0, 90.0, 95.0, 97.5, 99.0, 99.9]`.
2. The following SLOs are validated.
  * error rate is 0
  * mean latency is under 50 msec
  * 90th percentile latency is under 100 msec
  * 97.5th percentile latency is under 200 msec

