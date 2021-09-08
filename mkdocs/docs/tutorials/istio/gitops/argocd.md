---
template: main.html
---

# GitOps with Argo CD

!!! tip "Scenario: Iter8 Experiment + Gitops"
    GitOps is an approach increasingly adopted to simplify cluster management tasks.
    In this approach, the desired state of one or more clusters is kept in a Git *environment* repository.
    A CD tool, such as Argo CD, continuously monitors this repository for changes and synchronizes any detected changes to the target clusters.
    In the wider context, commits to the code repository trigger a CI pipeline to, for example, lint, build, test, and push newly built images to an image repository. It then writes configuration changes to the environment repository. The CD pipeline or tool then detects these changes and synchronizes them to the target clusters.
    Iter8 can be used in the context of GitOps so that new versions of an application can be progressively rolled out, or even rolled back when problems are detected. To do this, configuration to create an experiment is implemented in the CI pipeline.

    ![CICD+Iter8](../../../images/CICD+Iter8.png)

This tutorial assumes a basic understanding of Iter8. See, for example, the Istio [quick start tutorial](../quick-start.md).

??? warning "Setup the environment repository"
    In this tutorial, a fork of the [iter8 repository](https://github.com/iter8-tools/iter8) is used as the environment repository. To make changes to it, you will need your own [fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo) of the repository. For the purpose of this tutorial, we assume that your fork is at: https://github.com/[YOUR_ORG]/iter8

    Clone the forked repository and modify it to replace the generic `MY_ORG` with *YOUR_ORG*.

    * *Clone the forked repository*:

    ```shell
    git clone git@github.com:kalantar/iter8.git
    cd iter8
    export ITER8=$(pwd)
    ```

    * *Modify the clone*:

    === "MacOS"
        ```shell
        export YOUR_ORG=fill-in
        find $ITER8/samples/istio/gitops -name "*" -type f | xargs sed -i '' "s/MY_ORG/$YOUR_ORG/"
        git commit -a -m "update references"
        git push origin head
        ```
    === "Linux"
        ```shell
        export YOUR_ORG=fill-in
        find $ITER8/samples/istio/gitops -name "*" -type f | xargs sed -i "s/MY_ORG/$YOUR_ORG/"
        git commit -a -m "update references"
        git push origin head
        ```

??? warning "Setup a K8s cluster with Iter8, Istio and Argo CD"
    1. Setup [K8s cluster](../../../getting-started/setup-for-tutorials.md#local-kubernetes-cluster)
    2. [Install Iter8 in K8s cluster](../../../getting-started/install.md)
    3. Get [`iter8ctl`](../../../getting-started/install.md#get-iter8ctl)
    5. [Install Istio in K8s cluster](../setup-for-tutorials.md#install-istio)
    4. [Install Prometheus add-on](../setup-for-tutorials.md#install-optional-prometheus-add-on)
    6. [Install Argo CD in k8s cluster](../setup-for-tutorials.md#install-argo-cd)
    7. [Setup a GitHub token](../setup-for-tutorials.md#create-github-token) and give Iter8 permission to use it.


## 1. Create Baseline Version

Install the `bookinfo` application by creating an Argo CD application as follows:

```shell
kubectl apply -f $ITER8/samples/istio/gitops/argocd-app.yaml
```

The GitOps application defined by `argocd-app.yaml` identifies a part of your repository, https://github.com/[YOUR_ORG]/iter8, to be the environment repository.
Argo CD will immediately begin to synchronize the configuration (bookinfo application config) found there.
You can monitor the progress of the syncrhonization through the Argo CD UI.
When the state is both `Healthy` and `Synced`, it is ready; this might take a few minutes.


## 2. Create Candidate Version

When changes are merged into a code repository, a CI pipeline to, for example, lint, build, test, and push newly built images to an image repository runs.
It then writes configuration changes to the environment repository indicating changes are needed to the deployed application.
In this tutorial, we simulate the execution of a CI pipeline, by executing a simplified GitHub workflow: https://github.com/[YOUR_ORG]/iter8/actions/workflows/gitops-ci.yaml

Navigate to the workflow and click the button "Run workflow"

??? info "More about GitHub workflow gitops-ci.yaml"
    The "Simulate CI pipeline" creates the configuration changes that a typical CI pipeline would create when changes are made.
    In particular, it:

    1. Creates a new application configiration (by modifying a color property in a random way).
    2. Creates an Iter8 experiment to evaluate the new version.
    3. Configures load generation.

    These changes are pushed to the environment repository triggering Argo CD to deploy them.

You can use the Argo CD UI to monitor progress of the deployment.
By default, Argo CD is configured to run every three minutes. If you don't want to wait, manually refresh the application so that the changes are immediately synced to the cluster.

## 3. Observe experiment

The experiment should run for a few minutes once it starts, and you can track its progress with this command:

```shell
kubectl get experiments.iter8.tools --watch
```

More detail is available using the Iter8 CLI:

```shell
iter8ctl describe
```

## 4. Promote winner

Once the experiment finishes, review the pull requests on the environment repository, https://github.com/[YOUR_ORG]/iter8/pulls. Iter8 will have created a new pull request titled `Deploy version recommended by Iter8`.
Changes to `productpage.yaml` capture the recommended version changes.
Other changes remove artifacts specific to the experiment including the candidate version, the experiment and any load generation.

Merge the pull request to complete the experiment. Argo CD will detect the changes and synchronize the cluster to the new desired state.

??? info "How Iter8 creates a pull request"
    Iter8 creates a pull request at the end of the experiment by running a GitHub workflow.
    Iter8 uses a [notification/http] task(../../../../reference/tasks/notification-http) to execute a GitHub workflow at the end of the experiment. This workflow creates the needed pull request.


## 5. Cleanup
```shell
kubectl delete -f $ITER8/samples/istio/gitops/
kubectl delete ns istio-system
kubectl delete ns iter8-system
kubectl delete ns argocd
```


## Additional details

### Environment repository

It is generally considered good practice to use different Git repositories for code and for environment configuration. In cases where the same repository is being used for both, one needs to be careful when configuring CI/CD pipeline tools so code changes can be differentiated from configuration changes, so that one doesn't inadvertently create infinite loops.

The environment repository can be organized in many different ways. With tools such as Helm and Kustomize becoming widely used, it becomes even simpler for CI pipeline tools to update an environment repository to roll out new app versions. In this tutorial, we consciously decided to use the simplest directory structure (i.e., all YAML files within a single directory without subdirectories) without the use of any higher level templating tools. Adapting the basic directory structure to Helm/Kustomize is straight forward.

When organizing the directory structure, one needs to keep in mind that the CI pipeline tool will be creating new resources in the environment repository to start an Iter8 experiment. And when the experiment finishes, Iter8 (specifically, Iter8 tasks) will delete the added resources and update the baseline version in the environment repository. In other words, the invariant here is the directory structure, which should stay the same before and after an experiment.

### GitOps support for multiple environments

Some users might use GitOps to manage multiple environments, e.g., dev, staging, prod, so changes can always propagate from environment to environment, minimizing the chance of defects from reaching the prod environment. In this case, the Iter8 GitOps task would need to be modified so that environment repository changes are done at the correct places. For example, if different environments are managed by different environment repositories, the task would need to make multiple git commits, one for each of the respositories. This could be done all within a single task, or across multiple tasks.

### GitOps Guarantees
    
Unlike other progressive delivery tools, Iter8 adheres to GitOps' guarantees by ensuring the actual state is always in sync with the desired state. App versions that fail promotion criteria will never get promoted, even if the cluster has to be recreated from scratch. This important GitOps property is often not guaranteed by other tools!

### Caveats

1. Both CI pipeline tools and Iter8 need to write to the environment repository. If not coordinated, race conditions can occur. In this tutorial, we assume repo changes are done via pull requests, which is a common practice, so the chance of having a race condition is minimized, if not eliminated. However, other means to coordinate writes to the environment repository by different entities can be done so Iter8 can operate in fully automated pipelines.

2. When a new app version becomes available while an experiment is still running, Iter8 will preempt the existing experiment with the new one. We currently don't support `test-every-commit` behavior by queuing new experiments, but this could be supported in the future if it turrned out to be more common than we are currently expecting.

3. Iter8 task could fail, just like everything else. Iter8 tasks are currently `fail-stop` without retries. Please take this into account when writing Iter8 tasks and error handling code.
