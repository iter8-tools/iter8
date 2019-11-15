# Metrics

This document describes the metrics made available by iter8 out of the box, the anatomy of a metric definition, and how users can define their own metrics.

## Metrics defined by iter8

By default, iter8 leverages the metrics collected by Istio telemetry and stored in Prometheus. Users relying on iter8-defined metrics can simply reference them in the success criteria of an _experiment_ specification, as shown in the [`Experiment` CRD documentation](iter8_crd.md).

During an `experiment`, every time _iter8-controller_ calls _iter8-analytics_ the latter retrieves from Prometheus the values of the metrics referenced by the corresponding Kubernetes `experiment` resource. _Iter8-analytics_ analyzes the service versions that are part of the experiment with respect to the metric values, assessing the current outcome of the experiment based on the defined success criteria. It then returns that assessment to _iter8-controller_.

In particular, the following metrics, based on telemetry data collected by Istio, are available to iter8 users:

1. _iter8_latency_: mean latency, that is, average time that it takes a service version to respond to HTTP requests.

2. _iter8_error_count_: total error count (~5** HTTP status codes), that is, the cumulative number of HTTP requests that resulted in an error.

3. _iter8_error_rate_: error rate, that is, the total error count divided by the total number of HTTP requests.

When iter8 is installed, a Kubernetes `ConfigMap` named _iter8-metrics_ is populated with a definition for each of the above metrics. You can see the metric definitions in [this file](https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/install/helm/iter8-controller/templates/metrics/iter8_metrics.yaml). A few things to note in the definitions:

- Each metric is defined under the `metrics` section.

- They each refer back to a Prometheus query template defined under the `query_templates` section. Iter8 uses that template to compute the value of the metric for a service version.

- The metrics have a type, which can be either `Correctness` or `Performance`, depending on what they measure.

- Finally, each metric is associated with another query template assigned to the attribute `sample_size_query_template`. Iter8 relies on the notion of a sample-size query template to compute the total number of data points used in the computation of the metric values. Each of the iter8-defined metrics is associated with the `iter8_sample_size` query template defined under `query_templates`, which computes the total number of requests received by a service version. For the default iter8 metrics (mean latency, error count, and error rate), the total number of requests is the correct sample size measure.

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

  - _sample_size_query_template_: As explained earlier, this is the query template iter8 uses to compute the total number of data points used in the computation of the metric value. Since the metric we are defining here is supposed to measure the total number of requests that resulted in a 400 HTTP code, the sample size from which that value is computed is represented by the total number of HTTP requests received. Hence, we are relying on the pre-defined sample size query template `iter8_sample_size`, which computes the total umber of HTTP requests. If you are defining a metric that requires the sample size to be computed differently, you can define a new sample-size query template (with a different name) in the _query_templates_ section and reference it in the metric declaration.
  
