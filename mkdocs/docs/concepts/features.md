---
template: overrides/main.html
hide:
- toc
---

# Iter8 Features at a Glance

Iter8 enables app/ML model developers, service operators, SREs, ML engineers, and data scientists to achieve the following goals.

- Automate validation/releases over **any** cloud stack; tutorials are documented for **Knative**, **KFServing**[^1] and **Istio**[^2].
- Declaratively specify validation/release goals using an **experiment** - Kubernetes custom resource defined by Iter8.
- **Conformance** and **Canary** testing.
- **Progressive**, **fixed-split**, and **dark-launched** deployments.
- Specify SLOs and SLIs.
- **Traffic mirroring** and **traffic segmentation**.
- Use Helm, Kustomize, and plain YAML/JSON manifests.
- Use Out-of-the-box metrics custom metrics that can be defined using metrics in **Prometheus**.
- Statistically rigorous evaluation of versions, traffic splitting, and promotion/rollback decisions using **Bayesian learning** and **multi-armed bandit** algorithms.
- Observe experiments in realtime.


[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.