---
template: main.html
title: Install Iter8
hide:
- toc
---

# Install Iter8
Install Iter8 in your Kubernetes cluster as follows. This step requires [Kustomize v3+](https://kubectl.docs.kubernetes.io/installation/kustomize/).

```shell
export TAG=v0.7.14
kustomize build https://github.com/iter8-tools/iter8/install/core/?ref=${TAG} | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build https://github.com/iter8-tools/iter8/install/builtin-metrics/?ref=${TAG} | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system
```

## Install `iter8ctl`
Install `iter8ctl` CLI on your local machine as follows. This step requires [Go 1.16+](https://golang.org/doc/install).
```shell
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.5
```

<!-- ## Pinning the Iter8 version
To select the version of Iter8 during installation, select any Iter8 version (>= v0.6.0) from  [Iter8's release history](https://github.com/iter8-tools/iter8/releases) and use it as the `TAG` above.

## RBAC rules
As part of Iter8 installation, the following RBAC rules are also installed in your cluster. You can Kustomize Iter8 installation in order to install Iter8 only for the K8s environments of your choice, and eliminate RBAC rules not needed in your environment.

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
    | seldondeployments.machinelearning.seldon.io | get, list, patch, update | Cluster-wide |
    | services | get, list, watch | Cluster-wide |
    | deployments | get, list, watch | Cluster-wide | -->
