---
template: main.html
---

# SLO Validation (Helmex, GitOps)

!!! tip "Scenario: Safely rollout new version of a Knative app with SLO validation"
    This tutorial builds on the [Knative Helmex tutorial for SLO validation](slovalidation-helmex.md), and illustrates the [Helmex pattern](../../concepts/whatisiter8.md#what-is-helmex) in the context of GitOps.

    ![SLO validation](../../images/helmexgitops.png)

    In this tutorial, a Helm `values.yaml` file in a Git repo will be the *source of truth* about your deployed K8s application, and changes to application will be managed declaratively through this file. Steps **a)** and **b)** in the figure will be illustrated manually and Step **c)** will be automated by Iter8.

??? warning "Setup K8s cluster with Knative and local environment"
    1. Follow the setup in the [Knative Helmex tutorial for SLO validation](slovalidation-helmex.md).
    2. If you haven't done so already, try the [Knative Helmex tutorial for SLO validation](slovalidation-helmex.md), and cleanup. This will promote a better understanding of the current tutorial.

## 1. Fork and clone
Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Clone your fork, and set the `$ITER8` environment variable as follows.

```shell
export USERNAME=<your GitHub username>
```
```shell
git clone git@github.com:$USERNAME/iter8.git
```

```
cd iter8
export BRANCH=gitops-test
git checkout -b $BRANCH
export ITER8=$(pwd)
```

## 2. Sync K8s cluster
Deploy the baseline version of the `hello world` Knative app using Helm.

```shell
# curl https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/knative/second-exp/values.yaml
helm install my-app iter8/knslo-gitops \
  -f https://raw.githubusercontent.com/$USERNAME/iter8/master/samples/knative/second-exp/values.yaml
# helm get manifest my-app
```

## 3. Enable Iter8 to update Git repo

### 3.a) Create GitHub token
Create a [personal access token on GitHub](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token). In Step 8 of this process, select `repo`. This will ensure that the token can be used by Iter8 to update the `values.yaml` file in GitHub.

### 3.b) Create K8s secret
```shell
# replace $GHTOKEN with GitHub token created above
kubectl create secret generic ghtoken --from-literal=token=$GHTOKEN
```

### 3.c) Provide RBAC permission
```shell
kubectl create role ghtoken-reader \
  --verb=get \
  --resource=secrets \
  --resource-name=ghtoken
kubectl create rolebinding ghtoken-reader-binding \
  --role=ghtoken-reader \
  --serviceaccount=iter8-system:iter8-handlers
```

Iter8 can now read the GitHub token.

## 4. Create candidate version
When a new candidate arrives, a deployment pipeline would typically update the `values.yaml` file in the GitHub repo. In this tutorial, simulate this pipeline as follows.

### 4.a) Update `values.yaml` locally

```shell
cat <<EOF > $ITER8/samples/knative/second-exp/values.yaml
common:
  application: hello
  repo: "gcr.io/google-samples/hello-app"

baseline:
  dynamic:
    tag: "1.0"
    id: "v1"

candidate:
  dynamic:
    tag: "2.0"
    id: "v2"

experiment:
  # Iter8 will update this values.yaml file in the $BRANCH branch of your repo
  helmexGitOps:
    gitRepo: "https://github.com/$USERNAME/iter8.git"
    filePath: "samples/knative/second-exp/values.yaml"
    username: $USERNAME
    branch: $BRANCH
EOF
```

### 4.b) Git push `values.yaml`

```shell
git commit -a -m "update values.yaml with candidate version" --allow-empty
git push origin $BRANCH -f
# curl https://raw.githubusercontent.com/$USERNAME/iter8/$BRANCH/samples/knative/second-exp/values.yaml
```

### 4.c) Sync K8s cluster
Deploy the candidate version of the `hello world` Knative app and the Iter8 experiment using Helm.
```shell
helm upgrade my-app iter8/knslo-gitops \
  -f https://raw.githubusercontent.com/$USERNAME/iter8/$BRANCH/samples/knative/second-exp/values.yaml \
  --install
# helm get manifest my-app
```

