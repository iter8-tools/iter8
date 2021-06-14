---
template: main.html
---

# Metrics mock

!!! tip "Mocking the value of a metric"
    Iter8 enables you to mock the values of a metric. This is useful for learning purposes and quickly trying out sample Iter8 experiments without having to set up metric databases.

## Examples

```yaml linenums="1"
apiVersion: iter8.tools/v2alpha2
kind: Metric
metadata:
  name: user-engagement
spec:
  mock:
	- name: default
	  level: "20.0"
	- name: canary
	  level: "15.0"
```

```yaml linenums="1"
apiVersion: iter8.tools/v2alpha2
kind: Metric
metadata:
  name: request-count
spec:
	type: Counter
  mock:
	- name: default
	  level: "20.0"
	- name: canary
	  level: "15.0"
```

## Explanation
1. Use the `mock` field within metrics spec to indicate that the value of this metric needs to be mocked.
2. The name within the `name-level` pairs must match the name of some version in the `versionInfo` field of the experiment.
3. You can mock both `Counter` and `Gauge` metrics.
4. The semantics of `level` field are as follows:
	* If the metric is a counter, if level is `x`, and time elapsed since the start of the experiment is `y`, then `xy` is the metric value. This will keep increasing the metric value over time.
	* If the metric is gauge, if level is `x`, the metric value is a random value with mean `x`. The expected value of the metric will be `x` but its observed value may increase or decrease over time.