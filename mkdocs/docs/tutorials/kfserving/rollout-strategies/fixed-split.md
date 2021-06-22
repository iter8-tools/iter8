---
template: main.html
---

# Fixed % Split

!!! tip "Scenario: Canary rollout with fixed-%-split"

    [Fixed-%-split](../../../../concepts/buildingblocks/#rollout-strategy) is a type of canary rollout strategy. It enables you to experiment while sending a fixed percentage of traffic to each version as shown below.

    ![Fixed % split](../../../../images/canary-%-based.png)
    
## 1. Setup
    
Please follow [Step 1 of the quick start tutorial](../../../../getting-started/quick-start/kfserving/tutorial/#1-setup).

## 2. Create ML model versions
```shell
kubectl apply -f $ITER8/samples/kfserving/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/kfserving/quickstart/candidate.yaml
kubectl apply -f $ITER8/samples/kfserving/fixed-split/routing-rule.yaml
kubectl wait --for=condition=Ready isvc/flowers -n ns-baseline
kubectl wait --for=condition=Ready isvc/flowers -n ns-candidate
```

??? info "Virtual service with traffic split"
    ```yaml linenums="1"
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: routing-rule
      namespace: default
    spec:
      gateways:
      - knative-serving/knative-ingress-gateway
      hosts:
      - example.com
      http:
      - route:
        - destination:
            host: flowers-predictor-default.ns-baseline.svc.cluster.local
          headers:
            request:
              set:
                Host: flowers-predictor-default.ns-baseline
            response:
              set:
                version: flowers-v1
          weight: 60
        - destination:
            host: flowers-predictor-default.ns-candidate.svc.cluster.local
          headers:
            request:
              set:
                Host: flowers-predictor-default.ns-candidate
            response:
              set:
                version: flowers-v2
          weight: 40
    ```

## 3. Steps 3 and 4
Please follow [Steps 3 and 4 of the quick start tutorial](../../../../getting-started/quick-start/kfserving/tutorial/#3-generate-requests).

## 5. Launch experiment
```shell
kubectl apply -f $ITER8/samples/kfserving/fixed-split/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: fixedsplit-exp
    spec:
      target: flowers
      strategy:
        testingPattern: A/B
        deploymentPattern: FixedSplit
        actions:
          # when the experiment completes, promote the winning version using kubectl apply
          finish:
          - task: common/exec
            with:
              cmd: /bin/bash
              args: [ "-c", "kubectl apply -f {{ .promote }}" ]
      criteria:
        rewards: # Business rewards
        - metric: iter8-kfserving/user-engagement
          preferredDirection: High # maximize user engagement
      duration:
        intervalSeconds: 5
        iterationsPerLoop: 20
      versionInfo:
        # information about model versions used in this experiment
        baseline:
          name: flowers-v1
          variables:
          - name: ns
            value: ns-baseline
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v1.yaml
        candidates:
        - name: flowers-v2
          variables:
          - name: ns
            value: ns-candidate
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v2.yaml
    ```

## 6. Observe experiment
Follow [Step 6 of the quick start tutorial for KFServing](../../../../getting-started/quick-start/kfserving/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`fixedsplit-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/kfserving/fixed-split/experiment.yaml
kubectl delete -f $ITER8/samples/kfserving/fixed-split/routing-rule.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/candidate.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/baseline.yaml
```