---
template: main.html
---

# A/B Testing

!!! tip "Scenario: A/B testing and progressive traffic shift"
    [A/B testing](../../../concepts/buildingblocks.md#ab-testing) enables you to compare two versions of an ML model, and select a winner based on a (business) reward metric. In this tutorial, you will:

    1. Perform A/B testing.
    2. Specify *user-engagement* as the reward metric. This metric will be mocked by Iter8 in this tutorial.
    3. Combine A/B testing with [progressive traffic shifting](../../../concepts/buildingblocks.md#progressive-traffic-shift). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.

    ![Quickstart KFServing](../../../images/quickstart-ab.png)

???+ warning "Platform setup"
    1. Setup [K8s cluster](../../../getting-started/setup-for-tutorials.md#local-kubernetes-cluster)
    2. Get [Linkerd](../setup-for-tutorials.md)
    3. [Install Iter8 in K8s cluster](../../../getting-started/install.md)
    4. Get [`iter8ctl`](../../../getting-started/install.md#install-iter8ctl)

## 1. Create application versions
Create a new namespace, enable Linkerd proxy injection, deploy two Hello World applications, and create a traffic split. 

```shell
kubectl create deployment web --image=gcr.io/google-samples/hello-app:1.0
kubectl expose deployment web --type=NodePort --port=8080

kubectl create deployment web2 --image=gcr.io/google-samples/hello-app:2.0
kubectl expose deployment web2 --type=NodePort --port=8080

kubectl wait --for=condition=Ready pods --all

kubectl apply -f $ITER8/samples/linkerd/quickstart/traffic-split.yaml
```

??? info "Look inside traffic-split.yaml"
    ```yaml linenums="1"
    apiVersion: split.smi-spec.io/v1alpha2
    kind: TrafficSplit
    metadata:
      name: web-traffic-split
    spec:
      service: web
      backends:
      - service: web
        weight: 100
      - service: web2
        weight: 0
    ```

## 2. Generate requests
Generate requests to your app using [Fortio](https://github.com/fortio/fortio) as follows.

```shell
kubectl apply -f $ITER8/samples/linkerd/quickstart/fortio.yaml
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
            command: ["fortio", "load", "-allow-initial-errors", "-t", "6000s", "-qps", "16", "-json", "/shared/fortiooutput.json", $(URL)]
            env:
            - name: URL
              value: web.test:8080
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

## 3. Define metrics
Iter8 defines a custom K8s resource called *Metric* that makes it easy to use metrics from RESTful metric providers like Prometheus, New Relic, Sysdig and Elastic during experiments. 

For the purpose of this tutorial, you will [mock](../../../metrics/mock.md) a number of metrics as follows.

```shell
kubectl apply -f $ITER8/samples/linkerd/quickstart/metrics.yaml
```

??? info "Look inside metrics.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      labels:
        creator: iter8
      name: user-engagement
    spec:
      description: Number of error responses
      type: Gauge
      mock:
      - name: web
        level: 5
      - name: web2
        level: 10
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      labels:
        creator: iter8
      name: error-count 
    spec:
      description: Number of error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(response_total{status_code=~'5..',deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      type: Counter
      urlTemplate: http://prometheus.linkerd-viz:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      labels:
        creator: iter8
      name: error-rate
    spec:
      description: Fraction of requests with error responses
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(response_total{status_code=~'5..',deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s])) or on() vector(0)) / sum(increase(request_total{deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s]))
      provider: prometheus
      sampleSize: request-count
      type: Gauge
      urlTemplate: http://prometheus.linkerd-viz:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      labels:
        creator: iter8
      name: le5ms-latency-percentile
    spec:
      description: Less than 5 ms latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(response_latency_ms_bucket{le='5',deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s])) or on() vector(0)) / sum(increase(response_latency_ms_bucket{le='+Inf',deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s])) or on() vector(0)
      provider: prometheus
      sampleSize: iter8-linkerd/request-count
      type: Gauge
      urlTemplate: http://prometheus.linkerd-viz:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      labels:
        creator: iter8
      name: mean-latency
    spec:
      description: Mean latency
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          (sum(increase(response_latency_ms_sum{deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(request_total{deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s])) or on() vector(0))
      provider: prometheus
      sampleSize: request-count
      type: Gauge
      units: milliseconds
      urlTemplate: http://prometheus.linkerd-viz:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      labels:
        creator: iter8
      name: request-count
    spec:
      description: Number of requests
      jqExpression: .data.result[0].value[1] | tonumber
      params:
      - name: query
        value: |
          sum(increase(request_total{deployment='$name',namespace='$namespace',direction='inbound',tls='true'}[${elapsedTime}s]))
      provider: prometheus
      type: Counter
      urlTemplate: http://prometheus.linkerd-viz:9090/api/v1/query
    ```

## 4. Launch experiment
Launch the A/B testing & progressive traffic shift experiment as follows. This experiment also promotes the winning version of the model at the end.

```shell
kubectl apply -f $ITER8/samples/linkerd/quickstart/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      # target identifies the service under experimentation using its fully qualified name
      target: test/web-traffic-split
      strategy:
        # this experiment will perform an A/B test
        testingPattern: A/B
        # this experiment will progressively shift traffic to the winning version
        deploymentPattern: Progressive
        actions:
          # when the experiment completes, promote the winning version using kubectl apply
          finish:
          - if: CandidateWon()
            run: kubectl -n test apply -f https://raw.githubusercontent.com/alan-cha/iter8/linkerd/samples/linkerd/quickstart/vs-for-v2.yaml
          - if: not CandidateWon()
            run: kubectl -n test apply -f https://raw.githubusercontent.com/alan-cha/iter8/linkerd/samples/linkerd/quickstart/vs-for-v1.yaml
      criteria:
        rewards:
        # (business) reward metric to optimize in this experiment
        - metric: default/user-engagement
          preferredDirection: High
        objectives: # used for validating versions
        - metric: default/mean-latency
          upperLimit: 300
        - metric: default/error-rate
          upperLimit: "0.01"
        requestCount: default/request-count
      duration: # product of fields determines length of the experiment
        intervalSeconds: 10
        iterationsPerLoop: 5
      versionInfo:
        # information about the app versions used in this experiment
        baseline:
          name: web
          variables:
          - name: namespace # used by final action if this version is the winner
            value: test
          weightObjRef:
            apiVersion: split.smi-spec.io/v1alpha2
            kind: TrafficSplit
            name: web-traffic-split
            fieldPath: .spec.backends[0].weight
        candidates:
        - name: web2
          variables:
          - name: namespace # used by final action if this version is the winner
            value: test
          weightObjRef:
            apiVersion: split.smi-spec.io/v1alpha2
            kind: TrafficSplit
            name: web-traffic-split
            fieldPath: .spec.backends[1].weight
    ```

## 3. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 4. Cleanup
```shell
kubectl delete -f $ITER8/samples/linkerd/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/linkerd/quickstart/experiment.yaml
```
