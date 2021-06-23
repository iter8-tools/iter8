---
template: main.html
---

# Hybrid (A/B + SLOs) testing

!!! tip "Scenario: Hybrid (A/B + SLOs) testing and progressive traffic shift of KFServing models"
    [Hybrid (A/B + SLOs) testing](../../../concepts/buildingblocks.md#hybrid-ab-slos-testing) enables you to combine A/B or A/B/n testing with a reward metric on the one hand with SLO validation using objectives on the other. Among the versions that satisfy objectives, the version which performs best in terms of the reward metric is the winner. In this tutorial, you will:

    1. Perform hybrid (A/B + SLOs) testing.
    2. Specify *user-engagement* as the reward metric.
    3. Specify *latency* and *error-rate* based objectives, for which data will be provided by Prometheus.
    4. Combine hybrid (A/B + SLOs) testing with [progressive traffic shift](../../../concepts/buildingblocks.md#progressive-traffic-shift). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.
    
    ![Hybrid testing](../../../images/quickstart-hybrid.png)

???+ warning "Before you begin, you will need... "
    1. The [kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
    2. [Kustomize 3+](https://kubectl.docs.kubernetes.io/installation/kustomize/).
    3. [Go 1.13+](https://golang.org/doc/install).
    
## 1. Steps 1, 2, and 3
* Follow [Steps 1, 2, and 3 of the KFServing quick start tutorial](../../../getting-started/quick-start/kfserving/tutorial.md). 

## 4. Define metrics
```shell
kubectl apply -f $ITER8/samples/kfserving/hybrid/metrics.yaml
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
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: 95th-percentile-tail-latency
      namespace: iter8-kfserving
    spec:
      description: 95th percentile tail latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          histogram_quantile(0.95, sum(rate(revision_app_request_latencies_bucket{namespace_name='$ns'}[${elapsedTime}s])) by (le))
      provider: prometheus
      sampleSize: iter8-kfserving/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-count
      namespace: iter8-kfserving
    spec:
      description: Number of error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',namespace_name='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-rate
      namespace: iter8-kfserving
    spec:
      description: Fraction of requests with error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',namespace_name='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{namespace_name='$ns'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-kfserving/request-count
      type: Gauge
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: mean-latency
      namespace: iter8-kfserving
    spec:
      description: Mean latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(revision_app_request_latencies_sum{namespace_name='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{namespace_name='$ns'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-kfserving/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
      namespace: iter8-kfserving
    spec:
      description: Number of requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{namespace_name='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ```

## 5. Launch experiment
Launch the hybrid (A/B + SLOs) testing & progressive traffic shift experiment as follows.

```shell
kubectl apply -f $ITER8/samples/kfserving/hybrid/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: hybrid-exp
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

## 6. Understand the experiment
Follow [Step 6 of the quick start tutorial for KFServing](../../../getting-started/quick-start/kfserving/tutorial.md#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`hybrid-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/kfserving/hybrid/experiment.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/baseline.yaml
kubectl delete -f $ITER8/samples/kfserving/quickstart/candidate.yaml
```
