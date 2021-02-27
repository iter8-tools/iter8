---
template: overrides/main.html
---

# Mirroring

## 1. Create live and mirrored Knative services
```shell
kubectl apply -f $ITER8/samples/knative/mirroring/service.yaml
kubectl wait --for=condition=Ready ksvc/sample-app
```

## 2. Create Istio virtual service
```shell
kubectl apply -f $ITER8/samples/knative/mirroring/routing-rules.yaml
```
<!-- 
# 2.1 Use minikube set up with Istio for this test

# 2.2 Inject sleep deployment so you can curl

TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.1 sh -
cat <<EOF | istio-1.8.1/bin/istioctl kube-inject -f - | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sleep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sleep
  template:
    metadata:
      labels:
        app: sleep
    spec:
      containers:
      - name: sleep
        image: tutum/curl
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: IfNotPresent
EOF
cd $ITER8

# 2.3 Send some traffic
export SLEEP_POD=$(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name})
kubectl exec "${SLEEP_POD}" -c sleep -- curl -sS customdomain.com
# repeat the above a few times!

# 2.4 Minikube tunnel
minikube tunnel --cleanup # in a separate terminal

# 2.5 Get GATEWAY_IP
INGRESSGATEWAY=istio-ingressgateway

export GATEWAY_IP=`kubectl get svc $INGRESSGATEWAY --namespace istio-system \
    --output jsonpath="{.status.loadBalancer.ingress[*]['ip']}"`

# 2.6 curl
curl http://${GATEWAY_IP} --header "Host: example.com"

# 2.7 Quit Minikube tunnel

# 2.8 Cleanup sleep pod
kubectl delete deploy/sleep
 -->


## 3. Generate traffic with fortio
```shell
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.8.1 sh -
istio-1.8.1/bin/istioctl kube-inject -f $ITER8/samples/knative/mirroring/fortio.yaml | kubectl apply -f -
cd $ITER8
```

## 4. Create experiment
```shell
kubectl apply -f $ITER8/samples/knative/mirroring/experiment.yaml
```

## 5. Observe experiment
You can observe the experiment in realtime. Open two terminals and follow instructions in the two tabs below.

=== "iter8ctl"
    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment mirroring -o yaml | iter8ctl describe -f -
    sleep 2
    done
    ```

    ??? info "iter8ctl output"
        iter8ctl output will be similar to the following.
        ```shell
        ****** Overview ******
        Experiment name: canary-progressive
        Experiment namespace: default
        Target: default/sample-app
        Testing pattern: Canary
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Completed
        Number of completed iterations: 7

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
        | mean-latency (milliseconds)    |   1.201 |     1.322 |
        +--------------------------------+---------+-----------+
        | 95th-percentile-tail-latency   |   4.776 |     4.750 |
        | (milliseconds)                 |         |           |
        +--------------------------------+---------+-----------+
        | error-rate                     |   0.000 |     0.000 |
        +--------------------------------+---------+-----------+
        | request-count                  | 448.800 |    89.352 |
        +--------------------------------+---------+-----------+
        ```
        When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.   

=== "kubectl get experiment"
    ```shell
    kubectl get experiment mirroring --watch
    ```

    ??? info "kubectl get experiment output"
        kubectl output will be similar to the following.
        ```shell
        NAME                 TYPE     TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
        canary-progressive   Canary   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
        canary-progressive   Canary   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
        canary-progressive   Canary   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
        canary-progressive   Canary   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
        canary-progressive   Canary   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
        canary-progressive   Canary   default/sample-app   Running   6                      IterationUpdate: Completed Iteration 6
        canary-progressive   Canary   default/sample-app   Finishing   7                      TerminalHandlerLaunched: Finish handler 'finish' launched
        canary-progressive   Canary   default/sample-app   Completed   7                      ExperimentCompleted: Experiment completed successfully
        ```
        When the experiment completes (in ~ 4 mins), you will see the experiment stage change from `Running` to `Completed`.    

## 6. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/canaryprogressive/experiment.yaml
kubectl delete -f $ITER8/samples/knative/canaryprogressive/fortio.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/mesh-routing.yaml
kubectl delete -f $ITER8/samples/knative/mirroring/services.yaml
```

