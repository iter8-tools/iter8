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
NETWORK_LAYERS="istio contour gloo kourier"
if [[ ! " ${NETWORK_LAYERS[@]} " =~ " ${1} " ]]; then
    echo "Network Layer ${1} unsupported"
    echo "Use one of istio, gloo, kourier, contour"
    exit 1
fi

# Step 1: Export correct tags for install artifacts
export TAG="${TAG:-v0.5.1}"
export KNATIVE_TAG="${KNATIVE_TAG:-v0.21.0}"
echo "TAG = $TAG"
echo "KNATIVE_TAG = $KNATIVE_TAG"

# Step 2: Install Knative (https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-serving-component)

# 2(a). Install the Custom Resource Definitions (aka CRDs):
echo "Installing Knative CRDs"

kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_TAG}/serving-crds.yaml

# 2(b). Install the core components of Serving (see below for optional extensions):

echo "Installing Knative core components"

kubectl apply --filename https://github.com/knative/serving/releases/download/${KNATIVE_TAG}/serving-core.yaml


# Step 3: Ensure readiness of Knative-serving pods
echo "Waiting for all Knative-serving pods to be running..."
sleep 10 # allowing enough time for resource creation
kubectl wait --for condition=ready --timeout=300s pods --all -n knative-serving

# Step 4: Install a network layer
if [[ "istio" == ${1} ]]; then
    ##########Installing ISTIO ###########
    echo "Installing Istio for Knative"
    WORK_DIR=$(pwd)
    TEMP_DIR=$(mktemp -d)
    cd $TEMP_DIR
    curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.2 sh -
    cd istio-1.8.2
    export PATH=$PWD/bin:$PATH
    cd $WORK_DIR
    curl -L https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/istio-minimal-operator.yaml | istioctl install -y -f -

    kubectl apply --filename https://github.com/knative/net-istio/releases/download/${KNATIVE_TAG}/release.yaml
    echo "Istio installed successfully"
    

elif [[ "contour" == ${1} ]]; then
    ##########Installing CONTOUR ###########
    echo "Installing Contour for Knative"
    # Install a properly configured Contour:
    kubectl apply --filename https://github.com/knative/net-contour/releases/download/${KNATIVE_TAG}/contour.yaml

    # Install the Knative Contour controller:
    kubectl apply --filename https://github.com/knative/net-contour/releases/download/${KNATIVE_TAG}/net-contour.yaml

    # Configure Knative Serving to use Contour by default:
    kubectl patch configmap/config-network \
    -n knative-serving \
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
    kubectl apply --filename https://github.com/knative/net-kourier/releases/download/${KNATIVE_TAG}/kourier.yaml

    # Configure Knative Serving to use Kourier by default:
    kubectl patch configmap/config-network \
    -n knative-serving \
    --type merge \
    --patch '{"data":{"ingress.class":"kourier.ingress.networking.knative.dev"}}'
    echo "Kourier installed successfully"
fi

### Note: the preceding steps perform domain install; following steps perform Iter8 install

# Step 5: Install Iter8
echo "Installing Iter8 with Knative support"
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/core/build.yaml

# Step 6: Install Iter8's Prometheus add-on
echo "Installing Iter8's Prometheus add-on"
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus-operator/build.yaml

kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s

kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus/build.yaml

kubectl apply -f ${ITER8}/samples/knative/quickstart/service-monitor.yaml

# Step 7: Install Iter8's mock New Relic service
echo "Installing Iter8's mock New Relic service"
kubectl apply -f ${ITER8}/samples/knative/quickstart/metrics-mock.yaml

# Step 8: Verify platform setup
echo "Verifying platform setup"
kubectl wait --for condition=ready --timeout=300s pods --all -n knative-serving
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system