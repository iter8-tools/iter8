---
template: main.html
hide:
- navigation
- toc
---

# Roadmap

1. **Design simplifications (v0.8)**
    * CRD and controller-less experimentation using Kubernetes configmaps and jobs
    * *Hub:* Helm repo with reusable namespace-scoped experiment charts
    * Multi-tenancy
2. **GitOps**
    * Integration with ArgoCD, Flux, JenkinsX and other GitOps operators
3. **Enhanced experiments**
    * Support for progressive rollout of statefulsets
    * Support for distributed Helm applications
    * Enhanced examples for mirroring/shadow deployments
    * Enhanced examples for experiments with fixed traffic splits
    * Enhanced examples for experiments with progressive traffic shifting
4. **Metrics**
    * Enhanced examples with built-in metrics, including for apps using gRPC
    * Support for more metric providers like MySQL, PostgreSQL, CouchDB, MongoDB, and Google Analytics.
5. **CNCF Sandbox project**
6. **Enhanced MLOps experiments**
    * Customized experiments/metrics for serving frameworks like TorchServe and TFServing
7. **Analytics improvements**
    * Pareto testing strategies with single and multiple reward metrics
    * Early termination of experiments
    * Extension points for traffic shifting and version assessment algorithms
    * Experiments with `support` and `confidence` estimation
