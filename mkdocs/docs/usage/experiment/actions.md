---
template: overrides/main.html
---

# Actions

> An action is a sequence of tasks that can be executed by iter8. `spec.strategy.actions` can be used to specify `start` and `finish` actions that will be run at the start and end of an experiment respectively.

??? example "Sample experiment with start and finish actions"
    ```yaml
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
            value: base
        # candidate version(s) of the app
        # there is a single candidate in this experiment 
        # we will name it `candidate`
        candidates: 
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2
          - name: promote
            value: candid
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
      indicators:
      # report values for the following metrics in addition those in spec.criteria.objectives
      - 99th-percentile-tail-latency
      - 90th-percentile-tail-latency
      - 75th-percentile-tail-latency
      strategy:
        # canary testing => candidate `wins` if it satisfies objectives
        testingPattern: Canary
        # progressively shift traffic to candidate, assuming it satisfies objectives
        deploymentPattern: Progressive
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

iter8 currently implements two tasks that help in setting up and finishing up experiments. These tasks are organized into the `knative` and `common` task libraries.

## Tasks

### `knative/init-experiment`

The `knative` task library provides the `init-experiment` task. Use this task as part of the `start` action when experimenting with a Knative service. This task will do the following.

1. Verify that the target Knative service resource specified in the experiment is available. The target string in the experiment must be formatted as `namespace/name` of the Knative service.[^1]

2. Verify that the target Knative service resource meets three conditions: `Ready`, `ConfigurationsReady` and `RoutesReady`.[^2]

3. Verify that `revision` information supplied for app versions in the experiment can be found in the Knative service. For example, the sample experiment above refers to two revisions, namely, `sample-app-v1` and `sample-app-v2`. The `init-experiment` task will inspect the `status.traffic` stanza of the target Knative service to verify that the revisions are found.

4. Add the `namespace` variable to the `spec.versionInfo` stanza in the experiment. The value of this variable is the namespace of the target Knative service.

5. Add `weightObjRef` clause within the `spec.versionInfo` stanza in the experiment.

??? info "`spec.versionInfo` before and after `init-experiment` is executed"
    === "Before"
        ``` yaml
        versionInfo:         
          baseline: 
            name: current
            variables:
            - name: revision 
              value: sample-app-v1 
            - name: promote
              value: baseline
          candidates: 
          - name: candidate
            variables:
            - name: revision
              value: sample-app-v2
            - name: promote
              value: candidate 
        ```

    === "After"
        ``` yaml
        versionInfo:         
          baseline: 
            name: current
            variables:
            - name: revision 
              value: sample-app-v1 
            - name: promote
              value: base
            - name: namespace
              value: default
            weightObjRef:
              apiVersion: serving.knative.dev/v1
              kind: Service
              name: sample-app
              namespace: default
              fieldPath: /spec/traffic/0/percent  
          candidates: 
          - name: candidate
            variables:
            - name: revision
              value: sample-app-v2
            - name: promote
              value: candid
            - name: namespace
              value: default
            weightObjRef:
              apiVersion: serving.knative.dev/v1
              kind: Service
              name: sample-app
              namespace: default
              fieldPath: /spec/traffic/1/percent  
        ```

### `common/exec`

The `common` task library provides the `exec` task. Use this task to execute shell commands, in particular, the `kubectl`, `helm` and `kustomize` commands. Use the `exec` task as part of the `finish` action to promote the winning version at the end of an experiment. Use it as part of the `start` action to set up resources required for the experiment.

=== "kubectl"
    ``` yaml
    spec:
      strategy:
        actions:
          finish:
          - library: common
            task: exec # promote the winning version
            with:
              cmd: kubectl
              args:
              - "apply"
              - "-f"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
    ```


=== "Helm"

=== "Kustomize"

## Interpolation of task inputs

Inputs to tasks can container placeholders, or template variables which will be dynamically substituted when the task is executed by iter8. Variable interpolation works as follows.

1. iter8 will find the version recommended for promotion. This information is stored in the `status.recommendedBaseline` stanza of the experiment. The version recommended for promotion is the `winner`, if a `winner` has been found in the experiment. Otherwise, it is the baseline version supplied in the `spec.versionInfo` stanza of the experiment.

2. If the placeholder is `{{ .name }}`, iter8 will substitute it with the name of the version recommended for promotion. Else, if it is any other variable, iter8 will substitute it with the value of this corresponding variable for the version recommended for promotion. Note that variable values could have been supplied by the creator of the experiment, or by other tasks such as `init-experiment` that may be executed by iter8 as part of the experiment.

??? example "Interpolation Example 1"

    Consider the sample experiment above. Suppose the `winner` of this experiment was `candidate`. Then:
    
    1. The version recommended for promotion is `candidate`.
    2. The placeholder in the `exec` task in the `finish` action is `{{ .promote }}`.
    3. The value of the placeholder for the version recommended for promotion is `candid`.
    4. The command executed by the `exec` task is `kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/candid.yaml`.
    
??? example "Interpolation Example 2"

    Consider the sample experiment above. Suppose the `winner` of this experiment was `current`. Then:
    
    1. The version recommended for promotion is `current`.
    2. The placeholder in the `exec` task in the `finish` action is `{{ .promote }}`.
    3. The value of the placeholder for the version recommended for promotion is `base`.
    4. The command executed by the `exec` task is `kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/base.yaml`.

??? example "Interpolation Example 3"

    Consider the sample experiment above. Suppose the experiment did not yield a `winner`. Then:
    
    1. The version recommended for promotion is `current`.
    2. The placeholder in the `exec` task in the `finish` action is `{{ .promote }}`.
    3. The value of the placeholder for the version recommended for promotion is `base`.
    4. The command executed by the `exec` task is `kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/base.yaml`.

## Task failure
When a task exits with a failure, it will result in the failure of the experiment to which it belongs.

[^1]: The `init-experiment` task will repeatedly attempt to find the target Knative service resource in the cluster over a period of 180 seconds. If it cannot find the service at the end of this period, it will exit with a failure.

[^2]: The `init-experiment` task will repeatedly attempt to verify that the conditions are met over a period of 180 seconds. If it finds that the conditions are not met at the end of this period, it will exit with a failure.