---
menuTitle: Metrics
title: Metrics
weight: 62
summary: Iter8 metrics and customization
---

This document describes iter8's out-of-the-box metrics and how to extend iter8's metrics.

## Iter8's out-of-the-box metrics

<!-- TODO: What is thistutorial.md? -->

Iter8 leverages the metrics collected by Istio telemetry and stored in Prometheus. Users relying on iter8's out-of-the-box metrics can simply reference them in the criteria section of an _experiment_ specification, as illustrated in [this tutorial](thistutorial.md) and documented in the [`Experiment` CRD documentation]({{< ref "experiment" >}}). Iter8's out-of-the-box metrics are as follows.

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
*preferred_direction*    | *higher* or *lower* | This field indicates if higher values of the metric or preferred or lower values are preferred. It is of type enum with two possible values, *higher* or *lower*. For example, the *iter8_error_count* metric has a preferred direction which is *lower*. Preferred direction needs to be specified if you intend to use this as a reward metric or a metric with thresholds within experiment criteria (see [`Experiment` CRD documentation]({{< ref "experiment" >}})) | no
*units*    | *string* | Unit of measurement for this metric. For example, *iter8_latency* is a metric available out-of-the-box in iter8 and is measured in milli seconds. This field is used by iter8's KUI and Kiali integrations to format display. | no
*description*    | *string* | A description of this metric. This field is used by iter8's KUI and Kiali integrations to format display. | no

#### Prometheus query template for a counter metric {#query-template}
The Prometheus query template for the counter metric *iter8_error_count* is shown below.

```
sum(increase(istio_requests_total{response_code=~'5..',reporter='source',job='envoy-stats'}[$interval])) by ($version_labels)
```

The query template has two placeholders (i.e., terms beginning with $). These placeholders are substituted with actual values by iter8 in order to construct a Prometheus query. 

1. The query template has a `group by` clause (specified using the _by_ keyword) with the placeholder `$version_labels` as the group key, which ensures that each item in the Prometheus response vector corresponds to a distinct version. *Iter8* internally maps service versions to Prometheus entities using this placeholder.

2. The length of the recent time window over which this metric is computed is captured by the placeholder `$interval`. 

Both these placeholders are required in the query template. When a template is instantiated (i.e., placeholders are substituted with values), it results in a Prometheus query expression. An example of a query instantiated from the above template is shown below. In this example, since distinct versions correspond to distinct deployments of a service, iter8 has substituted `$version_labels` with the prometheus labels `destination_workload, destination_workload_namespace`. Each combination of these prometheus labels corresponds to a distinct version. 

```
sum(increase(istio_requests_total{response_code=~'5..',reporter='source',job='envoy-stats'}[300s])) by (destination_workload, destination_workload_namespace)
```
<!-- The  *iter8* queries Prometheus using this expression, the response from Prometheus needs to be an [instant vector](https://prometheus.io/docs/prometheus/latest/querying/basics/). -->

### Ratio metrics

A ratio metric is a ratio of two counter metrics. An example of a ratio metric that is available out-of-the-box in iter8 is *iter8_latency*, which is the average time taken by a service version to respond to HTTP requests. Iter8 ratio metrics have the following fields.

Field | Type | Description | Required
------|-------|--------|--------------
*name*    | *string* | Name of the metric | yes
*numerator*    | *string* | The counter metric in the numerator of the ratio | yes
*denominator*    | *string* | The counter metric in the denominator of the ratio | yes
*preferred_direction*    | *higher* or *lower* | This field indicates if higher values of the metric or preferred or lower values are preferred. It is of type enum with two possible values, *higher* or *lower*. For example, the *iter8_latency* metric has a preferred direction which is *lower*. Preferred direction needs to be specified if you intend to use this as a reward metric or a metric with thresholds within experiment criteria (see [`Experiment` CRD documentation]({{< ref "experiment" >}})) | no
*zero_to_one*    | *boolean* | This field indicates if the ratio metric always takes value in the range [0, 1]. For example, the *iter8_error_rate* metric has zero_to_one set to true. This field is optional and false by default. However, setting this field to true for metrics which possess this property helps iter8 provide better assessments. | no
*units*    | *string* | Unit of measurement for this metric. For example, *iter8_latency* has milli seconds as its units. This field is used by iter8's KUI and Kiali integrations to format display. | no
*description*    | *string* | A description of this metric. This field is used by iter8's KUI and Kiali integrations to format display. | no

### Adding new metrics in iter8

You can add new counter metrics in iter8 by extending the `counter_metrics.yaml` section of the configmap and new ratio metrics in iter8 by extending the `ratio_metrics.yaml` section of the configmap. For example, in the [A/B/n rollout tutorial]({{< ref "abn" >}}), during the step where you defined new metrics, you added the three new counter metrics and two new ratio metrics and extended the configmap as shown below.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iter8config-metrics
  namespace: iter8
data:
  # by convention, metrics with names beginning with iter8_ are defined by iter8
  # a counter metric is monotonically increasing or decreasing
  counter_metrics.yaml: |-
    - name: iter8_request_count
      query_template: sum(increase(istio_requests_total{reporter='source',job='istio-mesh'}[$interval])) by ($version_labels)
    - name: iter8_total_latency
      query_template: (sum(increase(istio_request_duration_seconds_sum{reporter='source',job='istio-mesh'}[$interval])) by ($version_labels))*1000
      units: msec # optional
    - name: iter8_error_count
      query_template: sum(increase(istio_requests_total{response_code=~'5..',reporter='source',job='istio-mesh'}[$interval])) by ($version_labels)
      preferred_direction: lower
    - name: books_purchased_total
      query_template: sum(increase(number_of_books_purchased_total{}[$interval])) by ($version_labels)
    - name: le_500_ms_latency_request_count
      query_template: (sum(increase(istio_request_duration_seconds_bucket{le='0.5',reporter='source',job='istio-mesh'}[$interval])) by ($version_labels))
    - name: le_inf_latency_request_count
      query_template: (sum(increase(istio_request_duration_seconds_bucket{le='+Inf',reporter='source',job='istio-mesh'}[$interval])) by ($version_labels))
  # the value of a ratio metric equals value of numerator divided by denominator
  ratio_metrics.yaml: |-
    - name: iter8_mean_latency
      numerator: iter8_total_latency
      denominator: iter8_request_count
      preferred_direction: lower
    - name: iter8_error_rate
      numerator: iter8_error_count
      denominator: iter8_request_count
      preferred_direction: lower
      zero_to_one: true
    - name: mean_books_purchased
      numerator: books_purchased_total
      denominator: iter8_request_count
      preferred_direction: higher
    - name: le_500_ms_latency_percentile
      numerator: le_500_ms_latency_request_count
      denominator: le_inf_latency_request_count
      preferred_direction: higher
      zero_to_one: true
```

**Note** Iter8 metrics are built on top of Prometheus metrics and rely on them being populated correctly in the Prometheus instance used in the iter8 experiment. All of iter8's out-of-the-box metrics rely on Prometheus metrics created by Istio's telemetry. In the extended configmap example, the newly defined counter metrics *le_500_ms_latency_request_count* and *le_inf_latency_request_count*, and the ratio metric *le_500_ms_latency_percentile* also rely on Prometheus metrics created by Istio's telemetry. The counter metric *books_purchased_total* and the ratio metric *mean_books_purchased* are created by directly instrumenting the [underlying application](https://github.com/iter8-tools/bookinfoapp-productpage).