#!/bin/bash

set -e

# create kind cluster
kind create cluster --wait 5m
kubectl cluster-info --context kind-kind

# platform setup
echo "Setting up platform"
$ITER8/samples/knative/quickstart/iter8-setup.sh
$ITER8/samples/knative/quickstart/platform-setup.sh istio

# create app versions
echo "Creating live and dark versions"
kubectl apply -f $ITER8/samples/knative/user-segmentation/services.yaml

# create Istio virtual service
echo "Creating Istio virtual service"
kubectl apply -f $ITER8/samples/knative/user-segmentation/routing-rule.yaml 

# Generate requests
echo "Generating requests"
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.7.0 sh -
istio-1.7.0/bin/istioctl kube-inject -f $ITER8/samples/knative/user-segmentation/curl.yaml | kubectl create -f -
cd $ITER8
    
# Create Iter8 experiment
echo "Creating Iter8 experiment"
kubectl wait --for=condition=Ready ksvc/sample-app-v1
kubectl wait --for=condition=Ready ksvc/sample-app-v2
kubectl apply -f $ITER8/samples/knative/user-segmentation/experiment.yaml  

export EXPERIMENT=user-segmentation-exp

# Wait for experiment to complete
kubectl wait experiment $EXPERIMENT --for=condition=Completed --timeout=360s

# Check
source  $ITER8/samples/knative/user-segmentation/check.sh

# Cleanup .. not needed since cluster is getting deleted; just forming a good habit!
kubectl delete -f $ITER8/samples/knative/user-segmentation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/curl.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/services.yaml

# delete kind cluster
kind delete cluster

set +e