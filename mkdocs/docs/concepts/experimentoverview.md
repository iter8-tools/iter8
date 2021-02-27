---
template: overrides/main.html
---

# Experiment overview

!!! abstract "iter8 experiment"
    **iter8** defines a Kubernetes CRD called **experiment** to automate metrics-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps.

??? example "Sample experiment"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      # `sample-app` Knative service in `default` namespace is the target of this experiment
      target: default/sample-app
      # information about app versions participating in this experiment
      versionInfo:         
        # every experiment has a baseline version
        # we will name it `current`
        baseline: 
          name: current
          variables:
          # `revision` variable is used for fetching metrics from Prometheus
          - name: revision 
            value: sample-app-v1 
          # `promote` variable is used by the finish task
          - name: promote
            value: baseline
        # candidate version(s) of the app
        # there is a single candidate in this experiment 
        # we will name it `candidate`
        candidates: 
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2
          - name: promote
            value: candidate 
      criteria:
        objectives: 
        # mean latency should be under 50 milliseconds
        - metric: mean-latency
          upperLimit: 50
        # 95th percentile latency should be under 100 milliseconds
        - metric: 95th-percentile-tail-latency
          upperLimit: 100
        # error rate should be under 1%
        - metric: error-rate
          upperLimit: "0.01"
      strategy:
        # canary testing => candidate `wins` if it satisfies objectives
        testingPattern: Canary
        # progressively shift traffic to candidate, assuming it satisfies objectives
        deploymentPattern: Progressive
        weights: # fine-tune traffic increments to candidate
          # candidate weight will not exceed 75 in any iteration
          maxCandidateWeight: 75
          # candidate weight will not increase by more than 20 in a single iteration
          maxCandidateWeightIncrement: 20
        actions:
          # run tasks under the `start` action at the start of an experiment   
          start:
          # the following task verifies that the `sample-app` Knative service in the `default` namespace is available and ready
          # it then updates the experiment resource with information needed to shift traffic between app versions
          - library: knative
            task: init-experiment
          # run tasks under the `finish` action at the end of an experiment   
          finish:
          # promote an app version
          # `https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/candidate.yaml` will be applied if candidate satisfies objectives
          # `https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/baseline.yaml` will be applied if candidate fails to satisfy objectives
          - library: common
            task: exec # promote the winning version
            with:
              cmd: kubectl
              args:
              - "apply"
              - "-f"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
      duration: # 12 iterations, 20 seconds each
        intervalSeconds: 20
        iterationsPerLoop: 12
    ```

## How iter8 runs an experiment
1. iter8 determines if it is safe to start an experiment using its *target acquisition* algorithm.

2. When the experiment starts, iter8 runs tasks specified under `spec.actions.start` such as setting up or updating resources needed for the experiment.

3. During each iteration, iter8 evaluates app versions based on `spec.criteria`, determines the `winner`, and optionally shifts traffic towards the `winner`.

4. When the experiment finishes, iter8 runs tasks specified under `spec.actions.finish` such as version promotion.

## Experiment spec in-brief
A brief explanation of the key stanzas in an experiment spec is given below.

### spec.target

`spec.target` is a string that identifies the app under experimentation and determines which experiments can run concurrently.

### spec.versionInfo

`spec.versionInfo` is an object that describes the app versions involved in the experiment. Every experiment involves a `baseline` version, and may involve zero or more `candidates`.

### spec.criteria

`spec.criteria` is an object that specifies the metrics used for evaluating versions along with acceptable limits for their values.

### spec.strategy.testingPattern

`spec.strategy.testingPattern` is a string enum that determines the logic used to evaluate the app versions and determine the `winner` of the experiment. iter8 supports two testing patterns, namely, `Canary` and `Conformance`.

### spec.strategy.deploymentPattern

`spec.strategy.deploymentPattern` is a string enum that determines if and how traffic is shifted during an experiment[^1]. iter8 supports two deployment patterns, namely, `Progressive` and `FixedSplit`.

### spec.strategy.weights

`spec.strategy.weights` is an object with  two integer fields, namely, `maxCandidateWeight` and `maxCandidateWeightIncrement`, that can be used to fine-tune traffic increments to the candidate. This stanza is applicable only for `Progressive` experiments. `maxCandidateWeight` specifies the maximum candidate weight that can be set by iter8 during an iteration. `maxCandidateWeightIncrement` specifies the maximum increase in candidate weight during a single iteration.

### spec.strategy.actions

An action is a sequence of tasks executed during an experiment. `spec.strategy.actions` is an object that can be used to specify `start` and `finish` actions that will be executed at the start and end of an experiment respectively.

### spec.duration

`spec.duration` is an object with two integer fields, namely, `iterationsPerLoop` and `intervalSeconds`. The former specifies the number of iterations in the experiment. The latter specifies the time interval in seconds between successive iterations.

[^1]: Traffic shifting is relevant only when an experiment involves two or more versions. `Conformance` testing experiments involve a single version. Hence, `spec.strategy.deploymentPattern` is ignored in these experiments.