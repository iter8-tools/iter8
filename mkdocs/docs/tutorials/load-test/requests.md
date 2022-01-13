---
template: main.html
---

# Request Generation

!!! tip "Control the request generation process during a load test"
    While running a load test, you can set the total number of requests/the duration of the load test, the number of requests sent per second, and the number of parallel connections used to send requests. This provides fine-grained control over the request generation process.

## Example
Follow the [quick start tutorial](../../getting-started/your-first-experiment.md). In the step where you run the experiment, replace the `iter8 run` command with either of the following commands.

### Number of queries
Set the number of requests sent during the load-test to 200, the number of requests per second to 10, and the number of parallel connections used to send the requests to 5, as follows.

```shell
iter8 run --set url=https://example.com \
          --set numQueries=200 \
          --set qps=10 \
          --set connections= 5
```

### Duration
Set the duration of the load test to 20 seconds, the number of requests per second to 10, and the number of parallel connections used to send the requests to 5, as follows. The duration value may be any [Go duration string](https://pkg.go.dev/maze.io/x/duration#ParseDuration).

```shell
iter8 run --set url=https://example.com \
          --set duration=20s \
          --set qps=10 \
          --set connections= 5
```

When you set `numQueries` and `qps` parameters, the duration of the load-test is automatically determined. Similarly, when you set `duration` and `qps` parameters, the number of requests is automatically determined. If you set both `numQueries` and `duration` parameters, the latter is ignored.

***

The default values for all parameters that can be set during the load test experiment are documented in the `values.yaml` file.
