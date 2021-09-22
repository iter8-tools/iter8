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
helm upgrade -n default my-exp $ITER8/samples/deployments/payload \
  --set URL='http://httpbin.default.svc.cluster.local/post' \
  --set payloadURL='https://raw.githubusercontent.com/sriumcp/iter8/post/samples/deployments/httpbin/payload.json' \
  --set contentType='application/json' \
  --set limitMeanLatency=100.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=200.0 \
  --install  
```

The above command creates [an Iter8 experiment](../../concepts/whatisiter8.md#what-is-an-iter8-experiment) that generates HTTP requests, collects latency and error rate metrics for the app, and verifies that the app satisfies mean latency (100 msec), error rate (0.0), 95th percentile tail latency (200 msec) SLOs. These HTTP requests are POST requests and use the JSON data from the `payloadURL` specified in the command as request payload.

??? note "View Iter8 experiment"
    View the Iter8 experiment as follows.
    ```shell
    helm get manifest -n default my-exp
    ```

    Notice the `metrics/collect` task in the experiment manifest and also the `objectives` stanza that specifies SLOs.

## 3. Observe experiment
Observe the experiment as described [here](../../getting-started/first-experiment.md#3-observe-experiment).

## 4. Cleanup
```shell
# remove experiment
helm uninstall -n default my-exp
# remove app
kubectl delete -n default -f $ITER8/samples/deployments/httpbin/service.yaml
kubectl delete -n default -f $ITER8/samples/deployments/httpbin/deploy.yaml
```
***

**Next Steps**

!!! tip "Try in your environment"
    1. Run the above experiment with your app by setting the `URL` value in the Helm command to the URL of your app, and also by using a `payloadURL` that is appropriate for your application. You can run this experiment in any Kubernetes environment such as a dev, test, staging, or production cluster.
    
    2. You can also customize the mean latency, error rate, and tail latency limits in the SLOs.

    3. Iter8 makes it possible to [promote the winning version](../../concepts/buildingblocks.md#version-promotion) in a number of different ways. For example, you may have a stable version running in production, have a candidate version deployed in a staging environment, perform this experiment, ensure that the candidate is successful, and promote it as the latest stable version in a GitOps-y manner as described [here](../deployments/slo-validation-gitops.md).