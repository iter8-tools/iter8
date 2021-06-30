---
template: main.html
---

# Hybrid (A/B + SLOs) testing

!!! tip "Scenario: Hybrid (A/B + SLOs) testing and progressive rollout of Knative services"
    [Hybrid (A/B + SLOs) testing](../../../concepts/buildingblocks.md#hybrid-ab-slos-testing) enables you to combine A/B or A/B/n testing with a reward metric on the one hand with SLO validation using objectives on the other. Among the versions that satisfy objectives, the version which performs best in terms of the reward metric is the winner. In this tutorial, you will:

    1. Perform hybrid (A/B + SLOs) testing.
    2. Specify *user-engagement* as the reward metric.
    3. Specify *latency* and *error-rate* based objectives; these metrics will be collected using Iter8's built-in metrics collection feature.
    4. Combine hybrid (A/B + SLOs) testing with [progressive rollout](../../../concepts/buildingblocks.md#progressive-traffic-shift). Iter8 will progressively shift traffic towards the winner and promote it at the end as depicted below.
    
    ![Hybrid testing](../../../images/quickstart-hybrid.png)

???+ warning "Platform setup"
    Follow [these steps](../platform-setup.md) to install Iter8 and Knative in your K8s cluster.

## 1. Create app versions
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

## 2. Define metrics
```shell
kubectl apply -f $ITER8/samples/knative/hybrid/metrics.yaml
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
      urlTemplate: https://my-newrelic-service.com
      mock:
      - name: sample-app-v1
        level: 15.0
      - name: sample-app-v2
        level: 20.0
    ```

??? Note "Metrics in your environment"
    You can define and use custom metrics from any database in Iter8 experiments. 
       
    For your application, replace the mocked user-engagement metric used in this tutorial with any custom metric you wish to optimize in the hybrid (A/B + SLOs) test. Documentation on defining custom metrics is [here](../../../metrics/custom.md).

## 3. Launch experiment
```shell
kubectl apply -f $ITER8/samples/knative/hybrid/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: hybrid-exp
    spec:
      target: default/sample-app
      strategy:
        testingPattern: A/B
        deploymentPattern: Progressive
        actions:
          loop:
          - task: metrics/collect
            with:
              versions:
              - name: sample-app-v1
                url: http://sample-app-v1.default.svc.cluster.local
              - name: sample-app-v2
                url: http://sample-app-v2.default.svc.cluster.local
          finish: # run the following sequence of tasks at the end of the experiment
          - task: common/exec # promote the winning version      
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
                kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml
      criteria:
        rewards: # Business rewards
        - metric: iter8-knative/user-engagement
          preferredDirection: High # maximize user engagement
        objectives: 
        - metric: iter8-system/mean-latency
          upperLimit: 50
        - metric: iter8-system/latency-95th-percentile
          upperLimit: 100
        - metric: iter8-system/error-rate
          upperLimit: "0.01"
        requestCount: iter8-system/request-count
      duration:
        maxLoops: 10
        intervalSeconds: 1
        iterationsPerLoop: 1
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

## 4. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 5. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/hybrid/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```