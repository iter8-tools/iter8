---
template: overrides/main.html
---

# Mirroring / Shadowing / Dark Launch

> Traffic mirroring, also called shadowing, enables dark-launch of app versions, and performing experiments with dark versions with zero-impact on production. Mirrored traffic is a replica of the real (production) traffic sent to dark versions. Metrics are collected and evaluated for the dark versions, but their responses are ignored.

Perform a `conformance` experiment on a dark version with mirrored traffic using the following.

1. A **Knative sample app** with live and dark versions.
2. **Istio virtual services** which send all requests to the live version, mirrors 40% of requests, and sends the mirrored requests to the dark version.
3. A **curl-based traffic generator** which simulates user requests.
4. An **Iter8 `conformance` experiment** which verifies that the dark version satisfies mean latency, 95th percentile tail latency, and error rate objectives.
    
??? warning "Before you begin"
    **Kubernetes cluster with Iter8, Knative and Istio:** Ensure that you have Kubernetes cluster with Iter8 and Knative installed, and ensure that Knative uses the Istio networking layer. You can do this by following Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/), and selecting `Istio` during Step 3.

    **Cleanup:** If you ran an Iter8 tutorial earlier, run the cleanup (last) step associated with it.

    **ITER8:** Ensure that you have cloned the Iter8 GitHub repo, and set the `ITER8` environment variable in your terminal to the root of the cloned repo. See [Step 2 of the quick start tutorial](/getting-started/quick-start/with-knative/#2-clone-repo) for example.

## 1. Create live and dark versions
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

## 2. Create Istio virtual services
```shell
kubectl apply -f $ITER8/samples/knative/mirroring/routing-rules.yaml
```

??? info "Look inside routing-rules.yaml"
    ```yaml linenums="1"
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    metadata:
      name: customdomain-mirroring
    spec:
      gateways:
      - mesh
      - knative-serving/knative-ingress-gateway
      hosts:
      - customdomain.com
      http:
      - rewrite:
          authority: customdomain.com
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
      name: customdomain-routing
    spec:
      gateways:
      - knative-serving/knative-local-gateway
      hosts:
      - "*"
      http:
      - match:
        - authority:
            prefix: customdomain.com-shadow
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
            prefix: customdomain.com
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


## 3. Generate traffic

```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.1 sh -
istio-1.8.1/bin/istioctl kube-inject -f $ITER8/samples/knative/mirroring/curl.yaml | kubectl create -f -
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
          - name: curl
            image: tutum/curl
            command: 
            - /bin/sh
            - -c
            - |
              sleep 10.0
              while true; do
              curl -sS customdomain.com
              sleep 0.5
              done
          restartPolicy: Never
    ```

## 4. Create experiment
```shell
kubectl wait --for=condition=Ready ksvc/sample-app
kubectl wait --for=condition=Available deploy/curl
kubectl apply -f $ITER8/samples/knative/mirroring/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: mirroring
    spec:
      target: default/sample-app
      strategy:
        testingPattern: Conformance
        actions:
          start:
          - library: knative
            task: init-experiment
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
        iterationsPerLoop: 7
      versionInfo:
        # information about app version used in this experiment
        baseline:
          name: current
          variables:
          - name: revision
            value: sample-app-v2
    ```

## 5. Observe experiment
You can observe the experiment in realtime. Open two terminals and follow instructions in the two tabs below.

=== "iter8ctl"
    ```shell
    while clear; do
    kubectl get experiment mirroring -o yaml | iter8ctl describe -f -
    sleep 4
    done
    ```

    ??? info "iter8ctl output"
        iter8ctl output will be similar to the following.

        ```shell
        ****** Overview ******
        Experiment name: mirroring
        Experiment namespace: default
        Target: default/sample-app
        Testing pattern: Conformance
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Completed
        Number of completed iterations: 7

        ****** Winner Assessment ******
        Winning version: not found
        Recommended baseline: current

        ****** Objective Assessment ******
        +--------------------------------+---------+
        |           OBJECTIVE            | CURRENT |
        +--------------------------------+---------+
        | mean-latency <= 50.000         | true    |
        +--------------------------------+---------+
        | 95th-percentile-tail-latency   | true    |
        | <= 100.000                     |         |
        +--------------------------------+---------+
        | error-rate <= 0.010            | true    |
        +--------------------------------+---------+

        ****** Metrics Assessment ******
        +--------------------------------+---------+
        |             METRIC             | CURRENT |
        +--------------------------------+---------+
        | request-count                  | 136.084 |
        +--------------------------------+---------+
        | mean-latency (milliseconds)    |   0.879 |
        +--------------------------------+---------+
        | 95th-percentile-tail-latency   |   4.835 |
        | (milliseconds)                 |         |
        +--------------------------------+---------+
        | error-rate                     |   0.000 |
        +--------------------------------+---------+
        ```

        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.   

=== "kubectl get experiment"
    ```shell
    kubectl get experiment mirroring --watch
    ```

    ??? info "kubectl get experiment output"
        kubectl output will be similar to the following.

        ```shell
        NAME        TYPE          TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
        mirroring   Conformance   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
        mirroring   Conformance   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
        mirroring   Conformance   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
        mirroring   Conformance   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
        mirroring   Conformance   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
        ```

        When the experiment completes (in ~ 2 mins), you will see the stage change from `Running` to `Completed`.

## 6. Cleanup

```shell
kubectl delete -f $ITER8/samples/knative/mirroring/experiment.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/curl.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/routing-rules.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/service.yaml
```

??? info "Understanding what happened"
    1. You configured a Knative service with two versions of your app. In the `service.yaml` manifest, you specified that the live version, `sample-app-v1`, should receive 100% of the production traffic and the dark version, `sample-app-v2`, should receive 0% of the production traffic.

    2. You used `customdomain.com` as the HTTP host in this tutorial. In your production cluster, use domain(s) that you own in the setup of the virtual services.

    3. You set up Istio virtual services which mapped the Knative revisions to this custom domain. The virtual services specified the following routing rules: all HTTP requests with their `Host` header or `:authority` pseudo-header set to `customdomain.com` would be be sent to `sample-app-v1`. 40% of these requests would be mirrored and sent to `sample-app-v2` and responses from `sample-app-v2` would be ignored.

    4. You generated traffic for `customdomain.com` using a `curl`-based job. You injected Istio sidecar injected into it to simulate traffic generation from within the cluster. The sidecar was needed in order to correctly route traffic. 

    5. You used Istio version 1.8.1 to inject the sidecar. This version of Istio corresponds to the one installed in [Step 3 of the quick start tutorial](http://localhost:8000/getting-started/quick-start/with-knative/#3-install-knative-and-iter8). If you have a different version of Istio installed in your cluster, change the Istio version during sidecar injection appropriately.
    
    6. You can also curl the Knative service from outside the cluster. See [here](https://knative.dev/docs/serving/samples/knative-routing-go/#access-the-services) for a related example where the Knative service and Istio virtual service setup is similar to this tutorial.

    7. You created an Iter8 `conformance` experiment to evaluate the dark version. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics for the dark version collected by Prometheus, and verified that the dark version satisfied all the objectives specified in `experiment.yaml`.
