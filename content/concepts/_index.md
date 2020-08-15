---
menuTitle: Concepts
title: Concepts
weight: 10
summary: Learn more about iter8
---

## What is iter8?

You are developing a distributed microservices-based application on Kubernetes and have created alternative versions of a service. You want to identify the `best` version of your service using a live experiment and rollout this version in a safe and reliable manner.

Enter `iter8`.

<!-- Software engineers use a variety of continuous experimentation strategies to achieve desirable outcomes for their businesses. The desired outcome could be safely releasing a new version in production while ensuring it meets various service level objectives (SLOs), or finding the `best` version among multiple competing versions of a microservice using live application traffic, or performance testing a version in a dev/test/staging environment before releasing in production.  -->

Iter8 is an open source toolkit for `continuous experimentation` on Kubernetes. Iter8 enables you to deliver high-impact code changes within your microservices applications in an agile manner while eliminating the risk. Using iter8's machine learning (ML)-driven experimentation capabilities, you can safely and rapidly orchestrate various types of live experiments, gain key insights into the behavior of your microservices, and rollout the best versions of your microservices in an automated, principled, and statistically robust manner.

## What is an iter8 experiment?

Use an `iter8 experiment` to safely expose alternative versions of a service to application traffic and intelligently rollout the best version of your service. Iter8's expressive model of experimentation supports a diverse variety of experiments. The four main kinds of experiments in iter8 are as follows.

1. Perform a **canary release** with two versions, a baseline and a candidate. Iter8 will shift application traffic in a safe and progressive manner to the candidate, if the candidate meets the criteria you specify in the experiment.
2. Perform an **A/B/n rollout** with multiple versions -- a baseline and multiple candidates. Iter8 will identify and shift application traffic safely and gradually to the `winner`, where the winning version is defined by the criteria you specify in your experiments.
3. Perform an **A/B rollout** -- this is a special case of the A/B/n rollout described above with a baseline and a single candidate.
4. Run a **performance test** with a single version of a microservice. `iter8` will verify if the version meets the criteria you specify in the experiment.

<!-- ## What knobs are available to control an iter8 experiment?
Included above is a) Bayesian assessments b) relative criteria, c) comparing more than two versions, d) roll forward, roll back, or split traffic, e) ability to extent iter8 your own custom metrics, and f) ability to specify versions as distinct services or distinct deployments of the same service in k8s

## What insights are available from an iter8 experiment?
Iter8 continually computes a variety of assessments for each version throughout the course of an experiment. Key assessments include the version's observed metric values, the credible intervals for each metric for this version, the probability of the version outperforming the baseline with respect to a given metric, the `win probability` of the version (i.e., the probability that this version is the best version among all versions), whether the version failed to satisfy any criteria, and the probability of a version satisfying all criteria. The concept of a `winner` is applicable in canary releases, A/B and A/B/n rollouts, and is determined by iter8 based on the criteria you specify in the experiment. -->

## How does iter8 work?

{{< figure src="/images/iter8pic.png" title="An iter8 experiment in progress" caption="Iter8 is composed of a Kubernetes controller and an analytics service which are jointly responsible for orchestrating an experiment, and making iterative decisions about how to shift application traffic, how to identify a winner, and when to terminate the experiment. Under the hood, iter8 uses advanced Bayesian learning techniques coupled with multi-armed bandit approaches for statistical assessments and decision making.">}}
