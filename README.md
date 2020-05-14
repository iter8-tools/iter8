# iter8: Analytics-driven canary releases and A/B testing
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## What is iter8 about?

<img src=img/iter8-overview.png width=65%>

Iter8 supports cloud-native, automated canary releases and A/B testing, driven by analytics based on robust statistical techniques. It comprises two components:

* _iter8-analytics_: A service that assesses the behavior of different microservice versions by analyzing metrics associated with each version using robust statistical techniques to determine which version is the best one with respect to the metrics of interest and which versions pass a set of success criteria. Multiple success criteria can be defined by users; each criterion refers to a metric and specifies absolute or relative thresholds which define how much a candidate version can deviate from a baseline (stable) version. The _iter8-analytics_ service exposes a REST API; each time it is called, the service returns the result of the data analysis along with a recommendation for how the traffic should be split across all microservice versions. The _iter8-analytics_' REST API is used by _iter8-controller_, which is described next.

* _iter8-controller_: A Kubernetes controller that automates canary releases and A/B testing by adjusting the traffic across different versions of a microservice as recommended by _iter8-analytics_. For instance, what happens in the case of a canary release is that the controller will shift the traffic towards the canary version if it is performing as expected, until the canary replaces the baseline (previous) version. If the canary is found not to be satisfactory, the controller rolls back by shifting all the traffic to the baseline version. Traffic decisions are made by _iter8-analytics_ and honored by _iter8-controller_.

## The `Experiment` CRD

When iter8 is installed, a new Kubernetes CRD is added to your cluster. This CRD _kind_ is `Experiment` and it is documented [here](doc_files/iter8_crd.md).

## Metrics

To assess the behavior of microservice versions, iter8 supports a few metrics out-of-the-box without requiring users to do any extra work. In addition, users can define their own custom metrics. Iter8's out-of-the-box metrics as well as user-defined metrics can be referenced in the success criteria of an _experiment_. More details about metrics are documented [here](doc_files/metrics.md).

## Supported environments

The _iter8-controller_ currently uses the [Istio service mesh](https://istio.io) traffic management capabilities to automate the user traffic split across multiple microservice versions.

## Installing iter8

These [instructions](doc_files/iter8_install.md) will guide you to install the two iter8 components (_iter8-analytics_ and _iter8-controller_) on Kubernetes with Istio.

## Tutorials

The following tutorials will help you get started with iter8:

* [Automated canary releases with iter8 on Kubernetes and Istio](doc_files/iter8_bookinfo_istio.md)
* [Automated canary release with iter8 on Kubernetes and Istio using Tekton](doc_files/iter8_tekton_task.md)

## Algorithms behind iter8

A key goal of this project is to introduce statistically robust algorithms for decision making during cloud-native canary releases and A/B testing experiments. We currently support [four algorithms](doc_files/algorithms.md).

## Integrations

Iter8 is integrated with [Tekton Pipelines](https://tekton.dev) for an end-to-end CI/CD experience, and with [KUI](https://www.kui.tools), for a richer Kubernetes command-line experience. Initial integrations with these two technologies already exist, but we are actively improving them.

**The upcoming iter8 1.0.0 release will feature a deep integration with KUI. Stay tuned!**

## Releases

The current iter8 release is **v0.1.0**. This documentation always points to the current release. Documentation of older releases can be found [here](releases.md).

**Release 1.0.0 is coming soon! Stay tuned!!**
