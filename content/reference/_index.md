---
title: Reference
weight: 60
summary: Reference 
---

### Experiment CRD

Automated A/B, A/B/n and canary tests are managed by a Kubernets controller. The behavior of a particular experiment is controlled by an `Experiment` custom resource. Details of the CRD are [here]{{< ref "experiment" >}}).

### Metrics

To assess the behavior of microservice versions, iter8 supports a few metrics out-of-the-box without requiring users to do any extra work. In addition, users can define their own custom metrics. Iter8's out-of-the-box metrics as well as user-defined metrics can be referenced in the success criteria of an experiment. More details about metrics are documented [here]{{< ref "metrics" >}}).

### Algorithms

A key goal of this project is to introduce statistically robust algorithms for decision making during cloud-native canary releases and A/B testing experiments. Details are provided [here]{{< ref "algorithms" >}}).
