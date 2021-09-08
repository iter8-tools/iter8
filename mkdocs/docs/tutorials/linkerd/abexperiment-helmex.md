---
template: main.html
---

# A/B Experiment

!!! tip "Scenario: Safely rollout a Kubernetes deployment with an A/B experiment"
    Launch a candidate version of your application (a K8s service and deployment), compare it against the baseline version, and promote the winner.
    
    ![SLO validation](../../images/yourfirstexperiment.png)

??? warning "Setup K8s cluster and local environment"
    1. Get [Helm 3.4+](https://helm.sh/docs/intro/install/) 
    2. Setup [K8s cluster](../../getting-started/setup-for-tutorials.md#local-kubernetes-cluster)
    3. Get [Linkerd](setup-for-tutorials.md)
    4. [Install Iter8 in K8s cluster](../../getting-started/install.md)
    5. Get [`iter8ctl`](../../getting-started/install.md#get-iter8ctl)
    6. Get [the Iter8 Helm repo](../../getting-started/setup-for-tutorials.md#iter8-helm-repo)

## 1. Create baseline version
Deploy the baseline version of the `hello world` application using Helm.

```shell
helm install my-app iter8/linkerd \
  --set baseline.dynamic.tag=1.0 \
  --set candidate=null  
```

??? note "Verify that baseline version is 1.0.0"
    ```shell
    # do this in a separate terminal
    kubectl port-forward svc/hello 8080:8080
    ```

    ```shell
    curl localhost:8080
    ```

    ```
    # output will be similar to the following (notice 1.0.0 version tag)
    # hostname will be different in your environment
    Hello, world!
    Version: 1.0.0
    Hostname: hello-bc95d9b56-xp9kv
    ```

<!-- 
```shell
kubectl create deploy hello --image=gcr.io/google-samples/hello-app:1.0
kubectl create svc clusterip hello --tcp=8080
``` 
-->

## 2. Create candidate version
Deploy the candidate version of the `hello world` application using Helm.

```shell
helm upgrade my-app iter8/linkerd \
  --set baseline.dynamic.tag=1.0 \
  --set baseline.weight=50 \
  --set candidate.dynamic.tag=2.0 \
  --set candidate.weight=50 \
  --install  
```

The above command creates [an Iter8 experiment](../../concepts/whatisiter8.md#what-is-an-iter8-experiment) alongside the candidate deployment of the `hello world` application. The experiment will collect latency and error rate metrics for the candidate, and verify that it satisfies the mean latency (50 msec), error rate (0.0), 95th percentile tail latency SLO (100 msec) SLOs.

??? note "View application and experiment resources"
    Use the command below to view your application and Iter8 experiment resources.
    ```shell
    helm get manifest my-app
    ```

??? note "Verify that candidate version is 2.0.0"
    ```shell
    # do this in a separate terminal
    kubectl port-forward svc/hello-candidate 8081:8080
    ```

    ```shell
    curl localhost:8081
    ```

    ```
    # output will be similar to the following (notice 2.0.0 version tag)
    # hostname will be different in your environment
    Hello, world!
    Version: 2.0.0
    Hostname: hello-bc95d9b56-xp9kv
    ```

<!-- 
```shell
kubectl create deploy hello-candidate --image=gcr.io/google-samples/hello-app:2.0
kubectl create svc clusterip hello-candidate --tcp=8080
``` 
-->

## 3. Observe experiment
Describe the results of the Iter8 experiment. Wait 20 seconds before trying the following command. If the output is not as expected, try again after a few more seconds.

```shell
iter8ctl describe
```

??? info "Experiment results will look similar to this ... "
    ```shell
    ****** Overview ******
    Experiment name: hello-experiment-57a46
    Experiment namespace: test
    Target: hello
    Testing pattern: A/B
    Deployment pattern: FixedSplit

    ****** Progress Summary ******
    Experiment stage: Completed
    Number of completed iterations: 10

    ****** Winner Assessment ******
    App versions in this experiment: [hello hello-candidate]
    Winning version: hello-candidate
    Version recommended for promotion: hello-candidate

    ****** Reward Assessment ******
    > Identifies values of reward metrics for each version. The best version is marked with a '*'.
    +--------------------------------+-------+-----------------+
    |             REWARD             | HELLO | HELLO-CANDIDATE |
    +--------------------------------+-------+-----------------+
    | test/user-engagement (higher   | 5.204 | 9.442 *         |
    | better)                        |       |                 |
    +--------------------------------+-------+-----------------+

    ****** Objective Assessment ******
    > Identifies whether or not the experiment objectives are satisfied by the most recently observed metrics values for each version.
    +------------------------------+-------+-----------------+
    |          OBJECTIVE           | HELLO | HELLO-CANDIDATE |
    +------------------------------+-------+-----------------+
    | test/mean-latency <= 300.000 | true  | true            |
    +------------------------------+-------+-----------------+
    | test/error-rate <= 0.010     | true  | true            |
    +------------------------------+-------+-----------------+

    ****** Metrics Assessment ******
    > Most recently read values of experiment metrics for each version.
    +--------------------------------+--------+-----------------+
    |             METRIC             | HELLO  | HELLO-CANDIDATE |
    +--------------------------------+--------+-----------------+
    | test/mean-latency              |  1.000 |           0.556 |
    | (milliseconds)                 |        |                 |
    +--------------------------------+--------+-----------------+
    | request-count                  | 10.213 |          10.213 |
    +--------------------------------+--------+-----------------+
    | test/error-rate                |  0.000 |           0.000 |
    +--------------------------------+--------+-----------------+
    | test/request-count             | 10.213 |          10.211 |
    +--------------------------------+--------+-----------------+
    | test/user-engagement           |  5.204 |           9.442 |
    +--------------------------------+--------+-----------------+
    ``` 

## 4. Promote winner
Assert that the experiment completed and found a winning version. If the conditions are not satisfied, try again after a few more seconds.

```shell
iter8ctl assert -c completed -c winnerFound
```

Promote the winner as follows.

```shell
helm upgrade my-app iter8/linkerd \
  --install \
  --set baseline.dynamic.tag=2.0 \
  --set candidate=null
```

??? note "Verify that baseline version is 2.0.0"
    ```shell
    # kill the port-forward commands from steps 1 and 2
    # do this in a separate terminal
    kubectl port-forward svc/hello 8080:8080
    ```

    ```shell
    curl localhost:8080
    ```

    ```
    # output will be similar to the following (notice 2.0.0 version tag)
    # hostname will be different in your environment
    Hello, world!
    Version: 2.0.0
    Hostname: hello-bc95d9b56-xp9kv
    ```

## 5. Cleanup

```shell
helm uninstall my-app
```

***

**Next Steps**

!!! tip "Use in production"
    The Helm chart source for this application is located in `$ITER8/helm/linkerd`. Modify the chart, including the experiment template, as needed by your application for production usage.
