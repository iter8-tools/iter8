---
template: main.html
---

# What is Iter8?

Kubernetes release engineering toolkit built for DevOps, MLOps, SRE and data science teams. 

## What is an Iter8 experiment?
Iter8 experiments make it simple to collect performance and business metrics for Kubernetes apps and ML models, assess and compare multiple app/ML model versions, safely rollout winning versions, and maximize business value with each release.

### Example
The following picture illustrates an Iter8 experiment that performs load testing with SLO validation of a gRPC service.

![Load testing gRPC](../tutorials/load-test-grpc/images/grpc-overview.png)

### Experiment chart
In order to enable reuse, Iter8 experiments are templated and packaged as specialized [Helm charts](https://helm.sh/docs/topics/charts/). Experiment charts can be combined with values to generate `experiments.yaml` files that provide fully defined experiment specifications.

Iter8 experiment charts enable to you to launch powerful release engineering experiments in a matter of seconds.

## Features at a glance

- **Load testing with SLOs** 
    
    Iter8 experiments can generate requests for HTTP and gRPC services, collect built-in latency and error-related metrics, and validate SLOs.

- **A/B(/n) testing** 
      
    Grow your business with every release. Iter8 experiments can compare multiple versions based on business value and promote a winner.

- **Simple to use** 
      
    Get started with Iter8 in seconds using pre-packaged experiment charts. Run Iter8 experiments locally, in a container, inside Kubernetes, or inside your CI/CD/GitOps pipelines.

- **Awesome integrations** 
      
    Use with any app, serverless, or ML framework. Iter8 works with Kubernetes deployments, statefulsets, Knative services, KServe/Seldon ML deployments, or custom Kubernetes resource types.

## Implementation

Iter8 is implemented as a `go` module and comes with a command line interface (CLI) that enables rapid experimentation.
