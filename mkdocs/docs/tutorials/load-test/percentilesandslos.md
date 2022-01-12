---
template: main.html
---

# Percentiles and SLOs

!!! tip "Control the latency percentiles computed and SLOs evaluated during the load test"
    While running a load test, you can control the latency percentiles that are computed and the SLOs that are evaluated by altering the `values.yaml` file. This tutorial shows you how.

## Latency percentiles and SLOs
Follow the [quick start tutorial](../../getting-started/your-first-experiment.md). Before you run the experiment, copy and paste the following YAML into the `values.yaml` file.

```yaml
percentiles: [50.0, 75.0, 90.0, 95.0, 97.5, 99.0, 99.9]
SLOs:
- metric: "built-in/error-rate"
  upperLimit: 0
- metric: "built-in/mean-latency"
  upperLimit: 100
- metric: "built-in/p50.0"
  upperLimit: 100
- metric: "built-in/p95.0"
  upperLimit: 250
- metric: "built-in/p97.5"
  upperLimit: 500
```

The above values ensure the following.

1.  The 50th, 75th, 90th, 95th, 97.5th, 99th, and 99.9th latency percentile values are computed.
2.  The following SLOs are evaluated.
    - error rate is 0
    - mean latency is under 100 msec
    - median (50th percentile) latency is under 100 msec
    - 95th percentile latency is under 250 msec
    - 97.5th percentile latency is under 500 msec

You can modify the `percentiles` and `SLOs` section of the `values.yaml` before the experiment run, as required by your testing needs.
