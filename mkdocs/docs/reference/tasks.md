---
template: main.html
---

# Tasks

Tasks are an extension mechanism for enhancing the behavior of Iter8 experiments and can be specified within the [spec.strategy.actions](../experiment/#strategy) field of the experiment.

## `common/exec`

Iter8 currently provides a single task type called `common/exec` that helps in setting up and finishing up experiments. Use `common/exec` tasks in experiments to execute shell commands, in particular, the `kubectl`, `helm` and `kustomize` commands. Use the `exec` task as part of the `finish` action to promote the winning version at the end of an experiment. Use it as part of the `start` action to set up resources required for the experiment.

=== "kubectl"
    ``` yaml linenums="1"
    spec:
      strategy:
        actions:
          start:
          # when using common/exec in a start action, always set disableInterpolation to true
          - task: common/exec # create a K8s resource
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
                kubectl apply -f https://raw.githubusercontent.com/my/favourite/resource.yaml
              disableInterpolation: true              
          finish:
          - task: common/exec # promote the winning version
            with:
              cmd: kubectl
              args:
              - "apply"
              - "-f"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
    ```

=== "Helm"
    ``` yaml linenums="1"
    spec:
      strategy:
        actions:
          start:
          # when using common/exec in a start action, always set disableInterpolation to true
          - task: common/exec # install a helm chart
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
                helm upgrade --install --repo https://raw.githubusercontent.com/my/favorite/helm-repo app --namespace=iter8-system app
              disableInterpolation: true
          finish:
          - task: common/exec
            with:
              cmd: helm
              args:
              - "upgrade"
              - "--install"
              - "--repo"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo" # repo url
              - "sample-app" # release name
              - "--namespace=iter8-system" # release namespace
              - "sample-app" # chart name
              - "--values=https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/{{ .promote }}-values.yaml" # placeholder is substituted dynamically
    ```

=== "Kustomize"
    ``` yaml linenums="1"
    spec:
      strategy:
        actions:
          start:
          # when using common/exec in a start action, always set disableInterpolation to true
          - task: common/exec # create kubernetes resources
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
                kustomize build github.com/my/favorite/kustomize/folder?ref=master | kubectl apply -f -
              disableInterpolation: true        
          finish: # run the following sequence of tasks at the end of the experiment
          - task: common/exec # promote the winning version using kustomize
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
              kustomize build github.com/iter8-tools/iter8/samples/knative/canaryfixedsplit/{{ .name }}?ref=master | kubectl apply -f -
    ```

### Placeholder substitution in task inputs

Inputs to tasks can contain placeholders, or template variables, which will be dynamically substituted when the task is executed by Iter8. For example, in the sample experiment above, one input is:

```bash 
"https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
```

In this case, the placeholder is `{{ .promote }}`. Placeholder substitution in task inputs works as follows.

1. Iter8 will find the version recommended for promotion. This information is stored in the `status.versionRecommendedForPromotion` field of the experiment. The version recommended for promotion is the `winner`, if a `winner` has been found in the experiment. Otherwise, it is the baseline version supplied in the `spec.versionInfo` field of the experiment.

2. If the placeholder is `{{ .name }}`, Iter8 will substitute it with the name of the version recommended for promotion. Else, if it is any other variable, Iter8 will substitute it with the value of the corresponding variable for the version recommended for promotion. Variable values are specified in the `variables` field of the version detail. Note that variable values could have been supplied by the creator of the experiment, or by other tasks such as `init-experiment` that may already have been executed by Iter8 as part of the experiment.

??? example "Placeholder substitution Example 1"

    Consider the sample experiment above. Suppose the `winner` of this experiment was `candidate`. Then:
    
    1. The version recommended for promotion is `candidate`.
    2. The placeholder in the argument to the `exec` task of the `finish` action is `{{ .promote }}`.
    3. The value of the placeholder for the version recommended for promotion is `candid`.
    4. The command executed by the `exec` task is then `kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/candid.yaml`.
    
??? example "Placeholder substitution Example 2"

    Consider the sample experiment above. Suppose the `winner` of this experiment was `current`. Then:
    
    1. The version recommended for promotion is `current`.
    2. The placeholder in the argument of the `exec` task of the `finish` action is `{{ .promote }}`.
    3. The value of the placeholder for the version recommended for promotion is `base`.
    4. The command executed by the `exec` task is then `kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/base.yaml`.

??? example "Placeholder substitution Example 3"

    Consider the sample experiment above. Suppose the experiment did not yield a `winner`. Then:
    
    1. The version recommended for promotion is `current`.
    2. The placeholder in the argument of the `exec` task of the `finish` action is `{{ .promote }}`.
    3. The value of the placeholder for the version recommended for promotion is `base`.
    4. The command executed by the `exec` task is then `kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/base.yaml`.

### Disable Interpolation (always do this in a `start` action)
By default, the `common/exec` task will attempt to find the version recommended for promotion, and use its values to substitute placeholders in the inputs to the task. However, this behavior will lead to task failure since version recommended for promotion will be generally undefined at this stage of the experiment. To use the `common/exec` task as part of an experiment `start` action, set `disableInterpolation` to `true` as illustrated in the `kubectl/Helm/Kustomize` samples above.

### Error handling in tasks
When a task exits with an error, it will result in the failure of the experiment to which it belongs.
