#!/bin/bash

set -e 

# Step 0: Ensure environment and arguments are well-defined

## 0(a). Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

## 0(b). Ensure Kubernetes cluster is available
KUBERNETES_STATUS=$(kubectl version | awk '/^Server Version:/' -)
if [[ -z ${KUBERNETES_STATUS} ]]; then
    echo "Kubernetes cluster is unavailable"
    exit 1
else
    echo "Kubernetes cluster is available"
fi

## 0(b). Ensure Kustomize v3 or v4 is available
KUSTOMIZE_VERSION=$(kustomize  version | cut -d. -f1 | tail -c 2)
if [[ ${KUSTOMIZE_VERSION} -ge "3" ]]; then
    echo "Kustomize v3+ available"
else
    echo "Kustomize v3+ is unavailable"
    exit 1
fi

# Step 1: Export correct tags for install artifacts
export LINKERD2_VERSION="stable-2.10.2"
echo "LINKERD2_VERSION=$LINKERD2_VERSION"

# Step 2: Install Linkerd (https://linkerd.io/2.10/getting-started/)
echo "Linkerd"
WORK_DIR=$(pwd)
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -sL https://run.linkerd.io/install | LINKERD2_VERSION=${LINKERD2_VERSION} sh -
export PATH=$PATH:~/.linkerd2/bin
cd $WORK_DIR

echo "Installing Linkerd..."
linkerd install | kubectl apply -f -
echo "Waiting for all Linkerd pods to be running..."
linkerd check
echo "Installing Viz extension..."
linkerd viz install | kubectl apply -f -
echo "Waiting for all Viz extension pods to be running..."
linkerd check
echo "Linkerd installed successfully"

### Note: the preceding steps perform domain install; following steps perform Iter8 install

# Step 3: Install Iter8
echo "Installing Iter8 with Istio Support"
kustomize build $ITER8/install/core | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build $ITER8/install/builtin-metrics | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system
