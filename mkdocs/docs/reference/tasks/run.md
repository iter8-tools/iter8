---
template: main.html
---

# `run`
The `run` task executes a bash script.

## Example

The following (partially-specified) experiment executes the one line script `kubectl apply` using a YAML manifest at the end of the experiment.

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

The script will be executed. If this script exits with a non-zero error code, the `run` task and therefore the experiment to which it belongs will fail.
