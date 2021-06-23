---
template: main.html
---

# SLO Validation with a single version

!!! tip "Scenario: SLO validation with builtin metrics"
    Iter8 enables you to perform SLO validation with a single version of your application (a.k.a. [Conformance testing](../../../../concepts/buildingblocks/#slo-validation)). In this tutorial, you will:

    1. Perform conformance testing.
    2. Specify *latency* and *error-rate* based service-level objectives (SLOs). If your version satisfies SLOs, Iter8 will declare it as the winner.
    3. Use Iter8's its builtin load generation and metrics collection capability.
    
    ![Conformance](../../../images/conformance.png)

## 1. Setup
* Setup your K8s cluster with Knative and Iter8 as described [here](../../../../getting-started/quick-start/knative/platform-setup/).
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

## 2. Create application version
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

## 3. Generate requests
Generation of requests is handled automatically by the Iter8 experiment.

## 4. Define metrics
Metrics collection is handled automatically by the Iter8 experiment.

## 5. Launch experiment
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
        actions:
          loop:
          - task: metrics/collect
            with:
              versions: 
              - name: sample-app-v1
                url: http://sample-app.default.svc.cluster.local
      criteria:
        objectives: 
        - metric: iter8-system/mean-latency
          upperLimit: 50
        - metric: iter8-system/error-count
          upperLimit: 0
      duration:
        maxLoops: 10
        intervalSeconds: 1
        iterationsPerLoop: 1
      versionInfo:
        # information about app versions used in this experiment
        baseline:
          name: sample-app-v1
    ```

## 6. Understand the experiment
Follow [Step 6 of the quick start tutorial for Knative](../../../../getting-started/quick-start/knative/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`conformance-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/conformance/experiment.yaml
kubectl delete -f $ITER8/samples/knative/conformance/baseline.yaml
```
