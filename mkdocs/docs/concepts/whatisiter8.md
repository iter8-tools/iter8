---
template: overrides/main.html
---

# What is Iter8?

You have developed multiple versions of a microservice app or an ML model. You want to identify the **winning version** and rollout the `winner` in a reliable manner.

!!! tip "Enter **Iter8**"
    **Iter8** helps you to unlock business value in your k8s apps and ML models rapidly by bringing together the best of AI, release engineering, and observability. 

Use Iter8's AI-driven experimentation capabilities to safely experiment with new versions of your apps and ML models on Kubernetes and OpenShift, gain key insights into their behavior with user requests, progressively shift traffic, and rollout the `winner` in a principled and robust manner.

## What is an Iter8 experiment?

!!! tip ""
    Iter8 defines a Kubernetes resource kind called **Experiment** that automates metrics and AI-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps / ML models.

A basic Iter8 experiment that automates `Canary` testing and `Progressive` deployment (traffic shifting) is illustrated below.

![Canary / Progressive / kubectl](/assets/images/canary-progressive-kubectl.png)

## Features at a glance

- Experimentation on any Kubernetes and OpenShift stack; stacks that are currently supported with documented code-samples are **Knative**,[^1] **KFServing**[^2] and **Istio**[^3].
- Testing patterns such as **Conformance**, **Canary**, **A/B**, **A/B/n** and **Pareto**[^4].
- Deployment patterns such as **Progressive**, **FixedSplit**, **DarkLaunch**, and **BlueGreen**[^5].
- Traffic shaping methods such as **mirroring**, **request routing**, and **sticky sessions**[^6].
- App config tools such as **Helm**, **Kustomize**, and plain **YAML/JSON** app manifests.
- **Metrics-based criteria** in experiments for evaluating app/model versions.
- Support for **Prometheus** metrics backend.
- Support for **custom metrics** based on any metric available in the backend.
- Statistically robust and principled assessment of app versions, traffic shifting, and version promotion using **Bayesian learning** and **multi-armed bandit algorithms**.
- The **iter8ctl** CLI for observing experiments in realtime.

## How does Iter8 work?

Iter8 consists of a [Kubernetes controller](https://github.com/iter8-tools/etc3), an [analytics service](https://github.com/iter8-tools/iter8-analytics), and an [action/task handler](https://github.com/iter8-tools/handler) which jointly orchestrate experiments. These components automate several functions including executing start up tasks that initialize a partially specified experiment, verifying that conditions needed for the experiment are satisfied, iteratively deciding how to split traffic between app versions, identifying a `winner`, error handling, deciding when to terminate the experiment, promoting the `winner`, and executing clean up tasks.

??? info "Deeper look into Iter8's component interactions"
    ![Under the hood](/assets/images/under-the-hood.png)

[^1]: Iter8 for Knative is supported on *Istio*, *Contour*, *Kourier*, and *Gloo* networking layers.
[^2]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^3]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.
[^4]: **A/B**, **A/B/n** and **Pareto** are works-under-progress. [Iter8 for Istio](https://github.com/iter8-tools/iter8) supports **A/B** and **A/B/n** testing patterns.
[^5]: **BlueGreen** deployment is work-under-progress.
[^6]: **Sticky sessions** traffic shaping feature is work-under-progress.
