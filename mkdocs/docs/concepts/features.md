---
template: main.html
hide:
- toc
---

# Iter8 Features at a Glance

Iter8 makes it easy to achieve the following goals.

- Automate releases, validation, and experiments over **any** cloud stack; tutorials available for **Knative**, **KFServing**[^1] and **Istio**[^2].
- Declaratively specify testing and deployment patterns, and metrics-based objectives (SLOs) and indicators (SLIs) for evaluating app/ML model versions using an **experiment** - the Kubernetes custom resource defined by Iter8.
- **Conformance** and **Canary** testing.
- **Progressive** traffic shifting.
- **Dark launches**, **traffic mirroring** and **traffic segmentation**.
- Use Helm, Kustomize, and plain YAML/JSON for defining your app manifests.
- Use metrics from any RESTful provider including **Prometheus**, **New Relic**, **Sysdig**, and **Elastic**.
- Statistically rigorous evaluation of versions, traffic splitting, and promotion/rollback decisions using **Bayesian learning** and **multi-armed bandit** algorithms.
- Observe experiments in realtime.


[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8-istio). An updated version is coming soon.
