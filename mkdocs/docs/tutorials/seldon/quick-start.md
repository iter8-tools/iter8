---
template: main.html
---

# Hybrid (A/B + SLOs) testing

!!! tip "Scenario: Hybrid (A/B + SLOs) testing and progressive traffic shift of Seldon models"
    [Hybrid (A/B + SLOs) testing](../../concepts/buildingblocks.md#hybrid-ab-slos-testing) enables you to combine A/B or A/B/n testing with a reward metric on the one hand with SLO validation using objectives on the other. Among the versions that satisfy objectives, the version which performs best in terms of the reward metric is the winner. In this tutorial, you will:

    1. Perform hybrid (A/B + SLOs) testing.
    2. Specify *user-engagement* as the reward metric; data for this metric will be provided by Prometheus.
    3. Specify *latency* and *error-rate* based objectives; data for these metrics will be provided by Prometheus.
    4. Combine hybrid (A/B + SLOs) testing with [progressive traffic shift](../../concepts/buildingblocks.md#progressive-traffic-shift). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.
    
    ![Quickstart Seldon](../../images/quickstart-hybrid.png)
    
???+ warning "Platform setup"
    Follow [these steps](platform-setup.md) to install Seldon and Iter8 in your K8s cluster.

## 1. Create ML model versions
Deploy two Seldon Deployments corresponding to two versions of an Iris classification model, along with an Istio virtual service to split traffic between them.

```shell
kubectl apply -f $ITER8/samples/seldon/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/seldon/quickstart/candidate.yaml
kubectl apply -f $ITER8/samples/seldon/quickstart/routing-rule.yaml
kubectl wait --for=condition=Ready --timeout=600s pods --all -n ns-baseline
kubectl wait --for=condition=Ready --timeout=600s pods --all -n ns-candidate
```

??? info "Look inside baseline.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
      name: ns-baseline
    ---
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      name: iris
      namespace: ns-baseline
    spec:
      predictors:
      - name: default
        graph:
          name: classifier
          modelUri: gs://seldon-models/sklearn/iris
          implementation: SKLEARN_SERVER
    ```

??? info "Look inside candidate.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
        name: ns-candidate
    ---
    apiVersion: machinelearning.seldon.io/v1
    kind: SeldonDeployment
    metadata:
      name: iris
      namespace: ns-candidate
    spec:
      predictors:
      - name: default
        graph:
          name: classifier
          modelUri: gs://seldon-models/xgboost/iris
          implementation: XGBOOST_SERVER
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
      - istio-system/seldon-gateway
      hosts:
      - iris.example.com
      http:
      - route:
        - destination:
            host: iris-default.ns-baseline.svc.cluster.local
            port:
              number: 8000
          headers:
            response:
              set:
                version: iris-v1
          weight: 100
        - destination:
            host: iris-default.ns-candidate.svc.cluster.local
            port:
              number: 8000
          headers:
            response:
              set:
                version: iris-v2
          weight: 0

    ```

## 2. Generate requests
Generate requests using [Fortio](https://github.com/fortio/fortio) as follows.

```shell
URL_VALUE="http://$(kubectl -n istio-system get svc istio-ingressgateway -o jsonpath='{.spec.clusterIP}'):80"
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/seldon/quickstart/fortio.yaml | sed "s/6000s/600s/g" | kubectl apply -f -
```

??? info "Look inside fortio.yaml"
    ```yaml linenums="1"
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio-requests
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "5", '-json', '/shared/fortiooutput.json', '-H', 'Host: iris.example.com', '-H', 'Content-Type: application/json', '-payload', '{"data": {"ndarray":[[6.8,2.8,4.8,1.4]]}}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/api/v1.0/predictions
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 6000']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never
    ---
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio-irisv1-rewards
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "0.7", '-json', '/shared/fortiooutput.json', '-H', 'Content-Type: application/json', '-payload', '{"reward": 1}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/seldon/ns-baseline/iris/api/v1.0/feedback
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 6000']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never
    ---
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: fortio-irisv2-rewards
    spec:
      template:
        spec:
          volumes:
          - name: shared
            emptyDir: {}    
          containers:
          - name: fortio
            image: fortio/fortio
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "1", '-json', '/shared/fortiooutput.json', '-H', 'Content-Type: application/json', '-payload', '{"reward": 1}',  "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE/seldon/ns-candidate/iris/api/v1.0/feedback
            volumeMounts:
            - name: shared
              mountPath: /shared         
          - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 6000']          
            volumeMounts:
            - name: shared
              mountPath: /shared       
          restartPolicy: Never
    
    ```

## 3. Define metrics
Iter8 defines a custom K8s resource called *Metric* that makes it easy to use metrics from RESTful metric providers like Prometheus, New Relic, Sysdig and Elastic during experiments. 
Define the Iter8 metrics used in this experiment as follows.

```shell
kubectl apply -f $ITER8/samples/seldon/quickstart/metrics.yaml
```

??? info "Look inside metrics.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
      name: iter8-seldon
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: 95th-percentile-tail-latency
      namespace: iter8-seldon
    spec:
      description: 95th percentile tail latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          histogram_quantile(0.95, sum(rate(seldon_api_executor_client_requests_seconds_bucket{seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) by (le))
      provider: prometheus
      sampleSize: iter8-seldon/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-count
      namespace: iter8-seldon
    spec:
      description: Number of error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(seldon_api_executor_server_requests_seconds_count{code!='200',seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query  
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-rate
      namespace: iter8-seldon
    spec:
      description: Fraction of requests with error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(seldon_api_executor_server_requests_seconds_count{code!='200',seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(seldon_api_executor_server_requests_seconds_count{seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-seldon/request-count
      type: Gauge
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query    
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: mean-latency
      namespace: iter8-seldon
    spec:
      description: Mean latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(seldon_api_executor_client_requests_seconds_sum{seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(seldon_api_executor_client_requests_seconds_count{seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-seldon/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query      
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
      namespace: iter8-seldon
    spec:
      description: Number of requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(seldon_api_executor_client_requests_seconds_sum{seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: user-engagement
      namespace: iter8-seldon
    spec:
      description: Number of feedback requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(seldon_api_executor_server_requests_seconds_count{service='feedback',seldon_deployment_id='$sid',kubernetes_namespace='$ns'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Gauge
      urlTemplate: http://seldon-core-analytics-prometheus-seldon.seldon-system/api/v1/query
    ```

??? Note "Metrics in your environment"
    You can define and use custom metrics from any database in Iter8 experiments. 
       
    For your application, replace the mocked user-engagement metric used in this tutorial with any custom metric you wish to optimize in the hybrid (A/B + SLOs) test. Documentation on defining custom metrics is [here](../../metrics/custom.md).

## 4. Launch experiment
Launch the hybrid (A/B + SLOs) testing & progressive traffic shift experiment as follows. This experiment also promotes the winning version of the model at the end.

```shell
kubectl apply -f $ITER8/samples/seldon/quickstart/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      target: iris
      strategy:
        testingPattern: A/B
        deploymentPattern: Progressive
        actions:
          # when the experiment completes, promote the winning version using kubectl apply
          finish:
          - if: CandidateWon()
            run: kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/seldon/quickstart/promote-v2.yaml
          - if: not CandidateWon()
            run: kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/seldon/quickstart/promote-v1.yaml
      criteria:
        requestCount: iter8-seldon/request-count
        rewards: # Business rewards
        - metric: iter8-seldon/user-engagement
          preferredDirection: High # maximize user engagement
        objectives:
        - metric: iter8-seldon/mean-latency
          upperLimit: 2000
        - metric: iter8-seldon/95th-percentile-tail-latency
          upperLimit: 5000
        - metric: iter8-seldon/error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 5
      versionInfo:
        # information about model versions used in this experiment
        baseline:
          name: iris-v1
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-rule
            namespace: default
            fieldPath: .spec.http[0].route[0].weight      
          variables:
          - name: ns
            value: ns-baseline
          - name: sid
            value: iris
        candidates:
        - name: iris-v2
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-rule
            namespace: default
            fieldPath: .spec.http[0].route[1].weight      
          variables:
          - name: ns
            value: ns-candidate
          - name: sid
            value: iris   
    ```

## 5. Observe experiment
Follow [these steps](../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.
    
## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/seldon/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/baseline.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/candidate.yaml
```
