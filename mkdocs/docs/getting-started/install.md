---
template: overrides/main.html
title: Install iter8
---

Follow these steps to install iter8 on Kubernetes using [Kustomize v3](https://kubectl.docs.kubernetes.io/installation/kustomize/). 

## Prerequisites
1. Kubernetes 1.15+
2. [KFServing](https://github.com/kubeflow/kfserving) v0.5.0-rc2+ with support for v1beta1 InferenceService APIs installed in your Kubernetes cluster. You can verify KFServing is up and running using the following command:

```
kubectl wait pods --all -n kfserving-system --for condition=ready --timeout=300s 
```

## Install iter8-kfserving
```shell
TAG=v0.1.0-alpha
kustomize build github.com/iter8-tools/iter8-kfserving/install?ref=$TAG | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build github.com/iter8-tools/iter8-kfserving/install/iter8-metrics?ref=$TAG | kubectl apply -f -
```

## Install kfserving-monitoring
```shell
TAG=v0.1.0-alpha
kustomize build github.com/iter8-tools/iter8-kfserving/install/monitoring/prometheus-operator?ref=$TAG | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build github.com/iter8-tools/iter8-kfserving/install/monitoring/prometheus?ref=$TAG | kubectl apply -f -
```

## Verify your installation
```shell
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system
kubectl wait --for condition=ready --timeout=300s pods --all -n kfserving-monitoring
```

## Removal
Use the following command to remove kfserving-monitoring from your Kubernetes cluster.
```shell
TAG=v0.1.0-alpha
kustomize build github.com/iter8-tools/iter8-kfserving/install/monitoring/prometheus?ref=$TAG | kubectl delete -f -
kustomize build github.com/iter8-tools/iter8-kfserving/install/monitoring/prometheus-operator?ref=$TAG | kubectl delete -f -
```

Use the following command to remove iter8-kfserving from your Kubernetes cluster.
```shell
TAG=v0.1.0-alpha
kustomize build github.com/iter8-tools/iter8-kfserving/install?ref=$TAG | kubectl delete -f -
```