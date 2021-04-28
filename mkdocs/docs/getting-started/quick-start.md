---
template: main.html
---

# Quick Start

!!! tip "Scenario: A/B testing and progressive deployment"
    [A/B testing](../../concepts/buildingblocks/#testing-pattern) enables you to compare two versions of an app/ML model, and select a winner based on a (business) reward metric and objectives (SLOs). In this tutorial, you will:

    1. Perform A/B testing.
    2. Specify `user-engagement` as the reward, and `latency` and `error-rate` based objectives. Iter8 will find a winner by comparing versions in terms of the reward, and by validating versions in terms of the objectives.
    3. The reward metric will by provided by New Relic and metrics used in objectives will be provided by Prometheus.
    4. Combine A/B testing with [progressive deployment](../../concepts/buildingblocks/#deployment-pattern).
    
    Assuming a winner is found, Iter8 will progressively shift the traffic towards the winner and promote it at the end as depicted below.

    ![Canary](../images/quickstart.png)

???+ warning "Before you begin, you will need... "
    1. The [kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
    2. [Go 1.13+](https://golang.org/doc/install).

## 1. Create Kubernetes cluster

Create a local cluster using Minikube or Kind as follows, or use a managed Kubernetes service. Ensure that the cluster has sufficient resources, for example, 8 CPUs and 12GB of memory.

=== "Minikube"

    ```shell
    minikube start --cpus 8 --memory 12288
    ```

=== "Kind"

    ```shell
    kind create cluster
    kubectl cluster-info --context kind-kind
    ```

    ??? info "Ensuring your Kind cluster has sufficient resources"
        Your Kind cluster inherits the CPU and memory resources of its host. If you are using Docker Desktop, you can set its resources as shown below.

        ![Resources](../images/ddresourcepreferences.png)

## 2. Clone Iter8 repo
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
```

## 3. Setup Kubernetes cluster

Choose the K8s stack over which you wish to perform the A/B testing experiment.
=== "Knative"
    Setup Knative, Iter8, a mock New Relic service, and Prometheus add-on within your cluster. 
    
    Knative can work with multiple networking layers. So can Iter8's Knative extension. Choose a networking layer for Knative.

    === "Contour"

        ```shell
        $ITER8/samples/knative/quickstart/platformsetup.sh contour
        ```

    === "Kourier"

        ```shell
        $ITER8/samples/knative/quickstart/platformsetup.sh kourier
        ```

    === "Gloo"
        This step requires Python. This will install `glooctl` binary under `$HOME/.gloo` folder.
        ```shell
        $ITER8/samples/knative/quickstart/platformsetup.sh gloo
        ```

    === "Istio"

        ```shell
        $ITER8/samples/knative/quickstart/platformsetup.sh istio
        ```

=== "Istio"
    Setup Istio, Iter8, and Prometheus add-on within your cluster. 

    ```shell
    $ITER8/samples/istio/quickstart/platformsetup.sh
    ```
    
=== "KFServing"
    Setup KFServing, Iter8, a mock New Relic service, and Prometheus add-on within your cluster.

    ```shell
    $ITER8/samples/kfserving/quickstart/platformsetup.sh
    ```

## 4. Create app versions

Choose the K8s stack over which you wish to perform the A/B testing experiment, and create baseline and candidate versions of your app.
=== "Knative"
    ```shell
    kubectl apply -f $ITER8/samples/knative/quickstart/baseline.yaml
    kubectl apply -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
    ```

    ??? info "Look inside baseline.yaml"
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
        ```

    ??? info "Look inside experimentalservice.yaml"
        ```yaml linenums="1"
        apiVersion: serving.knative.dev/v1
        kind: Service
        metadata:
        name: sample-app # name of the app
        namespace: default # namespace of the app
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
        # initially all traffic goes to sample-app-v1 and none to sample-app-v2
        - tag: current
            revisionName: sample-app-v1
            percent: 100
        - tag: candidate
            latestRevision: true
            percent: 0
        ```

=== "Istio"
    ```shell
    kubectl apply -f $ITER8/samples/knative/quickstart/baseline.yaml
    kubectl apply -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
    ```

## 5. Generate requests
In a production environment, your application would receive requests from end-users. For the purposes of this tutorial, simulate user requests using [Fortio](https://github.com/fortio/fortio) as follows.

```shell
kubectl wait --for=condition=Ready ksvc/sample-app
# URL_VALUE is the URL where your Knative application serves requests
URL_VALUE=$(kubectl get ksvc sample-app -o json | jq .status.address.url)
sed "s+URL_VALUE+${URL_VALUE}+g" $ITER8/samples/knative/quickstart/fortio.yaml | kubectl apply -f -
```

??? info "Look inside fortio.yaml"
    ```yaml linenums="1"
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
            command: ["fortio", "load", "-t", "6000s", "-qps", "16", "-json", "/shared/fortiooutput.json", $(URL)]
            env:
            - name: URL
            value: URL_VALUE
            volumeMounts:
            - name: shared
            mountPath: /shared         
        - name: busybox
            image: busybox:1.28
            command: ['sh', '-c', 'echo busybox is running! && sleep 600']
            volumeMounts:
            - name: shared
            mountPath: /shared       
        restartPolicy: Never
    ```

## 6. Define metrics
Define the Iter8 metrics used in this experiment.

```shell
kubectl apply -f $ITER8/samples/knative/quickstart/metrics.yaml
```

??? info "Look inside metrics.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
    labels:
        creator: iter8
    name: 95th-percentile-tail-latency
    namespace: iter8-knative
    spec:
    description: 95th percentile tail latency
    jqExpression: .data.result[0].value[1] | tonumber
    params:
    - name: query
        value: |
        histogram_quantile(0.95, sum(rate(revision_app_request_latencies_bucket{revision_name='$revision'}[${elapsedTime}s])) by (le))
    provider: prometheus
    sampleSize: request-count
    type: Gauge
    units: milliseconds
    urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
    labels:
        creator: iter8
    name: error-count
    namespace: iter8-knative
    spec:
    description: Number of error responses
    jqExpression: .data.result[0].value[1] | tonumber
    params:
    - name: query
        value: |
        sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',revision_name='$revision'}[${elapsedTime}s])) or on() vector(0)
    provider: prometheus
    type: Counter
    urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
    labels:
        creator: iter8
    name: error-rate
    namespace: iter8-knative
    spec:
    description: Fraction of requests with error responses
    jqExpression: .data.result[0].value[1] | tonumber
    params:
    - name: query
        value: |
        (sum(increase(revision_app_request_latencies_count{response_code_class!='2xx',revision_name='$revision'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[${elapsedTime}s])) or on() vector(0))
    provider: prometheus
    sampleSize: request-count
    type: Gauge
    urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
    labels:
        creator: iter8
    name: mean-latency
    namespace: iter8-knative
    spec:
    description: Mean latency
    jqExpression: .data.result[0].value[1] | tonumber
    params:
    - name: query
        value: |
        (sum(increase(revision_app_request_latencies_sum{revision_name='$revision'}[${elapsedTime}s])) or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[${elapsedTime}s])) or on() vector(0))
    provider: prometheus
    sampleSize: request-count
    type: Gauge
    units: milliseconds
    urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ---
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
    labels:
        creator: iter8
    name: request-count
    namespace: iter8-knative
    spec:
    description: Number of requests
    jqExpression: .data.result[0].value[1] | tonumber
    params:
    - name: query
        value: |
        sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[${elapsedTime}s])) or on() vector(0)
    provider: prometheus
    type: Counter
    urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query
    ```
The `urlTemplate` field in these metrics point to the Prometheus instance that was created in Step 3 above. If you wish to use these metrics in your production/staging/dev/test K8s cluster, change the `urlTemplate` values to match the URL of your Prometheus instance.

## 7. Launch experiment
Launch the Iter8 experiment. Iter8 will orchestrate the canary release of the new version with SLO validation and progressive deployment as specified in the experiment.

```shell
kubectl apply -f $ITER8/samples/knative/quickstart/experiment.yaml
```

??? info "Look inside experiment.yaml"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
    name: quickstart-exp
    spec:
    # target identifies the knative service under experimentation using its fully qualified name
    target: default/sample-app
    strategy:
        # this experiment will perform a canary test
        testingPattern: Canary
        deploymentPattern: Progressive
        actions:
        start: # run the following sequence of tasks at the start of the experiment
        - task: knative/init-experiment
        finish: # run the following sequence of tasks at the end of the experiment
        - task: common/exec # promote the winning version
            with:
            cmd: kubectl
            args:
            - "apply"
            - "-f"
            - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
    criteria:
        requestCount: iter8-knative/request-count
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
        # information about app versions used in this experiment
        baseline:
        name: current
        variables:
        - name: revision
            value: sample-app-v1
        - name: promote
            value: baseline
        candidates:
        - name: candidate
        variables:
        - name: revision
            value: sample-app-v2
        - name: promote
            value: candidate
    ```

The process automated by Iter8 during this experiment is depicted below.

![Iter8 automation](../images/canary-progressive-kubectl-iter8.png)

## 8. Observe experiment
Observe the experiment in realtime. Paste commands from the tabs below in separate terminals.

=== "Metrics-based analysis"
    Install `iter8ctl`. You can change the directory where `iter8ctl` binary is installed by changing `GOBIN` below.
    ```shell
    GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl@v0.1.3
    ```

    Periodically describe the experiment.
    ```shell
    while clear; do
    kubectl get experiment quickstart-exp -o yaml | iter8ctl describe -f -
    sleep 4
    done
    ```
    ??? info "Look inside `iter8ctl` output"
        The `iter8ctl` output will be similar to the following.
        ```shell
        ****** Overview ******
        Experiment name: quickstart-exp
        Experiment namespace: default
        Target: default/sample-app
        Testing pattern: Canary
        Deployment pattern: Progressive

        ****** Progress Summary ******
        Experiment stage: Running
        Number of completed iterations: 3

        ****** Winner Assessment ******
        > If the candidate version satisfies the experiment objectives, then it is the winner.
        > Otherwise, if the baseline version satisfies the experiment objectives, it is the winner.
        > Otherwise, there is no winner.
        App versions in this experiment: [current candidate]
        Winning version: candidate
        Version recommended for promotion: candidate

        ****** Objective Assessment ******
        > Identifies whether or not the experiment objectives are satisfied by the most recently observed metrics values for each version.
        +--------------------------------------------+---------+-----------+
        |                 OBJECTIVE                  | CURRENT | CANDIDATE |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/mean-latency <=              | true    | true      |
        |                                     50.000 |         |           |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/95th-percentile-tail-latency | true    | true      |
        | <= 100.000                                 |         |           |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/error-rate <=                | true    | true      |
        |                                      0.010 |         |           |
        +--------------------------------------------+---------+-----------+

        ****** Metrics Assessment ******
        > Most recently read values of experiment metrics for each version.
        +--------------------------------------------+---------+-----------+
        |                   METRIC                   | CURRENT | CANDIDATE |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/request-count                | 454.523 |    27.412 |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/mean-latency                 |   1.265 |     1.415 |
        | (milliseconds)                             |         |           |
        +--------------------------------------------+---------+-----------+
        | request-count                              | 454.523 |    27.619 |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/95th-percentile-tail-latency |   4.798 |     4.928 |
        | (milliseconds)                             |         |           |
        +--------------------------------------------+---------+-----------+
        | iter8-knative/error-rate                   |   0.000 |     0.000 |
        +--------------------------------------------+---------+-----------+
        ``` 

    As the experiment progresses, you should eventually see that all of the objectives reported as being satisfied by both versions. The candidate is identified as the winner and is recommended for promotion. When the experiment completes (in ~2 mins), you will see the experiment stage change from `Running` to `Completed`.

=== "Experiment progress"

    ```shell
    kubectl get experiment quickstart-exp --watch
    ```

    ??? info "kubectl get experiment output"
        The `kubectl` output will be similar to the following.
        ```shell
        NAME             TYPE     TARGET               STAGE     COMPLETED ITERATIONS   MESSAGE
        quickstart-exp   Canary   default/sample-app   Running   1                      IterationUpdate: Completed Iteration 1
        quickstart-exp   Canary   default/sample-app   Running   2                      IterationUpdate: Completed Iteration 2
        quickstart-exp   Canary   default/sample-app   Running   3                      IterationUpdate: Completed Iteration 3
        quickstart-exp   Canary   default/sample-app   Running   4                      IterationUpdate: Completed Iteration 4
        quickstart-exp   Canary   default/sample-app   Running   5                      IterationUpdate: Completed Iteration 5
        quickstart-exp   Canary   default/sample-app   Running   6                      IterationUpdate: Completed Iteration 6
        quickstart-exp   Canary   default/sample-app   Running   7                      IterationUpdate: Completed Iteration 7
        quickstart-exp   Canary   default/sample-app   Running   8                      IterationUpdate: Completed Iteration 8
        quickstart-exp   Canary   default/sample-app   Running   9                      IterationUpdate: Completed Iteration 9
        ```

    When the experiment completes (in ~ 2 mins), you will see the experiment stage change from `Running` to `Completed`.    

=== "Traffic split"

    ```shell
    kubectl get ksvc sample-app -o json --watch | jq .status.traffic
    ```

    ??? info "kubectl get ksvc output"
        The `kubectl` output will be similar to the following.
        ```shell
        [
        {
            "latestRevision": false,
            "percent": 45,
            "revisionName": "sample-app-v1",
            "tag": "current",
            "url": "http://current-sample-app.default.example.com"
        },
        {
            "latestRevision": true,
            "percent": 55,
            "revisionName": "sample-app-v2",
            "tag": "candidate",
            "url": "http://candidate-sample-app.default.example.com"
        }
        ]
        ```
    As the experiment progresses, you should see traffic progressively shift from `sample-app-v1` to `sample-app-v2`. When the experiment completes, all of the traffic will be sent to the winner, `sample-app-v2`.

???+ info "Understanding what happened"
    1. You created a Knative service with two revisions, `sample-app-v1` (baseline) and `sample-app-v2` (candidate).
    2. You generated requests for the Knative service using a Fortio job. At the start of the experiment, 100% of the requests are sent to the baseline and 0% to the candidate.
    3. You created an Iter8 experiment with canary testing and progressive deployment patterns. In each iteration, Iter8 observed the mean latency, 95th percentile tail-latency, and error-rate metrics collected by Prometheus, verified that the candidate satisfied all objectives, identified the candidate as the winner, progressively shifted traffic from the baseline to the candidate, and eventually promoted the candidate using the `kubectl apply` command embedded within its finish action.
    4. Had the candidate failed to satisfy objectives, then the baseline would have been promoted.

## 9. Cleanup
```shell
kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experiment.yaml
kubectl delete -f $ITER8/samples/knative/quickstart/experimentalservice.yaml
```
