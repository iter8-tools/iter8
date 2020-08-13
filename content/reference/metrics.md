---
menuTitle: Metrics
title: Iter8 metrics
weight: 10
summary: Iter8 metrics and customization
---

This document describes iter8's out-of-the-box metrics and how to extend iter8's metrics.

## Iter8's out-of-the-box metrics

Iter8 leverages the metrics collected by Istio telemetry and stored in Prometheus. Users relying on iter8's out-of-the-box metrics can simply reference them in the criteria section of an _experiment_ specification, as illustrated in [this tutorial](thistutorial.md) and documented in the [`Experiment` CRD documentation](../experiment). Iter8's out-of-the-box metrics are as follows.

Metric name        | Description 
-------------------|------------------------
*iter8_request_count*    | total number of HTTP requests to a service version
*iter8_latency*    | average time in milli seconds taken by a service version to respond to HTTP requests
*iter8_error_count*| number of HTTP requests that resulted in error (5xx HTTP status codes)
*iter8_error_rate* | fraction of HTTP requests that resulted in errors, i.e., *iter8_error_count / iter8_request_count*

## Extending iter8's metrics

When iter8 is installed, a Kubernetes `ConfigMap` named _iter8config-metrics_ is populated with a definition for each of the above out-of-the-box metrics. You can see the metric definitions in [this file](https://raw.githubusercontent.com/iter8-tools/iter8/f302de20acf0f026a63453657fe88ff0674bee65/install/helm/iter8-controller/templates/metrics/iter8_metrics.yaml). You can extend iter8's metrics by extending this configmap. Below, we describe the two types of metrics supported by iter8, namely, `counter` and `ratio` metrics and how to extend the configmap in order to add new counter and ratio metrics.

### Counter metrics

A counter metric is a metric whose initial value is zero and which can only increase over time. An example of a counter metric that is available out-of-the-box in iter8 is *iter8_request_count*, which is the total number of HTTP requests that were received by a service version. Iter8 counter metrics have the following fields.

Field | Type | Description | Required
------|-------|--------|--------------
*name*    | *string* | Name of the metric | yes
*query_template*    | *string* | Prometheus query template used to fetch this metric (see [below](#query-template)) | yes
*units*    | *string* | Unit of measurement for this metric. For example, *iter8_latency* is a metric available out-of-the-box in iter8 and is measured in milli seconds. This field is used by iter8's KUI and Kiali integrations to format display. | no
*preferred_direction*    | *higher* or *lower* | This field indicates if higher values of the metric or preferred or lower values are preferred. It is of type enum with two possible values, *higher* or *lower*. For example, the *iter8_error_count* metric has a preferred direction which is *lower*. Preferred direction needs to be specified if you intend to use this as a reward metric or a metric with thresholds within experiment criteria (see [`Experiment` CRD documentation](../experiment)) | no
*description*    | *string* | A description of this metric. This field is used by iter8's KUI and Kiali integrations to format display. | yes

#### Prometheus query template {#query-template}

### Ratio metrics

<!-- A sample query template is shown below:

```
sum(increase(istio_requests_total{response_code=~'4..',reporter='source'}[$interval]$offset_str)) by ($entity_labels)
```

As shown above, the query template has three placeholders (i.e., terms beginning with $). These placeholders are substituted with actual values by _iter8-analytics_ in order to construct a Prometheus query. 1) The query template has a `group by` clause (specified using the _by_ keyword) with the placeholder `$entity_labels` as the group key. Each group in a Prometheus response corresponds to a distinct entity. _Iter8-analytics_ maps service versions to Prometheus entities using this placeholder. 2) The time period of aggregation is captured by the placeholder `$interval`. 3) The placeholder `$offset_str` is used by _iter8-analytics_ to deal with historical data when available. All three placeholders are required in the query template. When a template is instantiated (i.e., placeholders are substituted with values), it results in a Prometheus query expression; when we query Prometheus using this expression, the response from Prometheus needs  to be an [instant vector](https://prometheus.io/docs/prometheus/latest/querying/basics/). -->


### Adding a new counter metric

<!-- .. refer back to the example in ... -->

### Adding a new ratio metric

<!-- .. refer back to the example in ...



- Each metric is defined under the `metrics` section.

- They refer back to a Prometheus query template defined under the `query_templates` section. Iter8 uses that template to query Prometheus and compute the value of the metric for every service version.

- If this metric is a counter (i.e., its value never decreases over time), then the `is_counter` key corresponding to this metric is set to `True`; otherwise, it is set to `False`.

- If the value of a metric is unavailable (for example, Prometheus returned `NaN` or a null value for the query corresponding to his metric), then, by default, iter8 sets the value of this metric to `0`. This can be changed to any other float value (specified by the user in string format, e.g., `"22.8"`) or to `"None"` using the `absent_value` key.

- Finally, each metric is associated with another key `sample_size_query_template` whose value is a Prometheus query template. Iter8 relies on the notion of a sample-size to compute the total number of data points used in the computation of the metric values. Each of the iter8-defined metrics is associated with the `iter8_sample_size` query template defined under `query_templates`, which computes the total number of requests received by a service version. For the default iter8 metrics (mean latency, error count, and error rate), the total number of requests is the correct sample size measure. -->

## Adding a new metric

<!-- Next, we describe how iter8 can be extended with new metrics through the _iter8-metrics_ `ConfigMap`. Any metric you define in the `ConfigMap` can then be referenced in the success criteria of experiments.

As an example, we will define a new metric called `error_count_400s` which computes the total count of HTTP requests that resulted in a 400 HTTP status code.
Adding a new metric involves creating a Prometheus query template and associating this template with the metric definition.  We now describe the structure of a Prometheus query template. -->

### Updating the _iter8-metrics_ `ConfigMap`

<!-- There are two steps involved in adding a new metric. Step 1: Extend the _query_templates_ section of the `ConfigMap`.

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

  - _sample_size_query_template_: As explained earlier, this is the sample size query template associated with this metric. The sample over which this value is computed for a version is the set of all HTTP requests received by the version. Hence, we are relying on the pre-defined sample-size query template `iter8_sample_size`, which computes the total number of HTTP requests for a version. If you are defining a metric that requires the sample size to be computed differently, you can create a new sample-size query template (with its own unique name) in the _query_templates_ section and reference it in the metric declaration. -->

  <!-- During an `experiment`, for every call made from  _iter8-controller_ to _iter8-analytics_, the latter in turn calls Prometheus to retrieve values of the metrics referenced by the Kubernetes `experiment` resource. _Iter8-analytics_ analyzes the service versions that are part of the experiment and arrives at an assessment based on their metric values. It returns this assessment to _iter8-controller_. -->

