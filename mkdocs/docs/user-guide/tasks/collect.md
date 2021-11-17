---
template: main.html
---

# `gen-load-and-collect-metrics`
The `gen-load-and-collect-metrics` task enables collection of [Iter8's built-in metrics](#built-in-metrics). It generates a stream of HTTP GET or POST requests to one or more app versions, and collects latency and error related metrics.

## Illustrative examples
Generate load and collect built-in metrics for a single version of an app.
```yaml
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: https://example.com
```

Generate load and collect built-in metrics for two app versions.
```yaml
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: http://iter8-app.default.svc:8000
    - url: http://iter8-app-candidate.default.svc:8000
```

Set the number of app versions to 2. Generate load and collect built-in metrics for the second version only.
```yaml
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - # set to null
    - url: http://iter8-app-candidate.default.svc:8000
```

## Inputs
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| numQueries | int | number of requests to be sent to each version. Default value is 100. | No |
| time | string | Duration of the `metrics/collect` task run. Specified in the [Go duration string format](https://golang.org/pkg/time/#ParseDuration) (example, `5s`). If both `time` and `numQueries` are specified, then `time` is ignored. | No |
| qps | float | Number of queries *per second* sent to each version. Default is 8.0. Setting this to 0 will maximizes query load without any wait time between queries. | No |
| connections | int | Number of parallel connection used for sending queries. Default is 4. | No |
| loadOnly | bool | If set to true, this task will send requests without collecting metrics. Default value is `false`. | No |
| payloadStr | string | String data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests to versions using data as the payload. | No |
| payloadURL | string | URL of payload. If this field is specified, Iter8 will send HTTP POST requests to versions using data downloaded from this URL as the payload. If both `payloadStr` and `payloadURL` is specified, the former is ignored. | No |
| contentType | string | [Content type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type) HTTP header value. This is intended to be used in conjunction with one of the `payload*` fields above. If this field is specified, Iter8 will send HTTP POST requests to versions using this content type header value.
| versions | [][Version](#version) | A non-empty list of versions. | Yes |

### Version
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| url | string | HTTP(S) URL where version receives GET or POST requests. | Yes |
| headers | map[string]string | HTTP headers to be used in requests sent to this version. | No |

## Built-in metrics
The following are the set of built-in Iter8 metrics.

| Namespace | Name         | Type | Description |
| ----- | ------------ | ----------- | -------- |
| iter8-system | request-count | Counter | Number of requests |
| iter8-system | error-count | Gauge | Number of responses with HTTP status code 4xx or 5xx |
| iter8-system | error-rate | Gauge | Fraction of responses with HTTP status code 4xx or 5xx |
| iter8-system | mean-latency | Gauge | Mean response latency |
| iter8-system | latency-50th-percentile | Gauge | 50th percentile (median) response latency |
| iter8-system | latency-75th-percentile | Gauge | 75th percentile response latency |
| iter8-system | latency-90th-percentile | Gauge | 90th percentile response latency |
| iter8-system | latency-95th-percentile | Gauge | 95th percentile response latency |
| iter8-system | latency-99th-percentile | Gauge | 99th percentile response latency |

## Number of app versions

Iter8 sets the number of app versions in the experiment as the length of the `versionInfo` input field of this task. If this equals `n`, the versions are numbered `0, ..., n-1`.