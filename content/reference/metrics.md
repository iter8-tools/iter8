---
menuTitle: Metrics
title: Iter8's metrics
weight: 10
summary: Iter8's out-of-the-box metrics and how users can define their own metrics
---

This document describes iter8's out-of-the-box metrics, the anatomy of a metric definition, and how users can define their own metrics.

## Metrics defined by iter8

By default, iter8 leverages the metrics collected by Istio telemetry and stored in Prometheus. Users relying on iter8's out-of-the-box metrics can simply reference them in the success criteria of an _experiment_ specification, as shown in the [`Experiment` CRD documentation](../experiment/).

During an `experiment`, for every call made from  _iter8-controller_ to _iter8-analytics_, the latter in turn calls Prometheus to retrieve values of the metrics referenced by the Kubernetes `experiment` resource. _Iter8-analytics_ analyzes the service versions that are part of the experiment and arrives at an assessment based on their metric values. It returns this assessment to _iter8-controller_.

In particular, the following metrics are available out-of-the-box from iter8. These metrics are based on the telemetry data collected by Istio.

1. _iter8_latency_: mean latency, that is, average time taken by the service version to respond to HTTP requests.

2. _iter8_error_count_: total error count, that is, number of HTTP requests that resulted in error (5xx HTTP status codes).

3. _iter8_error_rate_: error rate, that is, (total error count / total number of HTTP requests).

When iter8 is installed, a Kubernetes `ConfigMap` named _iter8config-metrics_ is populated with a definition for each of the above metrics. You can see the metric definitions in [this file](https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/install/helm/iter8-controller/templates/metrics/iter8_metrics.yaml). A few things to note in the definitions:

- Each metric is defined under the `metrics` section.

- They refer back to a Prometheus query template defined under the `query_templates` section. Iter8 uses that template to query Prometheus and compute the value of the metric for every service version.

- If this metric is a counter (i.e., its value never decreases over time), then the `is_counter` key corresponding to this metric is set to `True`; otherwise, it is set to `False`.

- If the value of a metric is unavailable (for example, Prometheus returned `NaN` or a null value for the query corresponding to his metric), then, by default, iter8 sets the value of this metric to `0`. This can be changed to any other float value (specified by the user in string format, e.g., `"22.8"`) or to `"None"` using the `absent_value` key.

- Finally, each metric is associated with another key `sample_size_query_template` whose value is a Prometheus query template. Iter8 relies on the notion of a sample-size to compute the total number of data points used in the computation of the metric values. Each of the iter8-defined metrics is associated with the `iter8_sample_size` query template defined under `query_templates`, which computes the total number of requests received by a service version. For the default iter8 metrics (mean latency, error count, and error rate), the total number of requests is the correct sample size measure.

## Adding a new metric

Next, we describe how iter8 can be extended with new metrics through the _iter8-metrics_ `ConfigMap`. Any metric you define in the `ConfigMap` can then be referenced in the success criteria of experiments.

As an example, we will define a new metric called `error_count_400s` which computes the total count of HTTP requests that resulted in a 400 HTTP status code.
Adding a new metric involves creating a Prometheus query template and associating this template with the metric definition.  We now describe the structure of a Prometheus query template.

### Prometheus query template

A sample query template is shown below:

```
sum(increase(istio_requests_total{response_code=~'4..',reporter='source'}[$interval]$offset_str)) by ($entity_labels)
```

As shown above, the query template has three placeholders (i.e., terms beginning with $). These placeholders are substituted with actual values by _iter8-analytics_ in order to construct a Prometheus query. 1) The query template has a `group by` clause (specified using the _by_ keyword) with the placeholder `$entity_labels` as the group key. Each group in a Prometheus response corresponds to a distinct entity. _Iter8-analytics_ maps service versions to Prometheus entities using this placeholder. 2) The time period of aggregation is captured by the placeholder `$interval`. 3) The placeholder `$offset_str` is used by _iter8-analytics_ to deal with historical data when available. All three placeholders are required in the query template. When a template is instantiated (i.e., placeholders are substituted with values), it results in a Prometheus query expression; when we query Prometheus using this expression, the response from Prometheus needs  to be an [instant vector](https://prometheus.io/docs/prometheus/latest/querying/basics/).


### Updating the _iter8-metrics_ `ConfigMap`

There are two steps involved in adding a new metric. Step 1: Extend the _query_templates_ section of the `ConfigMap`.

```yaml
error_count_400s: "sum(increase(istio_requests_total{job='istio-mesh',response_code=~'4..',reporter='source'}[$interval]$offset_str)) by ($entity_labels)"
```

For example, we have added our sample query template under a new key called `error_count_400s` in the _query_templates_ section of the `ConfigMap`. This query assumes that we are using Istio telemetry `v1`.

Step 2: We define a new metric in the _metrics_ section of the `ConfigMap` as follows:

```yaml
 -  name: error_count_400s
    is_counter: True
    sample_size_query_template: iter8_sample_size
```

The interpretation of the above definition is as follows.

  - _name_: Name of the new metric being defined. Its value is the key for the associated query template. In our example, the relevant key is `error_count_400s`.

  - _is_counter_: Set this to True if this metric is a counter metric. In our example, 4xx errors can never decrease over time, and therefore `is_counter` is set to True.

  - _sample_size_query_template_: As explained earlier, this is the sample size query template associated with this metric. The sample over which this value is computed for a version is the set of all HTTP requests received by the version. Hence, we are relying on the pre-defined sample-size query template `iter8_sample_size`, which computes the total number of HTTP requests for a version. If you are defining a metric that requires the sample size to be computed differently, you can create a new sample-size query template (with its own unique name) in the _query_templates_ section and reference it in the metric declaration.
