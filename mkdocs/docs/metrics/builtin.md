---
template: main.html
---

# Builtin Metrics

!!! tip "Builtin latency/error metrics"    
    Iter8 ships with a set of nine builtin metrics that measure your application's performance in terms of latency and errors. You can collect and use these metrics in experiments without the need to configure any external databases. 
    
    This feature enables you to get started with Iter8 experiments, especially, SLO validation experiments, quickly. As part of metrics collection, Iter8 will also generate HTTP requests to the application end-point.

## List of builtin metrics
The following are the set of builtin Iter8 metrics.

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

## Collecting builtin metrics
Use the [`metrics/collect` task](../reference/tasks/metrics-collect.md) in an experiment to collect builtin metrics for your app/ML model versions.

## Example
For an example of an experiment that uses builtin metrics, look inside the Knative experiment in [this tutorial](../../tutorials/knative/testing-strategies/conformance/#5-launch-experiment).