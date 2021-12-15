---
template: main.html
---

# `gen-load-and-collect-metrics`
The `gen-load-and-collect-metrics` task enables collection of [Iter8's built-in metrics](#built-in-metrics). It generates a stream of HTTP GET or POST requests to one or more app versions, and collects latency and error-related metrics.

## Examples
Generate load and collect built-in metrics for an app.
```yaml
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: https://example.com
```

Customize this task using various inputs.
```yaml
- task: gen-load-and-collect-metrics
  with:
    # Number of requests sent to each version.
    numQueries: 200
    # Duration for which requests are sent to each version. 
    # This field is ignored if `numQueries` is specified.
    duration: 10s
    # Number of queries *per second* sent to each version.
    qps: 10.0
    # Number of parallel connections used to send requests.
    connections: 8
    # String data to be sent as payload. 
    # If this field is specified, Iter8 will send HTTP POST requests.
    # with this string as the payload.
    # This field is ignored if `payloadURL` is specified.
    payloadStr: "abc123"
    # URL of payload. If this field is specified, 
    # Iter8 will send HTTP POST requests to versions with 
    # data downloaded from this URL as the payload.
    payloadURL: "https://www.my.image.com/image.png"
    # The type of the payload. Indicated using the Content-Type HTTP header value. 
    # This is intended to be used in conjunction with one of the `payload*` fields above. 
    # If this field is specified, Iter8 will send HTTP POST requests to versions 
    # with this content type header value.
    contentType: "image/png"
    # A list of error ranges. 
    # Each range specifies an upper and/or lower limit on HTTP status codes. 
    # HTTP responses that fall within these error ranges are considered error.
    # The following setting says that HTTP status codes 500 or above are errors.
    errorRanges:
    - lower: 500
    # A list of latency percentiles computed by this task.
    # Percentiles have single digit precision.
    percentiles: [50.0, 75.0, 90.0, 95.0]
    # Information about app versions.
    versionInfo:
    - # HTTP(S) URL where this version receives GET or POST requests.
      url: https://example.com
      # HTTP headers to be used in requests sent to this version.
      headers:
        "hello": "mars"
        "goodbye": "pluto"
```

Generate load and collect built-in metrics for two versions of an app.
```yaml
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: http://iter8-app.default.svc:8000
    - url: http://iter8-app-candidate.default.svc:8000
```

## Inputs
The following inputs are supported by this task.

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| numQueries | int | Number of requests to be sent to each version. Default value is 100. | No |
| duration | string | Duration of this task. Specified in the [Go duration string format](https://golang.org/pkg/time/#ParseDuration) (example, `5s`). If both `duration` and `numQueries` are specified, then `duration` is ignored. | No |
| qps | float | Number of queries *per second* sent to each version. Default value is 8.0. Setting this to 0 will maximizes query load without any wait time between queries. | No |
| connections | int | Number of parallel connections used to send requests. Default value is 4. | No |
| payloadStr | string | String data to be sent as payload. If this field is specified, Iter8 will send HTTP POST requests to versions with this string as the payload. | No |
| payloadURL | string | URL of payload. If this field is specified, Iter8 will send HTTP POST requests to versions with data downloaded from this URL as the payload. If both `payloadStr` and `payloadURL` are specified, the former is ignored. | No |
| contentType | string | [ContentType](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type) is the type of the payload. Indicated using the Content-Type HTTP header value. This is intended to be used in conjunction with one of the `payload*` fields above. If this field is specified, Iter8 will send HTTP POST requests to versions with this content type header value. | No |
| errorRanges | [][ErrorRange](#error-range) | A list of [error ranges](#error-range). Each range specifies an upper and/or lower limit on HTTP status codes. HTTP responses that fall within these error ranges are considered error. Default value is `{{lower: 400},}`, i.e., HTTP status codes 400 and above are considered as error. | No |
| percentiles | []float64 | Latency percentiles computed by this task. Percentile values have a single digit precision (i.e., rounded to one decimal place). Default value is `{50.0, 75.0, 90.0, 95.0, 99.0, 99.9,}`. | No |
| versionInfo | [][Version](#version) | A non-empty list of [version](#version) values. | Yes |

### Error range
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| lower | int | Lower limit of the error range. | No |
| upper | int | Upper limit of the error range. | No |

### Version
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| url | string | HTTP(S) URL where version receives GET or POST requests. | Yes |
| headers | map[string]string | HTTP headers to be used in requests sent to this version. | No |

## Built-in metrics
The following are the set of metrics collected by this task. All metrics collected by this task have their [backend name](../topics/metrics.md) set to `built-in`.

| Name         | Type | Description |
| ------------ | ----------- | -------- |
| request-count | [Counter](../topics/metrics.md#counter) | Number of requests |
| error-count | [Counter](../topics/metrics.md#counter) | Number of responses that are considered as error. The set of HTTP status codes that are considered as error is configurable using the `errorRanges` [input field](#inputs). By default, status codes 400 and above are considered as error. |
| error-rate | [Gauge](../topics/metrics.md#gauge) | Fraction of responses that are considered as error. |
| mean-latency | [Gauge](../topics/metrics.md#gaugee) | Mean response latency |
| pX, where X is a single precision floating point number (e.g., p95.0) | [Gauge](../topics/metrics.md#gauge) | Xth (e.g., 95.0th) percentile response latency. The set of latency percentiles is configurable using the `percentiles` [input field](#inputs). The default latency percentiles computed are `{50.0, 75.0, 90.0, 95.0, 99.0, 99.9,}`. |

## Number of app versions

Iter8 sets the [number of app versions](../topics/versionnumbering.md) in the experiment as the length of the `versionInfo` input field of this task. If this value equals `n`, the versions are numbered `0, ..., n-1`.