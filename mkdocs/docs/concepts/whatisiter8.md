---
template: overrides/main.html
---

# What is Iter8?

You are developing a microservice app / ML model on Kubernetes or OpenShift and have created alternative versions of the app / model. You want to identify the `winning version` using a live experiment and rollout this version reliably.

!!! tip ""
    Enter **Iter8**.

Iter8 is an open source toolkit for progressive delivery, automated rollouts, and metrics and AI-driven experiments on Kubernetes. Using Iter8's AI-driven experimentation capabilities, you can safely orchestrate various experiments on one, two, or more versions of your app / model, gain key insights into their behavior, and rollout the `winner` in a principled, automated, and statistically robust manner.
<!-- Iter8 enables delivery of high-impact code changes within your microservices applications in an agile manner while eliminating the risk.  -->


## What is an Iter8 experiment?

!!! tip ""
    Iter8 defines a Kubernetes resource kind called **Experiment** to automate metrics and AI-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps.

### Features at a glance
Iter8's expressive model of experimentation is designed to support the following.

- Any Kubernetes and OpenShift application stack; stacks that are currently supported with documentation and code-samples are **Knative**, **KFServing**[^1] and **Istio**[^2].
- Testing patterns such as **Conformance**, **Canary**, **A/B**, **A/B/n** and **Pareto**[^3].
- Deployment patterns such as **Progressive**, **FixedSplit** and **BlueGreen**[^4].
- App config tools such as **Helm**, **Kustomize**, and plain **YAML/JSON** app manifests.
- Traffic shaping methods like **mirroring**, **dark launch**, **request routing**, and **sticky sessions**[^5].

Iter8 experiments featuring a few combinations of the above capabilities are illustrated below.

=== "Canary / Progressive / kubectl"
    An Iter8 experiment that automates `Canary` testing, `Progressive` deployment (traffic shifting), and promotion of the `winner` (the `candidate` version in this illustration) using the `kubectl`.
    ![Canary / Progressive / kubectl](/assets/images/canary-progressive-kubectl.png)

## How does Iter8 work?

Iter8 consists of the [etc3 Kubernetes controller](https://github.com/iter8-tools/etc3), the [Iter8 analytics service](https://github.com/iter8-tools/iter8-analytics), and the [Iter8 action/task handler](https://github.com/iter8-tools/handler) which jointly orchestrate experiments. An Iter8 experiment can automate several key functions such as initializing a partially specified experiment, iteratively deciding how to split traffic between app versions, identifying a `winner`, deciding when to terminate the experiment, and promoting the `winner`.

![Canary / Progressive / kubectl](/assets/images/under-the-hood.png)

Under the hood, Iter8 uses advanced Bayesian learning techniques coupled with multi-armed bandit approaches for statistical assessments and decision making.


[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.
[^3]: **A/B**, **A/B/n** and **Pareto** are works-under-progress. [Iter8 for Istio](https://github.com/iter8-tools/iter8) supports **A/B** and **A/B/n** testing patterns.
[^4]: **BlueGreen** deployment is work-under-progress.
[^5]: **Sticky sessions** traffic shaping feature is work-under-progress.