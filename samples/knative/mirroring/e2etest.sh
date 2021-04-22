## For this test to succeed, do the following before launching the test.
## minikube start --cpus 6 --memory 12288
## Also install Kustomize v3 or v4: https://kustomize.io/

#!/bin/bash

set -e

cleanup () {
    status=$?
    if (( $status != 0 )); then
        kind delete cluster
        echo -e "\033[0;31mFAILED:\033[0m $0"
    fi
    exit $status
}
trap "cleanup" EXIT

# create kind cluster
kind create cluster
kubectl cluster-info --context kind-kind

# platform setup
echo "Setting up platform"
$ITER8/samples/knative/quickstart/platformsetup.sh istio

# create app with live and dark versions
echo "Creating live and dark versions"
kubectl apply -f $ITER8/samples/knative/mirroring/service.yaml

# create Istio virtual services
echo "Creating Istio virtual services"
kubectl apply -f $ITER8/samples/knative/mirroring/routing-rules.yaml        

# Generate requests
echo "Generating requests"
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.2 sh -
istio-1.8.2/bin/istioctl kube-inject -f $ITER8/samples/knative/mirroring/curl.yaml | kubectl create -f -
cd $ITER8
    
# Create Iter8 experiment
echo "Creating the Iter8 metrics and experiment"
kubectl apply -f $ITER8/samples/knative/quickstart/metrics.yaml
kubectl wait --for=condition=Ready ksvc/sample-app
kubectl apply -f $ITER8/samples/knative/mirroring/experiment.yaml

# Wait
kubectl wait experiment mirroring --for=condition=Completed --timeout=150s

# check experiment
source $ITER8/samples/knative/mirroring/check.sh

# Cleanup .. not needed since cluster is getting deleted; just forming a good habit!
kubectl delete -f $ITER8/samples/knative/mirroring/curl.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/experiment.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/routing-rules.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/service.yaml

# delete kind cluster
kind delete cluster

set +e

echo -e "\033[0;32mSUCCESS:\033[0m $0"