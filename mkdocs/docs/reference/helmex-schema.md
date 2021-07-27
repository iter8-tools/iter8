---
template: main.html
---

# Helmex Schema

The Helmex schema for the Helm `values.yaml` file is described below. It is intended for applications that are templated using Helm and use Iter8 experiments during releases. In addition to the below requirements, an application may impose additional application-specific schema requirements on `values.yaml`.

## Top-level fields
| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| baseline | [Version](#version) | Information about the baseline version of the application. The baseline version typically corresponds to the stable version of the application. | Yes |
| candidate | [Version](#version) | Information about the candidate version. Iter8 experiment resource will be created only if this field is present. If this field is modified, any existing experiment for the application will be replaced by a new experiment. | No |

### Version

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| dynamic | [Dynamic](#dynamic) | Information associated with a specific version. For example, each time the baseline version of the application changes, the `baseline.dynamic` field in the Helm values file should change. | Yes |

#### Dynamic
| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| id | string | Alpha-numeric string that uniquely identifies a version. This optional field is strongly recommended for every version. | No. |

## Example
The following Helm values file is an instance of the Helmex schema.

```yaml
# values meant for both baseline and candidate versions of the application;
common:
  application: hello
  repo: "gcr.io/google-samples/hello-app"
  serviceType: ClusterIP
  servicePortInfo:
    port: 8080
  regularLabels:
    app.kubernetes.io/managed-by: Iter8
  selectorLabels:
    app.kubernetes.io/name: hello

# values meant for baseline version of the application only;
# baseline version is required by Helmex schema
baseline:
  name: hello
  selectorLabels:
    app.kubernetes.io/track: baseline
  # required field for baseline version
  dynamic:
    # unique alpha-numeric version ID is strongly recommended
    id: "mn82l82"
    tag: "1.0"

# values meant for candidate version of the application only;
# optional section; Iter8 experiment will be deployed if this section is present
candidate:
  name: hello-candidate
  selectorLabels:
    app.kubernetes.io/track: candidate
  # required field for candidate version
  # if candidate is promoted, the dynamic field from candidate will be copied over to baseline, and candidate will be set to null
  dynamic:
    # unique alpha-numeric version ID is strongly recommended
    id: "8s72oa"
    tag: "2.0"

# this section is used in the creation of the Iter8 experiment
# the specific experiment section below is used in the context of an SLO validation experiment
experiment:
  # The SLO validation experiment will collect Iter8's built-in latency and error metrics.
  # There will be 8.0 * 5 = 40 queries sent during metrics collection.
  # time is the duration over which queries are sent during metrics collection.
  time: 5s
  # QPS is number of queries per second sent during metrics collection.
  QPS: 8.0
  # (msec) acceptable limit for mean latency of the application
  limitMeanLatency: 500.0
  # (msec) acceptable error rate for the application (1%)
  limitErrorRate: 0.01 
  # (msec) acceptable limit for 95th percentile latency of the application
  limit95thPercentileLatency: 1000.0
```
