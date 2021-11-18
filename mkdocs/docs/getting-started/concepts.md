---
template: main.html
---

# What is Iter8?
Iter8 is a metrics-driven experimentation platform that enables **DevOps/SRE/MLOps/data science teams** to maximize release velocity and business value of apps and ML models while protecting end-user experience.

Iter8 enables the following use-cases.

1.  Load testing/SLOs with built-in metrics
2.  A/B(/n) testing with business metrics from any backend
3.  SLOs with metrics from any backend
4.  Mirroring
5.  User segmentation
6.  Session affinity
7.  Gradual rollout

The traffic engineering use-cases (4 - 7 above) are achieved by using Iter8 along with a Kubernetes service-mesh or ingress.

## What is an Iter8 experiment?
An Iter8 experiment is a sequence of tasks. Iter8 provides a variety of tasks for the following purposes.

1.  Getting metrics for one or more versions of the app.
2.  Producing SLO validation and A/B/n testing insights based on metrics.
3.  Triggering a variety of useful events based on these insights. Events include sending a slack or HTTP notification, triggering a CI/CD/GitHub actions workflow, creating a pull request, and changing application state (including traffic splits for versions) inside a Kubernetes cluster.

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

Experiments are specified declaratively using a simple YAML file as shown below.
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
* As a step in your CI/CD/GitOps pipeline.

## Can I use Iter8 with ...?
Use Iter8 with

  * any app/serverless/ML framework
  * any metrics backend
  * any service mesh/ingress/networking technology, and 
  * any CI/CD/GitOps process.

## How is Iter8 implemented?

Iter8 is implemented as `go` module and comes with a command line interface that enables rapid experimentation.
