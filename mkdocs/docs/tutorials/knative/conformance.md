---
template: overrides/main.html
---

# Conformance Testing

!!! tip ""
    An experiment with [`Conformance`](/concepts/buildingblocks/#testing-pattern) testing.
    
    ![Canary](/assets/images/conformance.png)

You will create the following resources in this tutorial.

1. A **Knative app** (service) with a single version (revision).
2. A **Fortio-based traffic generator** that simulates user requests.
3. An **Iter8 experiment** that verifies that `baseline` satisfies mean latency, 95th percentile tail latency, and error rate `objectives`.

???+ warning "Before you begin, you will need... "
    **Kubernetes cluster:** Ensure that you have Kubernetes cluster with Iter8 and Knative installed. You can do this by following Steps 1, 2, and 3 of the [quick start tutorial for Knative](/getting-started/quick-start/with-knative/).

    **Cleanup:** If you ran an Iter8 tutorial earlier, run the associated cleanup step.

    **ITER8:** Ensure that `ITER8` environment variable is set to the root directory of your cloned Iter8 repo. See [Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-iter8-repo) for example.

    **[`iter8ctl`](/getting-started/install/#optional-step-3-iter8ctl):** This tutorial uses `iter8ctl`.

## 1. Create app
```shell
kubectl apply -f $ITER8/samples/knative/conformance/baseline.yaml
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

## 2. Generate requests
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/conformance/fortio.yaml | kubectl apply -f -
```

??? info "Look inside experiment.yaml"
    ```yaml
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
kubectl apply -f $ITER8/samples/knative/conformance/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml
    apiVersion: iter8.tools/v2alpha2
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
          - task: knative/init-experiment
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
          name: current
          variables:
          - name: revision
            value: sample-app-v1  
    ```

## 4. Observe experiment

Observe the experiment in realtime. Paste commands from the tabs below in separate terminals.

=== "iter8ctl"
    Periodically describe the experiment.
        ```shell
        while clear; do
        kubectl get experiment conformance-sample -o yaml | iter8ctl describe -f -
        sleep 4
        done
        ```

    The output will look similar to the [iter8ctl output](/getting-started/quick-start/with-knative/#7-observe-experiment) in the quick start instructions.

    As the experiment progresses, you should eventually see that all of the objectives reported as being satisfied by the version being tested. When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.

=== "kubectl get experiment"

    ```shell
    kubectl get experiment conformance-sample --watch
    ```

    The output will look similar to the [kubectl get experiment output](/getting-started/quick-start/with-knative/#7-observe-experiment) in the quick start instructions.

    When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.

## 5. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/conformance/fortio.yaml
kubectl delete -f $ITER8/samples/knative/conformance/experiment.yaml
kubectl delete -f $ITER8/samples/knative/conformance/baseline.yaml
```

???+ info "Understanding what happened"
    1. You created a Knative service with a single revision, sample-app-v1. 
    2. You generated requests for the Knative service using a Fortio job.
    3. You created an Iter8 `Conformance` experiment. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, and verified that `baseline` satisfied all the `objectives` specified in the experiment.