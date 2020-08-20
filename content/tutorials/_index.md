---
title: Tutorials
weight: 30
summary: Get started with tutorials
---

The following tutorials explain some of the basic concepts in iter8 and provide examples of basic building blocks to get started with your own experiments.

[Getting Started with Canary Testing]({{< ref "canary" >}}) shows how iter8 can be used to safely rollout a new version of a service. If the new version has errors or performs badly, the original version is restored.

[Getting Started with Canary Testing on Red Hat OpenShift]({{< ref "canary-openshift" >}}) is a Red Hat OpenShift specific version of the canary testing tutorial.

[Getting Started with A/B/n Rollout]({{< ref "abn" >}}) shows how iter8 can be used to experiment with several versions of a service to find the one that returns the highest reward and satisfies any functional and performance requirements. Once the experiment succeeds, iter8 will automatically shift all traffic to the winning version.

[Getting Started with A/B/n Rollout on Red Hat OpenShift]({{< ref "abn-openshift" >}}) is an OpenShift specific version of the A/B/n rollout tutorial.
