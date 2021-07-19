---
template: main.html
---

# Progressive Traffic Shift

!!! tip "Scenario: Progressive traffic shift"
    [Progressive traffic shift](../../../concepts/buildingblocks.md#progressive-traffic-shift) is a type of canary rollout strategy. It enables you to incrementally shift traffic towards the winning version over multiple iterations of an experiment as shown below.

    ![Progressive traffic shift](../../../images/progressive.png)

## Tutorials with progressive traffic shift

The [SLO validation](../testing-strategies/slovalidation.md) and [hybrid testing](../testing-strategies/hybrid.md) tutorials demonstrate progressive traffic shift.

## Specifying `weightObjRef`
Iter8 uses the `weightObjRef` field in the experiment resource to get the current traffic split between versions and/or modify the traffic split. Ensure that this field is specified correctly for each version. The following example demonstrates how to specify `weightObjRef` in experiments.

??? example "Example"
    The [SLO validation experiment](../testing-strategies/slovalidation.md) uses a Knative service for traffic shifting. Hence, the experiment manifest specifies the `weightObjRef` field for each version by referencing this Knative service and the traffic fields within the Knative service corresponding to the versions.

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