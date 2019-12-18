# Metrics

This document describes iter8's out-of-the-box metrics, the anatomy of a metric definition, and how users can define their own metrics.

## Metrics defined by iter8

By default, iter8 leverages the metrics collected by Istio telemetry and stored in Prometheus. Users relying on iter8's out-of-the-box metrics can simply reference them in the success criteria of an _experiment_ specification, as shown in the [`Experiment` CRD documentation](iter8_crd.md).

During an `experiment`, for every call made from  _iter8-controller_ to _iter8-analytics_, the latter in turn calls Prometheus to retrieve values of the metrics referenced by the Kubernetes `experiment` resource. _Iter8-analytics_ analyzes the service versions that are part of the experiment and arrives at an assessment based on their metric values. It returns this assessment to _iter8-controller_.

In particular, the following metrics are available out-of-the-box from iter8. These metrics are based on the telemetry data collected by Istio. 

1. _iter8_latency_: mean latency, that is, average time taken by the service version to respond to HTTP requests.

2. _iter8_error_count_: total error count, that is, number of HTTP requests that resulted in error (5xx HTTP status codes).

3. _iter8_error_rate_: error rate, that is, (total error count / total number of HTTP requests).

When iter8 is installed, a Kubernetes `ConfigMap` named _iter8-metrics_ is populated with a definition for each of the above metrics. You can see the metric definitions in [this file](https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/install/helm/iter8-controller/templates/metrics/iter8_metrics.yaml). A few things to note in the definitions:

- Each metric is defined under the `metrics` section.

- They refer back to a Prometheus query template defined under the `query_templates` section. Iter8 uses that template to query Prometheus and compute the value of the metric for every service version.

- If this metric is a counter (i.e., its value never decreases over time), then the `is_counter` key corresponding to this metric is set to `true`; otherwise, it is set to `false`.

- If the value a metric is unavailable (for instance, Prometheus returned `NaN` or a null value for the query corresponding to his metric), then, by default, iter8 sets the value of this metric to `0`. This can be changed to any other float value (specified by the user in string format, e.g., `"22.8"`) or to `"None"` using the `absent_value` key.

- Finally, each metric is associated with another key `sample_size_query_template` whose value is a Prometheus query template. Iter8 relies on the notion of a sample-size to compute the total number of data points used in the computation of the metric values. Each of the iter8-defined metrics is associated with the `iter8_sample_size` query template defined under `query_templates`, which computes the total number of requests received by a service version. For the default iter8 metrics (mean latency, error count, and error rate), the total number of requests is the correct sample size measure.

## Adding a new metric

Next, we describe how the _iter8-metrics_ `ConfigMap` must be extended to add to iter8 a new metric based on data you have available in your Prometheus database. Any metric you define in the `ConfigMap` can then be referenced in the success criteria of experiments.

As an example, we will define a new metric called `error_count_400s` which computes the total count of HTTP requests that resulted in a 400 HTTP status code.

### Prometheus query template

First, we need a new Prometheus query template that will be used to observe that metric. We will write a template as follows:

```
sum(increase(istio_requests_total{source_workload_namespace!='knative-serving',response_code=~'4..',reporter='source'}[$interval]$offset_str)) by ($entity_labels)
```

As shown above, the template needs to have a few variables that are used by _iter8-analytics_ when querying Prometheus. First, note that it has a `group by` clause (specified using the _by_ keyword) with the variable `$entity_labels` as the group key. In general, each group in a Prometheus response corresponds to a distinct entity. Iter8 maps service versions to Prometheus entities using this variable. Also, the time period of aggregation is captured by the variable `$interval`. Finally, the variable `$offset_str` is used by iter8 to deal with historical data when available.

All three variables are required in a query template and the values are calculated by iter8 during the course of an experiment. We also require each Prometheus response to be an [instant vector](https://prometheus.io/docs/prometheus/latest/querying/basics/).


### Updating the _iter8-metrics_ `ConfigMap`

Our sample query template above must be associated with a metric name and added to the _query_templates_ section of the `ConfigMap` as follows:

```yaml
error_count_400s: "sum(increase(istio_requests_total{source_workload_namespace!='knative-serving',response_code=~'4..',reporter='source'}[$interval]$offset_str)) by ($entity_labels)"
```

Next, we declare the metric and define its type in the _metrics_ section of the `ConfigMap` as follows:

```yaml
 -  name: error_count_400s
    metric_type: Correctness
    sample_size_query_template: iter8_sample_size
```

In the declaration above, here is how to interpret the metric attributes:

  - _name_: Name of the new metric being defined. This value should be same as the associated Prometheus query template.

  - _metric_type_: Currently, iter8 supports two kinds of metrics- _Performance_ and _Correctness_. _error_count_400s_ is a _Correctness_ metric, since it is a measure of how many errors are produced by the code.

  - _sample_size_query_template_: As explained earlier, this is the query template iter8 uses to compute the total number of data points used in the computation of the metric value. Since the metric we are defining here is supposed to measure the total number of requests that resulted in a 400 HTTP code, the sample size from which that value is computed is represented by the total number of HTTP requests received. Hence, we are relying on the pre-defined sample-size query template `iter8_sample_size`, which computes the total number of HTTP requests. If you are defining a metric that requires the sample size to be computed differently, you can create a new sample-size query template (with a different name) in the _query_templates_ section and reference it in the metric declaration.
