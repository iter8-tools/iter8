---
template: main.html
---

# Common Tasks

## `common/bash`

### Overview

The `common/bash` task executes a bash script. The script can be written to use placeholders that are [dynamically substituted at runtime](../../tasks#dynamic-variable-substitution). For example, the `common/bash` task can be used as part of a finish action to promote the winning version at the end of an experiment.

### Example

The following (partially-specified) experiment executes the one line script `kubectl apply` using a YAML manifest at the end of the experiment. The URL for the manifest contains a placeholder `{{ .promotionManifest }}`, which is dynamically substituted at the end of the experiment.

```yaml
kind: Experiment
...
spec:
  ...
  strategy:
    ...
    actions:
      finish: # run the following sequence of tasks at the end of the experiment
      - task: common/bash # promote the winning version      
        with:
          script: |
            kubectl apply -f {{ .promotionManifest }}
    ...
  versionInfo:
    # information about app versions used in this experiment
    baseline:
      name: sample-app-v1
      variables:
      - name: promotionManifest
        value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/baseline.yaml
    candidates:
    - name: sample-app-v2
      variables:
      - name: promotionManifest
        value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/candidate.yaml
```

### Inputs

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| script | string | The bash script that will be executed | Yes |

### Result

The script will be executed.

In the [example above](#example), a YAML file corresponding to the baseline or candidate version will be applied to the cluster.

If this task exits with a non-zero error code, the experiment to which it belongs will fail.

## `common/exec` (Deprecated)

### Overview

The `common/exec` task executes a shell command with arguments. Arguments may be specified using placeholders that are dynamically substituted at runtime. The `common/exec` task can be used as part of a finish action to promote the winning version at the end of an experiment.

`common/exec` is deprecated; use `common/bash` instead.

### Example

The following (partially-specified) experiment executes `kubectl apply` using a YAML manifest at the end of the experiment. The URL for the manifest contains a placeholder `{{ .promote }}`, which is dynamically substituted at the end of the experiment.

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

### Dynamic Variable Substitution

The `common/exec` task only supports [dynamic variable subsitution](../../tasks#dynamic-variable-substitution) for variables of the version recommended for promotion.

Instead of defaulting to a blank value when Iter8 has not determined a version to recommend for promotion (that is, in start tasks), this task supports the `disableInterpolation` option to prevent dynamic variable substitution.
