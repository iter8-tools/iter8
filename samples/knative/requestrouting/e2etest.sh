## For this test to succeed, do the following before launching the test.
## minikube start --cpus 6 --memory 12288
## Also install Kustomize v3 or v4: https://kustomize.io/

#!/bin/bash

set -e 

# platform setup
echo "Setting up platform"
$ITER8/samples/knative/quickstart/platformsetup.sh istio

# create app with live and dark versions
echo "Creating live and dark versions"
kubectl apply -f $ITER8/samples/knative/requestrouting/services.yaml

# create Istio virtual services
echo "Creating Istio virtual services"
kubectl apply -f $ITER8/samples/knative/requestrouting/routing-rule.yaml 

# Generate requests
echo "Generating requests"
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.2 sh -
istio-1.8.2/bin/istioctl kube-inject -f $ITER8/samples/knative/requestrouting/curl.yaml | kubectl create -f -
cd $ITER8
    
# Create Iter8 experiment
echo "Creating an Iter8 experiment"
kubectl wait --for=condition=Ready ksvc/sample-app-v1
kubectl wait --for=condition=Ready ksvc/sample-app-v2
kubectl apply -f $ITER8/samples/knative/requestrouting/experiment.yaml        

# Sleep
echo "Sleep for 125s"
sleep 125.0

# Check if experiment is complete and successful
echo "Checking if experiment is completed and successful"
completed="Completed"
if [[ $(kubectl get experiment request-routing -ojson | jq .status.stage)=="$completed" ]]; then
    echo "Experiment has Completed"
else
    echo "Experiment must be Completed. It is $(kubectl get experiment request-routing -ojson | jq .status.stage)"
    exit 1
fi

# Check if recommended baseline is candidate
echo "Checking if recommendedBaseline is candidate"
candidate="candidate"
if [[ $(kubectl get experiment request-routing -ojson | jq .status.recommendedBaseline)=="$candidate" ]]; then
    echo "recommendedBaseline is candidate"
else
    echo "recommendedBaseline must be candidate"
    exit 1
fi

# Cleanup
kubectl delete -f $ITER8/samples/knative/requestrouting/experiment.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/curl.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/services.yaml