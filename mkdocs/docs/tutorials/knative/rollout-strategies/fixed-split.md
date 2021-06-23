---
template: main.html
---

# Fixed % Split

!!! tip "Scenario: Canary rollout with fixed-%-split"

    [Fixed-%-split](../../../../concepts/buildingblocks/#rollout-strategy) is a type of canary rollout strategy. It enables you to experiment while sending a fixed percentage of traffic to each version as shown below.

    ![Fixed % split](../../../images/canary-%-based.png)
    
## 1. Setup
* Setup your K8s cluster with Knative and Iter8 as described [here](../../../../getting-started/quick-start/knative/platform-setup/).
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

## 2. Create versions and initialize traffic split
```shell
kubectl apply -f $ITER8/samples/knative/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/knative/fixed-split/experimentalservice.yaml
kubectl wait --for=condition=Ready ksvc/sample-app
```

??? info "Knative service with traffic split"
    ```yaml linenums="1"
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app
      namespace: default
    spec:
      template:
        metadata:
          name: sample-app-v2
        spec:
          containers:
          - image: gcr.io/knative-samples/knative-route-demo:green 
            env:
            - name: T_VERSION
              value: "green"
      traffic:
      - tag: current
        revisionName: sample-app-v1
        percent: 60
      - tag: candidate
        latestRevision: true
        percent: 40
    ```

## 3. Generate requests
Please follow [Step 3 of the quick start tutorial](../../../../getting-started/quick-start/knative/tutorial/#3-generate-requests).

## 4. Define metrics
Please follow [Step 4 of the quick start tutorial](../../../../getting-started/quick-start/knative/tutorial/#4-define-metrics).

## 5. Launch experiment
```shell
kubectl apply -f $ITER8/samples/knative/fixed-split/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: fixedsplit-exp
    spec:
      # target identifies the knative service under experimentation using its fully qualified name
      target: default/sample-app
      strategy:
        testingPattern: A/B
        deploymentPattern: FixedSplit
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
        rewards: # Business rewards
        - metric: iter8-knative/user-engagement
          preferredDirection: High # maximize user engagement
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
          variables:
          - name: promote
            value: baseline
        candidates:
        - name: sample-app-v2
          variables:
          - name: promote
            value: candidate
    ```

## 6. Understand the experiment
Follow [Step 6 of the quick start tutorial for Knative](../../../../getting-started/quick-start/knative/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`fixedsplit-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/fixed-split/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl apply -f $ITER8/samples/knative/fixed-split/experimentalservice.yaml
```
