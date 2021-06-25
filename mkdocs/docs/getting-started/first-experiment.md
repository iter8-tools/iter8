---
template: main.html
---

# SLO Validation

!!! tip "Scenario: SLO validation and safe rollout of a Kubernetes deployment"
    This tutorial illustrates a simple [SLO validation experiment](../../../concepts/buildingblocks.md#slo-validation). Dark launch a candidate deployment, and promote it after using Iter8 to validate that it satisfies service-level objectives (SLOs). You will:

    1. Specify *latency* and *error-rate* based service-level objectives (SLOs).
    2. Use Iter8's builtin capabilities for collecting latency and error-rate metrics.
    
    ![SLO validation](../../../images/darklaunchbuiltin.png)

???+ warning "Before you begin, you will need... "
    1. The [kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
    2. [Kustomize 3+](https://kubectl.docs.kubernetes.io/installation/kustomize/).
    3. [Go 1.13+](https://golang.org/doc/install).
    4. [Helm 3+](https://helm.sh/docs/intro/install/)    

## 1. Create K8s cluster
Use a managed K8s cluster, or create a local K8s cluster as follows.

=== "Kind"

    ```shell
    kind create cluster --wait 5m
    kubectl cluster-info --context kind-kind
    ```

=== "Minikube"

    ```shell
    minikube start
    ```

## 2. Install Iter8
Iter8 installation instructions are [here](install.md).

## 3. Create stable version of your application
```shell
kubectl create deployment hello --image=gcr.io/google-samples/hello-app:1.0
kubectl create service clusterip hello --tcp=8080
```

Port-forward your service in a new terminal.
```shell
kubectl port-forward svc/hello 8080:8080
```

You can now access it using `curl localhost:8080`.
```shell
# output of curl; hostname will differ in your environment
Hello, world!
Version: 1.0.0
Hostname: hello-bc95d9b56-xp9kv
```

## 4. Create experimental version of your application
```shell
kubectl create deployment hello-experimental --image=gcr.io/google-samples/hello-app:2.0
kubectl create service clusterip hello-experimental --tcp=8080
```

## 5. Launch experiment
Launch the SLO validation experiment. This experiment will generate requests for your experimental version, collect latency and error-rate metrics, and declare the experimental version as `winner` if it satisfies SLOs.

```shell
helm repo add iter8 https://iter8-tools.github.io/iter8/
helm install \
  --set URL=http://hello.default.svc.cluster.local:8080 \
  --set LimitMeanLatency='"50.0"' \
  --set LimitErrorRate='"0.0"' \
  --set Limit95thPercentileLatency='"100.0"' \
  experiment iter8/conformance
```

## 6. Understand the experiment

## 5. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/slovalidation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```
