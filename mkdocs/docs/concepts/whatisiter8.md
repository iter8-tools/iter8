---
template: main.html
---

# What is Iter8?

**Iter8** is a metrics-driven experimentation platform. Iter8 enables **DevOps/MLOps/SRE teams** to maximize release velocity and business value of apps and ML models while protecting end-user experience.

***

## What is an Iter8 experiment?
Iter8 defines a Kubernetes resource called **Experiment** that facilitates release engineering tasks such as SLO validation, A/B(/n) testing with business metrics, chaos testing, dark launch, canaries, progressive rollouts with advanced traffic engineering, and their hybrids. 

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

***

## Iter8 and Helm
While Helm is not a requirement for using Iter8, many Iter8 experiments are templated and packaged in the form of Helm charts for easy reusability.

***

## How does Iter8 work?

Iter8 consists of a [Go-based Kubernetes controller](https://github.com/iter8-tools/etc3) that orchestrates (reconciles) experiments in conjunction with a [Python-based analytics service](https://github.com/iter8-tools/iter8-analytics), and a [Go-based task runner](https://github.com/iter8-tools/handler).
