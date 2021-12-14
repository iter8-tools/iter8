---
template: main.html
---

# What is Iter8?
Iter8 is the open-source cloud-native platform for metrics-driven experiments and rollouts. Built for DevOps/SRE/MLOps/data science teams.

## Use cases

1.  Load testing with SLOs
2.  A/B(/n) testing with metrics from any backend
3.  SLOs with metrics from any backend
4.  Traffic mirroring
5.  User segmentation
6.  Session affinity
7.  Gradual rollout

The traffic engineering use-cases (4 - 7 above) are achieved by using Iter8 along with a Kubernetes service mesh or ingress.

## What is an Iter8 experiment?
An Iter8 experiment is a sequence of tasks that produce metrics-driven insights for your app/ML model, validates it, and optionally performs a rollout of your app/ML model. Iter8 provides a variety of highly customizable tasks that can be readily used within experiments to achieve the following.

1.  Generating load and getting built-in metrics for one or more versions of the app.
2.  Getting metrics by querying backends like Prometheus, New Relic, SysDig, or Elastic.
3.  SLO validation and A/B/n testing.
4.  Triggering events based on these insights. Events include:
      * sending a Slack or webhook notification
      * triggering a CI/CD/GitHub actions workflow
      * creating a pull request, and 
      * changing application state (including traffic splits) inside a Kubernetes cluster

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

Experiments are specified using a YAML file as shown below.
```yaml
# the following experiment performs a load test for https://example.com
# and validates error-rate and 95th percentile service level objectives (SLOs)
# 
# task 1: generate requests for the app and collect built-in metrics
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: https://example.com
# task 2: assess if app satisfies SLOs
# this experiment involves only one version of the app
- task: assess-app-versions
  with:
    SLOs:
    - metric: built-in/error-rate
      upperLimit: 0
    - metric: built-in/p95
      upperLimit: 100
```

## Where can I run experiments?

* On your local machine
* In a container
* Inside Kubernetes
* As a step in your CI/CD/GitOps pipeline

## Can I use Iter8 with ...?
Use Iter8 with

  * any app/serverless/ML framework
  * any metrics backend
  * any service mesh/ingress/networking technology, and 
  * any CI/CD/GitOps process

## How is Iter8 implemented?

Iter8 is implemented as a `go` module and comes with a command line interface that enables rapid experimentation.
