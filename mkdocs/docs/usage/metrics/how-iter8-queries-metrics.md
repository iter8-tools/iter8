---
template: overrides/main.html
---

# How iter8 queries metrics

> During an experiment, in each iteration, for each metric, and for each app-version, iter8 uses an HTTP query to retrieve the current metric value. The params of this HTTP query are constructed by interpolating the `spec.params` stanza of the metric. 

## How iter8 interpolates params

* For each object in the `spec.params` stanza of a metric, iter8 interpolates its `value` string.
    - `$interval` is replaced by the total time elapsed since the start of the experiment.
    - `$name` is replaced by the name of the app-version.
    - Any other placeholder (i.e., string beginning with `$`) is replaced by the value of the corresponding variable associated with the app version.[^1]

## An example

We will illustrate the end-to-end process of retrieving metric values in iter8 experiments using the following example.

??? example "A sample metric for illustrating param interpolation"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Metric
    metadata:
      name: my-metric
    spec:
      params:
      - name: query
        value: sum(increase(revision_app_request_latencies_count{name='$name', $namespace='$namespace'}[$interval])) or on() vector(0)
      description: my special metric
      type: counter
      provider: prometheus
    ```

[^1]: Recall that each app version in the experiment is associated with [variables](http://localhost:8000/usage/experiment/versioninfo/#variables).