---
template: main.html
---

# SLO Validation with GitOps
!!! tip "Scenario: Validate SLOs and promote a new version of a K8s app"
    **Problem:** You have a new version of a K8s app. You want to verify that it satisfies latency and error rate SLOs, and promote it to production as the stable version of your app in a GitOps-y manner.

    **Solution:** In this tutorial, you will [dark launch](../../concepts/buildingblocks.md#dark-launch) the new version of your K8s app along with an Iter8 experiment. Iter8 will [validate that the new satisfies latency and error-based objectives (SLOs)](../../concepts/buildingblocks.md#slo-validation) using [built-in metrics](../../metrics/builtin.md) and [promote the new version by raising a pull-request in a GitHub repo](../../concepts/buildingblocks.md#version-promotion).

    ![SLO Validation GitOps](../../images/slo-validation-gitops.png)

??? warning "Setup K8s cluster, local environment, and GitHub credentials"
    0. Complete the [Iter8 getting-started tutorial](../../getting-started/first-experiment.md) and then skip ahead to step 7 of setup.
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
    6. Enable Iter8 to update your fork.
        1. [Create a personal access token on GitHub](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token). In Step 8 of this process, select repo. This will ensure that the token can be used by Iter8 to update your app manifest in GitHub.
        2. Create K8s secret
        ```shell
        # replace $GHTOKEN with GitHub token created above
        kubectl create secret generic ghtoken --from-literal=token=$GHTOKEN
        ```
        3. Provide RBAC permission
        ```shell
        kubectl create role ghtoken-reader \
          --verb=get \
          --resource=secrets \
          --resource-name=ghtoken
        kubectl create rolebinding ghtoken-reader-binding \
          --role=ghtoken-reader \
          --serviceaccount=iter8-system:iter8-handlers
        ```


## 1. Create stable version
Create version `1.0` of the `hello world` app as follows.

```shell
# USERNAME is exported as part of setup steps.
kubectl apply -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/deployments/app/deploy.yaml
kubectl apply -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/deployments/app/service.yaml
```

## 2. Create candidate version
Create version `2.0` of the `hello world` app in the staging environment as follows. For the purpose of this tutorial, the production environment is the `default` namespace, and the staging environment is the `staging` namespace.

```shell
kubectl create ns staging
# version 2.0 of hello world app will in the staging namespace
kubectl set image --local -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/deployments/app/deploy.yaml hello='gcr.io/google-samples/hello-app:2.0' -o yaml | kubectl apply -n staging -f -
kubectl apply -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/deployments/app/service.yaml -n staging
```

Adapt [these instructions](../../getting-started/first-experiment.md#1-create-app) to verify that stable and candidate versions of your app are running.

## 3. Create Iter8 experiment
Deploy an Iter8 experiment for SLO validation and GitOps-y promotion of the app as follows.
```shell
helm upgrade -n staging my-exp $ITER8/samples/slo-gitops \
  --set URL='http://hello.staging.svc.cluster.local:8080' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --set username=$USERNAME \
  --set repo="github.com/$USERNAME/iter8.git" \
  --set newImage='gcr.io/google-samples/hello-app:2.0' \
  --install
```

The above command creates [an Iter8 experiment](../../concepts/whatisiter8.md#what-is-an-iter8-experiment) that generates requests, collects latency and error rate metrics for the candidate version of the app, and verifies that the candidate satisfies mean latency (50 msec), error rate (0.0), 95th percentile tail latency SLO (100 msec) SLOs. 

In the above command, the *USERNAME* environment variable was defined during setup. After the Iter8 experiment validates SLOs for the candidate, it uses the GitHub token (also provided during setup) to promote the candidate to production using a GitHub pull-request.

## 4. View and observe experiment
View the Iter8 experiment as described [here](../../getting-started/first-experiment.md#2-create-iter8-experiment). Observe the experiment by following [these steps](../../getting-started/first-experiment.md#3-observe-experiment). Make sure you supply the correct namespace (`staging`).

## 5. Review Iter8's PR

## 6. Cleanup

```shell
# remove Iter8 experiment and candidate version of the app
kubectl delete ns staging
# remove stable version of the app
kubectl delete deploy/hello
kubectl delete svc/hello
```

***

**Next Steps**

!!! tip "Use with your app"
    1. Replace hello

    2. Use with a GitOps operator like ArgoCD or FluxCD

    3. Promote from a staging to production cluster

