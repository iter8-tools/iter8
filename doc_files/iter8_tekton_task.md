# Automating canary releases with iter8 driven by Tekton

[Tekton Pipelines](https://github.com/tektoncd/pipeline/tree/master/docs) are an open source implementation of mechanisms to configure and run CI/CD style pipelines for a Kubernetes application. Custom resources are used to define pipelines and their building blocks.

This tutorial provides a sample task that drives a canary release with iter8. We can see how this task could be integrated into a pipeline that builds a new version, deploys it and drives a canary release of that new version.

This tutorial is similar to steps 1 - 5 of the tutorial [Successful canary release](iter8_bookinfo_istio.md#part-1-successful-canary-release-reviews-v2-to-reviews-v3). A Tekton task is used to replace step 4.

## Part I: Setup

If not already installed, install _iter8_ using [these instructions](iter8_install.md).

### 1. Install Tekton

Install Tekton using the instructions [here](https://github.com/tektoncd/pipeline/blob/master/docs/install.md).

### 2. Deploy the Bookinfo application and start load

Deploy the Bookinfo application by following [steps 1 - 3](iter8_bookinfo_istio.md#1-deploy-the-bookinfo-application) of the tutorial [Successful canary release](iter8_bookinfo_istio.md#part-1-successful-canary-release-reviews-v2-to-reviews-v3).

## Part II: Use Tekton to Start Canary Rollout

A Tekton task is a logical step in a pipeline that will be executed repeated as each new version of code is released. While strictly not necessary for demonstration purposes, we will write a task that relies on a git project that contains, in addition to the source code for the service, a template of the iter8 `Experiment` that should be executed when rolling out a new version of the service.

It will therefore be necessary to define a `PipelineResource` that identifies the git project in addition to a `Task`. Finally, a `TaskRun` can be defined that executes the task.

The YAML files used in this tutorial are in [this repository](https://github.com/iter8-tools/docs). Clone it:

    git clone git@github.com:iter8-tools/docs.git

### 1. Create a PipelineResource

We start by defining a `PipelineResource` to the git project for the Bookinfo reviews service. You can reference the project we created from [the Istio source](https://github.com/istio/istio/tree/master/samples/bookinfo/src/reviews) or you can duplicate it to make changes. In this case, if authorization is necessary, see the Tekton documentation for [Authentication](https://github.com/tektoncd/pipeline/blob/master/docs/auth.md).
You can reference the project [bookinfoapp-reviews](https://github.com/iter8-tools/bookinfoapp-reviews) or duplicate it to make changes.

    apiVersion: tekton.dev/v1alpha1
    kind: PipelineResource
    metadata:
    name: reviews-git
    spec:
    type: git
    params:
        - name: revision
        value: master
        - name: url
        value: https://github.com/iter8-tools/bookinfoapp-reviews

You can apply this with the following command:

    kubectl apply -f docs/doc_files/tekton/reviews-pipelineresource.yaml

### 2. Import Tekton `Task`

A Tekton `Task` that initiates a canary rollout using iter8 defines an `Experiment` from a template and applies it to the cluster. Such a task can be defined as follows:

    apiVersion: tekton.dev/v1alpha1
    kind: Task
    metadata:
    name: run-experiment
    spec:
    inputs:
        resources:
        - name: source
            type: git
        params:
        - name: experiment
            description: Path to experiment template (relative to source-git)
            default: experiment
        - name: experiment-id
            description: unique identifier of experiment 
            default: default
        - name: stable
            description: stable (default) version of service being tested
            default: stable
        - name: candidate
            description: candidate version of service being tested 
            default: candidate
        - name: target-namespace
            description: namespace in which to apply routing config
            default: default
    steps:
        - name: define-experiment
        image: mikefarah/yq
        command: [ "/bin/sh" ]
        args:
            - '-c'
            - |
            TEMPLATE="/workspace/source-git/${inputs.params.experiment}"
            NAME=$(yq read ${TEMPLATE} metadata.name)
            yq write --inplace ${TEMPLATE} metadata.name ${NAME}-${inputs.params.experiment-id}
            yq write --inplace ${TEMPLATE} spec.targetService.baseline ${inputs.params.stable}
            yq write --inplace ${TEMPLATE} spec.targetService.candidate ${inputs.params.candidate}
            cat ${TEMPLATE}
        - name: apply
        image: lachlanevenson/k8s-kubectl
        command: [ "kubectl" ]
        args:
            - "--namespace"
            - "${inputs.params.target-namespace}"
            - "apply"
            - "--filename"
            - "/workspace/source-git/${inputs.params.experiment}"

This task takes the following inputs:

- _source_git_ - reference to a github project containing the `Experiment` template (typically the same repository as the source code)
- _experiment_ - path to the `Experiment` template relative to the github project
- _experiment-id_ - a unique identifier for the experiment execution to allow unique naming when executed repeatedly
- _stable_ - the stable or baseline version of the service
- _candidate_ - the candidate version of the service
- _target-namespace_ - Kubernetes namespace into which the experiment should be run; that is, the namespace where the application is deployed

And has two steps:

- _define-experiment_: Assign a name and identify the baseline and candidate versions to define the `Experiment` to run.
- _apply_: apply the new `Experiment` to the cluster. This initiates an iter8 managed canary rollout.

To add to your cluster, apply as follows:

    kubectl apply -f docs/doc_files/tekton/iter8-task.yaml

### 3. Run the `Task` by defining a `TaskRun`

To run a task, define a `TaskRun` such as:

    apiVersion: tekton.dev/v1alpha1
    kind: TaskRun
    metadata:
    name: rollout-reviews-v3
    spec:
    taskRef:
        name: run-experiment
    inputs:
        resources:
        - name: source
            resourceRef:
            name: reviews-git
        params:
        - name: experiment
            value: 'iter8/experiment.yaml'
        - name: stable
            value: 'reviews-v2'
        - name: candidate
            value: 'reviews-v3'
        - name: target-namespace
            value: 'bookinfo-iter8'

You can apply this by:

    kubectl apply -f docs/doc_files/tekton/run-iter8-task.yaml

### 4. Deploy the canary version

Follow [step 5](iter8_bookinfo_istio.md#5-deploy-the-canary-version-and-start-the-rollout) of the tutorial [Successful canary release](iter8_bookinfo_istio.md#part-1-successful-canary-release-reviews-v2-to-reviews-v3).
You can watch the progress of the canary rollout using:

    kubectl get experiments -n bookinfo-iter8
