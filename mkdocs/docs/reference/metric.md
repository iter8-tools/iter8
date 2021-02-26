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
| sampleSize | string | Reference to a metric object in the `namespace/name` format or in the `name` format. The value of the sampleSize metric represents the number of data points over which the metric value is computed. This field applies only to `gauge` metrics. | No |
| provider | string | Type of the metrics database. Currently, `prometheus` is the only valid value. | No |

## Param

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Name of parameter. | Yes |
| value | string | Value that should be substitute for the parameter. | Yes |
