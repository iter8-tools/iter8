---
template: overrides/main.html
title: Installation
---

# Installation

## Step 1: Iter8

!!! example "Prerequisites"

    1. A Kubernetes cluster
    2. [kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

Install Iter8 in your Kubernetes cluster as follows.

```shell
export TAG=v0.3.0
curl -s https://raw.githubusercontent.com/iter8-tools/iter8-install/main/install.sh | bash
```

??? info "Look inside install.sh"
    ```shell
    #!/bin/bash

    set -e

    # Step 0: Export TAG
    export TAG="${TAG:-v0.3.0}"

    # Step 1: Install Iter8
    echo "Installing Iter8"
    kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/core/build.yaml
    kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
    kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/metrics/build.yaml

    echo "Verifying Iter8 installation"
    kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system

    set +e
    ```        

## (Optional) Step 2: Prometheus add-on

Install Iter8's Prometheus add-on in your cluster as follows. This step assumes you have installed Iter8 following Step 1 above.

```shell
export TAG=v0.3.0
curl -s https://raw.githubusercontent.com/iter8-tools/iter8-install/main/install-prom-add-on.sh | bash
```

??? info "Look inside install-prom-add-on.sh"
    ```shell
    #!/bin/bash

    set -e

    # Step 0: Export TAG
    export TAG="${TAG:-v0.3.0}"

    # Step 1: Install Prometheus add-on
    # This step assumes you have installed Iter8 using install.sh
    echo "Installing Prometheus add-on"
    kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus-operator/build.yaml
    kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
    kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/prometheus/build.yaml
    kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-install/${TAG}/prometheus-add-on/service-monitors/build.yaml

    echo "Verifying Prometheus-addon installation"
    kubectl wait --for condition=ready --timeout=300s pods --all -n iter8-system

    set +e
    ```

??? note "Running Iter8 tutorials without Iter8's Prometheus add-on"
    When you installed Iter8 in the first step above, you also installed several *out-of-the-box* Iter8 metric resources. They are required for running the tutorials documented on this site. 
    
    The out-of-the-box metric resources have a urlTemplate field. This field is configured as the URL of the Prometheus instance created in this step. 
    
    You can skip this step and still run Iter8 tutorials using your own Prometheus instance. To do so, ensure that your Prometheus instance scrapes the end-points that would have been scraped by the Prometheus instance created in this step, and configure the urlTemplate fields of Iter8 metric resources to match the URL of your Prometheus instance.

## (Optional) Step 3: iter8ctl
The iter8ctl CLI enables real-time observability of Iter8 experiments. 

!!! example "Prerequisites"

    Go 1.13+

```shell
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.2
```