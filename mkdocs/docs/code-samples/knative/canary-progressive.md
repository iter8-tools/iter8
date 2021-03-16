---
template: overrides/main.html
---

# Canary + Progressive + Helm Tutorial

!!! tip ""
    An experiment with [`Canary`](/concepts/experimentationstrategies/#testing-pattern) testing, [`Progressive`](/concepts/experimentationstrategies/#deployment-pattern) deployment, and [`Helm` based version promotion](/concepts/experimentationstrategies/#version-promotion).
    
    ![Canary](/assets/images/canary-progressive-helm.png)

You will create the following resources in this tutorial.

1. A **Knative app (service)** with two versions (revisions).
2. A **fortio-based traffic generator** that simulates user requests.
3. An **Iter8 experiment** that: 
    - verifies that `candidate` satisfies mean latency, 95th percentile tail latency, and error rate `objectives`
    - progressively shifts traffic from `baseline` to `candidate`, subject to the limits placed by the experiment's `spec.strategy.weights` field
    - eventually replaces `baseline` with `candidate` using `Helm`

??? warning "Before you begin, you will need ... "
    **Kubernetes cluster:** Ensure that you have Kubernetes cluster with Iter8 and Knative installed. You can do this by following Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/).

    **Cleanup:** If you ran an Iter8 tutorial earlier, run the associated cleanup step.

    **ITER8:** Ensure that `ITER8` environment variable is set to the root directory of your cloned Iter8 repo. See [Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo) for example.

    **[Helm v3](https://helm.sh/) and [iter8ctl](/getting-started/install/#step-4-install-iter8ctl):** This tutorial uses Helm v3 and iter8ctl.

## 1. Create versions
```shell
helm install --repo https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo sample-app sample-app --namespace=iter8-system
kubectl wait ksvc/sample-app --for condition=Ready --timeout=120s
helm upgrade --install --repo https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo sample-app sample-app --values=https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/experimental-values.yaml --namespace=iter8-system
```

??? info "Look inside values.yaml"
    ```yaml linenums="1"
    # default values used for installing sample-app Helm chart
    # using these values will create a baseline version (revision) that gets 100% of the traffic
    name: "sample-app-v1"
    image: "gcr.io/knative-samples/knative-route-demo:blue"
    tVersion: "blue"
    ```

??? info "Look inside experimental-values.yaml"
    ```yaml linenums="1"
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
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/canaryprogressive/fortio.yaml | kubectl apply -f -
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
            command: ["fortio", "load", "-t", "120s", "-json", "/shared/fortiooutput.json", $(URL)]
            env:
            - name: URL
              value: URL_VALUE
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 600']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never    
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
        # information about versions used in this experiment
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
Observe the experiment in realtime. Paste commands from the tabs below in separate terminals.

=== "iter8ctl"
    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment canary-progressive -o yaml | iter8ctl describe -f -
    sleep 4
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
    1. You created a Knative service using `helm install` subcommand and upgraded the service to have two revisions, sample-app-v1 (`baseline`) and sample-app-v2 (`candidate`) using `helm upgrade --install` subcommand. 
    2. The ksvc is created in the `default` namespace. Helm release information is located in the `iter8-system` namespace as specified by the `--namespace=iter8-system` flag.
    3. You generated requests for the Knative service using a fortio-job. At the start of the experiment, 100% of the requests are sent to baseline and 0% to candidate.
    4. You created an Iter8 `Canary` experiment with `Progressive` deployment pattern. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, verified that `candidate` satisfied all the objectives specified in the experiment, identified `candidate` as the `winner`, progressively shifted traffic from `baseline` to `candidate` and eventually promoted the `candidate` using `helm upgrade --install` subcommand.
        - **Note:** Had `candidate` failed to satisfy `objectives`, then `baseline` would have been promoted.
        - **Note:** You limited the maximum weight (traffic %) of `candidate` during iterations at 75% and maximum increment allowed during a single iteration to 20% using the field `spec.strategy.weights`. Traffic shifts during the experiment obeyed these limits.
