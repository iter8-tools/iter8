---
template: overrides/main.html
---

# What is Iter8?

**Iter8** is an AI-powered platform for cloud native release automation, validation, and experimentation. Iter8 makes it easy to unlock business value and guarantee SLOs by identifying the best performing app/ML model version and rolling it out safely.

Iter8 is designed for **developers, SREs, service operators, data scientists, and ML engineers** who wish to maximize release velocity and business value with their apps/ML models while protecting end-user experience.

## What is an Iter8 experiment?
Iter8 defines a Kubernetes resource called **Experiment** that automates the process of validating service-level objectives (SLOs), identifying the best version of your app/ML model version based on business and performance metrics, progressive traffic shifting, and promotion/rollback.[^1]

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

## How does Iter8 work?

Iter8 consists of a [Go-based Kubernetes controller](https://github.com/iter8-tools/etc3) that orchestrates (reconciles) experiments in conjunction with a [Python-based analytics service](https://github.com/iter8-tools/iter8-analytics), and a [Go-based task runner](https://github.com/iter8-tools/handler).

[^1]: Boxes with dashed boundaries in the picture are optional in an experiment.
