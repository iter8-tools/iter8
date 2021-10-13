---
template: main.html
---

# SLO Validation with Auto GH Actions Trigger
!!! tip "Validate SLOs and automatically trigger a GitHub Actions workflow"
    **Problem**: You have a Kubernetes app. You want to verify that it satisfies latency and error rate SLOs and automatically trigger a GitHub Actions workflow based on the result.

    **Solution**: In this tutorial, you will launch a Kubernetes app along with an Iter8 experiment. Iter8 will [validate that the app satisfies latency and error-based objectives (SLOs)](../../concepts/buildingblocks.md#slo-validation) using [built-in metrics](../../metrics/builtin.md). During this validation, Iter8 will generate HTTP GET requests for the app. Once Iter8 verifies that the app satisfies SLOs, it will [automatically trigger a GitHub Actions workflow](../../concepts/buildingblocks.md#version-promotion).

    ![SLO Validation GitHub Action Trigger](../../images/slo-validation-ghaction.png)

???+ warning "Setup Kubernetes cluster and local environment"
    1. If you completed the [Iter8 getting-started tutorial](../../getting-started/first-experiment.md) (highly recommended), you may skip the remaining steps of setup.
    2. Setup [K8s cluster](../../getting-started/setup-for-tutorials.md#local-kubernetes-cluster)
    3. [Install Iter8 in K8s cluster](../../getting-started/install.md)
    4. Get [Helm 3.4+](https://helm.sh/docs/intro/install/).
    5. Get [`iter8ctl`](../../getting-started/install.md#install-iter8ctl)
    6. Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Clone your fork, and set the `ITER8` environment variable as follows.
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

## 2. Enable GitOps
3.1) [Create a personal access token on GitHub](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token). In Step 8 of this process, grant `repo`, `workflow` and `read:org` permissions to this token. This will ensure that the token can be used by Iter8 to trigger GitHub Actions workflows.

3.2) Create K8s secret
```shell
# GHTOKEN environment variable contains the GitHub token created above
kubectl create secret generic ghtoken --from-literal=token=$GHTOKEN
```

## 3. Launch Iter8 experiment
Deploy an Iter8 experiment for SLO validation followed by a notification that triggers a GitHub Actions workflow.
```shell
# USERNAME environment variable contains your GitHub username
helm upgrade my-exp $ITER8/samples/slo-ghaction \
  --set URL='http://hello.default.svc.cluster.local:8080' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --set owner=$USERNAME \
  --set repo=iter8 \
  --set workflow=demo.yaml \
  --install
```

The above command creates [an Iter8 experiment](../../concepts/whatisiter8.md#what-is-an-iter8-experiment) that generates requests, collects latency and error rate metrics for the candidate version of the app, and verifies that the candidate satisfies mean latency (50 msec), error rate (0.0), 95th percentile tail latency (100 msec) SLOs. 

Once Iter8 verifies that the app satisfies SLOs, it will trigger the `demo.yaml` workflow in the `iter8` repo. It uses the `ghtoken` secret to do this.

View the manifest created by the Helm command, the default values used by the Helm chart, and the actual values used by the Helm release by following [the instructions in this step](../../getting-started/first-experiment.md#2a-view-manifest-and-values).

## 4. Observe experiment
Observe the experiment by following [these steps](../../getting-started/first-experiment.md#3-observe-experiment).Once the experiment completes, visit https://github.com/$USERNAME/iter8/actions to view your workflow run.

## 5. Cleanup
```shell
helm uninstall -n default my-exp
kubectl delete -n default -f $ITER8/samples/deployments/app/service.yaml
kubectl delete -n default -f $ITER8/samples/deployments/app/deploy.yaml
```

***

!!! tip "Reuse with your app"
    Reuse the above experiment with *your* app by replacing the `hello` app with *your* app, and modifying the Helm values appropriately.
