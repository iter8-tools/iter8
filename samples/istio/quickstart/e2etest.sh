#!/bin/bash

set -e -x

export EXPERIMENT=istio-quickstart

cleanup () {
    status=$?
    if (( $status != 0 )); then
        kind delete cluster
        echo -e "\033[0;31mFAILED:\033[0m $0"
    fi
    exit $status
}
trap "cleanup" EXIT

# create cluster
kind create cluster
kubectl cluster-info --context kind-kind

# platform setup
echo "Setting up platform"
$ITER8/samples/istio/quickstart/platformsetup.sh

echo "Create book-purchased metric"
kubectl apply -f $ITER8/samples/istio/quickstart/books-purchased.yaml

echo "Create bookinfo app with two productpage versions"
kubectl apply -f $ITER8/samples/istio/quickstart/namespace.yaml
kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/quickstart/bookinfo-app.yaml
kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/quickstart/productpage-v2.yaml
kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/quickstart/bookinfo-gateway.yaml
kubectl -n bookinfo-iter8 wait --for=condition=Ready pods --all --timeout=600s

echo "Generate requests"
URL_VALUE="http://$(kubectl -n istio-system get svc istio-ingressgateway -o jsonpath='{.spec.clusterIP}'):80/productpage"
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/istio/quickstart/fortio.yaml | sed "s/6000s/120s/g" | kubectl apply -f -
pod_name=$(kubectl get pods --selector=job-name=fortio -o jsonpath='{.items[*].metadata.name}')
kubectl wait --for=condition=Ready pods/"$pod_name" --timeout=240s

echo "Creating A/B experiment"
kubectl apply -f $ITER8/samples/istio/quickstart/experiment.yaml

# Wait
kubectl wait experiment $EXPERIMENT --for=condition=Completed --timeout=300s

# Log final experiment
kubectl get experiment $EXPERIMENT -o yaml

# Check
source $ITER8/samples/istio/quickstart/check.sh

# Cleanup .. not needed since cluster is getting deleted
# included to test manual instructions
# continue even if errors
set +e
kubectl delete -f $ITER8/samples/istio/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/istio/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/istio/quickstart/books-purchased.yaml
kubectl delete namespace bookinfo-iter8

# delete cluster
kind delete cluster

echo -e "\033[0;32mSUCCESS:\033[0m $0"
