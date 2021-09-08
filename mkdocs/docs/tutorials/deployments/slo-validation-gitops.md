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
        1. [Create a personal access token on GitHub](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token). In Step 8 of this process, select repo. This will ensure that the token can be used by Iter8 to update your application manifest in GitHub.
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
Create version `1.0` of the `hello world` application as follows.

```shell
kubectl apply -f https://github.com/iter8-tools/iter8/samples/deployments/baseline.yaml
```

## 2. Create candidate version
Create `hello` application version `2.0`.

```shell
kubectl apply -f https://github.com/iter8-tools/iter8/samples/deployments/candidate.yaml
```

## 3. Create Iter8 experiment
Create `hello` application version `2.0`.

```shell
kubectl apply -f https://github.com/iter8-tools/iter8/samples/deployments/candidate.yaml
```

## 4. Understand experiment results

## 5. Merge Iter8's pull-request

## 6. Use with your app

??? tip "Use with your app"
    Hello hello hello halo

??? tip "Use with a GitOps operator like ArgoCD or FluxCD"
    Hello hello hello halo

??? tip "Promote from a staging to production cluster"
    Hello hello hello halo

