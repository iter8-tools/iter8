---
template: overrides/main.html
title: Install Iter8
---

## Install Iter8

=== "Iter8 for Knative"

    !!! example "Prerequisites"

        1. **Kubernetes cluster**
        2. [**kubectl** CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

    Install Iter8 in your Kubernetes cluster using the following command.

    ```shell
    source <(curl -s https://raw.githubusercontent.com/iter8-tools/iter8/master/install/install.sh)
    ```

    ??? info "Look inside install.sh"
        ```shell
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
        ```        

    ## Optional: Install iter8ctl
    The iter8ctl CLI enables real-time observability of Iter8 experiments. Go 1.13+ is a pre-requisite for iter8ctl.

    ```shell
    GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.2-pre.1
    ```

    ## Optional: Customizing Prometheus URL
    The URL of the Prometheus metrics backend is supplied as part of [this configmap](https://github.com/iter8-tools/iter8/blob/v0.2.5/install/iter8-analytics/config.yaml) during the install process. This URL is intended to match the location of the [iter8-monitoring install](#step-2-install-iter8-monitoring) above. To use your own Prometheus backend, replace the value of the metrics backend URL in the configmap during the install process with the URL of your Prometheus backend. You can use Kustomize or `sed` or any tool of your choice for this customization.
    
=== "Iter8 for KFServing"
    An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving) along with installation instructions. An updated version is coming soon and will be documented here.

=== "Iter8 for Istio"
    An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8) along with installation instructions. An updated version is coming soon and will be documented here.