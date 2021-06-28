#!/bin/bash

set -e 

# Step 0: Ensure environment and arguments are well-defined

## 0(a). Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " ${ITER8}
fi

## 0(b). Ensure Kubernetes cluster is available
KUBERNETES_STATUS=$(kubectl version | awk '/^Server Version:/' -)
if [[ -z ${KUBERNETES_STATUS} ]]; then
    echo "Kubernetes cluster is unavailable"
    exit 1
else
    echo "Kubernetes cluster is available"
fi

## 0(c). Ensure Kustomize v3 or v4 is available
KUSTOMIZE_VERSION=$(kustomize  version | cut -d. -f1 | tail -c 2)
if [[ ${KUSTOMIZE_VERSION} -ge "3" ]]; then
    echo "Kustomize v3+ available"
else
    echo "Kustomize v3+ is unavailable"
    exit 1
fi

# Step 1: Install Iter8
echo "Installing Iter8 with KFServing support"
kustomize build $ITER8/install/core | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build $ITER8/install/builtin-metrics | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system

# Step 2: Install Iter8's Prometheus add-on
echo "Installing Iter8's Prometheus add-on"
kustomize build $ITER8/install/prometheus-add-on/prometheus-operator | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build $ITER8/install/prometheus-add-on/prometheus | kubectl apply -f -

kubectl apply -f ${ITER8}/samples/kfserving/quickstart/service-monitor.yaml

# Step 3: Verify platform setup
echo "Verifying platform setup"
kubectl wait --for=condition=Ready --timeout=300s pods --all -n iter8-system
