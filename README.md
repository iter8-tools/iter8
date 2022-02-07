# Iter8: Kubernetes Release Optimizer

[![Iter8 release (latest SemVer)](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
[![Test Status](https://github.com/iter8-tools/iter8/workflows/tests/badge.svg)](https://github.com/iter8-tools/iter8/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/iter8-tools/iter8/branch/master/graph/badge.svg)](https://codecov.io/gh/iter8-tools/iter8)
[![Slack channel](https://img.shields.io/badge/Slack-Join-purple)](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
[![Community meetups](https://img.shields.io/badge/meet-Iter8%20community%20meetups-brightgreen)](https://iter8.tools/0.8/getting-started/help/)


# What is Iter8?

Kubernetes release optimizer built for DevOps and MLOps teams.

## What is an Iter8 experiment?
Iter8 experiments make it simple to collect performance and business metrics for apps and ML models, assess, compare and validate multiple app/ML model versions, safely rollout winning version, and maximize business value in each release.

<p align='center'>
<img alt-text="Iter8 experiment" src="mkdocs/docs/images/iter8-intro-dark.png" width="70%" />
</p>

### Experiment chart
Specialized [Helm charts](https://helm.sh/docs/topics/charts/) that contain reusable experiment templates. Iter8 combines experiment charts with user supplied values to generate runnable `experiment.yaml` files.

#### Iter8 Hub
A specific location within in the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8) that hosts several pre-packaged and reusable charts. These charts enable to you to launch powerful release optimization experiments in seconds. Their usage is described in depth in various [Iter8 tutorials](https://iter8.tools/latest/tutorials/load-test-http/overview/).

## Features at a glance

- **Load testing with SLOs** 
    
    Iter8 experiments can generate requests for HTTP and gRPC services, collect built-in latency and error-related metrics, and validate SLOs.

- **A/B(/n) testing** 
      
    Grow your business with every release. Iter8 experiments can compare multiple versions based on business value and promote a winner.

- **Simple to use** 
      
    Get started with Iter8 in seconds using pre-packaged experiment charts. Run Iter8 experiments locally, inside Kubernetes, or inside your CI/CD/GitOps pipelines.

- **K8s app/serverless/ML frameworks** 
      
    Use with any app, serverless, or ML framework. Iter8 works with Kubernetes deployments, statefulsets, Knative services, KServe/Seldon ML deployments, or custom Kubernetes resource types.

## Installation
Install the latest stable release of the Iter8 CLI using `brew` as follows.

```shell
brew tap iter8-tools/iter8
brew install iter8
```

You can also install Iter8 using:
* [pre-compiled binaries](https://iter8.tools/latest/getting-started/install/)
* [source](https://iter8.tools/latest/getting-started/install/)
* [`go 1.17+`](https://iter8.tools/latest/getting-started/install/)

## Usage

### HTTP load test with SLOs

1.  [Load test an HTTP service and validate SLOs](https://iter8.tools/0.8/getting-started/your-first-experiment/).
2.  [Control the load characteristics during the HTTP load test experiment](https://iter8.tools/0.8/tutorials/load-test-http/loadcharacteristics/).
3.  [Load test an HTTP POST endpoint with request payload](https://iter8.tools/0.8/tutorials/load-test-http/payload/).
4.  [Learn more about built-in metrics and SLOs in an HTTP load test experiment](https://iter8.tools/0.8/tutorials/load-test-http/metricsandslos/).
5.  [Load test a Knative HTTP service](https://iter8.tools/0.8/tutorials/load-test-http/community/knative/loadtest/).


## Documentation
Iter8 documentation is available at https://iter8.tools.

## Implementation
Iter8 is primarily written in `go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [Fortio](https://github.com/fortio/fortio)
- [ghz](https://ghz.sh)
- [plotly.js](https://github.com/plotly/plotly.js)


## Contributing
See [here](CONTRIBUTING.md) for information about ways to contribute, Iter8 community meetings, finding an issue, asking for help, pull-request lifecycle, and more.

