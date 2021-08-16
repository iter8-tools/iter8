---
template: main.html
---

# Session Affinity

!!! tip "Scenario: Canary rollout with session affinity"

    [Session affinity](../../../../concepts/buildingblocks/#session-affinity) ensures that the version to which a particular user's request is routed remains consistent throughout the duration of the experiment. In this tutorial, you will use an experiment involving two user groups, 1 and 2. Reqeusts from user group 1 will have a `userhash` header value prefixed with `111` and will be routed to the baseline version. Requests from user group 2 will have a `userhash` header value prefixed with `101` and will be routed to the candidate version. The experiment is shown below.

    ![Session affinity](../../../images/session-affinity-exp.png)

???+ warning "Platform setup"
    Follow [these steps](../platform-setup.md) to install Iter8, KFServing and Prometheus in your K8s cluster.

## 1. Create ML model versions
Deploy two KFServing inference services corresponding to two versions of a TensorFlow classification model, along with an Istio virtual service to split traffic between them.

```shell
kubectl apply -f $ITER8/samples/kfserving/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/kfserving/quickstart/candidate.yaml
kubectl apply -f $ITER8/samples/kfserving/session-affinity/routing-rule.yaml
kubectl wait --for=condition=Ready isvc/flowers -n ns-baseline
kubectl wait --for=condition=Ready isvc/flowers -n ns-candidate
```

??? info "Istio virtual service"
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
      - match:
        - headers:
            userhash: # user hash is a 10-digit random binary string
              prefix: "101" # in expectation, 1/8th of user hashes will match this prefix
        route: # matching users will always go to v2
        - destination:
            host: flowers-predictor-default.ns-candidate.svc.cluster.local
          headers:
            request:
              set:
                Host: flowers-predictor-default.ns-candidate
            response:
              set:
                version: flowers-v2
      - route: # non-matching users will always go to v1
        - destination:
            host: flowers-predictor-default.ns-baseline.svc.cluster.local
          headers:
            request:
              set:
                Host: flowers-predictor-default.ns-baseline
            response:
              set:
                version: flowers-v1
    ```

## 2. Generate requests
Generate requests to your model as follows.

=== "Port forward (terminal one)"
    ```shell
    INGRESS_GATEWAY_SERVICE=$(kubectl get svc -n istio-system --selector="app=istio-ingressgateway" --output jsonpath='{.items[0].metadata.name}')
    kubectl port-forward -n istio-system svc/${INGRESS_GATEWAY_SERVICE} 8080:80
    ```

=== "Baseline requests (terminal two)"
    ```shell
    curl -o /tmp/input.json https://raw.githubusercontent.com/kubeflow/kfserving/master/docs/samples/v1beta1/rollout/input.json
    while true; do
    curl -v -H "Host: example.com" -H "userhash: 1111100000" localhost:8080/v1/models/flowers:predict -d @/tmp/input.json
    sleep 0.29
    done
    ```

=== "Candidate requests (terminal three)"
    ```shell
    curl -o /tmp/input.json https://raw.githubusercontent.com/kubeflow/kfserving/master/docs/samples/v1beta1/rollout/input.json
    while true; do
    curl -v -H "Host: example.com" -H "userhash: 1010101010" localhost:8080/v1/models/flowers:predict -d @/tmp/input.json
    sleep 2.0
    done
    ```

## 3. Define metrics
Please follow [Step 3 of the quick start tutorial](../quick-start.md#3-define-metrics).

## 4. Launch experiment
```shell
kubectl apply -f $ITER8/samples/kfserving/session-affinity/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: session-affinity-exp
    spec:
      target: flowers
      strategy:
        testingPattern: A/B
        deploymentPattern: FixedSplit
        actions:
          # when the experiment completes, promote the winning version using kubectl apply
          finish:
          - if: CandidateWon()
            run: kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v2.yaml
          - if: not CandidateWon()
            run: kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v1.yaml
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
        candidates:
        - name: flowers-v2
          variables:
          - name: ns
            value: ns-candidate
    ```

## 5. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/kfserving/session-affinity/experiment.yaml
kubectl delete -f $ITER8/samples/kfserving/session-affinity/routing-rule.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/candidate.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/baseline.yaml
```
