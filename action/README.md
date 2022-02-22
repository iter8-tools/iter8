# Run an Iter8 Experiment

This action runs an Iter8 experiment. For more details about Iter8 experiments, see <https://iter8.tools>.

## Usage

Experiments are expressed as helm charts. Create a values file to define experiment specific values. For example to run a load-test experiment against endpoint <http://example.com>:

``` yaml
    - run: |
        cat << EOF > values.yaml
          url: http://example.com
          SLOs:
            error-rate: 0
            mean-latency: 100
            p90: 150
            p97.5: 200
        EOF
    - uses: iter8-tools/iter8@v0.9
      with:
        chart: load-test-http
        valuesFile: ../values.yaml
```

For more options, see <https://iter8.tools>.

## Inputs

| Input Name | Description | Default |
| ---------- | ----------- | ------- |
| `chart` | Location of experiment chart. Must be specified. | None |
| `valuesFile` | Path to file of configuration values. | `values.yaml` |
| `logLevel` | Logging level; valid values are `trace`, `debug`, `info`, `warning`, `error`, `fatal` | `info` |
