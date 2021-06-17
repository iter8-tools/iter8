---
template: main.html
---

# User Segmentation

!!! tip "Scenario: SLO validation with user segmentation"
    [User segmentation](../../../concepts/buildingblocks/#traffic-engineering) is the ability to carve out a specific segment of users for an experiment, leaving the rest of the users unaffected by the experiment.

    In this tutorial, we will segment users into two groups: those from Wakanda, and others. Users from Wakanda will participate in the experiment: specifically, requests originating in Wakanda may be routed to baseline or candidate; requests that are originating from Wakanda will not participate in the experiment and are routed only to the baseline The experiment is depicted below.

    ![User segmentation](../../images/request-routing.png)    

???+ warning "Before you begin... "

    This tutorial is available for the following K8s stacks.

    [Knative](#before-you-begin){ .md-button }

    Please choose the same K8s stack consistently throughout this tutorial. If you wish to switch K8s stacks between tutorials, start from a clean K8s cluster, so that your cluster is correctly setup.

## Steps 1 to 3
    
Please follow steps 1 through 3 of the [quick start tutorial](../../../getting-started/quick-start/#1-create-kubernetes-cluster).

## 4. Create versions
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


## 5. Create routing rule
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

## 6. Generate traffic
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

## 7. Create Iter8 experiment
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
      name: request-routing
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

## 8. Observe experiment
Please follow [Step 8 of the quick start tutorial](../../../getting-started/quick-start/#8-observe-experiment) to observe the experiment in realtime. Note that the experiment in this tutorial uses a different name from the quick start one. Replace the experiment name `quickstart-exp` with `request-routing` in your commands. You can also observe traffic by suitably modifying the commands for observing traffic.

???+ info "Understanding what happened"
    1. You configured two Knative services corresponding to two versions of your app in `services.yaml`.

    2. You used `example.com` as the HTTP host in this tutorial.
        - **Note:** In your production cluster, use domain(s) that you own in the setup of the virtual service.

    3. You set up an Istio virtual service which mapped the Knative services to this custom domain. The virtual service specified the following routing rules: all HTTP requests to `example.com` with their Host header or :authority pseudo-header **not** set to `wakanda` would be routed to the `baseline`; those with `wakanda` Host header or :authority pseudo-header may be routed to `baseline` and `candidate`.
    
    4. The percentage of `wakandan` requests sent to `candidate` is 0% at the beginning of the experiment.

    5. You generated traffic for `example.com` using a `curl`-job with two `curl`-containers to simulate user requests. You injected Istio sidecar injected into it to simulate traffic generation from within the cluster. The sidecar was needed in order to correctly route traffic. One of the `curl`-containers sets the `country` header field to `wakanda`, and the other to `gondor`.
        - **Note:** You used Istio version 1.8.2 to inject the sidecar. This version of Istio corresponds to the one installed in [Step 3 of the quick start tutorial](http://localhost:8000/getting-started/quick-start/with-knative/#3-install-knative-and-iter8). If you have a different version of Istio installed in your cluster, change the Istio version during sidecar injection appropriately.
    
    6. You created an Iter8 `Canary` experiment with `Progressive` deployment pattern to evaluate the `candidate`. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, and verified that the `candidate` version satisfied all the `objectives` specified in the experiment. It progressively increased the proportion of traffic with `country: wakanda` header that is routed to the `candidate`.

## 9. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/user-segmentation/experiment.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/curl.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/user-segmentation/services.yaml
```