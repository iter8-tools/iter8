---
title: About
weight: 10
summary: Learn more about iter8
---

### Iter8 enables statistically robust continuous experimentation of microservices in your CI/CD pipelines

Use an `iter8 experiment` to safely expose competing versions of a service to application traffic, gather in-depth insights about key performance and business metrics for your microservice versions, and intelligently rollout the best version of your service.

Iter8's expressive model of cloud experimentation supports a variety of CI/CD scenarios. Using an iter8 experiment, you can:

1. Run a **performance test** with a single version of a microservice.
2. Perform a **canary release** with two versions, a baseline and a candidate. Iter8 will shift application traffic safely and gradually to the candidate, if it meets the criteria you specify in the experiment.
3. Perform an **A/B test** with two versions -- a baseline and a candidate. Iter8 will identify and shift application traffic safely and gradually to the `winner`, where the winning version is defined by the criteria you specify in the experiment.
4. Perform an **A/B/N test** with multiple versions -- a baseline and multiple candidates. Iter8 will identify and shift application traffic safely and gradually to the winner.

Under the hood, iter8 uses advanced Bayesian learning techniques coupled with multi-armed bandit approaches to compute a variety of statistical assessments for your microservice versions, and uses them to make robust traffic control and rollout decisions.

![iter8pic]({{< resourceAbsUrl path="images/iter8pic.png" >}})
