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

# Step 1: Export correct tags for install artifacts
export KFSERVING_VERSION="${KFSERVING_VERSION:-v0.5.1}"
echo "KFSERVING_VERSION=${KFSERVING_VERSION}"

# Step 2: Install KFServing (https://github.com/kubeflow/kfserving#install-kfserving)
WORK_DIR=$(pwd)
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
# git clone -b ${KFSERVING_TAG} https://github.com/kubeflow/kfserving.git
git clone https://github.com/kubeflow/kfserving.git
cd kfserving
set +e # hacks needed to overcome this very glitchy quick_install below
./hack/quick_install.sh
kubectl delete ns kfserving-system
kubectl apply -f ./install/${KFSERVING_VERSION}/kfserving.yaml
set -e
kubectl wait --for=condition=Ready --timeout=300s pods --all -n kfserving-system
cd $WORK_DIR
