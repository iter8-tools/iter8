---
template: main.html
---

# Metrics

## Fully-qualified metric names
Metrics in Iter8 are grouped according to the type of source (backend) from which they originate. They are uniquely identified through their fully-qualified names, which are of the form `backend-name/metric-name`.

## Example
Iter8's [built-in metrics](../tasks/collect.md) belong to the backend named `built-in`. One of the built-in metrics collected by Iter8 is `mean-latency`. Its fully qualified name is `built-in/mean-latency`.