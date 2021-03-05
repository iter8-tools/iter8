---
template: overrides/main.html
---

# Request Routing

> Request routing is the ability to route requests dynamically to different versions of the app based on attributes such as identity of the user, URI, or request origin. Use request routing in `Canary` experiments to specify the segment of the traffic that will participate in the experiment; requests within the segment may be routed to baseline or candidate; requests not in this segment will be routed to the baseline only.

Perform a `Progressive Canary` experiment with request routing using the following:

1. An app with baseline and candidate versions implemented as **Knative services**.
2. An  **Istio virtual service** which routes requests based an HTTP header called `country`. All requests are routed to the baseline, except those with their `country` header field set to `wakanda`; these may be routed to the baseline or candidate.
3. A **curl-based traffic generator** which simulates user requests, with `country` header field set to `wakanda` or `gondor`.
4. An **Iter8 experiment** which verifies that the candidate satisfies mean latency, 95th percentile tail latency, and error rate objectives, and progressively increases the proportion of traffic with `country: wakanda` header that is routed to the candidate.

??? warning "Before you begin"
    **Kubernetes cluster with Iter8, Knative and Istio:** Ensure that you have Kubernetes cluster with Iter8 and Knative installed, and ensure that Knative uses the Istio networking layer. You can do this by following Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/), and selecting `Istio` during Step 3.

    **Cleanup:** If you ran an Iter8 tutorial earlier, run the cleanup (last) step associated with it.

    **ITER8:** Ensure that you have cloned the Iter8 GitHub repo, and set the `ITER8` environment variable in your terminal to the root of the cloned repo. See [Step 2 of the quick start tutorial](/getting-started/quick-start/with-knative/#2-clone-repo) for example.


## 1. Create app versions
```shell
kubectl apply -f $ITER8/samples/knative/requestrouting/services.yaml
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


## 2. Create Istio virtual service
```shell
kubectl apply -f $ITER8/samples/knative/requestrouting/routing-rule.yaml
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
      - customdomain.com
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

## 3. Generate traffic
```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.2 sh -
istio-1.8.2/bin/istioctl kube-inject -f $ITER8/samples/knative/requestrouting/curl.yaml | kubectl create -f -
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
          activeDeadlineSeconds: 600
          containers:
          - name: curl-from-gondor
            image: tutum/curl
            command: 
            - /bin/sh
            - -c
            - |
              while true; do
              curl -sS customdomain.com -H "country: gondor"
              sleep 1.0
              done
          - name: curl-from-wakanda
            image: tutum/curl
            command: 
            - /bin/sh
            - -c
            - |
              while true; do
              curl -sS customdomain.com -H "country: wakanda"
              sleep 0.25
              done
          restartPolicy: Never
    ```

## 4. Create experiment
```shell
kubectl wait --for=condition=Ready ksvc/sample-app-v1
kubectl wait --for=condition=Ready ksvc/sample-app-v2
kubectl apply -f $ITER8/samples/knative/requestrouting/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha1
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
        - metric: mean-latency
          upperLimit: 50
        - metric: 95th-percentile-tail-latency
          upperLimit: 100
        - metric: error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        # information about app versions used in this experiment
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
            fieldPath: /spec/http/0/route/0/weight
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
            fieldPath: /spec/http/0/route/1/weight
    ```

## 5. Observe experiment
Observe the experiment in realtime. Paste commands from the tabs below in separate terminals.

=== "iter8ctl"
    ```shell
    while clear; do
    kubectl get experiment request-routing -o yaml | iter8ctl describe -f -
    sleep 4
    done
    ```

    ??? info "iter8ctl output"
        iter8ctl output will be similar to the following.

        ```shell
        ****** Overview ******
        Experiment name: request-routing
        Experiment namespace: default
        Target: default/routing-for-wakanda
        Testing pattern: Canary
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Completed
        Number of completed iterations: 10

        ****** Winner Assessment ******
        App versions in this experiment: [current candidate]
        Winning version: candidate
        Recommended baseline: candidate

        ****** Objective Assessment ******
        +--------------------------------+---------+-----------+
        |           OBJECTIVE            | CURRENT | CANDIDATE |
        +--------------------------------+---------+-----------+
        | mean-latency <= 50.000         | true    | true      |
        +--------------------------------+---------+-----------+
        | 95th-percentile-tail-latency   | true    | true      |
        | <= 100.000                     |         |           |
        +--------------------------------+---------+-----------+
        | error-rate <= 0.010            | true    | true      |
        +--------------------------------+---------+-----------+

        ****** Metrics Assessment ******
        +--------------------------------+---------+-----------+
        |             METRIC             | CURRENT | CANDIDATE |
        +--------------------------------+---------+-----------+
        | request-count                  | 374.500 |   137.107 |
        +--------------------------------+---------+-----------+
        | mean-latency (milliseconds)    |   0.752 |     0.741 |
        +--------------------------------+---------+-----------+
        | 95th-percentile-tail-latency   |   4.792 |     4.750 |
        | (milliseconds)                 |         |           |
        +--------------------------------+---------+-----------+
        | error-rate                     |   0.000 |     0.000 |
        +--------------------------------+---------+-----------+
        ```   

        When the experiment completes (in ~ 2 mins), you will see the stage change from `Running` to `Completed`.

=== "kubectl get experiment"
    ```shell
    kubectl get experiment request-routing --watch
    ```

    ??? info "kubectl get experiment output"
        kubectl output will be similar to the following.

        ```shell
        NAME              TYPE     TARGET                        STAGE     COMPLETED ITERATIONS   MESSAGE
        request-routing   Canary   default/routing-for-wakanda   Running   1                      IterationUpdate: Completed Iteration 1
        request-routing   Canary   default/routing-for-wakanda   Running   2                      IterationUpdate: Completed Iteration 2
        request-routing   Canary   default/routing-for-wakanda   Running   3                      IterationUpdate: Completed Iteration 3
        request-routing   Canary   default/routing-for-wakanda   Running   4                      IterationUpdate: Completed Iteration 4
        request-routing   Canary   default/routing-for-wakanda   Running   5                      IterationUpdate: Completed Iteration 5
        request-routing   Canary   default/routing-for-wakanda   Running   6                      IterationUpdate: Completed Iteration 6
        request-routing   Canary   default/routing-for-wakanda   Running   7                      IterationUpdate: Completed Iteration 7
        request-routing   Canary   default/routing-for-wakanda   Running   8                      IterationUpdate: Completed Iteration 8
        request-routing   Canary   default/routing-for-wakanda   Running   9                      IterationUpdate: Completed Iteration 9
        ```

        When the experiment completes (in ~ 2 mins), you will see the stage change from `Running` to `Completed`.

=== "kubectl get vs"
    ```shell
    kubectl get vs routing-for-wakanda -o json | jq .spec.http[0].route
    ```

    ??? info "kubectl output"
        kubectl output will be similar to the following.

        ```json
        [
          {
            "destination": {
              "host": "sample-app-v1.default.svc.cluster.local"
            },
            "headers": {
              "request": {
                "set": {
                  "Host": "sample-app-v1.default",
                  "Knative-Serving-Namespace": "default",
                  "Knative-Serving-Revision": "sample-app-v1-blue"
                }
              }
            },
            "weight": 15
          },
          {
            "destination": {
              "host": "sample-app-v2.default.svc.cluster.local"
            },
            "headers": {
              "request": {
                "set": {
                  "Host": "sample-app-v2.default",
                  "Knative-Serving-Namespace": "default",
                  "Knative-Serving-Revision": "sample-app-v2-green"
                }
              }
            },
            "weight": 85
          }
        ]
        ```

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/requestrouting/experiment.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/curl.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/routing-rule.yaml
kubectl delete -f $ITER8/samples/knative/requestrouting/services.yaml
```

??? info "Understanding what happened"
    1. You configured two Knative services corresponding to two versions of your app in `services.yaml`.

    2. You used `customdomain.com` as the HTTP host in this tutorial. In your production cluster, use domain(s) that you own in the setup of the virtual services.

    3. You set up an Istio virtual service, `routing-rule.yaml`, which mapped the Knative services to this custom domain. This virtual service determines how to route requests based on the `country` HTTP header. Before the experiment started, requests with their `country` header field set to `wakanda` as well as other requests were routed to the baseline.

    4. You generated traffic for `customdomain.com` using a `curl`-job. You injected Istio sidecar injected into it to simulate traffic generation from within the cluster. The sidecar was needed in order to correctly route traffic. The `curl`-job simulates user requests and sets the `country` header field to one of two values: `wakanda` or `gondor`.

    5. You used Istio version 1.8.2 to inject the sidecar. This version of Istio corresponds to the one installed in [Step 3 of the quick start tutorial](http://localhost:8000/getting-started/quick-start/with-knative/#3-install-knative-and-iter8). If you have a different version of Istio installed in your cluster, change the Istio version during sidecar injection appropriately.
    
    6. You can also curl the Knative services from outside the cluster. See [here](https://knative.dev/docs/serving/samples/knative-routing-go/#access-the-services) for a related example where the Knative service and Istio virtual service setup is similar to this tutorial.

    7. You created an Iter8 `Canary` experiment with `Progressive` deployment pattern to evaluate the candidate version. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics for the dark version collected by Prometheus, and verified that the candidate version satisfied all the objectives specified in `experiment.yaml`. It progressively increased the proportion of traffic with `country: wakanda` header that is routed to the candidate.

    8. You can add finish tasks to this experiment to accomplish version promotion. See examples [here](/getting-started/quick-start/with-knative/), [here](/code-samples/iter8-knative/canary-progressive/) and [here](/code-samples/iter8-knative/canary-fixedsplit/).


