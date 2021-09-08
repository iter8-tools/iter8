---
template: main.html
---

# Your First Experiment

!!! tip "Scenario: Safely rollout a Kubernetes deployment with SLO validation"
    [Dark launch](../concepts/buildingblocks.md#dark-launch) a candidate version of your application (a K8s service and deployment), [validate that the candidate satisfies latency and error-based objectives (SLOs)](../concepts/buildingblocks.md#slo-validation), and promote the candidate.
    
    ![SLO validation](../images/yourfirstexperiment.png)

??? warning "Setup K8s cluster and local environment"
    1. Get [Helm 3.4+](https://helm.sh/docs/intro/install/). This tutorial uses the [Helmex pattern](../concepts/whatisiter8.md#what-is-helmex)
    2. Setup [K8s cluster](setup-for-tutorials.md#local-kubernetes-cluster)
    3. [Install Iter8 in K8s cluster](install.md)
    4. Get [`iter8ctl`](install.md#install-iter8ctl)
    5. Get [the Iter8 Helm repo](setup-for-tutorials.md#iter8-helm-repo)

## 1. Create application
The `hello world` app consists of a K8s deployment and service. Deploy them as follows.

```shell
kubectl apply -f $ITER8/samples/deployments/app/deploy.yaml
kubectl apply -f $ITER8/samples/deployments/app/service.yaml
```

??? note "Verify app is running"
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

## 2. Create Iter8 experiment
Deploy the Iter8 experiment for SLO validation of the app as follows.
```shell
helm upgrade my-exp $ITER8/samples/first-exp \
  --set URL='http://hello.default.svc.cluster.local:8080' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --install  
```

The above command creates [an Iter8 experiment](../concepts/whatisiter8.md#what-is-an-iter8-experiment) that will generate requests, collect latency and error rate metrics for the application, and verify that it satisfies the mean latency (50 msec), error rate (0.0), 95th percentile tail latency SLO (100 msec) SLOs.

??? note "View Iter8 experiment deployed by Helm"
    Use the command below to view the Iter8 experiment deployed by the Helm command.
    ```shell
    helm get manifest my-exp
    ```

## 3. Observe experiment
Assert that the experiment completed and found a winning version. Wait 20 seconds before trying the following command. If the assertions are not satisfied, try again after a few seconds.

```shell
iter8ctl assert -c completed -c winnerFound
```

Describe the results of the Iter8 experiment. 
```shell
iter8ctl describe
```

??? info "Experiment results will look similar to this ... "
    ```shell
    ****** Overview ******
    Experiment name: my-experiment
    Experiment namespace: default
    Target: my-app
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
    > Identifies whether or not the experiment objectives are satisfied by the most recently observed metrics values for each version.
    +--------------------------------------+--------+
    |              OBJECTIVE               | MY-APP |
    +--------------------------------------+--------+
    | iter8-system/mean-latency <=         | true   |
    |                               50.000 |        |
    +--------------------------------------+--------+
    | iter8-system/error-rate <=           | true   |
    |                                0.000 |        |
    +--------------------------------------+--------+
    | iter8-system/latency-95th-percentile | true   |
    | <= 100.000                           |        |
    +--------------------------------------+--------+

    ****** Metrics Assessment ******
    > Most recently read values of experiment metrics for each version.
    +--------------------------------------+--------+
    |                METRIC                | MY-APP |
    +--------------------------------------+--------+
    | iter8-system/mean-latency            |  1.233 |
    +--------------------------------------+--------+
    | iter8-system/error-rate              |  0.000 |
    +--------------------------------------+--------+
    | iter8-system/latency-95th-percentile |  2.311 |
    +--------------------------------------+--------+
    | iter8-system/request-count           | 40.000 |
    +--------------------------------------+--------+
    | iter8-system/error-count             |  0.000 |
    +--------------------------------------+--------+
    ``` 

## 4. Cleanup
```shell
# remove experiment
helm uninstall my-exp
# remove application
kubectl delete -f $ITER8/samples/deployments/app/service.yaml
kubectl delete -f $ITER8/samples/deployments/app/deploy.yaml
```
***

**Next Steps**

!!! tip "Use with your application"
    1. Run the above experiment with your application by setting the `URL` value in the Helm command to the URL of your application. 
    
    2. You can also customize the mean latency, error rate, and tail latency limits.

    3. This experiment can be run in any K8s environment such as a dev, test, staging, or production cluster.