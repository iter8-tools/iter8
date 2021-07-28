#!/bin/bash

set -e

# Install Argo CD
echo "Installing latest Argo CD"
kubectl create namespace argocd --dry-run -o yaml | kubectl apply -f -
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Verify Argo CD installation
echo "Verifying Argo CD installation"
kubectl wait --for=condition=Ready --timeout=300s pods --all -n argocd
echo "Your Argo CD installation is complete"
echo "Run the following commands: "
echo "  1. kubectl port-forward svc/argocd-server -n argocd 8080:443"
echo "  2. Open a browser with URL: http://localhost:8080 with the following credential"
echo "     Username: 'admin', Password: '`kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d`'"
