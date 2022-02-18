# Iter8: Kubernetes Release Optimizer

[![Iter8 release (latest SemVer)](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
[![Unit test Coverage](https://codecov.io/gh/iter8-tools/iter8/branch/master/graph/badge.svg)](https://codecov.io/gh/iter8-tools/iter8)
[![Unit test Status](https://github.com/iter8-tools/iter8/workflows/tests/badge.svg)](https://github.com/iter8-tools/iter8/actions?query=workflow%3Atests)
[![e2e test Status](https://github.com/iter8-tools/hub/actions/workflows/tests.yaml/badge.svg)](https://github.com/iter8-tools/hub/actions?query=workflow%3Atests)
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
brew install iter8
```

Benchmark an HTTP service.

```shell
iter8 launch -c load-test-http --set url=https://httpbin.org/get
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
1.  [Load test an HTTP service and validate SLOs](https://iter8.tools/0.8/getting-started/your-first-experiment/).
2.  [Control the load characteristics during the HTTP load test experiment](https://iter8.tools/0.8/tutorials/load-test-http/loadcharacteristics/).
3.  [Load test an HTTP POST endpoint with request payload](https://iter8.tools/0.8/tutorials/load-test-http/payload/).
4.  [Learn more about built-in metrics and SLOs in an HTTP load test experiment](https://iter8.tools/0.8/tutorials/load-test-http/metricsandslos/).
5.  [Load test a Knative HTTP service](https://iter8.tools/0.8/tutorials/load-test-http/community/knative/loadtest/).


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
