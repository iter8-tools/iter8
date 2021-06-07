---
template: main.html
title: Installation
hide:
- toc
---

# Installation

!!! tip "Kustomize"
    Kustomize is a pre-requisite for Iter8 installation. Get Kustomize by following [these instructions](https://kubectl.docs.kubernetes.io/installation/kustomize/).

Install Iter8 in your Kubernetes cluster as follows.

```shell
export TAG=master
kustomize build https://github.com/iter8-tools/iter8/install/core/?ref=${TAG} | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build https://github.com/iter8-tools/iter8/install/builtin-metrics/?ref=${TAG} | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system
```

??? info "Default RBAC Rules"
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

## Pinning the Iter8 version

Iter8 release history is available [here](https://github.com/iter8-tools/iter8/releases). To pin the version of Iter8 during the installation, select any Iter8 version >= v0.5.13 and change the `TAG` above. 

For example, to install version `v0.5.14`, do `export TAG=v0.5.14` in the command above.