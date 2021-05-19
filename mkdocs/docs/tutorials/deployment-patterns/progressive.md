---
template: main.html
---

# Progressive Deployment

!!! tip "Scenario: Progressive deployment"
    [Progressive deployment](../../../concepts/buildingblocks/#deployment-pattern) enables you to incrementally shift traffic towards the winning version over multiple iterations of an experiment. 
    
    Progressive deployment is the default deployment pattern for Iter8 experiments. Progressive deployment during an A/B testing experiment is depicted below.

    ![Canary](../../images/quickstart.png)

## Tutorials with progressive deployment

The [A/B testing (quick start)](../../../getting-started/quick-start/) and [canary testing](../../testing-patterns/canary/) tutorials demonstrate progressive deployment.

## Specifying `weightObjRef`

Iter8 uses the `weightObjRef` field in the experiment resource to get the current traffic split between versions and/or modify the traffic split. Ensure that this field is specified correctly for each version. Below are a few examples that demonstrate how to specify `weightObjRef` in experiments.

=== "Istio"
    The [A/B testing experiment for Istio app](../../../getting-started/quick-start/#7-launch-experiment) uses an Istio virtual service for traffic shifting. Hence, the experiment manifest specifies the `weightObjRef` field for each version by referencing this Istio virtual service and the traffic fields within the Istio virtual service corresponding to the versions.

    ```yaml
    versionInfo:
      baseline:
        name: productpage-v1
        weightObjRef:
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          namespace: bookinfo-iter8
          name: bookinfo
          fieldPath: .spec.http[0].route[0].weight
      candidates:
      - name: productpage-v2
        weightObjRef:
          apiVersion: networking.istio.io/v1beta1
          kind: VirtualService
          namespace: bookinfo-iter8
          name: bookinfo
          fieldPath: .spec.http[0].route[1].weight
    ```

=== "KFServing"
    The [A/B testing experiment for KFServing model](../../../getting-started/quick-start/#7-launch-experiment) uses an Istio virtual service for traffic shifting. Hence, the experiment manifest specifies the `weightObjRef` field for each version by referencing this Istio virtual service and the traffic fields within the Istio virtual service corresponding to the versions.

    ```yaml
    versionInfo:
      baseline:
        name: flowers-v1
        weightObjRef:
          apiVersion: networking.istio.io/v1alpha3
          kind: VirtualService
          name: routing-rule
          namespace: default
          fieldPath: .spec.http[0].route[0].weight      
      candidates:
      - name: flowers-v2
        weightObjRef:
          apiVersion: networking.istio.io/v1alpha3
          kind: VirtualService
          name: routing-rule
          namespace: default
          fieldPath: .spec.http[0].route[1].weight 
    ```

=== "Knative"
    The [A/B testing experiment for Knative app](../../../getting-started/quick-start/#7-launch-experiment) uses a Knative service for traffic shifting. Hence, the experiment manifest specifies the `weightObjRef` field for each version by referencing this Knative service and the traffic fields within the Knative service corresponding to the versions.

    ```yaml
    versionInfo:
      baseline:
        name: sample-app-v1
        weightObjRef:
          apiVersion: serving.knative.dev/v1
          kind: Service
          name: sample-app
          namespace: default
          fieldPath: .spec.traffic[0].percent
      candidates:
      - name: sample-app-v2
        weightObjRef:
          apiVersion: serving.knative.dev/v1
          kind: Service
          name: sample-app
          namespace: default
          fieldPath: .spec.traffic[1].percent
    ```

## Traffic controls

You can specify the maximum traffic percentage that is allowed for a candidate version during the experiment. You can also specify the maximum increase in traffic percentage that is allowed for a candidate version during a single iteration of the experiment. You can specify these two controls in the `strategy` section of an experiment as follows.

```yaml
strategy:
  weights: # additional traffic controls to be used during an experiment
    # candidate weight will not exceed 75 in any iteration
    maxCandidateWeight: 75
    # candidate weight will not increase by more than 20 in a single iteration
    maxCandidateWeightIncrement: 20
```