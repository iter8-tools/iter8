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
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: canary-exp
    spec:
      target: default/sample-app
      strategy:
        testingPattern: Canary
        deploymentPattern: Progressive
        actions:
          finish: # run the following sequence of tasks at the end of the experiment
          - task: common/exec # promote the winning version      
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
                kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml
      criteria:
        requestCount: iter8-knative/request-count
        objectives: 
        - metric: iter8-knative/mean-latency
          upperLimit: 50
        - metric: iter8-knative/95th-percentile-tail-latency
          upperLimit: 100
        - metric: iter8-knative/error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        # information about app versions used in this experiment
        baseline:
          name: sample-app-v1
          weightObjRef:
            apiVersion: serving.knative.dev/v1
            kind: Service
            name: sample-app
            namespace: default
            fieldPath: .spec.traffic[0].percent
          variables:
          - name: promote
            value: baseline
        candidates:
        - name: sample-app-v2
          weightObjRef:
            apiVersion: serving.knative.dev/v1
            kind: Service
            name: sample-app
            namespace: default
            fieldPath: .spec.traffic[1].percent
          variables:
          - name: promote
            value: candidate
    ```

## 4. Understand the experiment
Follow [Step 6 of the quick start tutorial](../../../../getting-started/quick-start/kfserving/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`slovalidation-exp`) in your `iter8ctl` and `kubectl` commands.

## 5. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/slovalidation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```
