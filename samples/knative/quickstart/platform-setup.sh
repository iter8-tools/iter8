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

## 0(c). Ensure network layer is supported
NETWORK_LAYERS="istio kourier"
if [[ ! " ${NETWORK_LAYERS[@]} " =~ " ${1} " ]]; then
    echo "Network Layer ${1} unsupported"
    echo "Use one of kourier or istio"
    exit 1
fi

# Step 4: Install Knative using operator
if [[ "istio" == ${1} ]]; then
    kubectl apply -f https://github.com/knative/operator/releases/download/v0.24.0/operator.yaml
    kubectl wait --for=condition=available deploy/knative-operator --timeout=120s
    kubectl apply -f $ITER8/samples/knative/quickstart/withistio.yaml
    kubectl wait --for=condition=available deployment --all -n knative-serving --timeout=120s
    kubectl wait crd --all --for condition=established --timeout=120s
    kubectl wait --for condition=Ready --timeout=300s pods --all -n knative-serving
    
elif [[ "kourier" == ${1} ]]; then
    kubectl apply -f https://github.com/knative/operator/releases/download/v0.24.0/operator.yaml
    kubectl wait --for=condition=available deploy/knative-operator --timeout=120s
    kubectl apply -f $ITER8/samples/knative/quickstart/withkourier.yaml
    kubectl wait --for=condition=available deployment --all -n knative-serving --timeout=120s
    kubectl wait crd --all --for condition=established --timeout=120s    
    kubectl wait --for condition=Ready --timeout=300s pods --all -n knative-serving
fi
