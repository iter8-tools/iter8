---
template: main.html
---

# What is Iter8?

**Iter8** is a metrics-driven experimentation and release engineering platform.

Iter8 is designed for **DevOps, SRE, and MLOps teams** interested in maximizing release velocity and business value of their apps while protecting end-user experience. 

***

Use Iter8 for SLO validation, A/B(/n) testing with business metrics, chaos injection, dark launch, canaries, progressive rollouts with advanced traffic engineering, and their hybrids. 

Integrate with Helm, Istio, Linkerd, Litmus, Knative, KFServing, Seldon, and more.

***

## What is an Iter8 experiment?
Iter8 defines a Kubernetes resource called **Experiment** that automates the release engineering process as shown below.

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

***

## Iter8 and Helm
While Helm is not a requirement for using Iter8, many Iter8 experiments are templated and packaged in the form of Helm charts for easy reusability.

***

## How does Iter8 work?

Iter8 consists of a [Go-based Kubernetes controller](https://github.com/iter8-tools/etc3) that orchestrates (reconciles) experiments in conjunction with a [Python-based analytics service](https://github.com/iter8-tools/iter8-analytics), and a [Go-based task runner](https://github.com/iter8-tools/handler).
