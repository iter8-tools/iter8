# iter8: Analytics-driven canary releases and A/B testing
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## What is iter8 about?

Iter8's mission is to support cloud-native, automated canary releases and A/B testing, driven by analytics based on robust statistical techniques. The project comprises two components:

* _iter8-analytics_: An analytics API server that analyzes metrics from different versions of a microservice. Based on success criteria provided by the user, the metric analysis can assess the health of a canary version both absolutely and comparatively to a baseline version. The analytics API is called by _iter8_controller_ described next.

* _iter8-controller_: A Kubernetes controller that automates canary releases and A/B testing by gradually adjusting the traffic across different versions of a microservice. For instance, in the case of a canary release, the controller will gradually shift the traffic to the canary version if it is performing as expected, until the canary replaces the baseline (previous) version. If the canary is found not to be satisfactory, the controller rolls back by shifting all the traffic to the baseline version. Traffic decisions are made by _iter8-analytics_ and honored by the controller.

## Supported environments

The _iter8-controller_ currently supports the following Kubernetes-based environments, whose traffic-management capabilities are used:

* [Istio service mesh](https://istio.io)
* [Knative](https://knative.dev)

## Integrations

Iter8 is integrated with [Tekton Pipelines](https://tekton.dev) for an end-to-end CI/CD experience, and with [KUI](https://github.com/IBM/kui), for a richer Kubernetes command-line experience. Initial integrations with these two technologies already exist, but we are actively improving them. Stay tuned!

## Installing iter8

The following instructions will guide you to install the two iter8 components (_iter8-analytics_ and _iter8-controller_) on your desired target environment:

* [iter8 on Kubernetes and Istio](doc_files/istio_install)
* [iter8 on Knative](doc_files/knative_install)

## Tutorials

The following tutorials will help you get started with iter8:

* [Canary releases with iter8 on Kubernetes and Istio](doc_files/istio_canary)
* [Canary releases with iter8 on Knative](doc_files/knative_canary)
