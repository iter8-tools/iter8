---
template: overrides/main.html
---

# Metric Resource Object

!!! abstract ""
    Iter8 defines a Kubernetes custom resource kind called `Metric` to automate metrics-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps. This document describes Metric version `v2alpha1`. Metric CRD is defined by Iter8's `etc3` controller repo. For documentation on etc3 and the Go client for `Metric` API, see [here](https://pkg.go.dev/github.com/iter8-tools/etc3@v0.1.14/).


## MetricSpec

!!! abstract ""
    Fields in the `spec` stanza of a Metric resource object.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| params | [][Param](#param) | List of name/value pairs corresponding to the name and value of the HTTP query parameters used by Iter8 when querying the metrics backend. Each name represents a parameter name; the corresponding value is a template, which will be instantiated by Iter8 at query time. For examples and more details, see [here](/usage/metrics/how-iter8-queries-metrics/).| No |
| description | string | Human-readable description of the metric. | No |
| units | string | Units of measurement. Units are used only for display purposes. | No |
| type | string | Metric type. Valid values are `counter` and `gauge`. Default value = `gauge`. | No |
| sampleSize | [MetricReference](#metricreference) | Reference to a metric that represents the number of data points over which the metric value is computed. This field applies only to `gauge` metrics. | No |
| provider | string | Type of the metrics database. Currently, `prometheus` is the only valid value. | No |

## Param

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Name of the HTTP query parameter. | Yes |
| value | string | Value of the HTTP query parameter. See [here](/usage/metrics/how-iter8-queries-metrics/) for documentation on how Iter8 interpolates params. | Yes |

## MetricReference

!!! abstract ""
    A reference to a metric in the cluster.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| namespace | string | Namespace containing the referred metric. Defaults to the namespace of the referring metric. | No |
| name | string | Name of the referred metric. | Yes |
