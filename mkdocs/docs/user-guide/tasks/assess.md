---
template: main.html
---

# `assess-app-versions`
This task assesses if app versions satisfy service level objectives (SLOs). SLO inputs are specified in the form of metrics along with acceptable upper and lower limits on their values.

This task should be preceded in the experiment spec by other tasks that collect metrics such as the [`gen-load-and-collect-metrics` task](collect.md).

## Example
Validate service level objectives (SLOs) for app versions based on [Iter8's built-in metrics](collect.md).

```yaml
- task: assess-app-versions
  with:
    SLOs:
      # error rate must be 0
    - metric: built-in/error-rate
      upperLimit: 0
      # mean latency must be under 50 msec
    - metric: built-in/mean-latency
      upperLimit: 50
      # 95th percentile latency must be under 100 msec
    - metric: built-in/p95.0
      upperLimit: 100
      # 99th percentile latency must be under 300 msec
    - metric: built-in/p99.0
      upperLimit: 300
```

## Inputs
The following inputs are supported by this task.

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| SLOs | [][SLO](#slo) | A list of [service level objectives (SLOs)](#slo) | No |

### SLO

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| metric | string | [Fully-qualified metric name](../topics/metrics.md). | Yes |
| upperLimit | float64 | Acceptable upper limit on the value of the metric. | No |
| lowerLimit | float64 | Acceptable lower limit on the value of the metric. | No |
