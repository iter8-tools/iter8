#!/bin/bash

#set -e

# Step 0: Ensure environment and arguments are well-defined

## 0(a). Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

## 0(b). Ensure Openshift cluster is available
OPENSHIFT_STATUS=$(oc version | awk '/^Server /' -)
if [[ -z ${OPENSHIFT_STATUS} ]]; then
    echo "Openshift cluster is unavailable"
    exit 1
else
    echo "Openshift cluster is available"
fi

## 0(c). Ensure Kustomize v3 or v4 is available
KUSTOMIZE_VERSION=$(kustomize  version | cut -d. -f1 | tail -c 2)
if [[ ${KUSTOMIZE_VERSION} -ge "3" ]]; then
    echo "Kustomize v3+ available"
else
    echo "Kustomize v3+ is unavailable"
    exit 1
fi

# Step 0: Install Istio (https://istio.io/latest/docs/setup/getting-started/)
#export ISTIO_VERSION="${ISTIO_VERSION:-1.10.1}"
#echo "Installing Istio"
#WORK_DIR=$(pwd)
#TEMP_DIR=$(mktemp -d)
#cd $TEMP_DIR
#curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VERSION} sh -
#cd istio-${ISTIO_VERSION}
#export PATH=$PWD/bin:$PATH
#cd $WORK_DIR
#oc adm policy add-scc-to-group anyuid system:serviceaccounts:istio-system
#istioctl install --set profile=openshift -y
#oc -n istio-system expose svc/istio-ingressgateway --port=http2
#echo "Istio installed successfully"

# Step 1: Install Iter8
echo "Installing Iter8"
kustomize build $ITER8/install/core | oc apply -f -
oc wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build $ITER8/install/builtin-metrics | oc apply -f -
oc wait --for=condition=Ready --timeout=300s pods --all -n iter8-system

