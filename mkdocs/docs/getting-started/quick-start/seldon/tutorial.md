---
template: main.html
---

# Hybrid (A/B + SLOs) testing

!!! tip "Scenario: Hybrid (A/B + SLOs) testing and progressive rollout of Seldon models"
    [Hybrid (A/B + SLOs) testing](../../concepts/buildingblocks/#testing-pattern) enables you to combine A/B or A/B/n testing with a reward metric on the one hand with SLO validation using objectives on the other. Among the versions that satisfy objectives, the version which performs best in terms of the reward metric is the winner. In this tutorial, you will:

    1. Perform hybrid (A/B + SLOs) testing.
    2. Specify *user-engagement* as the reward metric; data for this metric will be provided by Prometheus.
    3. Specify *latency* and *error-rate* based objectives; data for these metrics will be provided by Prometheus.
    4. Combine hybrid (A/B + SLOs) testing with [progressive rollout](../../../../../concepts/buildingblocks/#deployment-pattern). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.
    
    ![Quickstart Seldon](../../../images/quickstart-hybrid.png)

???+ warning "Before you begin, you will need... "
    1. The [kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
    2. [Kustomize 3+](https://kubectl.docs.kubernetes.io/installation/kustomize/).
    3. [Go 1.13+](https://golang.org/doc/install).
    4. [Helm 3+](https://helm.sh/docs/intro/install/)
    
## 1. Setup
* Setup your K8s cluster with Seldon and Iter8 as described [here](../platform-setup/). 
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

## 2. Create ML model versions
Deploy two Seldon Deployments corresponding to two versions of an Iris classification model, along with an Istio virtual service to split traffic between them.

```shell
kubectl apply -f $ITER8/samples/seldon/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/seldon/quickstart/candidate.yaml
kubectl apply -f $ITER8/samples/seldon/quickstart/routing-rule.yaml
kubectl wait --for condition=ready --timeout=600s pods --all -n ns-baseline
kubectl wait --for condition=ready --timeout=600s pods --all -n ns-candidate
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

## 3. Generate requests
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

## 4. Define metrics
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
       
    For your application, replace the mocked user-engagement metric used in this tutorial with any custom metric you wish to optimize in the hybrid (A/B + SLOs) test. Documentation on defining custom metrics is [here](../../../../metrics/custom/).

## 5. Launch experiment
Iter8 defines a custom K8s resource called *Experiment* that automates a variety of release engineering and experimentation strategies for K8s applications and ML models. Launch the hybrid (A/B + SLOs) testing & progressive rollout experiment as follows.

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
          - task: common/exec
            with:
              cmd: /bin/bash
              args: [ "-c", "kubectl apply -f {{ .promote }}" ]
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
        iterationsPerLoop: 10
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
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/seldon/quickstart/promote-v1.yaml
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
          - name: promote
            value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/seldon/quickstart/promote-v2.yaml     
    ```

## 6. Understand the experiment
The process automated by Iter8 in this experiment is as follows.
    
![Iter8 automation](../../../images/quickstart-iter8-process.png)

Observe the results of the experiment in real-time as follows.
### a) Observe results
Install `iter8ctl`. You can change the directory where `iter8ctl` binary is installed by changing `GOBIN` below.
```shell
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.4
```

Periodically describe the experiment.
```shell
watch -x iter8ctl describe -f - <(kubectl get experiment quickstart-exp -o yaml)
```

??? info "Experiment results will look similar to this"
    ```shell
    ****** Overview ******
    Experiment name: quickstart-exp
    Experiment namespace: default
    Target: default/sample-app
    Testing pattern: A/B
    Deployment pattern: Progressive

    ****** Progress Summary ******
    Experiment stage: Running
    Number of completed iterations: 8

    ****** Winner Assessment ******
    App versions in this experiment: [sample-app-v1 sample-app-v2]
    Winning version: sample-app-v2
    Version recommended for promotion: sample-app-v2

    ****** Objective Assessment ******
    > Identifies whether or not the experiment objectives are satisfied by the most recently observed metrics values for each version.
    +--------------------------------------------+---------------+---------------+
    |                 OBJECTIVE                  | SAMPLE-APP-V1 | SAMPLE-APP-V2 |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/mean-latency <=              | true          | true          |
    |                                     50.000 |               |               |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/95th-percentile-tail-latency | true          | true          |
    | <= 100.000                                 |               |               |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/error-rate <=                | true          | true          |
    |                                      0.010 |               |               |
    +--------------------------------------------+---------------+---------------+

    ****** Metrics Assessment ******
    > Most recently read values of experiment metrics for each version.
    +--------------------------------------------+---------------+---------------+
    |                   METRIC                   | SAMPLE-APP-V1 | SAMPLE-APP-V2 |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/request-count                |      1213.625 |       361.962 |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/user-engagement              |        10.023 |        14.737 |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/mean-latency                 |         1.133 |         1.175 |
    | (milliseconds)                             |               |               |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/95th-percentile-tail-latency |         4.768 |         4.824 |
    | (milliseconds)                             |               |               |
    +--------------------------------------------+---------------+---------------+
    | iter8-knative/error-rate                   |         0.000 |         0.000 |
    +--------------------------------------------+---------------+---------------+
    ``` 

Observe how traffic is split between versions in real-time as follows.
### b) Observe traffic
```shell
kubectl get vs routing-rule -o json --watch | jq ".spec.http[0].route"
```

??? info "Look inside traffic summary"
    ```json
    [
      {
        "destination": {
          "host": "iris-default.ns-baseline.svc.cluster.local",
          "port": {
            "number": 8000
          }
        },
        "headers": {
          "response": {
            "set": {
              "version": "iris-v1"
            }
          }
        },
        "weight": 25
      },
      {
        "destination": {
          "host": "iris-default.ns-candidate.svc.cluster.local",
          "port": {
            "number": 8000
          }
        },
        "headers": {
          "response": {
            "set": {
              "version": "iris-v2"
            }
          }
        },
        "weight": 75
      }

    ```

### c) Observe progress
```shell
kubectl get experiment quickstart-exp --watch
```

??? info "Look inside progress summary"
    The `kubectl` output will be similar to the following.
    ```shell
    NAME             TYPE     TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
    quickstart-exp   Canary   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
    quickstart-exp   Canary   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
    quickstart-exp   Canary   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
    quickstart-exp   Canary   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
    quickstart-exp   Canary   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
    quickstart-exp   Canary   default/sample-app   Running   6                      IterationUpdate: Completed Iteration 6
    quickstart-exp   Canary   default/sample-app   Running   7                      IterationUpdate: Completed Iteration 7
    quickstart-exp   Canary   default/sample-app   Running   8                      IterationUpdate: Completed Iteration 8
    quickstart-exp   Canary   default/sample-app   Running   9                      IterationUpdate: Completed Iteration 9
    ```
    
## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/seldon/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/baseline.yaml
kubectl delete -f $ITER8/samples/seldon/quickstart/candidate.yaml
```
