---
template: overrides/main.html
---

# Quick Start with Knative Tutorial

!!! tip ""
    Perform an Iter8-Knative experiment with [`Canary`](/concepts/experimentationstrategies/#testing-pattern) testing, [`Progressive`](/concepts/experimentationstrategies/#deployment-pattern) deployment, and [`kubectl` based version promotion](/concepts/experimentationstrategies/#version-promotion).
    
    ![Canary](/assets/images/canary-progressive-kubectl.png)

You will create the following resources in this tutorial.

1. A **Knative app (service)** with two versions (revisions).
2. A **fortio-based traffic generator** that simulates user requests.
3. An **Iter8 experiment** that: 
    - verifies that `candidate` satisfies mean latency, 95th percentile tail latency, and error rate `objectives`
    - progressively shifts traffic from `baseline` to `candidate`
    - eventually replaces `baseline` with `candidate` using `kubectl`
??? warning "Before you begin, you will need ... "

    1. **Kubernetes cluster.** You can setup a local cluster using [Minikube](https://minikube.sigs.k8s.io/docs/) or [Kind](https://kind.sigs.k8s.io/)
    2. [**kubectl**](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
    3. [**Kustomize v3**](https://kubectl.docs.kubernetes.io/installation/kustomize/), and 
    4. [**Go 1.13+**](https://golang.org/doc/install)

## 1. Create Kubernetes cluster

Create a local Kubernetes cluster or use a managed Kubernetes service from your cloud provider. Ensure that the cluster has sufficient resources, for example, 2 cpus and 4GB of memory.

=== "Minikube"

    ```shell
    minikube start
    ```

=== "Kind"

    ```shell
    kind create cluster
    kubectl cluster-info --context kind-kind
    ```
    

## 2. Clone Iter8 repo
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
```

## 3. Install Knative and Iter8
Choose a networking layer for Knative.

=== "Istio"

    ```shell
    $ITER8/samples/knative/quickstart/platformsetup.sh istio
    ```

=== "Contour"

    ```shell
    $ITER8/samples/knative/quickstart/platformsetup.sh contour
    ```

=== "Kourier"

    ```shell
    $ITER8/samples/knative/quickstart/platformsetup.sh kourier
    ```

=== "Gloo"
    This step requires Python. This will install `glooctl` binary under `$HOME/.gloo` folder.
    ```shell
    $ITER8/samples/knative/quickstart/platformsetup.sh gloo
    ```

## 4. Create app versions
```shell
kubectl apply -f $ITER8/samples/knative/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```

??? info "Look inside baseline.yaml"
    ```yaml linenums="1"
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app
      namespace: default
    spec:
      template:
        metadata:
          name: sample-app-v1
        spec:
          containers:
          - image: gcr.io/knative-samples/knative-route-demo:blue 
            env:
            - name: T_VERSION
              value: "blue"
    ```

??? info "Look inside experimentalservice.yaml"
    ```yaml linenums="1"
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app # name of the app
      namespace: default # namespace of the app
    spec:
      template:
        metadata:
          name: sample-app-v2
        spec:
          containers:
          - image: gcr.io/knative-samples/knative-route-demo:green 
            env:
            - name: T_VERSION
              value: "green"
      traffic:
      # initially all traffic goes to sample-app-v1 and none to sample-app-v2
      - tag: current
        revisionName: sample-app-v1
        percent: 100
      - tag: candidate
        latestRevision: true
        percent: 0
    ```

## 5. Generate requests
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/quickstart/fortio.yaml | kubectl apply -f -
```

??? info "Look inside fortio.yaml"
    ```yaml linenums="1"
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: ["fortio", "load", "-t", "100s", "-qps", "16","-json", "/shared/fortiooutput.json", $(URL)]
            env:
            - name: URL
              value: URL_VALUE
            volumeMounts:
            - name: shared
              mountPath: /shared
          # useful for extracting fortiooutput.json out of here
          - name: busybox 
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 600']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never
    ```

## 6. Create Iter8 experiment
```shell
kubectl apply -f $ITER8/samples/knative/quickstart/experiment.yaml
```
??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
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
          finish: # run the following sequence of tasks at the end of the experiment
          - library: common
            task: exec # promote the winning version
            with:
              cmd: kubectl
              args: 
              - "apply"
              - "-f"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
      criteria:
        # mean latency of version should be under 50 milliseconds
        # 95th percentile latency should be under 100 milliseconds
        # error rate should be under 1%
        objectives: 
        - metric: mean-latency
          upperLimit: 50
        - metric: 95th-percentile-tail-latency
          upperLimit: 100
        - metric: error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        # information about app versions used in this experiment
      baseline:
        name: current
        variables:
        # variables are used when querying metrics and when interpolating task inputs
        - name: revision
          value: sample-app-v1 
        - name: promote
          value: baseline
      candidates:
      - name: candidate
        variables:
        # variables are used when querying metrics and when interpolating task inputs
        - name: revision
          value: sample-app-v2
        - name: promote
          value: candidate 
    ```

## 7. Observe experiment
Observe the experiment in realtime. Paste commands from the tabs below in separate terminals.

=== "iter8ctl"
    Install **iter8ctl**. You can change the directory where iter8ctl binary is installed by changing GOBIN below.
    ```shell
    GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.0
    ```

    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment quickstart-exp -o yaml | iter8ctl describe -f -
    sleep 4
    done
    ```
    ??? info "iter8ctl output"
        iter8ctl output will be similar to the following.
        ```shell
        ****** Overview ******
        Experiment name: quickstart-exp
        Experiment namespace: default
        Target: default/sample-app
        Testing pattern: Canary
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Running
        Number of completed iterations: 3

        ****** Winner Assessment ******
        App versions in this experiment: [current candidate]
        Winning version: candidate
        Recommended baseline: candidate

        ****** Objective Assessment ******
        +--------------------------------+---------+-----------+
        |           OBJECTIVE            | CURRENT | CANDIDATE |
        +--------------------------------+---------+-----------+
        | mean-latency <= 50.000         | true    | true      |
        +--------------------------------+---------+-----------+
        | 95th-percentile-tail-latency   | true    | true      |
        | <= 100.000                     |         |           |
        +--------------------------------+---------+-----------+
        | error-rate <= 0.010            | true    | true      |
        +--------------------------------+---------+-----------+

        ****** Metrics Assessment ******
        +--------------------------------+---------+-----------+
        |             METRIC             | CURRENT | CANDIDATE |
        +--------------------------------+---------+-----------+
        | request-count                  | 429.334 |    16.841 |
        +--------------------------------+---------+-----------+
        | mean-latency (milliseconds)    |   0.522 |     0.712 |
        +--------------------------------+---------+-----------+
        | 95th-percentile-tail-latency   |   4.835 |     4.750 |
        | (milliseconds)                 |         |           |
        +--------------------------------+---------+-----------+
        | error-rate                     |   0.000 |     0.000 |
        +--------------------------------+---------+-----------+
        ``` 
        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.   

=== "kubectl get experiment"

    ```shell
    kubectl get experiment quickstart-exp --watch
    ```

    ??? info "kubectl get experiment output"
        kubectl output will be similar to the following.
        ```shell
        NAME             TYPE     TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
        quickstart-exp   Canary   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
        quickstart-exp   Canary   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
        quickstart-exp   Canary   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
        quickstart-exp   Canary   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
        quickstart-exp   Canary   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
        quickstart-exp   Canary   default/sample-app   Running   6                      IterationUpdate: Completed Iteration 6
        quickstart-exp   Canary   default/sample-app   Running   7                      IterationUpdate: Completed Iteration 7
        quickstart-exp   Canary   default/sample-app   Running   8                      IterationUpdate: Completed Iteration 8
        quickstart-exp   Canary   default/sample-app   Running   9                      IterationUpdate: Completed Iteration 9
        ```
        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.

=== "kubectl get ksvc"

    ```shell
    kubectl get ksvc sample-app -o json --watch | jq .status.traffic
    ```

    ??? info "kubectl get ksvc output"
        kubectl output will be similar to the following.
        ```shell
        [
          {
            "latestRevision": false,
            "percent": 45,
            "revisionName": "sample-app-v1",
            "tag": "current",
            "url": "http://current-sample-app.default.example.com"
          },
          {
            "latestRevision": true,
            "percent": 55,
            "revisionName": "sample-app-v2",
            "tag": "candidate",
            "url": "http://candidate-sample-app.default.example.com"
          }
        ]
        ```

## 8. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```

??? info "Understanding what happened"
    1. You created a Knative service with two revisions, sample-app-v1 (`baseline`) and sample-app-v2 (`candidate`).
    2. You generated requests for the Knative service using a fortio-job. At the start of the experiment, 100% of the requests are sent to `baseline` and 0% to `candidate`.
    3. You created an Iter8 `Canary` experiment with `Progressive` deployment pattern. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, verified that `candidate` satisfied all the `objectives` specified in the experiment, identified `candidate` as the `winner`, progressively shifted traffic from `baseline` to `candidate` and eventually promoted the `candidate` using `kubectl`.
        - **Note:** Had `candidate` failed to satisfy `objectives`, then `baseline` would have been promoted.
