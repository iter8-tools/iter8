# Chaos Testing with SLO Validation

!!! tip "Scenario: Inject Chaos into Kubernetes cluster and verify if application can satisfy SLOs"
    **Problem:** You have a Kubernetes app. You want to stress test it by injecting chaos, and verify that it can satisfy service-level objectives (SLOs). This helps you guarantee that your application is resilient, and works well even under periods of stress (like intermittent pod failures).

    **Solution:** You will launch a Kubernetes application along with a composite experiment, consisting of [Litmus Chaos](https://litmuschaos.io/) experiment resource and an Iter8 experiment resource. The chaos experiment will delete pods of the application periodically, while the Iter8 experiment will send requests to the application and verify if it is able to satisfy SLOs.

    ![Chaos with SLO Validation](../../images/chaos-slo-validation.png)

??? warning "Setup Kubernetes cluster and local environment"
    0. If you completed the [Iter8 getting-started tutorial](../../getting-started/first-experiment.md) (highly recommended), you may skip to step number 6.
    1. Setup [K8s cluster](../../getting-started/setup-for-tutorials.md#local-kubernetes-cluster)
    2. [Install Iter8 in K8s cluster](../../getting-started/install.md)
    3. Get [Helm 3.4+](https://helm.sh/docs/intro/install/).
    4. Get [`iter8ctl`](../../getting-started/install.md#install-iter8ctl)
    5. Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Clone your fork, and set the ITER8 environment variable as follows.
    ```shell
    export USERNAME=<your GitHub username>
    ```
    ```shell
    git clone git@github.com:$USERNAME/iter8.git
    cd iter8
    export ITER8=$(pwd)
    ```
    6. Setup [Litmus].

## 1. Create app
The `hello world` app consists of a K8s deployment and service. Deploy the app as follows.

```shell
kubectl apply -n default -f $ITER8/samples/deployments/app/deploy.yaml
kubectl apply -n default -f $ITER8/samples/deployments/app/service.yaml
```

Use [these instructions](../../getting-started/first-experiment.md#verify-app) to verify that your app is running.

## 2. Create composite experiment

```shell
helm upgrade -n default my-exp $ITER8/samples/chaos \
  --set URL='http://hello.staging.svc.cluster.local:8080' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --set username=$USERNAME \
  --set newImage='gcr.io/google-samples/hello-app:2.0' \
  --install
```

## 3. Observe Experiment
View the Iter8 experiment as described [here](../../getting-started/first-experiment.md#2-create-iter8-experiment). Observe the experiment by following [these steps](../../getting-started/first-experiment.md#3-observe-experiment).

## 4. Cleanup
```shell
# remove chaos + Iter8 experiments
helm uninstall -n default my-exp
# remove app
kubectl delete -n default -f $ITER8/samples/deployments/app/service.yaml
kubectl delete -n default -f $ITER8/samples/deployments/app/deploy.yaml
```

***

**Next Steps**

???+ tip "Reconfigure app to ensure success, use with your own app, and try other types of Chaos"
    1. Increase replicaCount to 2. Repeat the same experiment and watch it succeed.

    2. I have my own app (not hello-world); how can I use these instructions with my own app.

    3. I want to inject some other type of chaos. How can I do that?

    4. I want to couple this experiment with version promotion. For example, I want to do this in staging, and once the experiment succeeds, I want to push it to production.