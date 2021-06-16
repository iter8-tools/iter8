---
template: main.html
---

# Platform Setup for Istio

## 1. Create Kubernetes cluster

Create a local cluster using Kind or Minikube as follows, or use a managed Kubernetes cluster. Ensure that the cluster has sufficient resources, for example, 8 CPUs and 12GB of memory.

=== "Kind"

    ```shell
    kind create cluster --wait 5m
    kubectl cluster-info --context kind-kind
    ```

    ??? info "Ensuring your Kind cluster has sufficient resources"
        Your Kind cluster inherits the CPU and memory resources of its host. If you are using Docker Desktop, you can set its resources as shown below.

        ![Resources](../../images/ddresourcepreferences.png)

=== "Minikube"

    ```shell
    minikube start --cpus 8 --memory 12288
    ```

## 2. Clone Iter8 repo
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
```

## 3. Install Istio, Iter8 and Telemetry
Setup Istio, Iter8, a mock New Relic service, and Prometheus add-on within your cluster.

```shell
$ITER8/samples/istio/quickstart/platformsetup.sh
```