## 5. Observe experiment
```shell
iter8ctl describe
```

## 6. Promote winner

### 6.a) Assert winner
Assert that the experiment completed and found a winning version. If the conditions are not satisfied, try again after a few seconds.
```shell
iter8ctl assert -c completed -c winnerFound
```

### 6.b) GitOps-y promotion by Iter8
This Iter8 experiment automates **version promotion in the GitHub `values.yaml` file**. 

??? note "GitHub `values.yaml` file before and after version promotion by Iter8"
    === "Before"
        ```yaml
        common:
          application: hello
          repo: "gcr.io/google-samples/hello-app"

        baseline:
          dynamic:
            tag: "1.0"
            id: "v1"

        candidate:
          dynamic:
            tag: "2.0"
            id: "v2"

        experiment:
          helmexGitOps:
            gitRepo: "https://github.com/$USERNAME/iter8.git"
            filePath: "samples/knative/second-exp/values.yaml"
            username: $USERNAME
            branch: $BRANCH
        ```

    === "After"
        ```yaml
        common:
          application: hello
          repo: "gcr.io/google-samples/hello-app"

        baseline:
          dynamic:
            tag: "2.0"
            id: "v2"

        experiment:
          helmexGitops:
            repo: "https://github.com/$USERNAME/iter8.git"
            path: "samples/knative/second-exp/values.yaml"
            branch: $BRANCH
            username: $USERNAME
        ```

### 6.b) Sync K8s cluster
Deploy the winning version of the `hello world` Knative app using Helm.
```shell
# curl https://raw.githubusercontent.com/$USERNAME/iter8/$BRANCH/samples/knative/second-exp/values.yaml
helm upgrade my-app iter8/knslo-gitops \
  -f https://raw.githubusercontent.com/$USERNAME/iter8/$BRANCH/samples/knative/second-exp/values.yaml \
  --install
# helm get manifest my-app
```

## 7. Cleanup
```shell
helm uninstall my-app
git push -d origin $BRANCH
```

***

**Next Steps**

!!! tip "Use in production"
    The `knslo-gitops` Helm chart is located in the `$ITER8/helm` folder. Modify the chart as needed by your application for production usage.

!!! tip "Use with ArgoCD (or any pull-based GitOps operator)"
    ArgoCD can automatically deploy and sync a Helm chart in a Git repo into a K8s cluster. See [this example](https://argoproj.github.io/argo-cd/operator-manual/cluster-bootstrapping/#helm-example) and [these details](https://argoproj.github.io/argo-cd/user-guide/helm/). Try a flavor of this tutorial with ArgoCD by placing the `knslo-gitops` chart in your Git repo. Update its `values.yaml` file in the same manner as in the tutorial.

!!! tip "Use Iter8 notifications"
    Iter8 experiments can emit notifications, for example, once the experiment reaches the `finishing` stage. The notifications can include details about winning version. See [here](../../reference/tasks/notification-http.md) for more details. You can use Iter8 notifications to trigger CI/CD workflows.

!!! tip "Use with GitHub Actions (or any push-based GitOps pipeline tool)"
    Suppose you want a GitHub Actions workflow to automatically detect changes to `values.yaml` and deploy to a K8s cluster. The following snippet shows how you can structure this workflow.
    ```yaml
    on:
      push:
        paths:
        - '/path/to/values.yaml'
        # run the jobs in the GitHub actions workflow whenever `values.yaml` is modified.
        # see https://github.com/iter8-tools/iter8/blob/master/.github/workflows/e2e-tests.yaml 
        # for examples of GitHub Actions workflows perform Helm releases.
    ```

!!! tip "Try other Iter8 Knative tutorials"
    * [SLO validation with progressive traffic shift](testing-strategies/slovalidation.md)
    * [Hybrid testing](testing-strategies/hybrid.md)
    * [Fixed traffic split](rollout-strategies/fixed-split.md)
    * [User segmentation based on HTTP headers](rollout-strategies/user-segmentation.md)