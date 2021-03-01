---
template: overrides/main.html
---

# Progressive Canary Deployment using Helm

Perform **zero-downtime progressive canary release of a Knative app**. This tutorial is similar to the [Iter8 quick start tutorial for Knative](/getting-started/quick-start/with-knative/). You will create:

1. A Knative service with two versions of your app, namely, `baseline` and `candidate`
2. A traffic generator which sends HTTP GET requests to the Knative service.
3. An **Iter8 experiment** that automates the following: 
    - verifies that latency and error-rate metrics for the `candidate` satisfy the given objectives
    - iteratively shifts traffic from `baseline` to `candidate`,
    - fine-tunes traffic shifting behavior during iterations using `spec.strategy.weights`, and
    - replaces `baseline` with `candidate` in the end using a `helm install` command

??? warning "Before you begin"
    **Kubernetes cluster:** Ensure that you have Kubernetes cluster with Iter8 and Knative installed. You can do this by following Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/).

    **Cleanup:** If you ran an Iter8 tutorial earlier, run the associated cleanup step. For example, [Step 8](/getting-started/quick-start/with-knative/#8-cleanup) is the cleanup step for the Iter8-Knative quick start tutorial.

    **ITER8:** Ensure that `ITER8` environment variable is set to the root directory of your cloned Iter8 repo. See [Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo) for example.

    **[Helm v3](https://helm.sh/) and [iter8ctl](/getting-started/install/#step-4-install-iter8ctl):** This tutorial uses Helm v3 and iter8ctl. Install if needed.

## 1. Create Knative app with canary
```shell
helm install --repo https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo sample-app sample-app --namespace=iter8-system
kubectl wait ksvc/sample-app --for condition=Ready --timeout=120s
helm upgrade --install --repo https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo sample-app sample-app --values=https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/experimental-values.yaml --namespace=iter8-system
```

??? info "Look inside values.yaml"
    ```yaml
    # default values used for installing sample-app Helm chart
    # using these values will create a baseline version (revision) that gets 100% of the traffic
    name: "sample-app-v1"
    image: "gcr.io/knative-samples/knative-route-demo:blue"
    tVersion: "blue"
    ```

??? info "Look inside experimental-values.yaml"
    ```yaml
    # values file used for upgrading sample-app Helm chart for use in Iter8 experiment
    # using these values will create a candidate version (revision)
    # baseline still gets 100% of the traffic
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

## 2. Generate requests
Verify Knative service is ready and generate requests to app using fortio.
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/canaryprogressive/fortio.yaml | kubectl apply -f -
```

## 3. Create Iter8 experiment
```shell
kubectl apply -f $ITER8/samples/knative/canaryprogressive/experiment.yaml
```
??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: canary-progressive
    spec:
      # target identifies the knative service under experimentation using its fully qualified name
      target: default/sample-app
      strategy:
        # this experiment will perform a canary test
        testingPattern: Canary
        deploymentPattern: Progressive
        weights: # fine-tune traffic increments to candidate
          # candidate weight will not exceed 75 in any iteration
          maxCandidateWeight: 75
          # candidate weight will not increase by more than 20 in a single iteration
          maxCandidateWeightIncrement: 20
        actions:
          start: # run the following sequence of tasks at the start of the experiment
          - library: knative
            task: init-experiment
          finish: # run the following sequence of tasks at the end of the experiment
          - library: common
            task: exec # promote the winning version using Helm upgrade
            with:
              cmd: helm
              args:
              - "upgrade"
              - "--install"
              - "--repo"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo" # repo url
              - "sample-app" # release name
              - "--namespace=iter8-system" # release namespace
              - "sample-app" # chart name
              - "--values=https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/{{ .promote }}-values.yaml" # values URL dynamically interpolated
              # - "--reset-values" # seems necessary to avoid ownership annotation metadata errors
              # - "--force" # seems necessary to avoid ownership annotation metadata errors
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
        iterationsPerLoop: 7
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
    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment canary-progressive -o yaml | iter8ctl describe -f -
    sleep 10
    done
    ```

    ??? info "iter8ctl output"
        iter8ctl output will be similar to the following.
        ```shell
        ****** Overview ******
        Experiment name: canary-progressive
        Experiment namespace: default
        Target: default/sample-app
        Testing pattern: Canary
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Completed
        Number of completed iterations: 7

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
        | mean-latency (milliseconds)    |   1.201 |     1.322 |
        +--------------------------------+---------+-----------+
        | 95th-percentile-tail-latency   |   4.776 |     4.750 |
        | (milliseconds)                 |         |           |
        +--------------------------------+---------+-----------+
        | error-rate                     |   0.000 |     0.000 |
        +--------------------------------+---------+-----------+
        | request-count                  | 448.800 |    89.352 |
        +--------------------------------+---------+-----------+
        ```
        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.   

=== "kubectl get experiment"
    ```shell
    kubectl get experiment canary-progressive --watch
    ```

    ??? info "kubectl get experiment output"
        kubectl output will be similar to the following.
        ```shell
        NAME                 TYPE     TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
        canary-progressive   Canary   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
        canary-progressive   Canary   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
        canary-progressive   Canary   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
        canary-progressive   Canary   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
        canary-progressive   Canary   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
        canary-progressive   Canary   default/sample-app   Running   6                      IterationUpdate: Completed Iteration 6
        canary-progressive   Canary   default/sample-app   Finishing   7                      TerminalHandlerLaunched: Finish handler 'finish' launched
        canary-progressive   Canary   default/sample-app   Completed   7                      ExperimentCompleted: Experiment completed successfully
        ```
        When the experiment completes (in ~ 4 mins), you will see the experiment stage change from `Running` to `Completed`.    

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
            "percent": 25,
            "revisionName": "sample-app-v1",
            "tag": "current",
            "url": "http://current-sample-app.default.example.com"
          },
          {
            "latestRevision": true,
            "percent": 75,
            "revisionName": "sample-app-v2",
            "tag": "candidate",
            "url": "http://candidate-sample-app.default.example.com"
          }
        ]
        ```
        
## 5. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/canaryprogressive/experiment.yaml
kubectl delete -f $ITER8/samples/knative/canaryprogressive/fortio.yaml
helm uninstall sample-app --namespace=iter8-system
```

??? info "Understanding what happened"
    1. You created a Knative service using `helm install` subcommand and upgraded the service to have both `baseline` and `candidate` versions (revisions) using `helm upgrade --install` subcommand. The ksvc is created in the `default` namespace. Helm release information is located in the `iter8-system` namespace specified by the `--namespace=iter8-system` flag.
    2. You created a load generator that sends requests to the Knative service. At this point, 100% of requests are sent to the baseline and 0% to the candidate.
    3. You created an Iter8 experiment with the above Knative service as the `target` of the experiment. In each iteration, Iter8 observed the `mean-latency`, `95th-percentile-tail-latency`, and `error-rate` metrics for the revisions (collected by Prometheus), ensured that the candidate satisfied all objectives specified in `experiment.yaml`, and progressively shifted traffic from baseline to candidate. You restricted the maximum weight (traffic percentage) of candidate during iterations at 75% and maximum increment allowed during a single iteration to 20% using the `spec.strategy.weights` stanza.
    4. At the end of the experiment, Iter8 identified the candidate as the `winner` since it passed all objectives. Iter8 decided to promote the candidate (roll forward) by using a `helm upgrade --install` command. Had the candidate failed to satisfy objectives, Iter8 would have promoted the baseline (rolled back) instead.
