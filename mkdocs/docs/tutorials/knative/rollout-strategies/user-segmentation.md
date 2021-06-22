---
template: main.html
---

# User Segmentation

!!! tip "Scenario: SLO validation with user segmentation"
    [User segmentation](../../../../concepts/buildingblocks/#user-segmentation_1) is the ability to carve out a specific segment of users for an experiment, leaving the rest of the users unaffected by the experiment.

    In this tutorial, you will:

    1. Segment users into two groups: those from Wakanda, and others. 
    2. Users from Wakanda will participate in the experiment: specifically, requests originating in Wakanda may be routed to baseline or candidate versions; requests that are originating from Wakanda will not participate in the experiment and will be routed to the baseline only. The experiment is shown below.

    ![User segmentation](../../../../images/canary-progressive-segmentation.png)

## 1. Setup with Istio
* Setup your K8s cluster with Knative and Iter8 as described [here](../../../../getting-started/quick-start/knative/platform-setup/).
* Ensure that the `ITER8` environment variable is set to the root of your local Iter8 repo.

> Knative with Istio is required in this tutorial. During this setup, choose [Istio](../../../../getting-started/quick-start/knative/platform-setup/#3-install-knative-iter8-and-telemetry) as the networking layer for Knative.


## 2. Create versions of your application
```shell
kubectl apply -f $ITER8/samples/knative/user-segmentation/services.yaml
```

??? info "Look inside services.yaml"
    ```yaml linenums="1"
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: sample-app-v1
      namespace: default
    spec:
      template:
        metadata:
          name: sample-app-v1-blue
          annotations:
            autoscaling.knative.dev/scaleToZeroPodRetentionPeriod: "10m"
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
      name: sample-app-v2
      namespace: default
    spec:
      template:
        metadata:
          name: sample-app-v2-green
          annotations:
            autoscaling.knative.dev/scaleToZeroPodRetentionPeriod: "10m"
        spec:
          containers:
          - image: gcr.io/knative-samples/knative-route-demo:green 
            env:
            - name: T_VERSION
              value: "green"
    ```


## 3. Create Istio virtual service to segment users
```shell
kubectl apply -f $ITER8/samples/knative/user-segmentation/routing-rule.yaml
```

??? info "Look inside routing-rule.yaml"
    ```yaml linenums="1"
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: routing-for-wakanda
    spec:
      gateways:
      - mesh
      - knative-serving/knative-ingress-gateway
      - knative-serving/knative-local-gateway
      hosts:
      - example.com
      http:
      - match:
        - headers:
            country:
              exact: wakanda
        route:
        - destination:
            host: sample-app-v1.default.svc.cluster.local
          headers:
            request:
              set:
                Knative-Serving-Namespace: default
                Knative-Serving-Revision: sample-app-v1-blue
                Host: sample-app-v1.default
          weight: 100
        - destination:
            host: sample-app-v2.default.svc.cluster.local
          headers:
            request:
              set:
                Knative-Serving-Namespace: default
                Knative-Serving-Revision: sample-app-v2-green
                Host: sample-app-v2.default
          weight: 0
      - route:
        - destination:
            host: sample-app-v1.default.svc.cluster.local
          headers:
            request:
              set:
                Knative-Serving-Namespace: default
                Knative-Serving-Revision: sample-app-v1-blue
                Host: sample-app-v1.default
    ```

## 4. Generate requests
```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.2 sh -
istio-1.8.2/bin/istioctl kube-inject -f $ITER8/samples/knative/user-segmentation/curl.yaml | kubectl create -f -
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
          - name: curl-from-gondor
            image: tutum/curl
            command: 
            - /bin/sh
            - -c
            - |
              while true; do
              curl -sS example.com -H "country: gondor"
              sleep 1.0
              done
          - name: curl-from-wakanda
            image: tutum/curl
            command: 
            - /bin/sh
            - -c
            - |
              while true; do
              curl -sS example.com -H "country: wakanda"
              sleep 0.25
              done
          restartPolicy: Never
    ```

## 5. Create Iter8 experiment
```shell
kubectl wait --for=condition=Ready ksvc/sample-app-v1
kubectl wait --for=condition=Ready ksvc/sample-app-v2
kubectl apply -f $ITER8/samples/knative/user-segmentation/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: user-segmentation-exp
    spec:
      # this experiment uses the fully-qualified name of the Istio virtual service as the target
      target: default/routing-for-wakanda
      strategy:
        # this experiment will perform a canary test
        testingPattern: Canary
        deploymentPattern: Progressive
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
        # information about versions used in this experiment
        baseline:
          name: current
          variables:
          - name: revision
            value: sample-app-v1-blue
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-for-wakanda
            namespace: default
            fieldPath: .spec.http[0].route[0].weight
        candidates:
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2-green
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-for-wakanda
            namespace: default
            fieldPath: .spec.http[0].route[1].weight
    ```

## 6. Observe experiment
Follow [Step 6 of the quick start tutorial for Knative](../../../../getting-started/quick-start/knative/tutorial/#6-understand-the-experiment) to observe metrics, traffic and progress of the experiment. Ensure that you use the correct experiment name (`user-segmentation-exp`) in your `iter8ctl` and `kubectl` commands.

## 7. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/user-segmentation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/curl.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/services.yaml
```