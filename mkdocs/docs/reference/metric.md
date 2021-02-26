---
template: overrides/main.html
---

# Metric Resource Object

## MetricSpec

Fields in an iter8 metric resource object `spec` are documented here.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| params | [][Param](#param) | List of name/value pairs. Each name represents a parameter name; the corresponding value is a template, which will be instantiated by iter8 while querying the metrics backend. For examples and more details, see [here](metrics_custom.md#instantiation-of-templated-http-query-params).| No |
| description | string | Human-readable description of the metric. | No |
| units | string | Units of measurement. Units are used only for display purposes. | No |
| type | string | Metric type. Valid values are `counter` and `gauge`. Default value = `gauge`. | No |
| sampleSize | [MetricReference](#metricreference) | Reference to a metric that represents the number of data points over which the metric value is computed. This field applies only to `gauge` metrics. | No |
| provider | string | Type of the metrics database. Currently, `prometheus` is the only valid value. | No |

## Param

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Name of parameter. | Yes |
| value | string | Value that should be substitute for the parameter. | Yes |

## MetricReference

A reference to another metric in the cluster. Used to refer to the metric that is used to count the number of data points over which a gauge metric is computed.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| namespace | string | Namespace containing another metric. Defaults to the namespace of the referring object. | No |
| name | string | Name of another metric. | Yes |
