# :zap: Iter8: Kubernetes Release Optimizer

[![Iter8 release](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
[![Unit test Coverage](https://codecov.io/gh/iter8-tools/iter8/branch/master/graph/badge.svg)](https://codecov.io/gh/iter8-tools/iter8)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/iter8-tools/iter8/tests?label=Unit%20tests)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/iter8-tools/hub/tests?label=Integration%20tests)

> - Safely rollout apps
> - Maximize business value
> - Use with any app/serverless/ML framework
> - Simplify CI/CD/GitOps
> - Get started in seconds

<p align='center'>
<img alt-text="Iter8 experiment" src="https://iter8-tools.github.io/docs/0.9/images/iter8-intro-dark.png" width="70%" />
</p>

## :rocket: Getting started

### Using Iter8 Action

Configure experiment with Helm values file and call:

``` yaml
    - run: |
        cat << EOF > expConfig.yaml
          url: https://httpbin.org/get
          SLOs:
            http/error-rate: 0
            http/mean-latency: 100
        EOF
    - uses: iter8-tools/iter8@v0.9
      with:
        chart: load-test-http
        valuesFile: expConfig.yaml
        validateSLOs: true
```

#### Action Inputs

| Input Name | Description | Default |
| ---------- | ----------- | ------- |
| `chart` | Location of experiment chart. Must be specified. | None |
| `valuesFile` | Path to file of configuration values. | None |
| `validateSLOs` | Validate any specified SLOs. | `true` |
| `chartVersion` | Chart version constraint | None |
| `logLevel` | Logging level; valid values are `trace`, `debug`, `info`, `warning`, `error`, `fatal` | `info` |

### Using Iter8 CLI

#### Install Iter8 CLI
```shell
brew tap iter8-tools/iter8
brew install iter8@0.9
```
[See here](https://iter8.tools/latest/getting-started/install) for more ways to install.

#### Benchmark an HTTP service
Launch Iter8 experiment and view report.
```shell
iter8 launch -c load-test-http --set url=https://httpbin.org/get
iter8 report
```

#### Benchmark a gRPC service
Start a sample gRPC service in a separate terminal.

```shell
docker run -p 50051:50051 docker.io/grpc/java-example-hostname:latest
```

Launch Iter8 experiment and view report.
```shell
iter8 launch -c load-test-grpc \
--set host="127.0.0.1:50051" \
--set call="helloworld.Greeter.SayHello" \
--set protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto"
iter8 report
```

## :dart: Usage examples
1.  [Load test, benchmark and validate HTTP services with SLOs](https://iter8.tools/0.9/tutorials/load-test-http/usage/).
2.  [Load test, benchmark and validate gRPC services with SLOs](https://iter8.tools/0.9/tutorials/load-test-grpc/usage/).
3.  Load test, benchmark and validate Knative services with SLOs: [HTTP](https://iter8.tools/0.9/tutorials/integrations/knative/load-test-http/) and [gRPC](https://iter8.tools/0.9/tutorials/integrations/knative/load-test-grpc/).

### Documentation
Iter8 documentation is available at https://iter8.tools.

## :wrench: Developing Iter8
We welcome PRs!

See [here](CONTRIBUTING.md) for information about ways to contribute, Iter8 community meetings, finding an issue, asking for help, pull-request lifecycle, and more.

## :hibiscus: Credits
Iter8 is primarily written in `Go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [ghz](https://ghz.sh)
- [Fortio](https://github.com/fortio/fortio)
- [plotly.js](https://github.com/plotly/plotly.js)

