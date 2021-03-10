---
template: overrides/main.html
title: Install Iter8
---

=== "Iter8 for Knative"
    Follow these steps to install Iter8 for Knative. 


    !!! example "Prerequisites"
        1. Kubernetes cluster with [Knative Serving](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-serving-component)
        2. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
        3. [Kustomize v3](https://kubectl.docs.kubernetes.io/installation/kustomize/), and 
        4. [Go 1.13+](https://golang.org/doc/install)

    ## Step 1: Export TAG
    ```shell
    export TAG=v0.2.5
    ```

    ## Step 2: Install iter8-monitoring
    ```shell
    kustomize build github.com/iter8-tools/iter8/install/monitoring/prometheus-operator/?ref=${TAG} | kubectl apply -f -
    kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
    kustomize build github.com/iter8-tools/iter8/install/monitoring/prometheus/?ref=${TAG} | kubectl apply -f - 
    ```

    ## Step 3: Install Iter8
    ```shell
    kustomize build github.com/iter8-tools/iter8/install/?ref=${TAG} | kubectl apply -f -
    kubectl wait crd -l creator=iter8 --for condition=established --timeout=120s
    kustomize build github.com/iter8-tools/iter8/install/iter8-metrics/?ref=${TAG} | kubectl apply -f -
    ```

    ## Step 4: Install iter8ctl
    ```shell
    GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.1-pre
    ```

    ## Optional: Customizing Iter8 install

    ### Prometheus URL
    The URL of the Prometheus metrics backend is supplied as part of [this configmap](https://github.com/iter8-tools/iter8/blob/v0.2.5/install/iter8-analytics/config.yaml) during the install process. This URL is intended to match the location of the [iter8-monitoring install](#step-2-install-iter8-monitoring) above. To use your own Prometheus backend, replace the value of the metrics backend URL in the configmap during the install process with the URL of your Prometheus backend. You can use `Kustomize` or `sed` or any tool of your choice for this customization.
    
=== "Iter8 for KFServing"
    An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving) along with installation instructions. An updated version is coming soon and will be documented here.

=== "Iter8 for Istio"
    An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8) along with installation instructions. An updated version is coming soon and will be documented here.