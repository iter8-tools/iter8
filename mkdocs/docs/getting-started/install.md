---
template: main.html
title: Install Iter8
hide:
- toc
---

# Install Iter8

Install Iter8 in your Kubernetes cluster as follows.

```shell
# Kustomize v3+ is required for the following steps
export TAG=v0.7.3 
kustomize build https://github.com/iter8-tools/iter8/install/core/?ref=${TAG} | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build https://github.com/iter8-tools/iter8/install/builtin-metrics/?ref=${TAG} | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system
```

## Install `iter8ctl`
The `iter8ctl` CLI enables observing the results of Iter8 experiments in real-time. Install `iter8ctl` on your local machine as follows. You can change the directory where `iter8ctl` is installed by changing `GOBIN` below.

```shell
# Go 1.13+ is required for the following step
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.4
```


<!-- ??? info "As part of Iter8 install, these RBAC rules are installed in your cluster."
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
    | seldondeployments.machinelearning.seldon.io | get, list, patch, update | Cluster-wide | -->
