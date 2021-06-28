---
template: main.html
---

# Your First Experiment

!!! tip "Scenario: Safely rollout a Kubernetes deployment with SLO validation"
    In this tutorial, you will:

    1. Deploy stable and candidate versions of an application. The latter will be [dark launched](../concepts/buildingblocks.md#dark-launch).
    2. Create an [Iter8 SLO validation experiment](../concepts/buildingblocks.md#slo-validation) with *latency* and *error-rate* SLOs.
    3. Use Iter8's [builtin latency and error-rate metrics collection](../metrics/builtin.md) feature to evaluate candidate.
    4. Verify that candidate satisfies SLOs and promote it. It will no longer be a dark version but will serve end-user requests as the latest stable version.
    
    ![SLO validation](../../../images/yourfirstexperiment.png)

???+ warning "Before you begin, you will need... "
    1. [Helm 3+](https://helm.sh/docs/intro/install/). Used to install the application.
    2. [Go 1.13+](https://golang.org/doc/install). Used to install Iter8's client side observability tool `iter8ctl`.

## 1. Create K8s cluster
Create a local K8s cluster as follows. You may skip this step if you already have a K8s cluster to work with.

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
Install Iter8 using [these steps](install.md).

## 3. Create stable version
```shell
helm repo add iter8 https://iter8-tools.github.io/iter8/
```

```shell
helm install \
  --set stable=gcr.io/google-samples/hello-app:1.0 \
  my-app iter8/deploy
```

??? note "Verify that stable version is up"
    ```shell
    # do this in a separate terminal
    kubectl port-forward svc/hello 8080:8080
    ```

    ```shell
    curl localhost:8080
    ```

    ```
    # output will be similar to the following (notice 1.0.0 version tag)
    # hostname will be different in your environment
    Hello, world!
    Version: 1.0.0
    Hostname: hello-bc95d9b56-xp9kv
    ```

<!-- 
```shell
kubectl create deploy hello --image=gcr.io/google-samples/hello-app:1.0
kubectl create svc clusterip hello --tcp=8080
``` 
-->

## 4. Create candidate version
```shell
helm upgrade --install \
  --set candidate=gcr.io/google-samples/hello-app:2.0 \
  --set LimitMeanLatency=50.0 \
  --set LimitErrorRate=0.0 \
  --set Limit95thPercentileLatency=100.0 \
  my-app iter8/deploy
```

??? note "Verify that candidate version is up"
    ```shell
    # do this in a separate terminal
    kubectl port-forward svc/hello-candidate 8081:8080
    ```

    ```shell
    curl localhost:8081
    ```

    ```
    # output will be similar to the following (notice 2.0.0 version tag)
    # hostname will be different in your environment
    Hello, world!
    Version: 2.0.0
    Hostname: hello-bc95d9b56-xp9kv
    ```

A high-level overview of the process automated by an Iter8 experiment is shown [here](../concepts/whatisiter8.md#what-is-an-iter8-experiment).

<!-- 
```shell
kubectl create deploy hello-candidate --image=gcr.io/google-samples/hello-app:2.0
kubectl create svc clusterip hello-candidate --tcp=8080
``` 
-->

## 5. Observe experiment results
To observe the results of the experiment in real-time, install `iter8ctl`. You can change the directory where `iter8ctl` binary is installed by changing `GOBIN` below.
```shell
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.4
```

Periodically describe the experiment results.
```shell
watch -x iter8ctl describe last
```

??? info "Verify SLOs are satisfied (see `Objective Assessment` section of the output) ..."
    ```shell
    ****** Overview ******
    Experiment name: my-experiment
    Experiment namespace: default
    Target: my-app
    Testing pattern: Conformance
    Deployment pattern: Progressive

    ****** Progress Summary ******
    Experiment stage: Completed
    Number of completed iterations: 1

    ****** Winner Assessment ******
    > If the version being validated; i.e., the baseline version, satisfies the experiment objectives, it is the winner.
    > Otherwise, there is no winner.
    Winning version: my-app

    ****** Objective Assessment ******
    > Identifies whether or not the experiment objectives are satisfied by the most recently observed metrics values for each version.
    +--------------------------------------+--------+
    |              OBJECTIVE               | MY-APP |
    +--------------------------------------+--------+
    | iter8-system/mean-latency <=         | true   |
    |                               50.000 |        |
    +--------------------------------------+--------+
    | iter8-system/error-rate <=           | true   |
    |                                0.000 |        |
    +--------------------------------------+--------+
    | iter8-system/latency-95th-percentile | true   |
    | <= 100.000                           |        |
    +--------------------------------------+--------+

    ****** Metrics Assessment ******
    > Most recently read values of experiment metrics for each version.
    +--------------------------------------+--------+
    |                METRIC                | MY-APP |
    +--------------------------------------+--------+
    | iter8-system/mean-latency            |  1.233 |
    +--------------------------------------+--------+
    | iter8-system/error-rate              |  0.000 |
    +--------------------------------------+--------+
    | iter8-system/latency-95th-percentile |  2.311 |
    +--------------------------------------+--------+
    | iter8-system/request-count           | 40.000 |
    +--------------------------------------+--------+
    | iter8-system/error-count             |  0.000 |
    +--------------------------------------+--------+
    ``` 

## 6. Promote winner
Promote the winning version at the end of an experiment as follows.

```shell
helm install \
  --set stable=gcr.io/google-samples/hello-app:2.0 \
  my-app iter8/deploy
```

??? note "Verify that candidate is the latest stable version ..."
    ```shell
    curl localhost:8080
    ```

    ```
    # output will be similar to the following (notice 2.0.0 version tag)
    # hostname will be different in your environment
    Hello, world!
    Version: 2.0.0
    Hostname: hello-bc95d9b56-xp9kv
    ```

## 7. Cleanup
```shell
helm uninstall my-app
```

## 8. Use in production
1. Details about the Helm chart... 

2. Iter8 can integrate with any CI/CD/GitOps pipeline. Click on the links below to see variations of this tutorial in GitHub actions and GitOps (ArgoCD) settings.

[GitHub Actions](#){ .md-button .md-button--primary }
[GitOps (ArgoCD)](#){ .md-button .md-button--primary }

## 9. Try more Iter8 tutorials
[KFServing](#){ .md-button .md-button--primary }
[Seldon](#){ .md-button .md-button--primary }
[Knative](#){ .md-button .md-button--primary }
[Istio](#){ .md-button .md-button--primary }

