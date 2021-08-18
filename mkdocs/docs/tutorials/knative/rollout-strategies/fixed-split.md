---
template: main.html
---

# Fixed % Split

!!! tip "Scenario: Canary rollout with fixed-%-split"

    [Fixed-%-split](../../../concepts/buildingblocks.md#fixed-split) is a type of canary rollout strategy. It enables you to experiment while sending a fixed percentage of traffic to each version as shown below.

    ![Fixed % split](../../../images/canary-%-based.png)
    
???+ warning "Platform setup"
    Follow [these steps](../setup-for-tutorials.md) to install Iter8 and Knative in your K8s cluster.

## 1. Create versions and fix traffic split
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

## 2. Launch experiment
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
        testingPattern: Canary
        deploymentPattern: FixedSplit
        actions:
          loop:
          - task: metrics/collect
            with:
              versions:
              - name: sample-app-v1
                url: http://sample-app-v1.default.svc.cluster.local
              - name: sample-app-v2
                url: http://sample-app-v2.default.svc.cluster.local
          finish: # run the following sequence of tasks at the end of the experiment
          - if: CandidateWon()
            run: "kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/candidate.yaml"
          - if: not CandidateWon()
            run: "kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/baseline.yaml"
      criteria:
        requestCount: iter8-system/request-count
        objectives: 
        - metric: iter8-system/mean-latency
          upperLimit: 50
        - metric: iter8-system/latency-95th-percentile
          upperLimit: 100
        - metric: iter8-system/error-rate
          upperLimit: "0.01"
      duration:
        maxLoops: 3
        intervalSeconds: 1
        iterationsPerLoop: 1
      versionInfo:
        # information about app versions used in this experiment
        baseline:
          name: sample-app-v1
        candidates:
        - name: sample-app-v2
    ```

## 3. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 4. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/fixed-split/experiment.yaml
kubectl apply -f $ITER8/samples/knative/fixed-split/experimentalservice.yaml
```
