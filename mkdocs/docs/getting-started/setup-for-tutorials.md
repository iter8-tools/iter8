---
template: main.html
---

# Setup for Tutorials
Setup common components that are required in Iter8 tutorials.

## Local Kubernetes cluster
Use a managed K8s cluster or a local K8s cluster for running Iter8 tutorials. You can setup the latter using [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) or [Minikube](https://minikube.sigs.k8s.io/docs/) as follows.

=== "Kind"
    ```shell
    kind create cluster --wait 5m
    kubectl cluster-info --context kind-kind
    ```
=== "Minikube"
    ```shell
    minikube start
    ```

??? info "Setting CPU and memory resources for your local K8s cluster"
    Iter8 tutorials in certain K8s environments, especially those involving Istio, require additional CPU and memory resources. You can set this as follows.

    === "Kind"
        A Kind cluster inherits its CPU and memory resources of its host. If you are using Docker Desktop, you can set its resources as follows.

        ![Resources](../images/ddresourcepreferences.png)

    === "Minikube"
        Set CPU and memory resources while starting Minikube as follows.
        ```shell
        minikube start --cpus 8 --memory 12288
        ```

## Iter8 GitHub repo
Clone the Iter8 GitHub repo and set the ITER8 environment variable as follows:

```shell
git clone https://github.com/iter8-tools/iter8.git
export ITER8=./iter8
```

## Iter8 Helm repo
Iter8 Helm repo contains charts that support Iter8's Helmex tutorials. Get the repo as follows:

```shell
helm repo add iter8 https://iter8-tools.github.io/iter8/ --force-update
```