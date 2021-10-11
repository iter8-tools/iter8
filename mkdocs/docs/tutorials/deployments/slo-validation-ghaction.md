---
template: main.html
---

# SLO Validation with GitHub Actions Trigger
!!! tip "Scenario: Validate SLOs and promote a new version of a K8s app"
    **Problem:** You have a new version of a K8s app. You want to verify that it satisfies latency and error rate SLOs, and promote it to production as the stable version of your app in a GitOps-y manner.

    **Solution:** In this tutorial, you will [dark launch](../../concepts/buildingblocks.md#dark-launch) the new version of your K8s app along with an Iter8 experiment. Iter8 will [validate that the new satisfies latency and error-based objectives (SLOs)](../../concepts/buildingblocks.md#slo-validation) using [built-in metrics](../../metrics/builtin.md) and [promote the new version by triggering a GitHub Actions workflow](../../concepts/buildingblocks.md#version-promotion).

    ![SLO Validation GitHub Action Trigger](../../images/slo-validation-ghaction.png)

??? warning "Setup Kubernetes cluster and local environment"
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
# create version 2.0 of hello world app in the staging namespace
kubectl set image --local -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/deployments/app/deploy.yaml hello='gcr.io/google-samples/hello-app:2.0' -o yaml | kubectl apply -n staging -f -
kubectl apply -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/deployments/app/service.yaml -n staging
```

Adapt [these instructions](../../getting-started/first-experiment.md#verify-app) to verify that stable and candidate versions of your app are running.

## 3. Enable Iter8 GitOps
3.1) [Create a personal access token on GitHub](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token). In Step 8 of this process, grant `repo`, `workflow` and `read:org` permissions to this token. This will ensure that the token can be used by Iter8 to trigger GitHub Actions workflows.

3.2) Create K8s secret
```shell
# replace $GHTOKEN with GitHub token created above
kubectl create secret generic -n staging ghtoken --from-literal=token=$GHTOKEN
```

## 4. Create Iter8 experiment
Deploy an Iter8 experiment for SLO validation followed by a notification that triggers a GitHub Actions workflow.
```shell
helm upgrade -n staging my-exp $ITER8/samples/slo-ghaction \
  --set URL='http://hello.staging.svc.cluster.local:8080' \
  --set limitMeanLatency=50.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=100.0 \
  --set username=$USERNAME \
  --set newImage='gcr.io/google-samples/hello-app:2.0' \
  --install
```

The above command creates [an Iter8 experiment](../../concepts/whatisiter8.md#what-is-an-iter8-experiment) that generates requests, collects latency and error rate metrics for the candidate version of the app, and verifies that the candidate satisfies mean latency (50 msec), error rate (0.0), 95th percentile tail latency (100 msec) SLOs. 

In the above command, the *USERNAME* environment variable was defined during setup. After the Iter8 experiment validates SLOs for the candidate, it uses the GitHub token (also provided during setup) to trigger a GitHub Actions workflow. The workflow in turn creates a GitHub pull request for promoting the candidate as the latest stable version.

## 5. View and observe experiment
View the Iter8 experiment as described [here](../../getting-started/first-experiment.md#2-create-iter8-experiment). Observe the experiment by following [these steps](../../getting-started/first-experiment.md#3-observe-experiment). Ensure correct namespace (`staging`) is used.

## 6. Review PR
Once the experiment completes, you can visit your fork at https://github.com/$USERNAME/iter8/pulls to review the GitHub pull request created by the GitHub Actions workflow.

## 7. Cleanup

```shell
# remove Iter8 experiment and candidate version of the app
kubectl delete ns staging
# remove stable version of the app
kubectl delete deploy/hello
kubectl delete svc/hello
```

***

**Next Steps**

??? tip "Use with your git repo/app"
    1. The Helm chart used in this tutorial is located at $ITER8/samples/slo-gitops-ghaction. Within this folder, the file `templates/experiment.yaml` contains the Iter8 experiment template. The following lines in this template are responsible for cloning the git repo, modifying it locally, and pushing the changes.
    ```shell
    # clone repo using token
    git clone https://$USERNAME:$TOKEN@github.com/$USERNAME/iter8.git

    # commit changes locally and push to a branch
    cd iter8
    git checkout -b iter8-gitops
    yq e -i '.spec.template.spec.containers[0].image = "{{ required ".Values.newImage is required!" .Values.newImage }}"' samples/deployments/app/deploy.yaml 
    git commit --allow-empty -a -m "promote candidate version to production"
    git push -f origin iter8-gitops
    ```
    Change these lines in the template so that the correct repo is cloned and your app is modified correctly.

    2. Ensure that correct GitHub access token is supplied to Iter8.

    3. You may need to supplement or replace the Helm values `username` and `newImage` used in this tutorial with other Helm values as needed by your repo/app.
