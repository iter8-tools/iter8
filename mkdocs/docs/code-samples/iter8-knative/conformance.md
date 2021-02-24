---
template: overrides/main.html
---

# Conformance Testing

Perform a conformance test of a Knative app.
You will create:

1. A Knative service with a single version.
2. A traffic genereator which sends HTTP GET requests to the Knative service.
3. An **iter8 experiment** that verifies that latency and error-rate metrics for the service satisfy the given objectives.

!!! warning "Before you begin"

    ** Kubernetes cluster:** Do not have a Kubernetes cluster with iter8 and Knative installed? Follow Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/) to create a cluster with iter8 and Knative.

    **Cleanup from previous experiment:** Tried an iter8 tutorial earlier but forgot to cleanup? Run the cleanup step from your tutorial now. For example, [Step 8](/getting-started/quick-start/with-knative/#8-cleanup) performs cleanup for the iter8-Knative quick start tutorial.

    **ITER8 environment variable:** ITER8 environment variable is not exported in your terminal? Do so now. For example, this is the [last command in Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo).

## 1. Create Knative app

```shell
kubectl apply -f $ITER8/samples/knative/conformance/baseline.yaml
```

??? info "Look inside baseline.yaml"

    ```yaml
    # apply this yaml at the start of the experiment to create the revision to be tested
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
    ```

## 2. Send requests
Verify Knative service is ready and send requests to app.
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/conformance/fortio.yaml | kubectl apply -f -
```

## 3. Create iter8 experiment
```shell
kubectl apply -f $ITER8/samples/knative/conformance/experiment.yaml
```
??? info "Look inside experiment.yaml"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
    name: conformance-sample
    spec:
    # target identifies the knative service under experimentation using its fully qualified name
    target: default/sample-app
    strategy:
        # this experiment will perform a conformance test
        testingPattern: Conformance
        actions:
        start: # run the following sequence of tasks at the start of the experiment
        - library: knative
            task: init-experiment
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
        intervalSeconds: 20
        iterationsPerLoop: 12
    versionInfo:
        # information about app versions used in this experiment
        baseline:
        name: current
        variables:
        - name: revision
            value: sample-app-v1 
    ```

## 4. Observe experiment

You can observe the experiment in realtime. Open three *new* terminals and follow instructions in the three tabs below.

=== "iter8ctl"
    Install **iter8ctl**. You can change the directory where iter8ctl binary is installed by changing GOBIN below.
    ```shell
    GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.0-pre
    ```

    Periodically describe the experiment.
        ```shell
        while clear; do
        kubectl get experiment conformance-sample -o yaml | iter8ctl describe -f -
        sleep 15
        done
        ```

        You should see output similar to the following.
        ```shell
        ****** Overview ******
        Experiment name: conformance-sample
        Experiment namespace: default
        Target: default/sample-app
        Testing pattern: Conformance
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Running
        Number of completed iterations: 3

        ****** Winner Assessment ******
        Winning version: not found
        Recommended baseline: current

        ****** Objective Assessment ******
        +--------------------------------+---------+
        |           OBJECTIVE            | CURRENT |
        +--------------------------------+---------+
        | mean-latency <= 50.000         | true    |
        +--------------------------------+---------+
        | 95th-percentile-tail-latency   | true    |
        | <= 100.000                     |         |
        +--------------------------------+---------+
        | error-rate <= 0.010            | true    |
        +--------------------------------+---------+

        ****** Metrics Assessment ******
        +--------------------------------+---------+
        |             METRIC             | CURRENT |
        +--------------------------------+---------+
        | request-count                  | 448.000 |
        +--------------------------------+---------+
        | mean-latency (milliseconds)    |   1.338 |
        +--------------------------------+---------+
        | 95th-percentile-tail-latency   |   4.770 |
        | (milliseconds)                 |         |
        +--------------------------------+---------+
        | error-rate                     |   0.000 |
        +--------------------------------+---------+
        ```    

=== "kubectl get experiment"

    ```shell
    kubectl get experiment conformance-sample --watch
    ```

    You should see output similar to the following.
    ```shell
    conformance-sample   Conformance   default/sample-app   Running        0                      StartHandlerLaunched: Start handler 'start' launched
    conformance-sample   Conformance   default/sample-app   Running        1                      IterationUpdate: Completed Iteration 1
    conformance-sample   Conformance   default/sample-app   Running        2                      IterationUpdate: Completed Iteration 2
    conformance-sample   Conformance   default/sample-app   Running        3                      IterationUpdate: Completed Iteration 3
    conformance-sample   Conformance   default/sample-app   Running        4                      IterationUpdate: Completed Iteration 4
    conformance-sample   Conformance   default/sample-app   Running        5                      IterationUpdate: Completed Iteration 5
    conformance-sample   Conformance   default/sample-app   Running        6                      IterationUpdate: Completed Iteration 6
    conformance-sample   Conformance   default/sample-app   Running        7                      IterationUpdate: Completed Iteration 7
    conformance-sample   Conformance   default/sample-app   Running        8                      IterationUpdate: Completed Iteration 8
    conformance-sample   Conformance   default/sample-app   Running        9                      IterationUpdate: Completed Iteration 9
    conformance-sample   Conformance   default/sample-app   Running        10                     IterationUpdate: Completed Iteration 10
    conformance-sample   Conformance   default/sample-app   Running        11                     IterationUpdate: Completed Iteration 11
    conformance-sample   Conformance   default/sample-app   Completed      12                     ExperimentCompleted: Experiment completed successfully
    ```

When the experiment completes (in ~ 4 mins), you will see the experiment stage change from `Running` to `Completed`.

## 5. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/conformance/fortio.yaml
kubectl delete -f $ITER8/samples/knative/conformance/experiment.yaml
```

??? info "Understanding what happened"
    1. In Step 1, you created a Knative service with a single revision, `sample-app-v1`.
    2. In Step 2, you created a load generator that sends requests to the Knative service.
    3. In step 3, you created an iter8 experiment with 12 iterations with the above Knative service as the `target` of the experiment. In each iteration, iter8 observed the `mean-latency`, `95th-percentile-tail-latency`, and `error-rate` metrics for the revisions (collected by Prometheus).It ensured that the deployed revision satisfied all objectives specified in `experiment.yaml`.
    4. At the end of the experiment, iter8 did not identify a `winner` since there is no winner in conformance experiment.
