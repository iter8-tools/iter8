---
template: main.html
---

# Progressive Traffic Shift

!!! tip "Scenario: Progressive traffic shift"
    [Progressive traffic shift](../../../concepts/buildingblocks.md#progressive-traffic-shift) is a type of canary rollout strategy. It enables you to incrementally shift traffic towards the winning version over multiple iterations of an experiment as shown below.

    ![Progressive traffic shift](../../../images/progressive.png)

## Tutorials with progressive traffic shift

The [A/B testing (quick start)](../quick-start.md) and [hybrid (A/B + SLOs) testing](../testing-strategies/hybrid.md) tutorials demonstrate progressive traffic shift.

## Specifying `weightObjRef`

Iter8 uses the `weightObjRef` field in the experiment resource to get the current traffic split between versions and/or modify the traffic split. Ensure that this field is specified correctly for each version. The following example demonstrates how to specify `weightObjRef` in experiments.

??? example "Example"
    The [A/B testing quick start tutorial](../quick-start.md#4-launch-experiment) uses an Istio virtual service for traffic shifting. Hence, the experiment manifest specifies the `weightObjRef` field for each version by referencing this Istio virtual service and the traffic fields within the Istio virtual service corresponding to the versions.

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