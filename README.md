# Iter8: Kubernetes Release Optimizer

[![Iter8 release](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)

Iter8 is the Kubernetes release optimizer built for DevOps, MLOps, SRE and data science teams. Iter8 makes it easy to ensure that Kubernetes apps and ML models perform well and maximize business value. 

Iter8 supports the following use-cases:

1.  Performance testing of HTTP services.
2.  Performance testing of gRPC services.
3.  A/B/n testing of applications and ML models
4.  Reliable and automated routing: blue-green and canary

## :rocket: Iter8 performance test

Iter8 introduces a set of tasks which which can be composed in order to conduct a variety of performance tests.

<p align='center'>
<img alt-text="Iter8 performance test" src="https://iter8-tools.github.io/docs/0.9/images/iter8-intro-dark.png" width="70%" />
</p>

Iter8 packs a number of powerful features that facilitate Kubernetes app testin. They include the following.

1.  **Generating load and collecting built-in metrics for HTTP and gRPC services.** Simplifies performance testing by eliminating the need to setup and use metrics databases.
2.  **Readiness check.** The performance testing portion can be configured to start only after the service is ready.
3.  **Test anywhere.** Iter8 performance tests can be launched inside a Kubernetes cluster, in local environments, or inside a GitHub Actions pipeline.
4.  **Traffic controller.** Automatically and dynamically reconfigures routing resources based on the state of Kubernetes apps/ML models.
5. **Client-side SDK.** Facilitates routing and metrics collection task associated with distributed (i.e., client-server architecture-based) A/B/n testing in Kubernetes.

Please see [https://iter8.tools](https://iter8.tools) for the complete documentation.

## :maple_leaf: Issues
Iter8 issues are tracked [here](https://github.com/iter8-tools/iter8/issues).

## :tada: Contributing
We welcome PRs!

See [here](CONTRIBUTING.md) for information about ways to contribute, finding an issue, asking for help, pull-request lifecycle, and more.

## :hibiscus: Credits
Iter8 is primarily written in `Go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [ghz](https://ghz.sh)
- [Fortio](https://github.com/fortio/fortio)
- [plotly.js](https://github.com/plotly/plotly.js)