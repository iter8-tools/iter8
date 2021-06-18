---
template: main.html
---

# Hybrid (A/B + SLOs) testing

!!! tip "Scenario: Hybrid (A/B + SLOs) testing and progressive rollout of Knative services"
    [Hybrid (A/B + SLOs) testing](../../concepts/buildingblocks/#testing-pattern) enables you to combine A/B or A/B/n testing with a reward metric on the one hand with SLO validation using objectives on the other. Among the versions that satisfy objectives, the version which performs best in terms of the reward metric is the winner. In this tutorial, you will:

    1. Perform hybrid (A/B + SLOs) testing.
    2. Specify *user-engagement* as the reward metric.
    3. Specify *latency* and *error-rate* based objectives; data for these metrics will be provided by Prometheus.
    4. Combine hybrid (A/B + SLOs) testing with [progressive rollout](../../../../../concepts/buildingblocks/#deployment-pattern). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.
    
    ![Quickstart Seldon](../../../images/quickstart-hybrid.png)

???+ warning "Before you begin, you will need... "
    1. The [kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
    2. [Kustomize 3+](https://kubectl.docs.kubernetes.io/installation/kustomize/).
    3. [Go 1.13+](https://golang.org/doc/install).
    
## 1. Setup
* Setup your K8s cluster with Knative and Iter8 as described [here](../platform-setup/). 
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

## 2. Create app versions
Deploy two versions of a Knative app.

```shell
kubectl apply -f $ITER8/samples/knative/quickstart/baseline.yaml
kubectl apply -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
kubectl wait --for=condition=Ready ksvc/sample-app
```

??? info "Look inside baseline.yaml"
    ```yaml linenums="1"
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app
      namespace: default
    spec:
      template:
        metadata:
          name: sample-app-v1
        spec:
          containers:
          - image: gcr.io/knative-samples/knative-route-demo:blue 
            env:
            - name: T_VERSION
              value: "blue"
    ```

??? info "Look inside experimentalservice.yaml"
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
        percent: 100
      - tag: candidate
        latestRevision: true
        percent: 0
    ```

## 3. Generate requests
Generate requests using [Fortio](https://github.com/fortio/fortio) as follows.

```shell
# URL_VALUE is the URL where your Knative application serves requests
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq ".status.address.url")
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/quickstart/fortio.yaml | kubectl apply -f -
```

??? info "Look inside fortio.yaml"
    ```yaml linenums="1"
    apiVersion: batch/v1
    kind: Job
    metadata:
    name: fortio
    spec:
    template:
        spec:
        volumes:
        - name: shared
            emptyDir: {}
        containers:
        - name: fortio
            image: fortio/fortio
            command: ["fortio", "load", "-t", "6000s", "-qps", "16", "-json", "/shared/fortiooutput.json", $(URL)]
            env:
            - name: URL
            value: URL_VALUE
            volumeMounts:
            - name: shared
            mountPath: /shared         
        - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 600']
            volumeMounts:
            - name: shared
            mountPath: /shared       
        restartPolicy: Never
    ```

## 4. Define metrics
Iter8 defines a custom K8s resource called *Metric* that makes it easy to use metrics from RESTful metric providers like Prometheus, New Relic, Sysdig and Elastic during experiments. 
Define the Iter8 metrics used in this experiment as follows.

```shell
kubectl apply -f $ITER8/samples/knative/quickstart/metrics.yaml
```

??? info "Look inside metrics.yaml"
    ```yaml linenums="1"
    apiVersion: v1
    kind: Namespace
    metadata:
      name: iter8-knative
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: user-engagement
      namespace: iter8-knative
    spec:
      params:
      - name: nrql
        value: |
          SELECT average(duration) FROM Sessions WHERE version='$name' SINCE $elapsedTime sec ago
      description: Average duration of a session
      type: Gauge
      headerTemplates:
      - name: X-Query-Key
        value: t0p-secret-api-key  
      provider: newrelic
      jqExpression: ".results[0] | .[] | tonumber"
      urlTemplate: http://metrics-mock.iter8-system.svc.cluster.local:8080/newrelic
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: 95th-percentile-tail-latency
      namespace: iter8-knative
    spec:
      description: 95th percentile tail latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          histogram_quantile(0.95, sum(rate(revision_app_request_latencies_bucket{revision_name='$name'}[${elapsedTime}s])) by (le))
      provider: prometheus
      sampleSize: iter8-knative/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-count
      namespace: iter8-knative
    spec:
      description: Number of error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',revision_name='$name'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: error-rate
      namespace: iter8-knative
    spec:
      description: Fraction of requests with error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',revision_name='$name'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{revision_name='$name'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-knative/request-count
      type: Gauge
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: mean-latency
      namespace: iter8-knative
    spec:
      description: Mean latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(revision_app_request_latencies_sum{revision_name='$name'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{revision_name='$name'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: iter8-knative/request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
      namespace: iter8-knative
    spec:
      description: Number of requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{revision_name='$name'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ```

??? Note "Metrics in your environment"
    You can define and use custom metrics from any database in Iter8 experiments. 
       
    For your application, replace the mocked user-engagement metric used in this tutorial with any custom metric you wish to optimize in the hybrid (A/B + SLOs) test. Documentation on defining custom metrics is [here](../../../../metrics/custom/).

## 5. Launch experiment
Iter8 defines a custom K8s resource called *Experiment* that automates a variety of release engineering and experimentation strategies for K8s applications and ML models. Launch the hybrid (A/B + SLOs) testing & progressive rollout experiment as follows.

```shell
kubectl apply -f $ITER8/samples/knative/quickstart/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      # target identifies the knative service under experimentation using its fully qualified name
      target: default/sample-app
      strategy:
        testingPattern: A/B
        deploymentPattern: Progressive
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
          weightObjRef:
            apiVersion: serving.knative.dev/v1
            kind: Service
            name: sample-app
            namespace: default
            fieldPath: .spec.traffic[0].percent
          variables:
          - name: promote
            value: baseline
        candidates:
        - name: sample-app-v2
          weightObjRef:
            apiVersion: serving.knative.dev/v1
            kind: Service
            name: sample-app
            namespace: default
            fieldPath: .spec.traffic[1].percent
          variables:
          - name: promote
            value: candidate
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
kubectl get ksvc sample-app -o json --watch | jq ".status.traffic"
```

??? info "Look inside traffic summary"
    The `kubectl` output will be similar to the following.
    ```shell
    [
    {
        "latestRevision": false,
        "percent": 45,
        "revisionName": "sample-app-v1",
        "tag": "current",
        "url": "http://current-sample-app.default.example.com"
    },
    {
        "latestRevision": true,
        "percent": 55,
        "revisionName": "sample-app-v2",
        "tag": "candidate",
        "url": "http://candidate-sample-app.default.example.com"
    }
    ]
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
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```