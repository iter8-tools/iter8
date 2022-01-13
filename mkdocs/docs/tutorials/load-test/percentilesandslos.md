---
template: main.html
---

# Error Codes, Percentiles, and SLOs

!!! tip "Specify error codes, latency percentiles and SLOs"
    While running a load test, you can specify the range of HTTP status codes that are considered as errors, the latency percentile values that are computed and reported, and the SLOs that are evaluated.

## Example
Follow the [quick start tutorial](../../getting-started/your-first-experiment.md). Before you run the experiment, copy and paste the following YAML into the `values.yaml` file.

```yaml
errorRanges:
- lower: 500
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

1.  HTTP status codes 400 and above are considered as errors.
2.  The following latency percentiles are computed and reported: `[50.0, 75.0, 90.0, 95.0, 97.5, 99.0, 99.9]`
3.  The following SLOs are evaluated.
    - error rate is 0
    - mean latency is under 100 msec
    - median (50th percentile) latency is under 100 msec
    - 95th percentile latency is under 250 msec
    - 97.5th percentile latency is under 500 msec

***

You may modify the `errorRanges`, `percentiles` and `SLOs` sections of the `values.yaml` files before running the experiment, as required by your load testing needs.
