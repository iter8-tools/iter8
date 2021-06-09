#!/bin/sh
echo "Hello from Iter8!"

# Loop until 'docker version' exits with 0,  meaning docker daemon is ready
until docker version > /dev/null 2>&1
do
  sleep 1
done

kind create cluster --wait 5m
kubectl cluster-info --context kind-kind

export TAG=master
kustomize build github.com/iter8-tools/iter8/install/core/?ref=${TAG} | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system
