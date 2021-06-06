#!/bin/bash

set -e

export EXPERIMENT=quickstart-exp

# create cluster
kind create cluster
kubectl cluster-info --context kind-kind

# platform setup
echo "Setting up platform"
$ITER8/samples/seldon/quickstart/platformsetup.sh

echo "Create app/ML model versions"
kubectl apply -f $ITER8/samples/seldon/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/seldon/quickstart/candidate.yaml
kubectl apply -f $ITER8/samples/seldon/quickstart/routing-rule.yaml
kubectl wait --for condition=ready --timeout=600s pods --all -n ns-baseline
kubectl wait --for condition=ready --timeout=600s pods --all -n ns-candidate

echo "Generate requests"
URL_VALUE="http://$(kubectl -n istio-system get svc istio-ingressgateway -o jsonpath='{.spec.clusterIP}'):80"
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/seldon/quickstart/fortio.yaml | sed "s/6000s/600s/g" | kubectl apply -f -
pod_name=$(kubectl get pods --selector=job-name=fortio-requests -o jsonpath='{.items[*].metadata.name}')
kubectl wait --for=condition=Ready pods/"$pod_name" --timeout=240s
pod_name=$(kubectl get pods --selector=job-name=fortio-irisv1-rewards -o jsonpath='{.items[*].metadata.name}')
kubectl wait --for=condition=Ready pods/"$pod_name" --timeout=240s
pod_name=$(kubectl get pods --selector=job-name=fortio-irisv2-rewards -o jsonpath='{.items[*].metadata.name}')
kubectl wait --for=condition=Ready pods/"$pod_name" --timeout=240s

echo "Define metrics"
kubectl apply -f $ITER8/samples/seldon/quickstart/metrics.yaml

echo "Launch experiment"
kubectl apply -f $ITER8/samples/seldon/quickstart/experiment.yaml

# Wait
kubectl wait experiment $EXPERIMENT --for=condition=Completed --timeout=300s

# Log final experiment
kubectl get experiment $EXPERIMENT -o yaml

# Check
source $ITER8/samples/seldon/quickstart/check.sh

kubectl delete -f $ITER8/samples/seldon/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/baseline.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/candidate.yaml

# delete cluster
kind delete cluster

set +e

echo -e "\033[0;32mSUCCESS:\033[0m $0"
