---
template: main.html
---

# User Segmentation

!!! tip "Scenario: SLO validation with user segmentation and builtin metrics"
    [User segmentation](../../../concepts/buildingblocks.md#progressive-traffic-shift-with-user-segmentation) is the ability to carve out a specific segment of users for an experiment, leaving the rest of the users unaffected by the experiment.

    In this tutorial, you will:

    1. Segment users into two groups: those from Wakanda, and others. 
    2. Users from Wakanda will participate in the experiment: specifically, requests originating from Wakanda may be routed to baseline or candidate versions; requests that are originating from outside Wakanda will not participate in the experiment and will be routed to the baseline only. The experiment is shown below.

    ![User segmentation](../../../images/canary-progressive-segmentation.png)

???+ warning "Platform setup"
    Follow [these steps](../platform-setup.md) to install Iter8 and Knative in your K8s cluster.

    **Note:** Knative needs to be installed with the Istio networking layer for this tutorial.

## 1. Create app versions
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


## 2. Create Istio virtual service to segment users
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

## 3. Generate requests
```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.7.0 sh -
istio-1.7.0/bin/istioctl kube-inject -f $ITER8/samples/knative/user-segmentation/curl.yaml | kubectl create -f -
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

## 4. Create Iter8 experiment
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
        actions:
          loop:
          - task: metrics/collect
            with:
              versions:
              - name: sample-app-v1
                url: http://sample-app-v1.default.svc.cluster.local
              - name: sample-app-v2
                url: http://sample-app-v2.default.svc.cluster.local
      criteria:
        # mean latency of version should be under 50 milliseconds
        # 95th percentile latency should be under 100 milliseconds
        # error rate should be under 1%
        objectives: 
        - metric: iter8-system/mean-latency
          upperLimit: 50
        - metric: iter8-system/latency-95th-percentile
          upperLimit: 100
        - metric: iter8-system/error-count
          upperLimit: "0.01"
      duration:
        maxLoops: 10
        intervalSeconds: 2
        iterationsPerLoop: 1
      versionInfo:
        # information about app versions used in this experiment
        baseline:
          name: sample-app-v1
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-for-wakanda
            namespace: default
            fieldPath: .spec.http[0].route[0].weight
        candidates:
        - name: sample-app-v2
          weightObjRef:
            apiVersion: networking.istio.io/v1alpha3
            kind: VirtualService
            name: routing-for-wakanda
            namespace: default
            fieldPath: .spec.http[0].route[1].weight
    ```

## 5. Observe experiment
Follow [these steps](../../../getting-started/first-experiment.md#3-observe-experiment) to observe your experiment.

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/user-segmentation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/curl.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/services.yaml
```