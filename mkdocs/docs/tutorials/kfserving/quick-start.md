---
template: main.html
---

# A/B Testing and Progressive traffic shift

!!! tip "Scenario: A/B testing and progressive traffic shift for KFServing models"
    [A/B testing](../../concepts/buildingblocks.md#ab-testing) enables you to compare two versions of an ML model, and select a winner based on a (business) reward metric. In this tutorial, you will:

    1. Perform A/B testing.
    2. Specify *user-engagement* as the reward metric. This metric will be mocked by Iter8 in this tutorial.
    3. Combine A/B testing with [progressive traffic shifting](../../concepts/buildingblocks.md#progressive-traffic-shift). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.

    ![Quickstart KFServing](../../images/quickstart-ab.png)

???+ warning "Platform setup"
    Follow [these steps](platform-setup.md) to install Iter8, KFServing and Prometheus in your K8s cluster.

## 1. Create ML model versions
Deploy two KFServing inference services corresponding to two versions of a TensorFlow classification model, along with an Istio virtual service to split traffic between them.

```shell
kubectl apply -f $ITER8/samples/kfserving/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/kfserving/quickstart/candidate.yaml
kubectl apply -f $ITER8/samples/kfserving/quickstart/routing-rule.yaml
kubectl wait --for=condition=Ready isvc/flowers -n ns-baseline
kubectl wait --for=condition=Ready isvc/flowers -n ns-candidate
```

??? info "Look inside baseline.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
      name: ns-baseline
    ---
    apiVersion: serving.kubeflow.org/v1beta1
    kind: InferenceService
    metadata:
      name: flowers
      namespace: ns-baseline
    spec:
      predictor:
        tensorflow:
          storageUri: "gs://kfserving-samples/models/tensorflow/flowers"
    ```

??? info "Look inside candidate.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
      name: ns-candidate
    ---
    apiVersion: serving.kubeflow.org/v1beta1
    kind: InferenceService
    metadata:
      name: flowers
      namespace: ns-candidate
    spec:
      predictor:
        tensorflow:
          storageUri: "gs://kfserving-samples/models/tensorflow/flowers-2"
    ```

??? info "Look inside routing-rule.yaml"
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
          weight: 100
        - destination:
            host: flowers-predictor-default.ns-candidate.svc.cluster.local
          headers:
            request:
              set:
                Host: flowers-predictor-default.ns-candidate
            response:
              set:
                version: flowers-v2
          weight: 0
    ```

## 2. Generate requests
Generate requests for your model as follows.

=== "Port forward Istio ingress in terminal one"
    ```shell
    INGRESS_GATEWAY_SERVICE=$(kubectl get svc -n istio-system --selector="app=istio-ingressgateway" --output jsonpath='{.items[0].metadata.name}')
    kubectl port-forward -n istio-system svc/${INGRESS_GATEWAY_SERVICE} 8080:80
    ```

=== "Send requests in terminal two"
    ```shell
    curl -o /tmp/input.json https://raw.githubusercontent.com/kubeflow/kfserving/master/docs/samples/v1beta1/rollout/input.json
    watch --interval 0.2 -x curl -v -H "Host: example.com" localhost:8080/v1/models/flowers:predict -d @/tmp/input.json
    ```

## 3. Define metrics
Iter8 defines a custom K8s resource called *Metric* that makes it easy to use metrics from RESTful metric providers like Prometheus, New Relic, Sysdig and Elastic during experiments. 

For the purpose of this tutorial, you will [mock](../../metrics/mock.md) the user-engagement metric as follows.

```shell
kubectl apply -f $ITER8/samples/kfserving/quickstart/metrics.yaml
```

??? info "Look inside metrics.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
      name: iter8-kfserving
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: user-engagement
      namespace: iter8-kfserving
    spec:
      mock:
      - name: flowers-v1
        level: "15.0"
      - name: flowers-v2
        level: "20.0"
    ```

??? Note "Metrics in your environment"
    You can define and use custom metrics from any database in Iter8 experiments. 
       
    For your application, replace the mocked metric used in this tutorial with any custom metric you wish to optimize in the A/B test. Documentation on defining custom metrics is [here](../../metrics/custom.md).

## 4. Launch experiment
Launch the A/B testing & progressive traffic shift experiment as follows. This experiment also promotes the winning version of the model at the end.

```shell
kubectl apply -f $ITER8/samples/kfserving/quickstart/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      target: flowers
      strategy:
        testingPattern: A/B
        deploymentPattern: Progressive
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
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-rule
            namespace: default
            fieldPath: .spec.http[0].route[0].weight      
          variables:
          - name: ns
            value: ns-baseline
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v1.yaml
        candidates:
        - name: flowers-v2
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-rule
            namespace: default
            fieldPath: .spec.http[0].route[1].weight      
          variables:
          - name: ns
            value: ns-candidate
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v2.yaml
    ```

## 5. Observe experiment
Follow [these steps](../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/kfserving/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/baseline.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/candidate.yaml
```
