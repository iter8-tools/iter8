---
template: overrides/main.html
---

# Canary with FixedSplit Deployment

Perform **zero-downtime fixed-split canary release of a Knative app**. You will create:

1. A Knative service with two versions of your app, namely, `baseline` and `candidate` using `kustomize`.
2. A traffic generator which sends HTTP GET requests to the Knative service.
3. An **iter8 experiment** that automates the following:
    - verifies that latency and error-rate metrics for the `candidate` satisfy the given objectives
    - traffic split between `baseline` and `candidate` will be remain unchanged during experiment iterations
    - replaces `baseline` with `candidate` in the end using a `kustomize` command

!!! warning "Before you begin"
    **Kubernetes cluster:** Do not have a Kubernetes cluster with iter8 and Knative installed? Follow Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/) to create a cluster with iter8 and Knative.

    **Cleanup from previous experiment:** Tried an iter8 tutorial earlier but forgot to cleanup? Run the cleanup step from your tutorial now. For example, [Step 8](/getting-started/quick-start/with-knative/#8-cleanup) performs cleanup for the iter8-Knative quick start tutorial.

    **ITER8 environment variable:** ITER8 environment variable is not exported in your terminal? Do so now. For example, this is the [last command in Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo).

    **Kustomize and iter8ctl:** [Kustomize](https://kustomize.io/) and [iter8ctl](/getting-started/install/#step-4-install-iter8ctl) are not installed locally? Do so now.

## 1. Create Knative app with canary

```shell
kustomize build $ITER8/samples/knative/canaryfixedsplit/baseline | kubectl apply -f -
kubectl wait ksvc/sample-app --for condition=Ready --timeout=120s
kustomize build $ITER8/samples/knative/canaryfixedsplit/experimentalservice | kubectl apply -f -
```

??? info "Look inside baseline/app.yaml"
    ```yaml
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
      ```

??? info "Look inside experimentalservice/app.yaml"
    ```yaml
    # This Knative service will be used for the iter8 experiment with traffic split between baseline and candidate revision
    # Traffic is split 75/25 between the baseline and candidate
    # Apply this after applying baseline.yaml in order to create the second revision
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
          # Docker image used by second revision
          - image: gcr.io/knative-samples/knative-route-demo:green 
            env:
            - name: T_VERSION
              value: "green"
      traffic: # 75% goes to sample-app-v1 and 25% to sample-app-v2
      - tag: current
        revisionName: sample-app-v1
        percent: 75
      - tag: candidate
        latestRevision: true
        percent: 25
    ```

## 2. Send requests

Verify Knative service is ready and send requests to app.
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/canaryfixedsplit/fortio.yaml | kubectl apply -f -
```

## 3. Create iter8 experiment

```shell
kubectl apply -f $ITER8/samples/knative/canaryfixedsplit/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: canary-fixedsplit
    spec:
      # target identifies the knative service under experimentation using its fully qualified name
      target: default/sample-app
      strategy:
        # this experiment will perform a canary test
        testingPattern: Canary
        deploymentPattern: FixedSplit
        actions:
          start: # run the following sequence of tasks at the start of the experiment
          - library: knative
            task: init-experiment
          finish: # run the following sequence of tasks at the end of the experiment
          - library: common
            task: exec # promote the winning version using Helm upgrade
            with:
              cmd: eval
              args:
              - "kustomize build github.com/iter8-tools/iter8/samples/knative/canaryfixedsplit/{{ .name }}?ref=master | kubectl apply -f -"
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
          name: baseline
          variables:
          - name: revision
            value: sample-app-v1
        candidates:
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2
    ```

## 4. Observe experiment

You can observe the experiment in realtime. Open three *new* terminals and follow instructions in the three tabs below.

=== "iter8ctl"
    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment canary-fixedsplit -o yaml | iter8ctl describe -f -
    sleep 2
    done
    ```

    You should see output similar to the following.
    ```shell
    ****** Overview ******
    Experiment name: canary-fixedsplit
    Experiment namespace: default
    Target: default/sample-app
    Testing pattern: Canary
    Deployment pattern: FixedSplit

    ****** Progress Summary ******
    Experiment stage: Running
    Number of completed iterations: 5

    ****** Winner Assessment ******
    App versions in this experiment: [baseline candidate]
    Winning version: candidate
    Recommended baseline: candidate

    ****** Objective Assessment ******
    +--------------------------------+----------+-----------+
    |           OBJECTIVE            | BASELINE | CANDIDATE |
    +--------------------------------+----------+-----------+
    | mean-latency <= 50.000         | true     | true      |
    +--------------------------------+----------+-----------+
    | 95th-percentile-tail-latency   | true     | true      |
    | <= 100.000                     |          |           |
    +--------------------------------+----------+-----------+
    | error-rate <= 0.010            | true     | true      |
    +--------------------------------+----------+-----------+

    ****** Metrics Assessment ******
    +--------------------------------+----------+-----------+
    |             METRIC             | BASELINE | CANDIDATE |
    +--------------------------------+----------+-----------+
    | 95th-percentile-tail-latency   |    4.798 |     4.825 |
    | (milliseconds)                 |          |           |
    +--------------------------------+----------+-----------+
    | error-rate                     |    0.000 |     0.000 |
    +--------------------------------+----------+-----------+
    | request-count                  |  652.800 |   240.178 |
    +--------------------------------+----------+-----------+
    | mean-latency (milliseconds)    |    1.270 |     1.254 |
    +--------------------------------+----------+-----------+
    ```    

=== "kubectl get experiment"

    ```shell
    kubectl get experiment canary-fixedsplit --watch
    ```

    You should see output similar to the following.
    ```shell
    NAME               TYPE     TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
    canary-fixesplit   Canary   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
    canary-fixesplit   Canary   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
    canary-fixesplit   Canary   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
    canary-fixesplit   Canary   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
    canary-fixesplit   Canary   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
    canary-fixesplit   Canary   default/sample-app   Running   6                      IterationUpdate: Completed Iteration 6
    canary-fixesplit   Canary   default/sample-app   Running   7                      IterationUpdate: Completed Iteration 7
    canary-fixesplit   Canary   default/sample-app   Running   8                      IterationUpdate: Completed Iteration 8
    canary-fixesplit   Canary   default/sample-app   Running   9                      IterationUpdate: Completed Iteration 9
    canary-fixesplit   Canary   default/sample-app   Running   10                     IterationUpdate: Completed Iteration 10
    canary-fixesplit   Canary   default/sample-app   Running   11                     IterationUpdate: Completed Iteration 11
    canary-fixesplit   Canary   default/sample-app   Finishing   12                     TerminalHandlerLaunched: Finish handler 'finish' launched
    canary-fixesplit   Canary   default/sample-app   Completed   12                     ExperimentCompleted: Experiment completed successfully
    ```

=== "kubectl get ksvc"

    ```shell
    kubectl get ksvc sample-app -o json --watch | jq .status.traffic
    ```

    You should see output similar to the following. The traffic percentage should remain the same during the experiment.
    ```shell
    [
      {
        "latestRevision": false,
        "percent": 75,
        "revisionName": "sample-app-v1",
        "tag": "current",
        "url": "http://current-sample-app.default.example.com"
      },
      {
        "latestRevision": true,
        "percent": 25,
        "revisionName": "sample-app-v2",
        "tag": "candidate",
        "url": "http://candidate-sample-app.default.example.com"
      }
    ]
    ```

When the experiment completes (in ~ 4 mins), you will see the experiment stage change from `Running` to `Completed`.

## 5. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/canaryfixedsplit/fortio.yaml
kubectl delete -f $ITER8/samples/knative/canaryfixedsplit/experiment.yaml
kustomize build $ITER8/samples/knative/canaryfixedsplit/experimentalservice | kubectl delete -f -
```

??? info "Understanding what happened"
    1. In Step 1, you created a Knative service which manages two revisions, `sample-app-v1` (`baseline`) and `sample-app-v2` (`candidate`).
    2. In Step 2, you created a load generator that sends requests to the Knative service. 75% of requests are sent to the baseline and 25% to the candidate. This distribution remains fixed throughout the experiment.
    3. In step 3, you created an iter8 experiment with 12 iterations with the above Knative service as the `target` of the experiment. In each iteration, iter8 observed the `mean-latency`, `95th-percentile-tail-latency`, and `error-rate` metrics for the revisions (collected by Prometheus).
    4. At the end of the experiment, iter8 identified the candidate as the `winner` since it passed all objectives. iter8 decided to promote the candidate (rollforward) using kustomize as part of its `finish` action. Had the candidate failed, iter8 would have decided to promote the baseline (rollback) instead.
