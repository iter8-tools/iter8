---
template: main.html
title: Installation
hide:
- toc
---

# Installation

Install Iter8 in your Kubernetes cluster as follows.

```shell
export TAG=v0.5.1
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/$TAG/core/build.yaml
```

The above command installs Iter8's controller and analytics services in the `iter8-system` namespace, the Experiment and Metric CRDs, and the following RBAC permissions.

??? info "Default RBAC Permissions"
    | Resource | Permissions | Scope |
    | ----- | ---- | ----------- |
    | experiments.iter8.tools | get, list, patch, update, watch | Cluster-wide |
    | experiments.iter8.tools/status | get, patch, update | Cluster-wide |
    | metrics.iter8.tools | get, list | Cluster-wide |
    | jobs.batch | create, delete, get, list, watch | Cluster-wide |
    | leases.coordination.k8s.io | get, list, watch, create, update, patch, delete | `iter8-system` namespace |
    | events | create | `iter8-system` namespace |
    | services.serving.knative.dev | get, list, patch, update | Cluster-wide |
    | inferenceservices.serving.knative.dev | get, list, patch, update | Cluster-wide |
    | virtualservices.networking.istio.io | get, list, patch, update, create, delete | Cluster-wide |
    | destinationrules.networking.istio.io | get, list, patch, update, create, delete | Cluster-wide |
