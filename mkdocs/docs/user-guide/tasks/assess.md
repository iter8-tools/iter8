---
template: main.html
---

# `assess-app-versions`
The `assess-app-versions` assesses if app versions satisfy service level objectives (SLOs). SLOs are specified as inputs to the task in the form of metrics, and acceptable upper and lower limits on the metric values.

This task is intended to be preceded by the [`gen-load-and-collect-metrics` task](collect.md). The latter task collects metrics for app versions, while the former task performs version assessments based on metrics.

## Illustrative example
Validate service level objectives (SLOs) for app versions  based on [Iter8's builtin metrics](collect.md).

```yaml
- task: assess-app-versions
  with:
    SLOs:
      # error rate must be 0
    - metric: iter8-fortio/error-rate
      upperLimit: 0
      # mean latency must be under 50 msec
    - metric: iter8-fortio/mean-latency
      upperLimit: 50
      # 95th percentile latency must be under 100 msec
    - metric: iter8-fortio/p95.0
      upperLimit: 100
```

## Inputs
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| SLOs | [][SLO](#slo) | A list of [service level objectives (SLOs)](#slo) | No |

### SLO
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| metric | string | Fully-qualified metric name, in the `backend-name/metric-name` format. | Yes |
| upperLimit | float64 | Acceptable upper limit on the value of the metric. | No |
| lowerLimit | float64 | Acceptable lower limit on the value of the metric. | No |
