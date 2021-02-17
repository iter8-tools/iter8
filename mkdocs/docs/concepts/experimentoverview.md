---
template: overrides/main.html
---

# Experiment overview

> **iter8** defines a Kubernetes CRD called **experiment** to enable metrics-driven experiments, progressive delivery, and automated rollout for Kubernetes and OpenShift apps.
 <!-- Use iter8 experiments to test your app versions, observe how they perform, and rollout the best version of your app in a safe, robust and automated manner. -->

## What happens during an experiment
1. At the beginning of an experiment, iter8 runs all the tasks specified as part of the `start` action in the experiment. A typical start task is the `init-experiment` task which verifies that the entity under experimentation (for example, a Knative service) is available and ready, and expands the experiment spec to include key details about app versions like metric labels.

2. An experiment spans a specified number of iterations. During these iterations, iter8 evaluates app versions based on the criteria specified in the experiment, determines the `winning version`, and optionally shifts traffic towards the `winner`.

3. At the end of an experiment, iter8 runs all the tasks specified as part of the `finish` action in the experiment. A typical finish task is to apply the rollout manifest associated with the `winner` or rollback to the baseline if no winner is found.

??? note "Sample experiment"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: sample-experiment
    spec:
      target: knative-test/sample-app
      strategy:
        testingPattern: Canary
        deploymentPattern: Progressive
        actions:
          start:
          - library: knative
            task: init-experiment
          finish:
          - library: common
            task: exec
            with:
              cmd: kubectl
              args:
              - apply
              - -f
              - https://github.com/myapp/{{ .pathFragment }}/app.yaml
        criteria:
          objectives:
          - metric: mean-latency
            upperLimit: 2000
          - metric: error-rate
            upperLimit: "0.01"
        duration:
          intervalSeconds: 15
          iterationsPerLoop: 8
        versionInfo:
          baseline:
            name: baseline
            variables: 
            - name: revision
              value: sample-app-v1
            - name: pathFragment
              value: stable
          candidates:
          - name: candidate
            variables:
            - name: revision
              value: sample-app-v2
            - name: pathFragment
              value: latest
    ```

## Experiment setup in-brief
The sample experiment above illustrates the key aspects in the setup of an iter8 experiment. A brief explanation of these aspects along with links to in-depth descriptions is given below.

### spec.target

Target is a string that is specified as part of an experiment. It serves two purposes: 1) identify the entity under experimentation, and 2) determine which experiments can run concurrently.

In the sample experiment above, the entity under experimentation is the Knative service named `sample-app` under the `default` namespace. Hence, the target is specified as the fully qualified name (namespace/name) of the Knative service which is `default/sample-app`. 

Experiments that share the same target are deemed to conflict with each other. When an experiment is being run, iter8 will not run a second experiment with the same target concurrently; the second experiment will be started after the first one completes. Experiments with different targets can be run by iter8 concurrently.

??? note "Links to in-depth description and code samples"
    1. In-depth description of deployment patterns is [here](aspects/deployment.md).
    2. Code samples...

### spec.strategy.testingPattern

Testing pattern determines the logic used to evaluate the app versions, and determine the `winner`. iter8 supports two testing patterns, namely, `canary` and `conformance`.

- Canary: Two app versions, namely, `baseline` and `candidate`, are evaluated. Candidate is declared the winner if it satisfies experiment objectives. If candidate fails to satisfy objectives but baseline does, then baseline is declared the winner.

- Conformance: A single app version is evaluated; it is declared the winner if it satisfies experiment objectives.

The sample experiment above uses the canary testing pattern.

??? note "Links to in-depth description and code samples"
    1. In-depth description of testing patterns is [here](aspects/testing.md).
    2. Code samples...

### spec.criteria

Criteria specify the metrics used for evaluating versions and acceptable limits for their values.

The sample experiment above specifies that mean latency of versions should be under 100 milliseconds, 95th percentile tail latency should be under 150 milliseconds, and error rate should be under 0.1%.

??? note "Links to in-depth description and code samples"
    1. In-depth description of criteria is [here](aspects/criteria.md).
    2. Code samples...


### spec.strategy.deploymentPattern

Deployment pattern determines if and how traffic is shifted during a canary experiment. iter8 supports two deployment patterns, namely, `progressive` and `fixed-split`.

- Progressive: Progressively shift traffic towards the winner during each iteration of the experiment.

- Fixed-split: The traffic split set at the start of the experiment is left unchanged during iterations.

The sample experiment above uses the progressive deployment pattern.

??? note "Links to in-depth description and code samples"
    1. In-depth description of deployment patterns is [here](aspects/deployment.md).
    2. Code samples...

### spec.strategy.actions

An action is a set of tasks that can be run by iter8. You can specify `start` and `finish` actions in an experiment that will be run at the start and end of an experiment respectively.

The sample experiment above consists of start and finish actions. The start action consists of a single task, namely `init-experiment`. This task verifies that the targeted Knative service is available and ready, and populates the experiment resource with details about the app versions such as JSON paths used for specifying traffic percentages. The finish action consists of a single task, namely `exec`. This task applies the Knative service manifest corresponding to the version to be promoted at the end of the experiment. Assuming that the candidate satisfies the experiment objectives, its manifest will be applied.

??? note "Links to in-depth description and code samples"
    1. In-depth description of deployment patterns is [here](aspects/actions.md).
    2. Code samples...

### spec.duration

The `duration` of an experiment is determined by `iterationsPerLoop` and `intervalSeconds`. The former specifies the number of iterations in an experiment. The latter specifies the time interval in seconds between two iterations.

The sample experiment runs for 8 iterations, with the interval between each iteration being 15 seconds.

??? note "Links to in-depth description and code samples"
    1. In-depth description of duration is [here](aspects/duration.md).
    2. Code samples...
