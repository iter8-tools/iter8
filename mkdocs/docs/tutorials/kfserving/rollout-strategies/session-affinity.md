---
template: main.html
---

# Session Affinity

!!! tip "Scenario: Canary rollout with session affinity"

    [Session affinity](../../../../concepts/buildingblocks/#session-affinity) ensures that the version to which a particular user's request is routed remains consistent throughout the duration of the experiment. In this tutorial, you will use an experiment involving two user groups, 1 and 2. Reqeusts from user group 1 will have a `userhash` header value prefixed with `111` and will be routed to the baseline version. Requests from user group 2 will have a `userhash` header value prefixed with `101` and will be routed to the candidate version. The experiment is shown below.

    ![Session affinity](../../../images/session-affinity-exp.png)

## 1. Setup
* Setup your K8s cluster with KFServing and Iter8 as described [here](../../../../getting-started/quick-start/kfserving/platform-setup/).
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

## 2. Create ML model versions
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

## 3. Generate requests
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

## 4. Define metrics
Please follow [Step 4 of the quick start tutorial](../../../../getting-started/quick-start/kfserving/tutorial/#4-define-metrics).

## 5. Launch experiment
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
          - task: common/exec
            with:
              cmd: /bin/bash
              args: [ "-c", "kubectl apply -f {{ .promote }}" ]
      criteria:
        requestCount: iter8-kfserving/request-count
        rewards: # Business rewards
        - metric: iter8-kfserving/user-engagement
          preferredDirection: High # maximize user engagement
        objectives:
        - metric: iter8-kfserving/mean-latency
          upperLimit: 2000
        - metric: iter8-kfserving/95th-percentile-tail-latency
          upperLimit: 5000
        - metric: iter8-kfserving/error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
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

## 6. Understand the experiment
Follow [Step 6 of the quick start tutorial for KFServing](../../../../getting-started/quick-start/kfserving/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`session-affinity-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/kfserving/session-affinity/experiment.yaml
kubectl delete -f $ITER8/samples/kfserving/session-affinity/routing-rule.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/candidate.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/baseline.yaml
```
