---
template: overrides/main.html
---

# Custom Metrics

!!! abstract ""
    Define custom Iter8 metrics based on Prometheus metrics, and use them in experiments.

This document illustrates custom metric creation using three examples. The first two examples illustrate `counter` metrics while the third illustrates `gauge` metrics. You may find it helpful to read documentation on [using metrics in experiments](/reference/metrics/using-metrics) and [how Iter8 queries metrics](/reference/metrics/how-iter8-queries-metrics) before creating custom metrics.

### Example 1: counter metric

#### Defining an Iter8 counter metric named `correct-predictions`
Suppose you are experimenting with ML models and you have a Prometheus counter metric named `correct_predictions`, which records the number of correct predictions made by each model version until now.
```shell
# Prometheus query to get the number of correct predictions for a model version in the past 30 seconds.
sum(increase(correct_predictions{revision_name='my-model-predictor-default-dlgm8'}[30s]))
```
```shell
# Prometheus query to get the number of correct predictions for another model version in the past 30 seconds.
sum(increase(correct_predictions{revision_name='my-model-predictor-default-h4bvl'}[30s]))
```

!!! tip "Metric labels"
    This example is motivated by Iter8-KFServing experiments. KFServing creates distinct Knative revisions for different model versions. Hence, as seen in the above examples, the `revision_name` label provides a convenient way to filter and select a specific model version.

You can turn this Prometheus metric into an Iter8 counter metric using the following yaml manifest.
```yaml
#correctpredictions.yaml
apiVersion: iter8.tools/v2alpha2
kind: Metric
metadata:
  name: correct-predictions
spec:
  params:
  - name: query
    value: sum(increase(correct_predictions{revision_name='$revision'}[$elapsedTime])) or on() vector(0)
  description: Number of correct predictions
  type: counter
  provider: prometheus
```

!!! tip "Dealing with `nodata`"
    Values may be unavailable for a metric in Prometheus, in which case, Prometheus may return a `nodata` response. For example, values may be unavailable for the `correct_predictions` metric for a model version if no requests have been sent to that model version until now, or if Prometheus has a large scrape interval and is yet to collect data. In such cases, the `on() or vector(0)` clause replaces the `nodata` response with a zero value. This is the recommended approach for defining Iter8 counter metrics.


Using the above YAML file, you can create an Iter8 metric in your Kubernetes cluster as follows.
```shell
kubectl apply -f correctpredictions.yaml -n your-metrics-namespace
```

You can now list this metric using `kubectl`.
```shell
kubectl get metrics.iter8.tools correct-predictions -n your-metrics-namespace
NAMESPACE                NAME                           TYPE      DESCRIPTION
your-metrics-namespace   correct-predictions            counter   Number of correct predictions
```

### Example 2: counter metric

#### `request-count` metric
The `request-count` metric is typically installed *out-of-the-box* as part of Iter8. Although not a custom metric, this example serves to further illustrate the concepts introduced in Example 1. There are no real differences between custom and out-of-the-box metrics other than the fact the latter are created as part of Iter8 installation.
```yaml
apiVersion: iter8.tools/v2alpha2
kind: Metric
metadata:
  name: request-count
spec:
  params:
  - name: query
    value: sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$elapsedTime])) or on() vector(0)
  description: Number of requests
  type: counter
  provider: prometheus
  jqExpression: ".data.result[0].value[1] | tonumber"
```

### Example 3: gauge metric

#### Defining an Iter8 gauge metric named `accuracy`
This example builds on Examples 1 and 2 to define a new Iter8 gauge metric called `accuracy`. This metric is intended to capture the ratio of correct predictions over request count.
```yaml
#accuracy.yaml
apiVersion: iter8.tools/v2alpha2
kind: Metric
metadata:
  name: accuracy
spec:
  description: Accuracy of the model version
  params:
  - name: query
    value: (sum(increase(correct_predictions{revision_name='$revision'}[$elapsedTime])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$elapsedTime])) or on() vector(0))
  type: gauge
  sampleSize: 
    name: iter8-kfserving-monitoring/request-count
  provider: prometheus
  jqExpression: ".data.result[0].value[1] | tonumber"
```

`spec.sampleSize` represents the number of data points over which the gauge metric value is computed. In this case, since `accuracy` is computed over all the requests received by a specific model version, the sampleSize metric is `request-count`.

## Prometheus response
The following is a sample response returned by Prometheus to Iter8 for a metric query.
```json
{
    "status": "success",
    "data": {
      "resultType": "vector",
      "result": [
        {
          "value": [1556823494.744, "21.7639"]
        }
      ]
    }
}
```
Whenever Iter8 queries Prometheus for a metric, it issues `n` queries, where `n` is the number of app versions involved in the experiment.[^1] For each query, Iter8 expects the schema of the Prometheus response to match the schema of the above sample. Specifically, `status` should equal `success`, `resultType` should equal `vector` (i.e., Prometheus should return an instant-vector), with a single `result` within it.

[^1]: `n=1` in `Conformance` experiments, and `n=2` in `Canary` experiments.
