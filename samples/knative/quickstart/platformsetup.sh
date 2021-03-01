#!/bin/bash

set -e 

## Ensure ITER8 environment variable is set.
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

## Ensure Kubernetes cluster is available.
KUBERNETES_STATUS=$(kubectl version | awk '/^Server Version:/' -)
if [[ -z ${KUBERNETES_STATUS} ]]; then
    echo "Kubernetes cluster is unavailable"
    exit 1
else
    echo "Kubernetes cluster is available"
fi


## Ensure Kustomize v3 is available
KUSTOMIZE_VERSION=$(kustomize version | cut -f 1 | cut -d/ -f 2 | cut -d. -f 1)
if [[ $KUSTOMIZE_VERSION == "v4" ]]; then
    echo "Kustomize v4 is available"
else
    echo "Kustomize Version found: $KUSTOMIZE_VERSION"
    echo "Kustomize v4 is not available"
    echo "Get Kustomize v4 from https://kubectl.docs.kubernetes.io/installation/kustomize/"
    exit 1
fi

## Ensure network layer is supported
NETWORK_LAYERS="istio contour gloo kourier"
if [[ ! " ${NETWORK_LAYERS[@]} " =~ " ${1} " ]]; then
    echo "Network Layer ${1} unsupported"
    echo "Use one of istio, gloo, kourier, contour"
    exit 1
fi


# Step1: Install Knative (https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-serving-component)

# 1(a). Install the Custom Resource Definitions (aka CRDs):
echo "Installing Knative CRDs"

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.20.0/serving-crds.yaml

# 1(b). Install the core components of Serving (see below for optional extensions):

echo "Installing Knative core components"

kubectl apply --filename https://github.com/knative/serving/releases/download/v0.20.0/serving-core.yaml


# Step 2: Monitor the Knative components until all of the components are `Running` or `Completed`:
echo "Waiting for all Knative-serving pods to be running..."
sleep 10 # allowing enough time for resource creation
kubectl wait --for condition=ready --timeout=300s pods --all -n knative-serving


# Step 3: Install a network layer
if [[ "istio" == ${1} ]]; then
    ##########Installing ISTIO ###########
    echo "Installing Istio for Knative"
    WORK_DIR=$(pwd)
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.1 sh -
    cd istio-1.8.1
    export PATH=$PWD/bin:$PATH
    cd $WORK_DIR
    curl -L https://raw.githubusercontent.com/iter8-tools/iter8-knative/main/quickstart/istio-minimal-operator.yaml | istioctl install -y -f -

    kubectl apply --filename https://github.com/knative/net-istio/releases/download/v0.20.0/release.yaml
    echo "Istio installed successfully"
    

elif [[ "contour" == ${1} ]]; then
    ##########Installing CONTOUR ###########
    echo "Installing Contour for Knative"
    # Install a properly configured Contour:
    kubectl apply --filename https://github.com/knative/net-contour/releases/download/v0.20.0/contour.yaml

    # Install the Knative Contour controller:
    kubectl apply --filename https://github.com/knative/net-contour/releases/download/v0.20.0/net-contour.yaml

    # Configure Knative Serving to use Contour by default:
    kubectl patch configmap/config-network \
    --namespace knative-serving \
    --type merge \
    --patch '{"data":{"ingress.class":"contour.ingress.networking.knative.dev"}}'
    echo "Contour installed successfully"

elif [[ "gloo" == ${1} ]]; then
    ##########Installing GLOO ###########
    echo "Installing Gloo for Knative"
    # Install Gloo and the Knative integration:
    curl -sL https://run.solo.io/gloo/install | sh
    export PATH=$HOME/.gloo/bin:$PATH
    glooctl install knative --install-knative=false
    echo "Gloo installed successfully"
    
elif [[ "kourier" == ${1} ]]; then
    ##########Installing KOURIER ###########
    echo "Installing Kourier for Knative"
    
    # Install the Knative Kourier controller:
    kubectl apply --filename https://github.com/knative/net-kourier/releases/download/v0.20.0/kourier.yaml

    # Configure Knative Serving to use Kourier by default:
    kubectl patch configmap/config-network \
    --namespace knative-serving \
    --type merge \
    --patch '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}'
    echo "Kourier installed successfully"
fi

# Step 4: Export correct tag for install artifacts
export TAG=v0.2.5

# Step 5: Install iter8-monitoring
echo "Installing iter8-monitoring"
kustomize build github.com/iter8-tools/iter8/install/monitoring/prometheus-operator/?ref=${TAG} | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build github.com/iter8-tools/iter8/install/monitoring/prometheus/?ref=${TAG} | kubectl apply -f - 

# Step 6: Install iter8 for Knative
echo "Installing iter8 for Knative"
kustomize build github.com/iter8-tools/iter8/install/?ref=${TAG} | kubectl apply -f -
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kustomize build github.com/iter8-tools/iter8/install/iter8-metrics/?ref=${TAG} | kubectl apply -f -

# Step 7: Verify iter8 installation
echo "Verifying installation"
kubectl wait --for condition=ready --timeout=300s pods --all -n knative-serving
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-monitoring