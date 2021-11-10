---
template: main.html
---

# What is Iter8?
Iter8 is a metrics-driven experimentation platform that enables **DevOps/SRE/MLOps/data science teams** to maximize release velocity and business value of apps and ML models while protecting end-user experience.

Iter8 enables the following use-cases.

1.  Load testing with SLO validation
2.  SLO validation
3.  A/B(/n) testing with business metrics
4.  Mirroring
5.  User segmentation
6.  Session affinity
7.  Gradual rollout

The traffic engineering use-cases (4 - 7 above) are achieved by using Iter8 along with a Kubernetes service-mesh or ingress.

## What is an Iter8 experiment?
An Iter8 experiment is a sequence of tasks. Iter8 provides a variety of tasks for the following purposes.

1.  Getting metrics
2.  Analyzing the app (or versions of the app), and producing assessments and recommendations based on metrics
3.  Achieving a variety of side effects such as sending a slack or HTTP notification, triggering a CI/CD/GitHub actions workflow, creating a pull request, waiting for a resource to become available or ready, and changing application state (including traffic splits) within a Kubernetes cluster.

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

Experiments are specified declaratively using a simple YAML file as shown below.
```yaml
# the following experiment performs a load test for https://example.com
# and validates error-rate and 95th percentile service level objectives (SLOs)
# 
# task 0: generate requests for the app and collect built-in metrics
- task: gen-load-and-collect-metrics
  with:
    versionInfo:
    - url: https://example.com
# task 1: assess how the app is performing relative to SLOs
# this experiment involves only one version of the app
- task: assess-app-versions
  with:
    criteria:
      SLOs:
      - metric: iter8-fortio/error-rate
        upperLimit: 0
      - metric: iter8-fortio/p95
        upperLimit: 100
```

## Where can I run experiments?

## Can I use Iter8 with ...?
Iter8 can be used with:

  * any app/serverless/ML framework
  * any metrics backend
  * any service mesh/ingress/networking technology, and 
  * any CI/CD/GitOps process.

## How is Iter8 implemented?

Iter8 is implemented as `go` module and comes with a command line interface that enables rapid experimentation.
