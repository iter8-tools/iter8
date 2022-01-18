---
template: main.html
---

# Metrics and SLOs

!!! tip "Learn more about metrics and SLOs"
    Learn more about the built-in metrics that are collected and the SLOs that are validated during the load test.

***

By default, the load test experiment collects the following built-in metrics: `error-count`, `error-rate`, `mean-latency`, and latency percentiles in the list `[50.0, 75.0, 90.0, 95.0, 99.0, 99.9]`. In addition, any other latency percentiles that are specified as part of SLOs are also collected.

***

Follow the [quick start tutorial](../../getting-started/your-first-experiment.md). In the step where you run the experiment, replace the `iter8 run` command with the following command.

```shell
iter8 run --set url=https://example.com \
          --set SLOs.error-rate=0 \
          --set SLOs.mean-latency=50 \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200
```

The above values ensure the following.

1.  The following latency percentiles are collected and reported: `[25.0, 50.0, 75.0, 90.0, 95.0, 97.5, 99.0, 99.9]`
2.  The following SLOs are validated.
    - error rate is 0
    - mean latency is under 50 msec
    - 90th percentile latency is under 100 msec
    - 97.5th percentile latency is under 200 msec

