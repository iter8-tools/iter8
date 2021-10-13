---
template: main.html
---

# Your First Experiment

!!! tip "Validate service level objectives (SLOs) for an app"
    **Problem**: You have a Kubernetes app. You want to verify that it satisfies latency and error rate SLOs.

    **Solution**: In this tutorial, you will launch a Kubernetes app along with an Iter8 experiment. Iter8 will [validate that the app satisfies latency and error-based objectives (SLOs)](../concepts/buildingblocks.md#slo-validation) using [built-in metrics](../metrics/builtin.md). During this validation, Iter8 will generate HTTP GET requests for the app.

    ![SLO Validation](../images/slo-validation.png)

    **Preview**:
    <script id="asciicast-439859" src="https://asciinema.org/a/439859.js" async></script>

???+ warning "Setup Kubernetes cluster and local environment"
    1. Setup [Kubernetes cluster](setup-for-tutorials.md#local-kubernetes-cluster)
    2. [Install Iter8 in Kubernetes cluster](install.md)
    3. Get [Helm 3.4+](https://helm.sh/docs/intro/install/).
    4. Get [`iter8ctl`](install.md#get-iter8ctl)
    5. Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Clone your fork, and set the `ITER8` environment variable as follows.
    ```shell
    export USERNAME=<your GitHub username>
    ```
    ```shell
    git clone git@github.com:$USERNAME/iter8.git
    cd iter8
    export ITER8=$(pwd)
    ```

## 1. Create app
The `hello world` app consists of a Kubernetes deployment and service. Deploy the app as follows.

```shell
kubectl apply -n default -f $ITER8/samples/deployments/app/deploy.yaml
kubectl apply -n default -f $ITER8/samples/deployments/app/service.yaml
```

### 1.a) Verify app is running

```shell
# create a curl pod and get shell
kubectl run curl --image=radial/busyboxplus:curl -i --tty --rm
```

```shell
# curl the app
curl hello.default.svc.cluster.local:8080
```

`Curl` output will be similar to the following (notice 1.0.0 version tag).
```
Hello, world!
Version: 1.0.0
Hostname: hello-bc95d9b56-xp9kv
```

```shell
# exit from the curl pod
exit
```

## 2. Launch Iter8 experiment
Deploy an Iter8 experiment for SLO validation of the app as follows.
```shell
helm upgrade -n default my-exp $ITER8/samples/slo-validation \
  --set URL='http://hello.default.svc.cluster.local:8080' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --install
```

The above command creates [an Iter8 experiment](../concepts/whatisiter8.md#what-is-an-iter8-experiment) that generates requests, collects latency and error rate metrics for the app, and verifies that the app satisfies mean latency (50 msec), error rate (0.0), 95th percentile tail latency (100 msec) SLOs.

### 2.a) View manifest and values
View the manifest created by the Helm command, the default values used by the Helm chart, and the actual values used by the Helm release using the following instructions.

??? note "View experiment manifest using these instructions"
    ```shell
    helm get manifest -n default my-exp
    ```

    Your manifest created by will be similar to the following.
    ```yaml linenums="1" hl_lines="14 15 16 17 18 19 20 25 26 27 28 29 30 31"
    # Source: slo/templates/experiment.yaml
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: slo-validation-mz8py
    spec:
      # target identifies application for which SLO validation experiment is performed
      target: app
      strategy:
        # this experiment will perform a conformance test
        testingPattern: Conformance
        actions:
          start:
          - task: metrics/collect
            with:
              time: "5s"
              qps: 8
              versions:
              - name: my-app
                url: "http://hello.default.svc.cluster.local:8080"
      criteria:
        requestCount: iter8-system/request-count
        indicators:
        - iter8-system/error-count
        objectives:
        - metric: iter8-system/mean-latency
          upperLimit: "50.0"
        - metric: iter8-system/error-rate
          upperLimit: "0.0"
        - metric: iter8-system/latency-95th-percentile
          upperLimit: "100.0"
      duration:
        intervalSeconds: 1
        iterationsPerLoop: 1
      versionInfo:
        # information about app versions used in this experiment
        baseline:
          name: my-app
    ```

    The [`metrics/collect` task](../reference/tasks/metrics-collect.md) highlighted above is responsible for generating and sending HTTP requests to the app's API endpoint, receiving responses, and creating latency and error-related metrics. The [`objectives`](../../concepts/buildingblocks/#objectives-slos) stanza describes the SLOs that the app needs to satisfy in order to be declared a winner.

??? note "View default values used by the Helm chart"
    ```shell
    helm show values $ITER8/samples/slo-validation
    ```

    This will display the defaults in the Helm chart's `values.yaml` file as follows.
    ```yaml linenums="1"
    # Default values for Iter8 SLO validation experiment

    # Number of queries sent for validation
    numQueries: 100

    # Queries per second
    QPS: 8.0

    # Upper limit on the mean latency. Values beyond this limit are not acceptable
    limitMeanLatency: 500.0 # 500 msec latency on an average is ok. Not more.

    # Upper limit on the error rate. Values beyond this limit are not acceptable
    limitErrorRate: 0.00 # 0%. Errors are not acceptable.

    # Upper limit on the 95th percentile tail latency. Values beyond this limit are not acceptable
    limit95thPercentileLatency: 1000.0 # One second tail latency is ok. Not more.
    ```

??? note "View values set in the Helm release"
    ```shell
    helm get values my-exp --all
    ```

    This will display the all the values set in the Helm release as follows.
    ```yaml linenums="1"
    COMPUTED VALUES:
    QPS: 8
    URL: http://hello.default.svc.cluster.local:8080
    limit95thPercentileLatency: "100.0"
    limitErrorRate: "0.0"
    limitMeanLatency: "50.0"
    numQueries: 100
    ```

## 3. Observe experiment
The `iter8ctl` CLI makes it easy to observe experiment progress and outcomes.

### 3.a) Assert outcomes
Assert that the experiment completed and found a winning version. Wait 20 seconds before trying the following command. If the assertions are not satisfied, try again after a few seconds.

```shell
iter8ctl assert -c completed -c winnerFound -n default
```

### 3.b) Describe results
Describe the results of the Iter8 experiment.

```shell
iter8ctl describe -n default
```

??? info "Sample experiment results"
    ```shell
    ****** Overview ******
    Experiment name: slo-validation-1dhq3
    Experiment namespace: default
    Target: app
    Testing pattern: Conformance
    Deployment pattern: Progressive

    ****** Progress Summary ******
    Experiment stage: Completed
    Number of completed iterations: 1

    ****** Winner Assessment ******
    > If the version being validated; i.e., the baseline version, satisfies the experiment objectives, it is the winner.
    > Otherwise, there is no winner.
    Winning version: my-app

    ****** Objective Assessment ******
    > Whether objectives specified in the experiment are satisfied by versions.
    > This assessment is based on last known metric values for each version.
    +--------------------------------------+------------+--------+
    |                METRIC                | CONDITION  | MY-APP |
    +--------------------------------------+------------+--------+
    | iter8-system/mean-latency            | <= 50.000  | true   |
    +--------------------------------------+------------+--------+
    | iter8-system/error-rate              | <= 0.000   | true   |
    +--------------------------------------+------------+--------+
    | iter8-system/latency-95th-percentile | <= 100.000 | true   |
    +--------------------------------------+------------+--------+

    ****** Metrics Assessment ******
    > Last known metric values for each version.
    +--------------------------------------+--------+
    |                METRIC                | MY-APP |
    +--------------------------------------+--------+
    | iter8-system/mean-latency            |  1.285 |
    +--------------------------------------+--------+
    | iter8-system/error-rate              |  0.000 |
    +--------------------------------------+--------+
    | iter8-system/latency-95th-percentile |  2.208 |
    +--------------------------------------+--------+
    | iter8-system/request-count           | 40.000 |
    +--------------------------------------+--------+
    | iter8-system/error-count             |  0.000 |
    +--------------------------------------+--------+
    ```

### 3.c) Debug
The `iter8ctl` debug command is especially useful in situations where the experiment failed to complete successfully, or produces an unexpected outcome (such as failure to find a winning version).

```shell
# print Iter8logs at priority levels 1 (error), 2 (warning), and 3 (info)
iter8ctl debug --priority 3 -n default
```

??? info "Sample output from iter8ctl debug"
    ```
    Debugging experiment slo-validation-rsuq7 in namespace default
    source: task-runner priority: 3 message: metrics collection completed for all versions
    ```

## 4. Cleanup
```shell
helm uninstall -n default my-exp
kubectl delete -n default -f $ITER8/samples/deployments/app/service.yaml
kubectl delete -n default -f $ITER8/samples/deployments/app/deploy.yaml
```
***

!!! tip "Next Steps"
    1. Run the above experiment with *your* app by replacing the `hello` app with *your* app, and modifying the Helm values appropriately.

    2. Customize the number of queries, and queries per second generated by Iter8, by setting the `numQueries` and `QPS` values.

    3. Try other variations of SLO validation that involve:
        - [HTTP POST API accepting a payload](../tutorials/deployments/slo-validation-payload.md)
        - [GitOps with automated pull request](../tutorials/deployments/slo-validation-pr.md)
        - [GitOps with automated GH Actions workflow trigger](../tutorials/deployments/slo-validation-ghaction.md)
        - [Chaos injection](../tutorials/deployments/slo-validation-chaos.md)
