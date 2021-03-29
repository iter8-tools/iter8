#!/bin/bash

set -e

# Step 0: Export TAG
export TAG="${TAG:-v0.3.0-pre.4}"

# Step 1: Install Iter8
echo "Installing Iter8"
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/core/build.yaml
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/metrics/build.yaml

echo "Verifying Iter8 installation"
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system

# Step 2: Install Prometheus add-on
# Comment out commands in this step if you wish to skip Prometheus add-on install
echo "Installing Prometheus add-on"
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus-operator/build.yaml
kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus/build.yaml
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/service-monitors/build.yaml

echo "Verifying Prometheus-addon installation"
kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system

set +e

return 0