---
template: overrides/main.html
---

# What is Iter8?

You are developing a distributed microservices-based application on Kubernetes and have created alternative versions of a service. You want to identify the `winning version` of your service using a live experiment and rollout this version in a safe and reliable manner.

!!! tip ""
    Enter **Iter8**.

Iter8 is an open source toolkit for progressive delivery, automated rollouts, and metrics and AI-driven experiments on Kubernetes. Iter8 enables delivery of high-impact code changes within your microservices applications in an agile manner while eliminating the risk. Using Iter8's AI-driven experimentation capabilities, you can safely and rapidly orchestrate various types of live experiments, gain key insights into the behavior of your microservices, and rollout the best versions of your microservices in an automated, principled, and statistically robust manner.

## What is an Iter8 experiment?

!!! tip ""
    Iter8 defines a Kubernetes resource kind called **Experiment** to automate metrics and AI-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps.

Iter8's expressive model of experimentation is designed to support the following.

- Diverse Kubernetes and OpenShift application stacks; currently supported stacks are **Knative**, **KFServing**[^1] and **Istio**[^2],
- Testing patterns such as **Conformance**, **Canary**, **A/B**, **A/B/n** and **Pareto**[^3], 
- Deployment patterns such as **Progressive**, **FixedSplit** and **BlueGreen**[^4],
- App config tools such as **Helm**, **Kustomize**, and plain **YAML/JSON** app manifests, and
- Traffic shaping methods like **mirroring**, **dark launch**, **request routing**, and **sticky sessions**[^5].

Iter8 experiments featuring a few combinations of the above capabilities are illustrated below.

## How does Iter8 work?

Iter8 is composed of a Kubernetes controller, an analytics service and a task runner framework which are jointly responsible for orchestrating an experiment, and making iterative decisions about how to setup the experiment, how to shift application traffic during the experiment, how to identify a `winner`, when to terminate the experiment, and how to promote the `winner`. 

Under the hood, Iter8 uses advanced Bayesian learning techniques coupled with multi-armed bandit approaches for statistical assessments and decision making.


[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.
[^3]: **A/B**, **A/B/n** and **Pareto** are works-under-progress. [Iter8 for Istio](https://github.com/iter8-tools/iter8) supports **A/B** and **A/B/n** testing patterns.
[^4]: **BlueGreen** deployment is work-under-progress.
[^5]: **Sticky sessions** traffic shaping feature is work-under-progress.