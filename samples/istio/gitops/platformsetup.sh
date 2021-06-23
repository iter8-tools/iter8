#!/bin/bash

set -e

# Ensure ITER8 environment variable is set
if [[ -z ${ITER8} ]]; then
    echo "ITER8 environment variable needs to be set to the root folder of Iter8"
    exit 1
else
    echo "ITER8 is set to " $ITER8
fi

# Install base platform components
${ITER8}/samples/istio/quickstart/platformsetup.sh

# Install Argo CD
echo "Installing latest Argo CD"
kubectl create namespace argocd --dry-run -o yaml | kubectl apply -f -
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
#arch=`uname | awk '{print tolower($0)}'`
#if [ $arch  = "linux" ];
#then
    #arch="linux"
#else if [ $arch != "darwin" ];
    #then
        #echo "\"$arch\" is not a supported archicture"
        #exit 1
    #fi
#fi
#echo "Installing Argo CD CLI"
#VERSION=$(curl --silent "https://api.github.com/repos/argoproj/argo-cd/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
#curl -sSL -o ${ITER8}/argocd https://github.com/argoproj/argo-cd/releases/download/$VERSION/argocd-${arch}-amd64
#chmod +x ${ITER8}/argocd

# Verify Argo CD installation
echo "Verifying Argo CD installation"
kubectl wait --for=condition=Ready --timeout=300s pods --all -n argocd
echo "Your Argo CD installation is complete"
echo "Run the following commands: "
echo "  1. kubectl port-forward svc/argocd-server -n argocd 8080:443"
echo "  2. Open a browser with URL: http://localhost:8080 with the following credential"
echo "     Username: 'admin', Password: '`kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d`'"
