# Run an Iter8 Load Test Experiment

This action runs an Iter8 [load test experiment](https://iter8.tools/0.8/tutorials/load-test/overview/). For more details about Iter8 experiments, see <https://iter8.tools>.

## Usage

Load test and validate the the HTTP service whose URL is <https://example.com>. We will specify that the error rate must be 0, the mean latency must be under 50 msec, the 90th percentile latency must be under 100 msec, and the 97.5th percentile latency must be under 200 msec.

``` yaml
- uses: iter8-tools/load-test@v1
  with:
    url: http://example.com
    error_rate: 0
    mean_latency: 50
    p90: 100
    p97_5: 200
```

Set the number of requests sent during the load-test to 200, the number of requests per second to 10, and the number of parallel connections used to send the requests to 5:

``` yaml
- uses: iter8-tools/load-test@v1
  with:
    url: http://example.com
    error_rate: 0
    mean_latency: 50
    p90: 100
    p97_5: 200
    numQueries: 200
    qps: 10
    connections: 5
```

Set the duration of the load test to 20 seconds:

``` yaml
- uses: iter8-tools/load-test@v1
  with:
    url: http://example.com
    error_rate: 0
    mean_latency: 50
    p90: 100
    p97_5: 200
    duration: 20s
    qps: 10
    connections: 5
```

## Inputs

| Input Name | Description | Default |
| ---------- | ----------- | ------- |
| `url` | HTTP(S) URL where the app receives GET or POST requests. | Must be specified |
| `numQueries` | 'Number of requests sent to the app. | 100 |
| `duration` | Duration for which requests are sent to the app. Value can be any [Go duration string](https://pkg.go.dev/maze.io/x/duration#ParseDuration). Ignored if `numQueries` is specified. | None |
| `qps` | Number of requests per second sent to each version. | 8.0 |
| `connections` | Number of parallel connections used to send requests. | 4 |
| `payloadStr` | String data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests with this string as the payload. This field is ignored if `payloadUrl` is specified. | None |
| `payloadUrl` | URL of payload. If this field is specified, Iter8 will send HTTP POST requests to versions with data downloaded from this URL as the payload. | None |
| `contentType` | The type of the payload. Indicated using the Content-Type HTTP header value. This is intended to be used in conjunction with one of the `payload*` fields above. If this field is specified, Iter8 will send HTTP POST requests to versions with this content type header value. | `application/octet-stream` if one of `payloadStr` or `payloadURL` is specified. Otherwise none. |
| `error_rate` | Maximum acceptable error rate | None |
| `mean_latency` | Maximum acceptable mean latency | None |
| `p25` | Maximum acceptable 25th percentile latency | None |
| `p50` | Maximum acceptable 50th percentile latency | None |
| `p75` | Maximum acceptable 75th percentile latency | None |
| `p90` | Maximum acceptable 90th percentile latency | None |
| `p95` | Maximum acceptable 95th percentile latency | None |
| `p97_5` | Maximum acceptable 97.5th percentile latency | None |
| `p99` | Maximum acceptable 99th percentile latency | None |
| `p99_9` | Maximum acceptable 99.9th percentile latency | None |
| `values` | Path to file of configuration values. | `values.yaml` |
