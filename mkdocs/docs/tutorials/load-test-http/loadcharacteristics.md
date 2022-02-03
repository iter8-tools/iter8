---
template: main.html
---

# Load Characteristics

!!! tip "Control the load characteristics during the HTTP load test experiment"
    Control the load characteristics during the HTTP load test experiment by setting the number of queries/duration, the number of queries sent per second, and the number of parallel connections used to send requests.

***

Follow the [quick start tutorial](../../getting-started/your-first-experiment.md). In the step where you run the experiment, replace the `iter8 run` command with either of the following commands.

### Number of queries
Set the number of requests sent during the load-test to 200, the number of requests per second to 10, and the number of parallel connections used to send the requests to 5, as follows.

```shell
iter8 run --set url=https://example.com \
          --set SLOs.error-rate=0 \
          --set SLOs.latency-mean=50 \
          --set SLOs.latency-p90=100 \
          --set SLOs.latency-p'97\.5'=200 \
          --set numQueries=200 \
          --set qps=10 \
          --set connections=5
```

### Duration
Set the duration of the load test to 20 sec, the number of requests per second to 10, and the number of parallel connections used to send the requests to 5, as follows. The duration value may be any [Go duration string](https://pkg.go.dev/maze.io/x/duration#ParseDuration).

```shell
iter8 run --set url=https://example.com \
          --set SLOs.error-rate=0 \
          --set SLOs.latency-mean=50 \
          --set SLOs.p90=100 \
          --set SLOs.p'97\.5'=200 \
          --set duration=20s \
          --set qps=10 \
          --set connections=5
```

***

When you set the `numQueries` and `qps` parameters, the duration of the load test is automatically determined. Similarly, when you set the `duration` and `qps` parameters, the number of requests is automatically determined. If you set both `numQueries` and `duration` parameters, the latter will be ignored.

