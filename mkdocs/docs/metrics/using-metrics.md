---
template: main.html
---

# Using Metrics in Experiments

!!! tip "Iter8 metric resources"    
    Iter8 defines a custom Kubernetes resource (CRD) called **Metric** that makes it easy to define and use metrics in experiments. 
    
    Iter8 installation includes a set of pre-defined [builtin metrics](builtin.md) that pertain to app/ML model latency/errors. You can also [define custom metrics](custom.md) that enable you to utilize data from Prometheus, New Relic, Sysdig, Elastic or any other database of your choice.

## List metrics
Find the set Iter8 metrics available in your cluster using `kubectl get`.

``` shell
kubectl get metrics.iter8.tools --all-namespaces
```

```shell
NAMESPACE         NAME                      TYPE      DESCRIPTION
iter8-kfserving   user-engagement           Gauge     Average duration of a session
iter8-system      error-count               Counter   Number of responses with HTTP status code 4xx or 5xx (Iter8 builtin metric)
iter8-system      error-rate                Gauge     Fraction of responses with HTTP status code 4xx or 5xx (Iter8 builtin metric)
iter8-system      latency-50th-percentile   Gauge     50th percentile (median) latency (Iter8 builtin metric)
iter8-system      latency-75th-percentile   Gauge     75th percentile latency (Iter8 builtin metric)
iter8-system      latency-90th-percentile   Gauge     90th percentile latency (Iter8 builtin metric)
iter8-system      latency-95th-percentile   Gauge     95th percentile latency (Iter8 builtin metric)
iter8-system      latency-99th-percentile   Gauge     99th percentile latency (Iter8 builtin metric)
iter8-system      mean-latency              Gauge     Mean latency (Iter8 builtin metric)
iter8-system      request-count             Counter   Number of requests (Iter8 builtin metric)
```

## Referencing metrics within experiments

Use metrics in experiments by referencing them in the criteria section of the experiment manifest. Reference metrics using the `namespace/name` or `name` [format](../reference/experiment.md#criteria).

??? example "Sample experiment illustrating the use of metrics"
    ```yaml
    kind: Experiment
    ... 
    spec:
      ...
      criteria:
        requestCount: iter8-knative/request-count
        # mean latency of version should be under 50 milliseconds
        # 95th percentile latency should be under 100 milliseconds
        # error rate should be under 1%
        objectives: 
        - metric: iter8-knative/mean-latency
          upperLimit: 50
        - metric: iter8-knative/95th-percentile-tail-latency
          upperLimit: 100
        - metric: iter8-knative/error-rate
          upperLimit: "0.01"
    ```

## Observing metric values
During an experiment, Iter8 reports the metric values observed for each version. Use `iter8ctl` to observe these metric values in realtime. See [here](../getting-started/quick-start/kfserving/tutorial.md#a-observe-results) for an example.