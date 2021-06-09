---
template: main.html
---

# Traffic Mirroring

!!! tip "Scenario: SLO validation for a dark launched version with mirrored traffic"

    [Traffic mirroring or shadowing](../../../concepts/buildingblocks/#traffic-engineering) enables experimenting with a dark launched version with zero-impact on end-users. Mirrored traffic is a replica of the real user requests that is routed to the dark version. Metrics are collected and evaluated for the dark version, but responses from the dark version are ignored.
    
    In this tutorial, you will use mirror traffic to a dark launched version as depicted below.

    ![Mirroring](../../images/mirroring.png)

    
???+ warning "Before you begin... "

    This tutorial is available for the following K8s stacks.

    [Knative](#before-you-begin){ .md-button }

    Please choose the same K8s stack consistently throughout this tutorial. If you wish to switch K8s stacks between tutorials, start from a clean K8s cluster, so that your cluster is correctly setup.
    
## Steps 1 to 3
    
Please follow steps 1 through 3 of the [quick start tutorial](../../../getting-started/quick-start/#1-create-kubernetes-cluster).

## 4. Create app with live and dark versions
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

## 5. Create Istio virtual services
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


## 6. Generate requests

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

## 7. Create Iter8 experiment
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

## 8. Observe experiment
Please follow [Step 8 of the quick start tutorial](../../../getting-started/quick-start/#8-observe-experiment) to observe the experiment in realtime. Note that the experiment in this tutorial uses a different name from the quick start one. Replace the experiment name `quickstart-exp` with `mirroring` in your commands. You can also observe traffic by suitably modifying the commands for observing traffic.

???+ info "Understanding what happened"
    1. You configured a Knative service with two versions of your app. In the `service.yaml` manifest, you specified that the live version, `sample-app-v1`, should receive 100% of the production traffic and the dark version, `sample-app-v2`, should receive 0% of the production traffic.

    2. You used `example.com` as the HTTP host in this tutorial.
        - **Note:** In your production cluster, use domain(s) that you own in the setup of the virtual services.

    3. You set up Istio virtual services which mapped the Knative revisions to the custom domain. The virtual services specified the following routing rules: all HTTP requests with their `Host` header or `:authority` pseudo-header set to `example.com` would be sent to `sample-app-v1`. 40% of these requests would be mirrored and sent to `sample-app-v2` and responses from `sample-app-v2` would be ignored.

    4. You generated traffic for `example.com` using a `curl`-based job. You injected Istio sidecar injected into it to simulate traffic generation from within the cluster. The sidecar was needed in order to correctly route traffic. 
        - **Note:** You used Istio version 1.8.2 to inject the sidecar. This version of Istio corresponds to the one installed in [Step 3 of the quick start tutorial](http://localhost:8000/getting-started/quick-start/with-knative/#3-install-knative-and-iter8). If you have a different version of Istio installed in your cluster, change the Istio version during sidecar injection appropriately.
    
    5. You created an Iter8 `Conformance` experiment to evaluate the dark version. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics for the dark version collected by Prometheus, and verified that the dark version satisfied all the objectives specified in `experiment.yaml`.

## 9. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/mirroring/curl.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/experiment.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/routing-rules.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/service.yaml
```
