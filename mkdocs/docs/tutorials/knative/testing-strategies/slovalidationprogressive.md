---
template: main.html
---

# SLO Validation

!!! tip "Scenario: SLO validation with progressive traffic shift"
    This tutorial illustrates an [SLO validation experiment with two versions](../../../concepts/buildingblocks.md#slo-validation); the candidate version will be promoted after Iter8 validates that it satisfies service-level objectives (SLOs). You will:

    1. Specify *latency* and *error-rate* based service-level objectives (SLOs). If the candidate version satisfies SLOs, Iter8 will declare it as the winner.
    2. Use Prometheus as the provider for latency and error-rate metrics.
    3. Combine SLO validation with [progressive traffic shifting](../../../concepts/buildingblocks.md#progressive-traffic-shift).
    
    ![SLO validation with progressive traffic shift](../../../images/slovalidationprogressive.png)

## Steps 1 to 4
    
Please follow steps 1 through 4 of the [Knative quick start tutorial](../../../getting-started/quick-start/knative/tutorial.md#1-setup).


## 5. Launch experiment
Launch the SLO validation experiment.
```shell
kubectl apply -f $ITER8/samples/knative/slovalidation/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: canary-exp
    spec:
      target: default/sample-app
      strategy:
        testingPattern: Canary
        deploymentPattern: Progressive
        actions:
          finish: # run the following sequence of tasks at the end of the experiment
          - task: common/exec # promote the winning version      
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
                kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml
      criteria:
        requestCount: iter8-knative/request-count
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
          weightObjRef:
            apiVersion: serving.knative.dev/v1
            kind: Service
            name: sample-app
            namespace: default
            fieldPath: .spec.traffic[0].percent
          variables:
          - name: promote
            value: baseline
        candidates:
        - name: sample-app-v2
          weightObjRef:
            apiVersion: serving.knative.dev/v1
            kind: Service
            name: sample-app
            namespace: default
            fieldPath: .spec.traffic[1].percent
          variables:
          - name: promote
            value: candidate
    ```

## 6. Understand the experiment
Follow [Step 6 of the quick start tutorial for KFServing](../../../../getting-started/quick-start/kfserving/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`slovalidation-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/slovalidation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```
