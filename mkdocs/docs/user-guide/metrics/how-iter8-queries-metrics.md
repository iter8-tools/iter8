---
template: overrides/main.html
---

# How Iter8 queries metrics

!!! abstract ""
    During an experiment, in each iteration, for each metric, and for each app-version, Iter8 uses an HTTP query to retrieve the current metric value. The params of this HTTP query are constructed by interpolating the `spec.params` field of the metric. 

## How Iter8 interpolates params

* For each object in the `spec.params` field of a metric, Iter8 interpolates its `value` string.
    - `$interval` is replaced by the total time elapsed since the start of the experiment.
    - `$name` is replaced by the name of the app-version.
    - Any other placeholder (i.e., string beginning with `$`) is replaced by the value of the corresponding variable associated with the app version.[^1]

## An example

This example illustrates the end-to-end process of retrieving metric values in Iter8 experiments using the following samples.

??? example "A sample metric for illustrating param interpolation"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Metric
    metadata:
      name: my-metric
      namespace: iter8-knative-monitoring
    spec:
      params:
      - name: query
        value: sum(increase(revision_app_request_latencies_count{name='$name', namespace='$namespace'}[$interval])) or on() vector(0)
      description: my special metric
      type: counter
      provider: prometheus
    ```

??? example "Parts of a sample experiment for illustrating param interpolation"
    ```yaml
    ...
    spec:
      ...
      versionInfo:         
        baseline: 
          name: current
          variables:
          - name: sample-app-v1
            value: sample-app-v1 
          - name: namespace
            value: iter8-knative-monitoring
          - name: promote
            value: baseline
        candidates: 
        - name: sample-app-v2
          variables:
          - name: revision
            value: sample-app-v2
          - name: namespace
            value: iter8-knative-monitoring
          - name: promote
            value: candidate 
      criteria:
        objectives: 
        - metric: iter8-knative-monitoring/my-metric
          upperLimit: 50
        - metric: iter8-knative-monitoring/95th-percentile-tail-latency
          upperLimit: 100
        - metric: iter8-knative-monitoring/error-rate
          upperLimit: "0.01"
        indicators:
        - iter8-knative-monitoring/75th-percentile-tail-latency
        - iter8-knative-monitoring/90th-percentile-tail-latency
        - iter8-knative-monitoring/99th-percentile-tail-latency
      ...
    status:
      initTime: "2020-12-27T21:55:48Z"
      ...
    ```

1. Iter8 retrieves the [URL of the prometheus database](/getting-started/install/prometheus-url), for instance, http://prometheus-operated.iter8-knative-monitoring:9090.
2. Consider a specific iteration of the experiment. Consider `my-metric` which is referenced in the `spec.criteria` field of the experiment. Consider the app version named `sample-app-v2`.
    - suppose the time elapsed since the start of the experiment (i.e, since `status.initTime`) equals 285 seconds. Iter8 substitutes `$interval` with `285s`.
    - Iter8 substitutes `$name` with `sample-app-v2`, the version name.
    - Iter8 substitutes `$namespace` with `iter8-knative-monitoring` which is the value of the `namespace` variable for `sample-app-v2`.
3. Step 2 yields an HTTP query with a single parameter named `query` with the following value:
```
sum(increase(revision_app_request_latencies_count{name='sample-app-v2', namespace='iter8-knative-monitoring'}[285s])) or on() vector(0)
```
Iter8 sends an HTTP GET request to `http://prometheus-operated.iter8-knative-monitoring:9090` with the `query` parameter. The response from Prometheus contains the current value of `my-metric` for `sample-app-v2`. 

[^1]: Recall that each app version in the experiment is associated with [variables](http://localhost:8000/usage/experiment/versioninfo/#variables).