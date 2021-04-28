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
export TAG="${TAG:-v0.5.1}"
export KFSERVING_TAG="${KFSERVING_TAG:-v0.5.1}"
echo "TAG = ${TAG}"
echo "KFSERVING_TAG = ${KFSERVING_TAG}"

# Step 2: Install KFServing (https://github.com/kubeflow/kfserving#install-kfserving)
WORK_DIR=$(pwd)
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
git clone -b ${KFSERVING_TAG} https://github.com/kubeflow/kfserving.git
cd kfserving
eval ./hack/quick_install.sh
cd $WORK_DIR

### Note: the preceding steps perform domain install; following steps perform Iter8 install

# Step 3: Install Iter8
echo "Installing Iter8 with KFServing support"
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/core/build.yaml

# Step 4: Install Iter8's Prometheus add-on
echo "Installing Iter8's Prometheus add-on"
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus-operator/build.yaml

kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s

kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus/build.yaml

kubectl apply -f ${ITER8}/samples/kfserving/quickstart/service-monitor.yaml

# Step 5: Install Iter8's mock New Relic service
echo "Installing Iter8's mock New Relic service"
kubectl apply -f ${ITER8}/samples/kfserving/quickstart/metrics-mock.yaml

# Step 6: Verify platform setup
echo "Verifying platform setup"
kubectl wait --for condition=ready --timeout=300s pods --all -n kfserving-system
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system