---
template: overrides/main.html
---

# Fixed Split Canary Release

!!! tip ""
    An experiment with [`Canary`](/concepts/experimentationstrategies/#testing-pattern) testing, [`FixedSplit`](/concepts/experimentationstrategies/#deployment-pattern) deployment, and [Kustomize based version promotion](/concepts/experimentationstrategies/#version-promotion).
    
    ![Canary](/assets/images/canary-fixedsplit-kustomize.png)

You will create the following resources in this tutorial.

1. A **Knative app** (service) with two versions (revisions).
2. A **Fortio-based traffic generator** that simulates user requests.
3. An **Iter8 experiment** that: 
    - verifies that `candidate` satisfies mean latency, 95th percentile tail latency, and error rate `objectives`
    - maintains a 75-25 split of traffic between `baseline` and `candidate` throughout the experiment
    - eventually replaces `baseline` with `candidate` using Kustomize

???+ warning "Before you begin, you will need... "
    **Kubernetes cluster:** Ensure that you have Kubernetes cluster with Iter8 and Knative installed. You can do this by following Steps 1, 2, and 3 of the [quick start tutorial for Knative](/getting-started/quick-start/with-knative/).

    **Cleanup:** If you ran an Iter8 tutorial earlier, run the associated cleanup step.

    **ITER8:** Ensure that `ITER8` environment variable is set to the root directory of your cloned Iter8 repo. See [Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo) for example.

    **[Kustomize v3+](https://kustomize.io/) and [`iter8ctl`](/getting-started/install/#step-4-install-iter8ctl):** This tutorial uses Kustomize v3+ and `iter8ctl`.

## 1. Create app versions

```shell
kustomize build $ITER8/samples/knative/canaryfixedsplit/baseline | kubectl apply -f -
kubectl wait ksvc/sample-app --for condition=Ready --timeout=120s
kustomize build $ITER8/samples/knative/canaryfixedsplit/experimentalservice | kubectl apply -f -
```

??? info "Look inside output of `kustomize build $ITER8/.../baseline`"
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
    ```

??? info "Look inside output of `kustomize build $ITER8/.../experimentalservice`"
    ```yaml linenums="1"
    # This Knative service will be used for the Iter8 experiment with traffic split between baseline and candidate revision
    # Traffic is split 75/25 between the baseline and candidate
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

## 2. Generate requests
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/canaryfixedsplit/fortio.yaml | kubectl apply -f -
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
kubectl apply -f $ITER8/samples/knative/canaryfixedsplit/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
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
          - task: knative/init-experiment
          finish: # run the following sequence of tasks at the end of the experiment
          - task: common/exec # promote the winning version using kustomize
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
              kustomize build github.com/iter8-tools/iter8/samples/knative/canaryfixedsplit/{{ .name }}?ref=master \
                | kubectl apply -f -
      criteria:
        # mean latency of version should be under 50 milliseconds
        # 95th percentile latency should be under 100 milliseconds
        # error rate should be under 1%
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
        # information about versions used in this experiment
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
Observe the experiment in realtime. Paste commands from the tabs below in separate terminals.

=== "iter8ctl"
    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment canary-fixedsplit -o yaml | iter8ctl describe -f -
    sleep 4
    done
    ```

    ??? info "iter8ctl output"
        The `iter8ctl` output will be similar to the following:
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
        Version recommended for promotion: candidate

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
        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.   

=== "kubectl get experiment"

    ```shell
    kubectl get experiment canary-fixedsplit --watch
    ```

    ??? info "kubectl get experiment output"
        The `kubectl` output will be similar to the following.
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
        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.   

=== "kubectl get ksvc"

    ```shell
    kubectl get ksvc sample-app -o json --watch | jq .status.traffic
    ```

    ??? info "kubectl get ksvc output"
        The `kubectl` output will be similar to the following. The traffic percentage should remain the same during the experiment.
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

## 5. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/canaryfixedsplit/fortio.yaml
kubectl delete -f $ITER8/samples/knative/canaryfixedsplit/experiment.yaml
kustomize build $ITER8/samples/knative/canaryfixedsplit/experimentalservice | kubectl delete -f -
```

???+ info "Understanding what happened"
    1. You created a Knative service with two revisions, sample-app-v1 (`baseline`) and sample-app-v2 (`candidate`) using Kustomize.
    2. You generated requests for the Knative service using a Fortio job. At the start of the experiment, 75% of the requests are sent to `baseline` and 25% to `candidate`.
    4. You created an Iter8 `Canary` experiment with `FixedSplit` deployment pattern. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, verified that `candidate` satisfied all the objectives specified in the experiment, identified `candidate` as the `winner`, and eventually promoted the `candidate` using `kustomize build ... | kubectl apply -f -` commands.
        - **Note:** Had `candidate` failed to satisfy `objectives`, then `baseline` would have been promoted.
        - **Note:** There was no traffic shifting during experiment iterations since this used a `FixedSplit` deployment pattern.