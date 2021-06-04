---
template: main.html
---

# What is Iter8?

**Iter8** is an AI-powered platform for cloud native release automation and experimentation platform that enables SLO validation, A/B testing and progressive delivery. 

Iter8 makes it easy to **unlock business value and guarantee SLOs** by identifying the best performing version of your app/ML model and promoting it safely.

Iter8 is designed for **DevOps and MLOps teams** interested in maximizing release velocity and business value with their apps/ML models while protecting end-user experience.

## What is an Iter8 experiment?
Iter8 defines a Kubernetes resource called **Experiment** that automates SLO validation, A/B, and A/B/n testing experiments. During an experiment, Iter8 can compare multiple versions, find, and safely promote the **winning version (winner)** based on business metrics and performance metrics like latency and error-rate.

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

## How does Iter8 work?

Iter8 consists of a [Go-based Kubernetes controller](https://github.com/iter8-tools/etc3) that orchestrates (reconciles) experiments in conjunction with a [Python-based analytics service](https://github.com/iter8-tools/iter8-analytics), and a [Go-based task runner](https://github.com/iter8-tools/handler).
