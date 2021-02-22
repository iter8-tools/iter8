---
template: overrides/main.html
---

# spec.versionInfo

> `spec.versionInfo` describes the app versions involved in the experiment. Every experiment involves a `baseline` version, and may involve zero or more `candidates`.

??? example "Sample experiment"
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

## Schema of versionInfo
The schema of the `versionInfo` object is as follows:

``` yaml
versionInfo:
  # details of baseline version; required
  baseline: <versionDetail> 
  # details of one or more candidate versions; optional
  candidates: 
  - <versionDetail>
  - ...
```

### Number of versions
[`Conformance`](testing.md) experiments involve only a single version (baseline). Hence, in `Conformance` experiments, the `candidates` stanza of `versionInfo` must be omitted. A [`Canary`](testing.md) experiment involves two versions, baseline and a candidate. Hence, in `Canary` experiments, the `candidates` stanza must be a list of length one and must contain a single `versionDetail` object.[^1]

## Schema of versionDetail

A `versionDetail` object that describes a `baseline` version is exemplified below; `versionDetail` object for candidate versions share the same schema.

``` yaml
baseline:
  # name of the version; must be unique; required
  name: current
  # a list of variables associated with this version; optional
  # each variable is a name-value pair    
  variables:
  - name: revision 
    value: sample-app-v1 
  - name: promote
    value: baseline
  # iter8 uses weightObjRef to get and set weight (traffic percentage); optional
  weightObjRef:
    apiVersion: serving.knative.dev/v1
    kind: Service
    name: sample-app
    namespace: default
    fieldPath: /spec/traffic/0/percent  
```

### name
Each version has a unique name.

### variables
`variables` are name-value pairs associated with a version. Metrics and tasks within experiment specs can contain strings with placeholders. iter8 uses `variables` to interpolate these strings.

### weightObjRef
weightObjRef contains a reference to a Kubernetes resource and a field-path within the resource. iter8 uses weightObjRef to get or set weight (traffic percentage) for the version.

## Auto-creation of versionInfo

iter8 ships with helper tasks that can inspect an experiment resource with no or partially specified `spec.versionInfo`, automatically generate the remaining portion of `spec.versionInfo` and update the experiment with this information. See the [`init-experiment` task in the `knative` task library](actions.md) for an example.

[^1]: `A/B/n` experiments involve more than one candidate. Their description is coming soon.




