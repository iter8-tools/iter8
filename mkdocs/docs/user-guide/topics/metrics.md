---
template: main.html
---

# Metrics

## Fully-qualified metric names
Metrics in Iter8 are grouped according to the type of source (backend) from which they originate. They are uniquely identified through their fully-qualified names, which are of the form `backend-name/metric-name`.

For example, Iter8's [built-in metrics](../tasks/collect.md) belong to the backend named `built-in`. One of the built-in metrics collected by Iter8 is `mean-latency`. Its fully qualified name is `built-in/mean-latency`.

## Metric types
Iter8 supports three types of metrics, `Counter`, `Gauge`, and `Histogram`.

### Counter
A Counter metric is one whose value never decreases over time. For example, the error count for an app version never decreases over the course of an experiment, and is a Counter metric. 

### Gauge
A Gauge metric is one whose value may increase or decrease over time. For example, the error rate for an app version may increase or decrease over the course of an experiment, and is a Gauge metric. 

### Histogram
A Histogram metric has a set of buckets, and counts associated with each bucket. For example, Iter8's built-in latency metric is stored in the form of a histogram. Each bucket represents a lower and upper value of observed latency value, and count represents the number of observations that fall within a bucket.