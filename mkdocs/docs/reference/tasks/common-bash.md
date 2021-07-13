---
template: main.html
---

# `common/bash`
The `common/bash` task executes a bash script. The script can be written to use placeholders that are [dynamically substituted at runtime](#dynamic-variable-substitution). For example, the `common/bash` task can be used as part of a finish action to promote the winning version at the end of an experiment.

## Example

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

## Inputs

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| script | string | The bash script that will be executed | Yes |

## Result

The script will be executed.

In the example above, a YAML file corresponding to the baseline or candidate version will be applied to the cluster.

If this task exits with a non-zero error code, the experiment to which it belongs will fail.

## Dynamic Variable Substitution
The script input to the `common/bash` task may contain placeholders, or template variables, which will be dynamically substituted when the task is executed by Iter8. For example, in the task:

```bash
- task: common/bash # promote the winning version      
  with:
    script: |
        kubectl apply -f {{ .promoteManifest }}
```

`{{ .promotionManifest}}` is a placeholder.

Placeholders are specified using the Go language specification for data-driven [templates](https://golang.org/pkg/html/template/). In particular, placeholders are specified between double curly braces.

The `common/bash` task supports placeholders for:

- Values of variables of the version recommended for promotion. To specify such placeholders, use the name of the variable as defined in the [`versionInfo` section](../experiment.md#versioninfo) of the experiment definition. For example, in the above example, `{{ .promotionManifest }}` is a placeholder for the value of the variable with the name `promotionManifest` of the version Iter8 recommends for promotion (see [`.status.versionRecommendedForPromotion`](../experiment.md#status)).

- Values defined in the experiment itself. To specify such placeholders, use the prefix `.this`. For example, `{{ .this.metadata.name }}` is a placeholder for the name of the experiment.

If Iter8 cannot evaluate a placeholder expression, a blank value ("") will be substituted.