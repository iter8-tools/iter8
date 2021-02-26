---
template: overrides/main.html
---

# Metric Resource Object

Fields in an iter8 metric resource object `spec` are documented here. For complete documentation, see the iter8 Experiment API [here](https://pkg.go.dev/github.com/iter8-tools/etc3@v0.1.13-pre/api/v2alpha1).

## MetricSpec

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| params | map[string][Param](#param) | A list of name/value pairs that can be substituted in queries to metrics backends. For examples and more details, see [here](https://iter8.tools/usage/metrics/how-iter8-queries-metrics/).| No |
| description | string | Human-readable description of the metric. | No |
| units | string | Units of measurement. Units are used only for display purposes. | No |
| type | string | Metric type. Valid values are `counter` and `gauge`. Default value = `gauge`. | No |
| sampleSize | string | Reference to a metric object in the `namespace/name` format or in the `name` format. The value of the sampleSize metric represents the number of data points over which the metric value is computed. This field applies only to `gauge` metrics. | No |
| provider | string | Type of the metrics database. Currently, `prometheus` is the only valid value. | No |

## Param

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | name of a parameter | No |
| value | string | value that should be substituted or name | No |