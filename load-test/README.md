# Run an Iter8 Load Test Experiment

This action runs an Iter8 [load test against an HTTP service](https://iter8.tools/0.8/tutorials/load-test-http/usage/). For more details about Iter8 experiments, see <https://iter8.tools>.

## Usage

Experiments are expressed as helm charts. Create a values file to define experiment specific values. For example, to load test and validate the the HTTP service whose URL is <https://httpbin.org/get> with requirements that the error rate must be 0, the mean latency must be under 50 msec, the 90th percentile latency must be under 100 msec, and the 97.5th percentile latency must be under 200 msec:

``` yaml
    - run: |
        cat << EOF > values.yaml
          url: https://httpbin.org/get
          SLOs:
            error-rate: 0
            latency-mean: 100
            latency-p90: 150
            latency-p97.5: 200
        EOF
    - uses: iter8-tools/iter8/load-test@v0.8
      with:
        valuesFile: ../values.yaml
```

For more options, see <https://iter8.tools>.

## Inputs

| Input Name | Description | Default |
| ---------- | ----------- | ------- |
| `valuesFile` | Path to file of configuration values. | `values.yaml` |
| `logLevel` | Logging level; valid values are `trace`, `debug`, `info`, `warning`, `error`, `fatal` | `info` |
