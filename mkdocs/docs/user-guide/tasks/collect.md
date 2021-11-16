---
template: main.html
---

# `gen-load-and-collect-metrics`
The `metrics/collect` task enables collection of [built-in metrics](../../metrics/builtin.md). It generates a stream of HTTP requests to one or more app versions, with payload (optional), and collects latency and error metrics.

## Example

The following start action contains a `metrics/collect` task which is executed at the start of the experiment. The task sends a certain number of HTTP requests to each version specified in the task, and collects built-in latency/error metrics for them.

```yaml
start:
- task: metrics/collect
  with:
    versions:
    - name: iter8-app
      url: http://iter8-app.default.svc:8000
    - name: iter8-app-candidate
      url: http://iter8-app-candidate.default.svc:8000
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
| name | string | Name of the version. Version names must be unique. If the version name matches the name of a version in the experiment's `versionInfo` section, then the version is considered *real*. If the version name does not match the name of a version in the experiment's `versionInfo` section, then the version is considered *pseudo*. Built-in metrics collected for real versions can be used within the experiment's `criteria` section. Pseudo versions are useful if the intent is only to generate load (`GET` and `POST` requests). Built-in metrics collected for pseudo versions cannot be used with the experiment's `criteria` section. | Yes |
| headers | map[string]string | Additional HTTP headers to be used in requests sent to this version. | No |
| url | string | HTTP URL of this version. | Yes |


## Result

This task will run for the specified duration (`time`), send requests to each version (`versions`) at the specified rate (`qps`), and will collect [built-in metrics]() for each version. Built-in metric values are stored in the metrics field of the experiment status in the same manner as custom metric values.

The task may result in an error, for instance, if one or more required fields are missing or if URLs are mis-specified. In this case, the experiment to which it belongs will fail.

## Start vs loop actions
If this task is embedded in start actions, it will run once at the beginning of the experiment.

If this task is embedded in loop actions, it will run in each loop of the experiment. The results from each run will be aggregated.

## Load generation without metrics collection
You can use this task to send HTTP GET and POST requests to app versions without collecting metrics by setting the [`loadOnly` input](#inputs) to `true`.