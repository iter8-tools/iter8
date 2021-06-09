#!/bin/sh
echo "Hello from Iter8!"

# Loop until 'docker version' exits with 0,  meaning docker daemon is ready
until docker version > /dev/null 2>&1
do
  sleep 1
done

echo "Docker in Docker is running..."

kind create cluster --wait 5m
kubectl cluster-info --context kind-kind

echo "Installing Iter8..."
export TAG=master
kustomize build github.com/iter8-tools/iter8/install/core/?ref=${TAG} | kubectl apply -f -
kubectl wait --for=condition=Ready pods --all -n iter8-system

echo "Iter8 in Docker is running..."
echo "All systems go."
