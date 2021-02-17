---
template: overrides/main.html
---

# Quick start in 5 minutes

Follow this tutorial to perform **progressive canary rollout of a Knative app**. You will create:

1. A Knative service with two versions of your app, namely, *baseline* and *candidate*
2. A traffic generator which sends HTTP GET requests to the Knative service.
3. An **iter8 experiment** that verifies that latency and error-rate metrics for the *candidate* satisfy the given objectives, iteratively shifts traffic from *baseline* to *candidate*, and finally replaces *baseline* with *candidate*.

!!! example "Before you begin, you will need:"

    1. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
    2. [Kustomize v3](https://kubectl.docs.kubernetes.io/installation/kustomize/), and 
    3. [Go 1.13+](https://golang.org/doc/install)

## Create a Kubernetes cluster

Create a local Kubernetes cluster using Minikube or Kind. You can also use a managed Kubernetes service from your cloud provider.

=== "Minikube"

    ```shell
    minikube start --cpus 6 --memory 12288
    ```

=== "Kind"

    ```shell
    kind create cluster
    ```
    Ensure that the cluster has sufficient resources (for example, 5 cpus and 12GB of memory).

## Clone this repository
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8/knative
```

## Install Knative and iter8
Choose a networking layer for Knative. Install Knative and iter8.

=== "Istio"

    ```shell
    ./quickstart/platformsetup.sh istio
    ```

=== "Contour"

    ```shell
    ./quickstart/platformsetup.sh contour
    ```

=== "Kourier"

    ```shell
    ./quickstart/platformsetup.sh kourier
    ```

=== "Gloo"
    This step requires Python. This will install `glooctl` binary under `$HOME/.gloo` folder.
    ```shell
    ./quickstart/platformsetup.sh gloo
    ```

## Create Knative service
```shell
kubectl apply -f ./quickstart/service.yaml
```

??? info "Service YAML"
    ```yaml
    # Service objects for progressive experiment
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app # The name of the app
      namespace: default # The namespace the app will use
    spec:
      template:
        metadata:
          name: sample-app-v1
        spec:
          containers:
          # The URL to the sample app docker image
          - image: gcr.io/knative-samples/knative-route-demo:blue 
            env:
            - name: T_VERSION
              value: "blue"
    ---
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app # The name of the app
      namespace: default # The namespace the app will use
    spec:
      template:
        metadata:
          name: sample-app-v2
        spec:
          containers:
          # The URL to the sample app docker image
          - image: gcr.io/knative-samples/knative-route-demo:green 
            env:
            - name: T_VERSION
              value: "green"
      traffic:
      - tag: current
        revisionName: sample-app-v1
        percent: 100
      - tag: candidate
        latestRevision: true
        percent: 0
    ```

## Generate traffic
Verify Knative service is ready and generates traffic.
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" quickstart/fortio.yaml | kubectl apply -f -
```

## Create iter8 experiment
```shell
kubectl apply -f ./quickstart/experiment.yaml
```
??? info "Experiment YAML"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      # target identifies the knative service under experimentation using its fully qualified name
      target: default/sample-app
      strategy:
        # this experiment will perform a canary test
        testingPattern: Canary
        actions:
          start: # run a sequence of tasks at the start of the experiment
          - library: knative
            task: init-experiment
      criteria:
        objectives:
        # mean latency of version should be under 500 milli seconds; error rate under 1%    
        - metric: mean-latency
          upperLimit: 500
        - metric: error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 20
        iterationsPerLoop: 12
      versionInfo:
        # information about app versions used in this experiment
        baseline:
          name: current
          variables:
          - name: revision
            value: sample-app-v1 
        candidates:
          - name: candidate
            variables:
            - name: revision
              value: sample-app-v2
    ```

## Observe experiment in realtime

Follow instructions in the three tabs below in *separate* terminals.

=== "using iter8ctl"
    Install **iter8ctl**. You can change the directory where iter8ctl binary is installed by changing GOBIN below.
    ```shell
    GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.0-pre
    ```

    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment quickstart-exp -o yaml | iter8ctl describe -f -
    sleep 15
    done
    ```

    You should see output similar to the following.
    ```shell
    ****** Overview ******
    Experiment name: quickstart-exp
    Experiment namespace: default
    Target: default/sample-app
    Testing pattern: Canary
    Deployment pattern: Progressive

    ****** Progress Summary ******
    Experiment stage: Running
    Number of completed iterations: 7

    ****** Winner Assessment ******
    App versions in this experiment: [current candidate]
    Winning version: candidate
    Recommended baseline: candidate

    ****** Objective Assessment ******
    +-------------------------+---------+-----------+
    |        OBJECTIVE        | CURRENT | CANDIDATE |
    +-------------------------+---------+-----------+
    | mean-latency <= 500.000 | true    | true      |
    +-------------------------+---------+-----------+
    | error-rate <= 0.010     | true    | true      |
    +-------------------------+---------+-----------+

    ****** Metrics Assessment ******
    +-----------------------------+---------+-----------+
    |           METRIC            | CURRENT | CANDIDATE |
    +-----------------------------+---------+-----------+
    | request-count               | 984.528 |   401.720 |
    +-----------------------------+---------+-----------+
    | mean-latency (milliseconds) |   1.187 |     1.208 |
    +-----------------------------+---------+-----------+
    | error-rate                  |   0.000 |     0.000 |
    +-----------------------------+---------+-----------+
    ```    

=== "using experiment object"

    ```shell
    kubectl get experiment quickstart-exp --watch
    ```

    You should see output similar to the following.
    ```shell
    experiment output
    ```

    

=== "using knative service object"

    ```shell
    kubectl get ksvc sample-app --watch
    ```

    You should see output similar to the following.
    ```shell
    ksvc output
    ```

When the experiment completes, you will see the experiment stage change from `Running` to `Completed` in the `iter8ctl` output.

## Cleanup
```shell
kubectl delete -f ./quickstart/fortio.yaml
kubectl delete -f ./quickstart/experiment.yaml
kubectl delete -f ./quickstart/service.yaml
```

??? info "Understanding what happened"
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod
    nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor
    massa, nec semper lorem quam in massa.

    ``` python
    def bubble_sort(items):
        for i in range(len(items)):
            for j in range(len(items) - 1 - i):
                if items[j] > items[j + 1]:
                    items[j], items[j + 1] = items[j + 1], items[j]
    ```

    Nunc eu odio eleifend, blandit leo a, volutpat sapien. Phasellus posuere in
    sem ut cursus. Nullam sit amet tincidunt ipsum, sit amet elementum turpis.
    Etiam ipsum quam, mattis in purus vitae, lacinia fermentum enim.
