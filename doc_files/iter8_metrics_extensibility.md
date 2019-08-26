# Extending Metrics for iter8

This tutorial shows how you can add custom metrics to iter8 to observe the canary and baseline versions of a microservice before a roll out analysis.

First we will take a look at the available metrics that iter8 comes pre-packaged with. Then we will learn how to modify a configuration file to add new metrics.

## Available metrics
Iter8, by default allows three kinds of metrics to assess the health of a canary version. They are:
1. _iter8_latency_: measures the mean latency of a service belonging to a namespace.
2. _iter8_error_rate_: measures the mean error rate (~5** HTTP Status codes) of the service belonging to a namespace.
3. _iter8_error_count_: measures total error count (~5** HTTP Status codes) of the service belonging to a namespace.

All these metrics query [Prometheus](https://prometheus.io) and are defined in `iter8-controller/install/helm/iter8-controller/templates/metrics/iter8_metrics.yaml`. We will henceforth refer to this file as the `Metrics Config File`.

Few things to note about configurations on this file:
- Each metric is defined under the `metrics` section of the Metrics Config File.
- They each refer back to a Prometheus Query template defined under the `query_templates` section in the same file.
- The metrics are also defined as either a `Correctness` metric or a `Performance` metric depending on what they measure.
- In addition to this, each metric is associated to another required parameter called `sample_size_query_template`. This refers back to a Prometheus query defined under `query_templates` which measures the  number of data points required to make a decision for that metric.

## Adding a new metric

iter8 comes with the metrics extensibility feature which allows users to define any new Prometheus query to assess the canary version of a micro service.

In this section we will describe how this feature can be put to use. As an example, we will define a metric called `error_count_400s` which measures the count of 400 HTTP Status codes from a service.

- First we add a new Prometheus Query that will be used to observe that metric. This is done under the _query_templates_ section in the Metrics Config File as follows:
```
error_count_400s: "sum(increase(istio_requests_total{source_workload_namespace!='knative-serving',response_code=~'4..',reporter='source'}[$interval]$offset_str)) by ($entity_labels)"
```
In the metrics defined thus far, notice that the Prometheus query templates have a group by clause (specified using the by keyword) with `$entity_labels` as the group key. In general, each group in a Prometheus response corresponds to a distinct entity which is defined in the CRD. We require each Prometheus response to be an [instant vector](https://prometheus.io/docs/prometheus/latest/querying/basics/).
In the query template, the time period of aggregation is the variable `$interval`. The variable `$offset_str` is used to analyze metrics in the past. It offsets the experiment by a time period which is the difference between the current time and the time at which we would like to stop measuring the metric in the past. For current experiments, iter8 appends an empty string here.
All three variables are _required parameters_ in a query template and the values are calculated by iter8 during the course of a roll out test.

-  Next we define the various features of this metric in the _metrics_ section of the same file:
```
 -  name: error_count_400s
      metric_type: Correctness
      sample_size_query_template: iter8_sample_size
```
  - _name_: Name of the new metric being defined. Should be same as the associated Prometheus query template
  - _metric_type_: Currently, iter8 supports two kinds of metrics- _Performance_ and _Correctness_. _error_count_400s_ is a Correctness metric.
  - _sample_size_query_template_: Sample Size measures the  number of data points required to make a decision for that metric. Like the three available iter8 metrics, _error_count_400s_ also refers back to the _iter8_sample_size_ query template defined in the _query_templates_ section of the file. If needed, a new sample size query template can be defined (with a different name) in the _query_templates_ section and can be referred to in this section.
