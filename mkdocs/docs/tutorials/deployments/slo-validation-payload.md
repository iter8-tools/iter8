---
template: main.html
---

# SLO Validation with Payload
!!! tip "Scenario: Validate SLOs for apps with POST APIs that receive payload"
    **Problem**: You have a Kubernetes app that receives a payload through an HTTP POST API. You want to verify that it satisfies latency and error rate SLOs.

    **Solution**: In this tutorial, you will launch a Kubernetes app that implements a POST API that receives a payload, along with an Iter8 experiment. Iter8 will [validate that the app satisfies latency and error-based objectives (SLOs)](../../concepts/buildingblocks.md#slo-validation) using [built-in metrics](../../metrics/builtin.md). During this validation, Iter8 will generate POST requests with payload for the app.

    ![SLO Validation](../../images/slo-validation.png)

??? warning "Setup Kubernetes cluster and local environment"
    1. Setup [Kubernetes cluster](../../getting-started/setup-for-tutorials.md#local-kubernetes-cluster)
    2. [Install Iter8 in Kubernetes cluster](../../getting-started/install.md)
    3. Get [Helm 3.4+](https://helm.sh/docs/intro/install/).
    4. Get [`iter8ctl`](../../getting-started/install.md#get-iter8ctl)
    5. Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Clone your fork, and set the ITER8 environment variable as follows.
    ```shell
    export USERNAME=<your GitHub username>
    ```
    ```shell
    git clone git@github.com:$USERNAME/iter8.git
    cd iter8
    export ITER8=$(pwd)
    ```

## 1. Create app
The `httpbin` app consists of a Kubernetes deployment and service. Deploy the app as follows.

```shell
kubectl apply -n default -f $ITER8/samples/deployments/httpbin/deploy.yaml
kubectl apply -n default -f $ITER8/samples/deployments/httpbin/service.yaml
```

### Verify app

??? note "Verify that the app is running using these instructions"
    ```shell
    # do this in a separate terminal
    kubectl port-forward -n default svc/httpbin 8080:80
    ```

    ```shell
    curl http://localhost:8080/post -X POST -d @$ITER8/samples/deployments/httpbin/payload.json -H "Content-Type: application/json"
    ```

    `Curl` output will be similar to the following.
    ```json
    {
      "args": {}, 
      "data": "{  \"hello\": \"world\",  \"goodbye\": \"world\"}", 
      "files": {}, 
      "form": {}, 
      "headers": {
        "Accept": "*/*", 
        "Content-Length": "41", 
        "Content-Type": "application/json", 
        "Host": "localhost:8080", 
        "User-Agent": "curl/7.64.1"
      }, 
      "json": {
        "goodbye": "world", 
        "hello": "world"
      }, 
      "origin": "127.0.0.1", 
      "url": "http://localhost:8080/post"
    }
    ```

## 2. Create Iter8 experiment
Deploy an Iter8 experiment for SLO validation of the app as follows.
```shell
helm upgrade -n default my-exp $ITER8/samples/first-exp \
  --set URL='http://httpbin.default.svc.cluster.local:8080' \
  --set payloadURL='http://payload.com' \
  --set contentType='application/json' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --install  
```

The above command creates [an Iter8 experiment](../../concepts/whatisiter8.md#what-is-an-iter8-experiment) that generates requests, collects latency and error rate metrics for the app, and verifies that the app satisfies mean latency (50 msec), error rate (0.0), 95th percentile tail latency SLO (100 msec) SLOs.

??? note "View Iter8 experiment"
    View the Iter8 experiment as follows.
    ```shell
    helm get manifest -n default my-exp
    ```

    There are two main aspects to this ... task and criteria.

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
helm uninstall -n default my-exp
# remove app
kubectl delete -n default -f $ITER8/samples/deployments/app/service.yaml
kubectl delete -n default -f $ITER8/samples/deployments/app/deploy.yaml
```
***

**Next Steps**

!!! tip "Use with your app"
    1. Run the above experiment with your app by setting the `URL` value in the Helm command to the URL of your app. 
    
    2. You can also customize the mean latency, error rate, and tail latency limits.

    3. This experiment can be run in any Kubernetes environment such as a dev, test, staging, or production cluster.    

