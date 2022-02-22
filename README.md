# Iter8: Kubernetes Release Optimizer

[![Iter8 release](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
[![Unit test Coverage](https://codecov.io/gh/iter8-tools/iter8/branch/master/graph/badge.svg)](https://codecov.io/gh/iter8-tools/iter8)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/iter8-tools/iter8/tests?label=Unit%20tests)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/iter8-tools/hub/tests?label=Integration%20tests)
[![Slack channel](https://img.shields.io/badge/Slack-Join-purple)](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
[![Community meetups](https://img.shields.io/badge/meet-Iter8%20community%20meetups-brightgreen)](https://iter8.tools/0.8/getting-started/help/)

> Kubernetes release optimizer built for DevOps, MLOps, SRE and data science teams.

<p align='center'>
<img alt-text="Iter8 experiment" src="https://iter8-tools.github.io/docs/0.9/images/iter8-intro-dark.png" width="70%" />
</p>

> Use Iter8 experiments to safely rollout apps and ML models, and maximize business value with each release. Use with any app/serverless/ML framework.

## Quick start

Install Iter8 CLI. [See here](https://iter8.tools/latest/getting-started/install) for more ways to install.

```shell
brew tap iter8-tools/iter8
brew install iter8@0.9
```

### Benchmark an HTTP service
```shell
iter8 launch -c load-test-http --set url=https://httpbin.org/get
iter8 report
```

### Benchmark a gRPC service
Start a sample gRPC service in a separate terminal.

```shell
docker run -p 50051:50051 docker.io/grpc/java-example-hostname:latest
```

Launch Iter8 experiment.
```shell
iter8 launch -c load-test-grpc \
--set host="127.0.0.1:50051" \
--set call="helloworld.Greeter.SayHello" \
--set protoURL="https://raw.githubusercontent.com/grpc/grpc-go/master/examples/helloworld/helloworld/helloworld.proto"
iter8 report
```

## Features at a glance

* **Load testing with SLOs**

  Iter8 experiments can generate requests for HTTP and gRPC services, collect built-in latency and error-related metrics, and validate SLOs.

* **A/B(/n) testing**

  Grow your business with every release. Iter8 experiments can compare multiple versions based on business value and identify a winner.

* **Use anywhere**

  Get started with Iter8 in seconds using pre-packaged experiment charts. Run Iter8 experiments locally, inside Kubernetes, or inside your CI/CD/GitOps pipelines.

* **App frameworks**
    
  Use with any app, serverless, or ML framework. Iter8 works with Kubernetes deployments, statefulsets, Knative services, KServe/Seldon ML deployments, or other custom Kubernetes resource types.

## Usage Examples
1.  [Load test, benchmark and validate HTTP services](https://iter8.tools/0.9/tutorials/load-test-http/usage/).
2.  [Load test, benchmark and validate gRPC services](https://iter8.tools/0.9/tutorials/load-test-grpc/usage/).
3.  Load test, benchmark and validate Knative services: [HTTP](https://iter8.tools/0.9/tutorials/integrations/knative/load-test-http/) and [gRPC](https://iter8.tools/0.9/tutorials/integrations/knative/load-test-grpc/)


## Documentation
Iter8 documentation is available at https://iter8.tools.

## Credits
Iter8 is primarily written in `Go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [ghz](https://ghz.sh)
- [Fortio](https://github.com/fortio/fortio)
- [plotly.js](https://github.com/plotly/plotly.js)

## Contributing
See [here](CONTRIBUTING.md) for information about ways to contribute, Iter8 community meetings, finding an issue, asking for help, pull-request lifecycle, and more.
