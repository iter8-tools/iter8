---
template: overrides/main.html
---

# How Iter8 queries metrics

!!! abstract ""
    During an experiment, in each iteration, for each metric, and for each app-version, Iter8 uses an HTTP query to retrieve the current metric value. The params of this HTTP query are constructed by interpolating the `spec.params` field of the metric. 

## URL of the Prometheus metrics backend

The Prometheus metrics backend URL is [configurable as part of Iter8 install](/getting-started/install/#prometheus-url).

## How Iter8 interpolates `params`

* Each HTTP query parameter in the [`spec.params` field](/reference/apispec/#spec_1) of the metric is represented as a name-value pair. The value string is treated as a [`Python` string template](https://docs.python.org/3/library/string.html#string.Template). Specifically it may contain placeholders which are special strings beginning with the symbol `$`. This template is `interpolated`, i.e., its placeholders are substituted with actual values during query time as follows.
    - The `$interval` placeholder is substituted with the total time elapsed since the start of the experiment. All metrics are expected to use `$interval`.
    - The `$name` placeholder is substituted with the name of the app version.
    - Any other placeholder is substituted with the value of the corresponding [variable associated with the app version](/reference/apispec/#variable). If no such variable is associated with the app version, then no substitution occurs.

## An end-to-end Iter8 metric query example

This example illustrates the end-to-end process of retrieving metric values in Iter8 experiments. Consider the sample metric and experiment snippets shown below.

??? example "Sample metric for illustrating param interpolation"
    ```yaml linenums="1"
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

??? example "Experiment snippet for illustrating param interpolation"
    ```yaml linenums="1" hl_lines="7 12 17 22"
    ...
    spec:
      ...
      versionInfo:         
        baseline: 
          # name of the version
          name: sample-app-v1
          variables:
          - name: sample-app-v1
            value: sample-app-v1 
          - name: namespace
            value: iter8-knative-monitoring
          - name: promote
            value: baseline
        candidates: 
          # name of the version
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

1. Iter8 determines the [URL of the Prometheus metrics backend](#url-of-the-prometheus-metrics-backend).

2. Suppose Iter8 is performing iteration 10 of the experiment. `iter8-knative-monitoring/my-metric` is referenced in the `spec.criteria` field of the experiment. When querying the value of this metric for the app version named `sample-app-v2`, Iter8 does the following.
    - Iter8 computes the time elapsed since the start of the experiment (i.e, `time.Now() - status.initTime`). Suppose this value equals 285 seconds; Iter8 substitutes `$interval` with `285s`.
    - Iter8 substitutes `$name` with `sample-app-v2`, the name associated with this version.
    - Iter8 substitutes `$namespace` with `iter8-knative-monitoring` which is the value of the `namespace` variable associated with the version named `sample-app-v2`. 
    - **Note:** The `name` and `namespace` values associated with versions are highlighted in the experiment snippet above.

3. Step 2 yields an HTTP query with a single HTTP query parameter named `query` with the following value:
``` shell
sum(increase(revision_app_request_latencies_count{name='sample-app-v2', namespace='iter8-knative-monitoring'}[285s])) or on() vector(0)
```
Iter8 sends an HTTP GET request to the Prometheus metrics backend (for example, `http://prometheus-operated.iter8-knative-monitoring:9090`) with an HTTP query parameter named `query`. The [response from Prometheus](/reference/metrics/custom-metrics/#prometheus-response) contains the current value of `my-metric` for `sample-app-v2`.
