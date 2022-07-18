# Iter8: Metrics-Driven Release Optimizer

[![Iter8 release](https://img.shields.io/github/v/release/iter8-tools/iter8?sort=semver)](https://github.com/iter8-tools/iter8/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/iter8-tools/iter8)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/iter8-tools/iter8/tests?label=Unit%20tests)

Iter8 is the Kubernetes release optimizer built for DevOps, MLOps, SRE and data science teams. Iter8 makes it easy to ensure that Kubernetes apps and ML models perform well and maximize business value. Use Iter8 for the following use-cases.

1.  Performance testing and SLO validation of HTTP services.
2.  Performance testing and SLO validation of gRPC services.
3.  SLO validation using custom metrics from any database.

## :rocket: Iter8 experiment

Iter8 introduces the notion of an experiment, which is a set of configurable tasks that are executed in a specific sequence.

<p align='center'>
<img alt-text="Iter8 experiment" src="https://iter8-tools.github.io/docs/0.9/images/iter8-intro-dark.png" width="70%" />
</p>

Iter8 packs a number of powerful features that facilitate Kubernetes app testing and experimentation. They include the following.

1.  **Generating load and collecting built-in metrics for HTTP and gRPC services.** Simplifies performance testing by eliminating the need to setup and use metrics databases.
2.  **Well-defined notion of service-level objectives (SLOs).** Makes it simple to define and verify SLOs in experiments.
3.  **Custom metrics.** Enables the use of custom metrics from any database(s) or REST API(s) in experiments.
4.  **Readiness checks.** The performance testing portion of the experiment begins only after the service is ready.
5.  **HTML/text reports.** Promotes human understanding of experiment results through visual insights.
6.  **Assertions.** Verifies whether the target app satisfies the specified SLOs or not after an experiment. Simplifies automation in CI/CD/GitOps pipelines: branch off into different paths depending upon whether the assertions are true or false.
7.  **Multi-loop experiments.** Experiment loops are executed periodically. Experiment tasks (such as collecting metrics and assessing app versions) are evaluated during each loop of the experiment, and experiment results are updated with the latest assessments.

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

