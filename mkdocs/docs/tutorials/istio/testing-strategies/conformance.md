---
template: main.html
---

# SLO Validation with a single version

!!! tip "Scenario: SLO validation with a single version"
    Iter8 enables you to perform SLO validation with a single version of your application (a.k.a. [conformance testing](../../../concepts/buildingblocks.md#slo-validation)). In this tutorial, you will:

    1. Perform conformance testing.
    2. Specify *latency* and *error-rate* based service-level objectives (SLOs). If your version satisfies SLOs, Iter8 will declare it as the winner.
    3. Use Prometheus as the provider for latency and error-rate metrics.
    
    ![Conformance](../../../images/conformance.png)

???+ warning "Platform setup"
    Follow [these steps](../platform-setup.md) to install Iter8 and Istio in your K8s cluster. 

## 1. Create application version
Deploy [bookinfo](https://istio.io/latest/docs/examples/bookinfo/) app:

```shell
kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/conformance/bookinfo-app.yaml
```

??? info "Look inside productpage-v1 defined in bookinfo-app.yaml"
    ```yaml linenums="1"
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: productpage-v1
      labels:
        app: productpage
        version: v1
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: productpage
          version: v1
      template:
        metadata:
          annotations:
            sidecar.istio.io/inject: "true"
            prometheus.io/scrape: "true"
            prometheus.io/path: /metrics
            prometheus.io/port: "9080"
          labels:
            app: productpage
            version: v1
        spec:
          serviceAccountName: bookinfo-productpage
          containers:
          - name: productpage
            image: iter8/productpage:demo
            imagePullPolicy: IfNotPresent
            ports:
            - containerPort: 9080
            env:
              - name: deployment
                value: "productpage-v1"
              - name: namespace
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.namespace
              - name: color
                value: "red"
              - name: reward_min
                value: "0"
              - name: reward_max
                value: "5"
              - name: port
                value: "9080"
    ```

## 2. Generate requests
Generate requests using Fortio as follows.

```shell
kubectl wait -n bookinfo-iter8 --for=condition=Ready pods --all
# URL_VALUE is the URL of the `bookinfo` application
URL_VALUE="http://$(kubectl -n istio-system get svc istio-ingressgateway -o jsonpath='{.spec.clusterIP}'):80/productpage"
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/istio/quickstart/fortio.yaml | kubectl apply -f -
```

??? info "Look inside fortio.yaml"
    ```yaml
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
            command: [ 'fortio', 'load', '-t', '6000s', '-qps', "16", '-json', '/shared/fortiooutput.json', '-H', 'Host: bookinfo.example.com', "$(URL)" ]
            env:
            - name: URL
              value: URL_VALUE
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
Please follow step 3 of the [quick start tutorial](../quick-start.md).

## 4. Launch experiment
```shell
kubectl apply -f $ITER8/samples/istio/conformance/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: conformance-exp
    spec:
      # target identifies the service under experimentation using its fully qualified name
      target: bookinfo-iter8/productpage
      strategy:
        # this experiment will perform a Conformance test
        testingPattern: Conformance
      criteria:
        objectives: # used for validating versions
        - metric: iter8-istio/mean-latency
          upperLimit: 100
        - metric: iter8-istio/error-rate
          upperLimit: "0.01"
        requestCount: iter8-istio/request-count
      duration: # product of fields determines length of the experiment
        intervalSeconds: 10
        iterationsPerLoop: 5
      versionInfo:
        # information about the app versions used in this experiment
        baseline:
          name: productpage-v1
          variables:
          - name: namespace # used by final action if this version is the winner
            value: bookinfo-iter8
    ```

## 5. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/istio/conformance/fortio.yaml
kubectl delete -f $ITER8/samples/istio/conformance/experiment.yaml
kubectl delete ns bookinfo-iter8
```
