---
template: main.html
---

# Common Tasks

## `common/exec`

### Overview

The `common/exec` task executes a shell commands with any specified arguments. Arguments are specified using placeholders, or templated variables, that are dynamically instantiated at runtime using values defined when the task is. The `common/exec` task is used in most tutorials as part of a finish action to promote the winning version at the end of an experiment.

### Examples

The following example executes `kubectl apply` using the file defined by the placeholder  `.promte` when the experiment has completed.

```yaml
- finish:
  task: common/exec
    with:
    - cmd: /bin/bash
    - args:
      - -c
      - |
        kubectl apply -f {{ .promote }}
```

### Result

The command will be executed; in this case, the yaml file will be applied to the cluster. If a task exits with a non-zero error code, the experiment to which it belongs will also fail.

### Arguments

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| cmd | string | The command that should be executed | Yes |
| args | []string | A list of command line arguments that should be passed to `cmd`. | No |
| disableInterpolation | bool | Flag indicating whether or not to disable placeholder subsitution. For details, see [below](#disabling-placeholder-substitution). Default is `false`. | No |

### Details

#### Dynamic Placeholder Substitution

Inputs to tasks can contain placeholders, or template variables, which will be dynamically substituted when the task is executed by Iter8. For example, in the sample experiment above, one input is:

    kubectl apply -f {{ .promote }}

In this case, the placeholder is `{{ .promote }}`. Placeholder substitution in task inputs works as follows.

Iter8 will find the version recommended for promotion. This information is set by Iter8 as the experiment exectuted and in stored in the `status.versionRecommendedForPromotion` field of the experiment. The version recommended for promotion is the winner, if a winner has been found in the experiment. Otherwise, it is the baseline version supplied in the `spec.versionInfo` field of the experiment.

If the placeholder is `{{ .name }}`, Iter8 will substitute it with the name of the version recommended for promotion. If it is any other variable, Iter8 will substitute it with the value of the corresponding variable for the version recommended for promotion. Variable values are specified in the variables field of the version detail. Note that variable values could have been supplied by the creator of the experiment, or by other tasks that may already have been executed by Iter8 as part of the experiment.

#### Disabling Placeholder Substitution

By default, the `common/exec` task will attempt to find the version recommended for promotion, and use the values defined for it (in the spec.versionInfo portion of the experiment). However, this behavior will lead to task failure when the version recommended for promotion is not available. This is usually the case when a start task is executed. To use the `common/exec` task as part of an experiment start action, set `disableInterpolation` to true.
