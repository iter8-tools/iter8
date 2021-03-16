---
template: overrides/main.html
---

# What is Iter8?

**Iter8** is an AI-powered platform for cloud native release automation and experimentation. Iter8 makes it easy to unlock business value and guarantee SLOs by identifying the top versions of your apps/ML models and rolling them out safely.

Use Iter8 to automate progressive delivery, validation, and promotion/rollback of new versions, and maximize release velocity with confidence while protecting end-user experience.

## Features at a glance

- Release automation and experimentation on **any** cloud stack; stacks that are currently supported with documented code-samples are **Knative**, **KFServing**[^1] and **Istio**[^2].
- **Conformance** and **Canary** testing.
- **Progressive**, **FixedSplit**, and **DarkLaunch** deployments.
- Traffic shaping methods such as **mirroring** and **traffic segmentation**.
- Integration with app config tools such as **Helm**, **Kustomize**, and **kubectl**.
- **Metrics-based criteria** in experiments for evaluating app/model versions.
- Support for **Prometheus** metrics backend.
- Support for **custom metrics** based on any metric available in **Prometheus**.
- Statistically robust and principled assessment and decision making during experiments using **Bayesian learning** and **multi-armed bandit** algorithms.
- The **iter8ctl** CLI for observing experiments in realtime.

## How does Iter8 work?

Iter8 consists of a [Kubernetes controller](https://github.com/iter8-tools/etc3), an [analytics service](https://github.com/iter8-tools/iter8-analytics), and an [action/task handler](https://github.com/iter8-tools/handler) which jointly orchestrate experiments. These components automate several functions including executing start up tasks that initialize a partially specified experiment, verifying that conditions needed for the experiment are satisfied, iteratively deciding how to split traffic between app versions, identifying a `winner`, error handling, deciding when to terminate the experiment, promoting the `winner`, and executing clean up tasks.

<!-- ??? info "Deeper look into Iter8's component interactions"
    ![Under the hood](/assets/images/under-the-hood.png) -->

[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.