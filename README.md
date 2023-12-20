# Iter8: Kubernetes Release Optimizer

[![Iter8 release](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)

Iter8 is the Kubernetes release optimizer built for DevOps, MLOps, SRE and data science teams. Iter8 makes it easy to ensure that Kubernetes apps and ML models perform well and maximize business value. 

Iter8 supports the following use-cases:

1.  Progressive release with automated traffic management
2.  A/B/n testing with a client SDK and business metrics
3.  Performance testing for HTTP and gRPC endpoints

Any Kubernetes resource type, included CRDs can be used with Iter8.

## :rocket: Features

Iter8 introduces a set of tasks which which can be composed in order to conduct tests.

<p align='center'>
<img alt-text="Iter8 performance test" src="https://iter8-tools.github.io/docs/0.18/images/iter8-intro-dark.png" width="70%" />
</p>

Iter8 packs a number of powerful features that facilitate Kubernetes application and ML model testing. They include the following:

1.  **Use any resource types.** Iter8 is easily extensible so that an application being tested can be composed of any resource types including CRDs.
2. **Client SDK.** A client SDK enables application frontend components to reliably associate business metrics with the contributing version of the backend thereby enabling A/B/n testing of backends.
3. **Composable test tasks.** Performance test tasks include load generation and metrics storage simplifing setup.

Please see [https://iter8.tools](https://iter8.tools) for the complete documentation.

## :maple_leaf: Issues
Iter8 issues are tracked [here](https://github.com/iter8-tools/iter8/issues).

## :tada: Contributing
We welcome PRs!

See [here](CONTRIBUTING.md) for information about ways to contribute, finding an issue, asking for help, pull-request lifecycle, and more.

## :hibiscus: Credits
Iter8 is primarily written in `Go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [Istio](https://istio.io)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
- [Fortio](https://github.com/fortio/fortio)
- [ghz](https://ghz.sh)
- [Grafana](https://grafana.com/)
