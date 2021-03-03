---
template: overrides/main.html
---

# What is Iter8?

You have developed multiple versions of a microservice app or an ML model. You want to identify the **winning version** and rollout this version reliably in production.

!!! tip ""
    Enter **Iter8**.

Iter8 is an open source toolkit for **progressive delivery**, **automated rollouts**, and **metrics and AI-driven experiments** on Kubernetes and OpenShift. 

Use Iter8's AI-driven experimentation capabilities to safely orchestrate experiments on one, two, or more versions of your app / model, gain key insights into their behavior, and (optionally) shift traffic and rollout the `winner` in a principled, automated, and statistically robust manner.
<!-- Iter8 enables delivery of high-impact code changes within your microservices applications in an agile manner while eliminating the risk.  -->


## What is an Iter8 experiment?

!!! tip ""
    Iter8 defines a Kubernetes resource kind called **Experiment** to automate metrics and AI-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps / ML models.

A basic Iter8 experiment that automates `Canary` testing and `Progressive` deployment (traffic shifting) is illustrated below.

![Canary / Progressive / kubectl](/assets/images/canary-progressive-kubectl.png)

## Features at a glance

Iter8 provides a powerful and expressive framework for experimentation. Combine the following features within experiments in a rapid and flexible manner.

- Experimentation on **any** Kubernetes and OpenShift stack; stacks that are currently supported with documented code-samples are **Knative**, **KFServing**[^1] and **Istio**[^2].
- Testing patterns such as **Conformance**, **Canary**, **A/B**, **A/B/n** and **Pareto**[^3].
- Deployment patterns such as **Progressive**, **FixedSplit** and **BlueGreen**[^4].
- Traffic shaping methods like **mirroring**, **dark launch**, **request routing**, and **sticky sessions**[^5].
- App config tools such as **Helm**, **Kustomize**, and plain **YAML/JSON** app manifests.
- Statistically robust assessment of app versions, traffic shifting, and version promotion. Iter8's analytics engine implements **Bayesian learning** and **multi-armed bandit algorithms** for principled decision making during experiments.
- The **iter8ctl** CLI for observing experiments in realtime.

## How does Iter8 work?

Iter8 consists of a [Kubernetes controller](https://github.com/iter8-tools/etc3), an [analytics service](https://github.com/iter8-tools/iter8-analytics), and an [action/task handler](https://github.com/iter8-tools/handler) which jointly orchestrate experiments. These components automate several functions including executing start up tasks that initialize a partially specified experiment, verifying that conditions needed for the experiment are satisfied, iteratively deciding how to split traffic between app versions, identifying a `winner`, error handling, deciding when to terminate the experiment, promoting the `winner`, and executing clean up tasks.

Interactions between these components during an experiment is illustrated below.

![Under the hood](/assets/images/under-the-hood.png)

[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.
[^3]: **A/B**, **A/B/n** and **Pareto** are works-under-progress. [Iter8 for Istio](https://github.com/iter8-tools/iter8) supports **A/B** and **A/B/n** testing patterns.
[^4]: **BlueGreen** deployment is work-under-progress.
[^5]: **Sticky sessions** traffic shaping feature is work-under-progress.