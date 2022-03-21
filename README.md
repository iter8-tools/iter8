# Iter8: Kubernetes Release Optimizer

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

## :dart: Features at a glance

**Load testing with SLOs**

  > Iter8 experiments can generate requests for HTTP and gRPC services, collect built-in latency and error-related metrics, and validate SLOs.

**A/B(/n) testing**

  > Grow your business with every release. Iter8 experiments can compare multiple versions based on business value and identify a winner.

**Simple to use**

  > Get started with Iter8 in seconds using pre-packaged experiment charts. Run Iter8 experiments locally, inside Kubernetes, or inside your CI/CD/GitOps pipelines.

**App frameworks**

  > Use with any app, serverless, or ML framework. Iter8 works with Kubernetes deployments, statefulSets, Knative services, KServe/Seldon ML deployments, or other custom Kubernetes resource types.

## :rocket: Usage examples
1.  [Load test, benchmark and validate HTTP services with SLOs](https://iter8.tools/0.9/tutorials/load-test-http/usage/).
2.  [Load test, benchmark and validate gRPC services with SLOs](https://iter8.tools/0.9/tutorials/load-test-grpc/usage/).
3.  Performance testing and SLO validation using Iter8 GitHub Action: [HTTP](https://iter8.tools/0.9/tutorials/load-test-http/ghaction/) and [gRPC](https://iter8.tools/0.9/tutorials/load-test-grpc/ghaction/).
4.  Performance testing and SLO validation for Knative services: [HTTP](https://iter8.tools/0.9/tutorials/integrations/knative/load-test-http/) and [gRPC](https://iter8.tools/0.9/tutorials/integrations/knative/load-test-grpc/).

Please see [https://iter8.tools](https://iter8.tools) for the complete documentation.

## :maple_leaf: Issues

Iter8 issues are tracked [here](https://github.com/iter8-tools/iter8/issues).

## :tada: Contributing
We welcome PRs!

See [here](CONTRIBUTING.md) for information about ways to contribute, Iter8 community meetings, finding an issue, asking for help, pull-request lifecycle, and more.

## :hibiscus: Credits
Iter8 is primarily written in `Go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [ghz](https://ghz.sh)
- [Fortio](https://github.com/fortio/fortio)
- [plotly.js](https://github.com/plotly/plotly.js)

