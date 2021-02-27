---
template: overrides/main.html
---

# Traffic Mirroring

!!! abstract "iter8 experiment"
    **iter8** defines a Kubernetes CRD called **experiment** to automate metrics-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps.
    
??? warning "Before you begin"
    **Kubernetes cluster:** Ensure that you have Kubernetes cluster with iter8 and Knative installed. You can do this by following Steps 1, 2, and 3 of [the quick start tutorial for Knative](/getting-started/quick-start/with-knative/).

    **Cleanup:** If you ran an iter8 tutorial earlier, run the associated cleanup step. For example, [Step 8](/getting-started/quick-start/with-knative/#8-cleanup) is the cleanup step for the iter8-Knative quick start tutorial.

    **ITER8:** Ensure that `ITER8` environment variable is set to the root directory of your cloned iter8 repo. See [Step 2 of the quick start tutorial for Knative](/getting-started/quick-start/with-knative/#2-clone-repo) for example.

    **[Helm v3](https://helm.sh/) and [iter8ctl](/getting-started/install/#step-4-install-iter8ctl):** This tutorial uses Helm v3 and iter8ctl. Install if needed.

## 1. Create live and mirrored Knative revisions
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


## 3. Deploy curl with sidecar
```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.1 sh -
istio-1.8.1/bin/istioctl kube-inject -f $ITER8/samples/knative/mirroring/curl.yaml | kubectl create -f -
cd $ITER8
```

??? info "Look inside curl.yaml"
    ```yaml linenums="1"
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: curl
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: curl
      template:
        metadata:
          labels:
            app: curl
        spec:
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
            imagePullPolicy: IfNotPresent
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
    Periodically describe the experiment.

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
    1. You created a Knative service using `helm install` subcommand and upgraded the service to have both `baseline` and `candidate` versions (revisions) using `helm upgrade --install` subcommand. The ksvc is created in the `default` namespace. Helm release information is located in the `iter8-system` namespace specified by the `--namespace=iter8-system` flag.

    2. You created a load generator that sends requests to the Knative service. At this point, 100% of requests are sent to the baseline and 0% to the candidate.

    3. You created an iter8 experiment with the above Knative service as the `target` of the experiment. In each iteration, iter8 observed the `mean-latency`, `95th-percentile-tail-latency`, and `error-rate` metrics for the revisions (collected by Prometheus), ensured that the candidate satisfied all objectives specified in `experiment.yaml`, and progressively shifted traffic from baseline to candidate. You restricted the maximum weight (traffic percentage) of candidate during iterations at 75% and maximum increment allowed during a single iteration to 20% using the `spec.strategy.weights` stanza.

    4. At the end of the experiment, iter8 identified the candidate as the `winner` since it passed all objectives. iter8 decided to promote the candidate (roll forward) by using a `helm upgrade --install` command. Had the candidate failed to satisfy objectives, iter8 would have promoted the baseline (rolled back) instead.


