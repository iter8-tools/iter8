---
template: main.html
---

# Common Tasks

## `common/exec`

### Overview

The `common/exec` task executes a shell command with arguments. Arguments are specified using placeholders that are dynamically substituted at runtime. The `common/exec` task can be used as part of a finish action to promote the winning version at the end of an experiment.

### Example

The following (partially-specified) experiment executes `kubectl apply` using a YAML manifest at the end of the experiment. The URL for the manifest contains a placeholder `.promote`, which is dynamically substituted at the end of the experiment.

```yaml
kind: Experiment
...
spec:
  ...
  strategy:
    ...
    actions:
      finish: # run the following sequence of tasks at the end of the experiment
      - task: common/exec # promote the winning version      
        with:
          cmd: /bin/sh
          args:
          - "-c"
          - |
            kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml
    ...
  versionInfo:
    # information about app versions used in this experiment
    baseline:
      name: sample-app-v1
      variables:
      - name: promote
        value: baseline
    candidates:
    - name: sample-app-v2
      variables:
      - name: promote
        value: candidate
```

### Inputs

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| cmd | string | The command that should be executed | Yes |
| args | []string | A list of command line arguments that should be passed to `cmd`. | No |
| disableInterpolation | bool | Flag indicating whether or not to disable placeholder subsitution. For details, see [below](#disabling-placeholder-substitution). Default is `false`. | No |

### Result

The command with the supplied arguments will be executed. 

In the [example above](#example), a YAML file corresponding to the baseline or candidate version will be applied to the cluster.

If this task exits with a non-zero error code, the experiment to which it belongs will fail.

### Dynamic placeholder substitution

Inputs to tasks can contain placeholders, or template variables, which will be dynamically substituted when the task is executed by Iter8. In the [example above](#example), one input is:
```shell
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml
```
In this case, the placeholder is `{{ .promote }}`. 

Placeholder substitution in task inputs works as follows. 

Iter8 will find the version recommended for promotion which is determined by Iter8 during the  course of the experiment, and stored in the `status.versionRecommendedForPromotion` field of the experiment resource. The version recommended for promotion is the winner, if a winner has been found in the experiment. Otherwise, it is the baseline version supplied in the `spec.versionInfo` field of the experiment.

If the placeholder is `{{ .name }}`, Iter8 will substitute it with the name of the version recommended for promotion. If it is any other variable, Iter8 will substitute it with the value of the corresponding variable for the version recommended for promotion. Variable values are specified in the variables field of the version detail when the experiment is created.

### Disabling placeholder substitution

By default, the `common/exec` task will attempt to find the version recommended for promotion, and use the values defined for it (in the spec.versionInfo portion of the experiment). However, this behavior will lead to task failure when the version recommended for promotion is not available. This is usually the case when a start task is executed. To use the `common/exec` task as part of an experiment start action, set `disableInterpolation` to true.
