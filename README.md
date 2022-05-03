# Iter8: Metrics-Driven Release Optimizer

[![Iter8 release](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/iter8-tools/iter8/tests?label=Unit%20tests)

> Iter8 is a metrics-driven release optimizer built for DevSecOps, MLOps, SRE and data science teams. Iter8 makes it easy to ensure that new versions of apps and ML models perform well, are secure, and maximize business value.

<p align='center'>
<img alt-text="Iter8 experiment" src="https://iter8-tools.github.io/docs/0.9/images/iter8-intro-dark.png" width="70%" />
</p>

## :rocket: Features at a glance
1. **Load test, benchmark and validate HTTP services with SLOs**
    * [In your local environment](https://iter8.tools/0.10/tutorials/load-test-http/basicusage/)
    * [Inside a Kubernetes cluster](https://iter8.tools/0.10/tutorials/load-test-http/kubernetesusage/)
2. **Load test, benchmark and validate gRPC services with SLOs**
    * [In your local environment](https://iter8.tools/0.10/tutorials/load-test-grpc/basicusage/)
    * [Inside a Kubernetes cluster](https://iter8.tools/0.10/tutorials/load-test-grpc/kubernetesusage/)

Please see [https://iter8.tools](https://iter8.tools) for the complete documentation.

## :white_check_mark: Installing Iter8 using GitHub Actions

Install the latest version of Iter8 using the GitHub Action `iter8/iter8@v0.10`. A specific version can be installed using the version as the action reference. For example, to install version v0.10.15, use `iter8/iter8@v0.10.15`.

Once Iter8 is installed, it can be used as documented (see [https://iter8.tools](https://iter8.tools)) in `run` actions. For example:

```yaml
- uses: iter8/iter8@v0.10 # install Iter8
- run: |
    iter8 version
    iter8 launch -c load-test-http --set url=http://httpbin.org/get
```

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

