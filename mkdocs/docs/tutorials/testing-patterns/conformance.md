---
template: main.html
---

# Conformance Testing

!!! tip "Scenario: Conformance testing"
    [Conformance testing](../../../concepts/buildingblocks/#testing-pattern) enables you to validate a  version of your app/ML model using service-level objectives (SLOs). In this tutorial, you will:

    1. Perform conformance testing.
    2. Specify *latency* and *error-rate* based service-level objectives (SLOs). If your version satisfies SLOs, Iter8 will declare it as the winner.
    3. Use Prometheus as the provider for latency and error-rate metrics.
    
    ![Conformance](../../images/conformance.png)

???+ warning "Before you begin, you will need... "
    > **Note:** Please choose the same K8s stack (for example, Istio, KFServing, or Knative) consistently throughout this tutorial. If you wish to switch K8s stacks between tutorials, start from a clean K8s cluster, so that your cluster is correctly setup.

## Steps 1, 2, and 3
    
Please follow steps 1, 2, and 3 of the [quick start tutorial](../../../getting-started/quick-start/#1-create-kubernetes-cluster).


## 4. Create app/ML model version
=== "Knative"
    Deploy a Knative app.

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

## 5. Generate requests
=== "Knative"
    Generate requests using Fortio as follows.

    ```shell
    kubectl wait --for=condition=Ready ksvc/sample-app
    URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
    sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/conformance/fortio.yaml | kubectl apply -f -
    ```

    ??? info "Look inside fortio.yaml"
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
                command: ["fortio", "load", "-t", "6000s", "-json", "/shared/fortiooutput.json", $(URL)]
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

## 6. Define metrics
Please follow step 6 of the [quick start tutorial](../../../getting-started/quick-start/#1-define-metrics).

## 7. Launch experiment
Launch the Iter8 experiment that orchestrates conformance testing for the app/ML model in this tutorial.

=== "Knative"
    ```shell
    kubectl apply -f $ITER8/samples/knative/conformance/experiment.yaml
    ```

    ??? info "Look inside experiment.yaml"
        ```yaml
        apiVersion: iter8.tools/v2alpha2
        kind: Experiment
        metadata:
          name: conformance-exp
        spec:
          # target identifies the knative service under experimentation using its fully qualified name
          target: default/sample-app
          strategy:
            # this experiment will perform a conformance test
            testingPattern: Conformance
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
            # information about app versions used in this experiment
            baseline:
              name: sample-app-v1
        ```

The process automated by Iter8 during this experiment is depicted below.

![Iter8 automation](../../images/conformance-iter8-process.png)

## 8. Observe experiment

Follow [step 8 of quick start tutorial](../../../getting-started/quick-start/#8-observe-experiment) to observe the experiment in realtime. Note that the experiment in this tutorial uses a different name from the quick start one. Replace the experiment name `quickstart-exp` with `conformance-exp` in your commands.


???+ info "Understanding what happened"
    1. You created a single version of an app/ML model.
    2. You generated requests for your app/ML model versions.
    3. You created an Iter8 experiment with conformance testing pattern. In each iteration, Iter8 observed the latency and error-rate metrics collected by Prometheus; Iter8 verified that the version (referred to as baseline in a conformance experiment) satisfied all the SLOs, and identified baseline as the winner.

## 9. Cleanup

=== "Knative"
    ```shell
    kubectl delete -f $ITER8/samples/knative/conformance/fortio.yaml
    kubectl delete -f $ITER8/samples/knative/conformance/experiment.yaml
    kubectl delete -f $ITER8/samples/knative/conformance/baseline.yaml
    ```
