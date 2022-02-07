---
template: main.html
---

# What is Iter8?

Kubernetes release optimizer built for DevOps and MLOps teams.

## What is an Iter8 experiment?
Iter8 experiments make it simple to collect performance and business metrics for apps and ML models, assess, compare and validate multiple app/ML model versions, safely rollout winning version, and maximize business value in each release.

<p align='center'>
  <img alt-text="load-test-http" src="../../images/iter8-intro-dark.png" width="70%" />
</p>

### Experiment chart
Experiment charts are specialized [Helm charts](https://helm.sh/docs/topics/charts/) that contain reusable experiment templates. Iter8 combines experiment charts with user supplied values to generate runnable `experiment.yaml` files.

#### Iter8 Hub
Iter8 hub is a specific location within in the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8) that hosts several pre-packaged and reusable charts. These charts enable to you to launch powerful release optimization experiments in seconds. Their usage is described in depth in various [Iter8 tutorials](https://iter8.tools/latest/tutorials/load-test-http/overview/).

## Features at a glance

- **Load testing with SLOs** 
    
    Iter8 experiments can generate requests for HTTP and gRPC services, collect built-in latency and error-related metrics, and validate SLOs.

- **A/B(/n) testing** 
      
    Grow your business with every release. Iter8 experiments can compare multiple versions based on business value and promote a winner.

- **Simple to use** 
      
    Get started with Iter8 in seconds using pre-packaged experiment charts. Run Iter8 experiments locally, inside Kubernetes, or inside your CI/CD/GitOps pipelines.

- **App/serverless/ML frameworks** 
      
    Use with any app, serverless, or ML framework. Iter8 works with Kubernetes deployments, statefulsets, Knative services, KServe/Seldon ML deployments, or custom Kubernetes resource types.

## Implementation
Iter8 is primarily written in `go` and builds on a few awesome open source projects including:

- [Helm](https://helm.sh)
- [Fortio](https://github.com/fortio/fortio)
- [ghz](https://ghz.sh)
- [plotly.js](https://github.com/plotly/plotly.js)
