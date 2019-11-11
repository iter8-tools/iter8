# iter8: Analytics-driven canary releases and A/B testing
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## What is iter8 about?

<img src=img/iter8-overview.png width=50%>

Iter8 supports cloud-native, automated canary releases and A/B testing, driven by analytics based on robust statistical techniques. It comprises two components:

* _iter8-analytics_: A service that assesses the behavior of different microservice versions by (1) looking at metrics associated with each version and (2) applying robust statistical techniques to analyze the data online in order to determine which version is the best one with respect to the metrics of interest and which versions pass a set of success criteria. Multiple success criteria can be defined by the users; each criterion can refer to a different metric and specify absolute acceptable thresholds as well as comparative thresholds meant to define how much a candidate version can deviate from a baseline (stable) version. The _iter8-analytics_ service exposes a REST API; each time it is called, the service returns the result of the data analysis along with a recommendation for how the traffic should be split across all microservice versions. The _iter8-analytics_' REST API is used by _iter8-controller_, which is described next.

* _iter8-controller_: A Kubernetes controller that automates canary releases and A/B testing by gradually adjusting the traffic across different versions of a microservice as recommended by _iter8-analytics_. For instance, what happens in the case of a canary release is that the controller will gradually shift the traffic to the canary version if it is performing as expected, until the canary replaces the baseline (previous) version. If the canary is found not to be satisfactory, the controller rolls back by shifting all the traffic to the baseline version. Traffic decisions are made by _iter8-analytics_ and honored by _iter8-controller_.

## The `Experiment` CRD

When iter8 is installed, a new Kubernetes CRD is added to your cluster. This CRD kind is `Experiment` and it is documented [here](doc_files/iter8_crd.md).

## Supported environments

The _iter8-controller_ currently supports the following Kubernetes-based environments, whose traffic-management capabilities are used:

* [Istio service mesh](https://istio.io)
* [Knative](https://knative.dev)

## Installing iter8

These [instructions](doc_files/iter8_install.md) will guide you to install the two iter8 components (_iter8-analytics_ and _iter8-controller_) on Kubernetes with Istio and/or Knative.

## Tutorials

The following tutorials will help you get started with iter8:

* [Automated canary releases with iter8 on Kubernetes and Istio](doc_files/iter8_bookinfo_istio.md)
* [Automated canary releases with iter8 on Knative](doc_files/iter8_bookinfo_knative.md)
* [Automated canary release with iter8 on Kubernetes and Istio using Tekton](doc_files/iter8_tekton_task.md)

## Integrations

Iter8 is integrated with [Tekton Pipelines](https://tekton.dev) for an end-to-end CI/CD experience, and with [KUI](https://www.kui.tools), for a richer Kubernetes command-line experience. Initial integrations with these two technologies already exist, but we are actively improving them. Stay tuned!