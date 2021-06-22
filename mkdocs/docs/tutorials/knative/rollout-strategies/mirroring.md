---
template: main.html
---

# Traffic Mirroring (Shadowing)

!!! tip "Scenario: Dark launch with traffic mirroring (shadowing)"

    [Traffic mirroring or shadowing](../../../concepts/buildingblocks/#traffic-mirroring-shadowing) enables experimenting with a dark launched version with zero-impact on end-users. Mirrored traffic is a replica of the real user requests that is routed to the dark version. Metrics are collected and evaluated for the dark version, but responses from the dark version are ignored.
    
    In this tutorial, you will use mirror traffic to a dark launched version as shown below.

    ![Mirroring](../../../../images/mirroring.png)
    
## 1. Setup with Istio
* Setup your K8s cluster with Knative and Iter8 as described [here](../../../../getting-started/quick-start/knative/platform-setup/).
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

> Knative with Istio is required in this tutorial. During this setup, choose [Istio](../../../../getting-started/quick-start/knative/platform-setup/#3-install-knative-iter8-and-telemetry) as the networking layer for Knative.

## 2. Create app with live and dark versions
```shell
kubectl apply -f $ITER8/samples/knative/mirroring/service.yaml
```

??? info "Look inside service.yaml"
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
    ---
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
      - revisionName: sample-app-v1
        percent: 100
      - latestRevision: true
        percent: 0
    ```

## 3. Create Istio virtual service to mirror traffic
```shell
kubectl apply -f $ITER8/samples/knative/mirroring/routing-rules.yaml
```

??? info "Look inside routing-rules.yaml"
    ```yaml linenums="1"
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: example-mirroring
    spec:
      gateways:
      - mesh
      - knative-serving/knative-ingress-gateway
      hosts:
      - example.com
      http:
      - rewrite:
          authority: example.com
        route:
        - destination:
            host: knative-local-gateway.istio-system.svc.cluster.local
        mirror:
          host: knative-local-gateway.istio-system.svc.cluster.local
        mirrorPercent: 40
    ---
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: example-routing
    spec:
      gateways:
      - knative-serving/knative-local-gateway
      hosts:
      - "*"
      http:
      - match:
        - authority:
            prefix: example.com-shadow
        route:
        - destination:
            host: sample-app-v2.default.svc.cluster.local
            port:
              number: 80
          headers:
            request:
              set:
                Knative-Serving-Namespace: default
                Knative-Serving-Revision: sample-app-v2
      - match:
        - authority:
            prefix: example.com
        route:
        - destination:
            host: sample-app-v1.default.svc.cluster.local
            port:
              number: 80
          headers:
            request:
              set:
                Knative-Serving-Namespace: default
                Knative-Serving-Revision: sample-app-v1
    ```

## 4. Generate requests

```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.2 sh -
istio-1.8.2/bin/istioctl kube-inject -f $ITER8/samples/knative/mirroring/curl.yaml | kubectl create -f -
cd $ITER8
```

??? info "Look inside curl.yaml"
    ```yaml linenums="1"
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: curl
    spec:
      template:
        spec:
          activeDeadlineSeconds: 6000
          containers:
          - name: curl
            image: tutum/curl
            command: 
            - /bin/sh
            - -c
            - |
              sleep 10.0
              while true; do
              curl -sS example.com
              sleep 0.5
              done
          restartPolicy: Never
    ```

## 5. Create Iter8 experiment
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
kubectl apply -f $ITER8/samples/knative/mirroring/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: mirroring
    spec:
      target: default/sample-app
      strategy:
        testingPattern: Conformance
        actions:
          start:
          - task: knative/init-experiment
      criteria:
        # mean latency of version should be under 50 milliseconds
        # 95th percentile latency should be under 100 milliseconds
        # error rate should be under 1%
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
        # information about version used in this experiment
        baseline:
          name: current
          variables:
          - name: revision
            value: sample-app-v2
    ```

## 6. Observe experiment
Follow [Step 6 of the quick start tutorial for Knative](../../../../getting-started/quick-start/knative/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`mirroring`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/mirroring/curl.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/experiment.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/routing-rules.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/service.yaml
```
