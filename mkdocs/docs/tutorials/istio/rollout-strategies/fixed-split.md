---
template: main.html
---

# Fixed % Split

!!! tip "Scenario: Canary rollout with fixed-%-split"

    [Fixed-%-split](../../../concepts/buildingblocks.md#rollout-strategy) is a type of canary rollout strategy. It enables you to experiment while sending a fixed percentage of traffic to each version as shown below.

    ![Fixed % split](../../../images/canary-%-based.png)
    
???+ warning "Platform setup"
    Follow [these steps](../platform-setup.md) to install Iter8 and Istio in your K8s cluster. 

## 1. Create versions and fix traffic split
```shell
kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/fixed-split/bookinfo-app.yaml
kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/quickstart/productpage-v2.yaml
kubectl wait -n bookinfo-iter8 --for=condition=Ready pods --all
```

??? info "Virtual service with traffic split"
    ```yaml linenums="1"
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: bookinfo
    spec:
      gateways:
      - mesh
      - bookinfo-gateway
      hosts:
      - productpage
      - "bookinfo.example.com"
      http:
      - match:
        - uri:
            exact: /productpage
        - uri:
            prefix: /static
        - uri:
            exact: /login
        - uri:
            exact: /logout
        - uri:
            prefix: /api/v1/products
        route:
        - destination:
            host: productpage
            port:
              number: 9080
            subset: productpage-v1
          weight: 60
        - destination:
            host: productpage
            port:
              number: 9080
            subset: productpage-v2
          weight: 40
    ```

## 2. Steps 2 and 3
Please follow [Steps 2 and 3 of the quick start tutorial](../quick-start.md).

## 4. Launch experiment
```shell
kubectl apply -f $ITER8/samples/istio/fixed-split/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: fixedsplit-exp
    spec:
      # target identifies the service under experimentation using its fully qualified name
      target: bookinfo-iter8/productpage
      strategy:
        # this experiment will perform an A/B test
        testingPattern: A/B
        # this experiment will not shift traffic during iterations
        deploymentPattern: FixedSplit
        actions:
          # when the experiment completes, promote the winning version using kubectl apply
          finish:
          - task: common/exec
            with:
              cmd: /bin/bash
              args: [ "-c", "kubectl -n bookinfo-iter8 apply -f {{ .promote }}" ]
      criteria:
        rewards:
        # (business) reward metric to optimize in this experiment
        - metric: iter8-istio/user-engagement 
          preferredDirection: High
        objectives: # used for validating versions
        - metric: iter8-istio/mean-latency
          upperLimit: 300
        - metric: iter8-istio/error-rate
          upperLimit: "0.01"
        requestCount: iter8-istio/request-count
      duration: # product of fields determines length of the experiment
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        # information about the app versions used in this experiment
        baseline:
          name: productpage-v1
          variables:
          - name: namespace # used by final action if this version is the winner
            value: bookinfo-iter8
          - name: promote # used by final action if this version is the winner
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/istio/quickstart/vs-for-v1.yaml
        candidates:
        - name: productpage-v2
          variables:
          - name: namespace # used by final action if this version is the winner
            value: bookinfo-iter8
          - name: promote # used by final action if this version is the winner
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/istio/quickstart/vs-for-v2.yaml
    ```

## 5. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/istio/fixed-split/experiment.yaml
kubectl delete -f $ITER8/samples/istio/quickstart/fortio.yaml
kubectl delete ns bookinfo-iter8
```
