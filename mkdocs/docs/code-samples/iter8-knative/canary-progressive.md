---
template: overrides/main.html
---

# Progressive Canary Deployment using Helm

Perform **zero-downtime progressive canary release of a Knative app**. This tutorial is similar to the [iter8 quick start tutorial for Knative](/getting-started/quick-start/with-knative/). You will create:

1. A Knative service with two versions of your app, namely, `baseline` and `candidate`
2. A traffic generator which sends HTTP GET requests to the Knative service.
3. An **iter8 experiment** that automates the following: 
    - verifies that latency and error-rate metrics for the `candidate` satisfy the given objectives
    - iteratively shifts traffic from `baseline` to `candidate`, and 
    - replaces `baseline` with `candidate` in the end using a `helm install` command

!!! warning "Before you begin"
    **Kubernetes cluster:** Do not have a Kubernetes cluster with iter8 and Knative installed? Follow Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/) to create a cluster with iter8 and Knative.

    **Cleanup from previous experiment:** Tried an iter8 tutorial earlier but forgot to cleanup? Run the cleanup step from your tutorial now. For example, [Step 8](/getting-started/quick-start/with-knative/#8-cleanup) performs cleanup for the iter8-Knative quick start tutorial.

    **ITER8 environment variable:** ITER8 environment variable is not exported in your terminal? Do so now. For example, this is the [last command in Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo).

    **Helm v3:** [Helm v3](https://helm.sh/) is not installed locally? Do so now. This sample uses Helm.

## 1. Create Knative app with canary
```shell
helm upgrade --install sample-app $ITER8/samples/knative/canaryprogressive/sample-app --values=$ITER8/samples/knative/canaryprogressive/sample-app/values.yaml 
kubectl wait ksvc/sample-app --for condition=Ready --timeout=120s
helm upgrade --install sample-app $ITER8/samples/knative/canaryprogressive/sample-app --values=$ITER8/samples/knative/canaryprogressive/sample-app/experimental-values.yaml 
```

??? info "Look inside values.yaml"
    ```yaml
    # values file used for installing sample-app Helm chart with baseline version
    name: "sample-app-v1"
    image: "gcr.io/knative-samples/knative-route-demo:blue"
    tVersion: "blue"
    ```

??? info "Look inside experimental-values.yaml"
    ```yaml
    # values file used for upgrading sample-app Helm chart for use in iter8 experiment
    name: "sample-app-v2"
    image: "gcr.io/knative-samples/knative-route-demo:green"
    tVersion: "green"
    traffic:
    - tag: current
      revisionName: sample-app-v1
      percent: 100
    - tag: candidate
      latestRevision: true
      percent: 0
    ```

## 2. Send requests
Verify Knative service is ready and send requests to app.
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/canaryprogressive/fortio.yaml | kubectl apply -f -
```

## 3. Create iter8 experiment
```shell
kubectl apply -f $ITER8/samples/knative/canaryprogressive/experiment.yaml
```
??? info "Look inside experiment.yaml"
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
          finish: # run the following sequence of tasks at the end of the experiment
          - library: common
            task: exec # promote the winning version
            with:
              cmd: kubectl
              args: ["apply", "-f", "https://github.com/iter8-tools/iter8/samples/knative/quickstart/{{ .promote }}.yaml"]
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
        - name: promote
          value: baseline
      candidates:
      - name: candidate
        variables:
        - name: revision
          value: sample-app-v2
        - name: promote
          value: candidate 
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

=== "kubectl get experiment"

    ```shell
    kubectl get experiment quickstart-exp --watch
    ```

    You should see output similar to the following.
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
    quickstart-exp   Canary   default/sample-app   Running   10                     IterationUpdate: Completed Iteration 10
    quickstart-exp   Canary   default/sample-app   Running   11                     IterationUpdate: Completed Iteration 11
    quickstart-exp   Canary   default/sample-app   Finishing   12                     TerminalHandlerLaunched: Finish handler 'finish' launched
    quickstart-exp   Canary   default/sample-app   Completed   12                     ExperimentCompleted: Experiment completed successfully
    ```

    

=== "kubectl get ksvc"

    ```shell
    kubectl get ksvc sample-app -o json --watch | jq .status.traffic
    ```

    You should see output similar to the following.
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

When the experiment completes (in ~ 4 mins), you will see the experiment stage change from `Running` to `Completed`.

## 5. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```

??? info "Understanding what happened"
    1. In Step 4, you created a Knative service which manages two revisions, `sample-app-v1` (`baseline`) and `sample-app-v2` (`candidate`).
    2. In Step 5, you created a load generator that sends requests to the Knative service. At this point, 100% of requests are sent to the baseline and 0% to the candidate.
    3. In step 6, you created an iter8 experiment with 12 iterations with the above Knative service as the `target` of the experiment. In each iteration, iter8 observed the `mean-latency`, `95th-percentile-tail-latency`, and `error-rate` metrics for the revisions (collected by Prometheus), ensured that the candidate satisfied all objectives specified in `experiment.yaml`, and progressively shifted traffic from baseline to candidate.
    4. At the end of the experiment, iter8 identified the candidate as the `winner` since it passed all objectives. iter8 decided to promote the candidate (rollforward) by applying `candidate.yaml` as part of its `finish` action. Had the candidate failed, iter8 would have decided to promote the baseline (rollback) by applying `baseline.yaml`.
