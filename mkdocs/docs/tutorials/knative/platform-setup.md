---
template: main.html
---

# Platform Setup

## 1. K8s cluster
Use a managed K8s cluster, or create a local K8s cluster as follows. If you wish to choose Istio as the networking layer for Knative, then ensure that the cluster has sufficient resources, for example, 8 CPUs and 12GB of memory.

=== "Kind"

    ```shell
    kind create cluster --wait 5m
    kubectl cluster-info --context kind-kind
    ```

    ??? info "Ensuring your Kind cluster has sufficient resources (needed for Istio networking layer)"
        Your Kind cluster inherits the CPU and memory resources of its host. If you are using Docker Desktop, you can set its resources as shown below.

        ![Resources](../../images/ddresourcepreferences.png)

=== "Minikube"

    ```shell
    minikube start # --cpus 8 --memory 12288 # (needed for Istio networking layer)
    ```

## 2. Install Knative Serving
=== "Learning/tutorial purposes"
    Use the Knative Serving install script, bundled as part of Iter8, as follows.

    * **Clone Iter8 repo**
    ```shell
    git clone https://github.com/iter8-tools/iter8.git
    cd iter8
    export ITER8=$(pwd)
    ```

    * **Install Knative Serving**
    Knative can work with multiple networking layers. So can Iter8. For a quick start with Knative and Iter8, we recommend Kourier.

        === "Kourier"

            ```shell
            $ITER8/samples/knative/quickstart/platform-setup.sh kourier
            ```

        === "Istio"

            ```shell
            $ITER8/samples/knative/quickstart/platform-setup.sh istio
            ```

=== "Production install of Knative / other networking layers"
    Refer to the [official Knative serving install instructions](https://knative.dev/docs/install/).

## 3. Install Iter8
Iter8 installation instructions are [here](../../getting-started/install.md).
