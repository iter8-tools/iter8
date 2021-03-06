---
template: overrides/main.html
---

# Iter8 API Specification

!!! abstract "Abstract"
    The Iter8 API provides two [Kubernetes custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) to automate metrics and AI-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps.

    1. The `Experiment` object provides expressive controls required by application developers and service operators who wish to automate new releases of their apps in a robust, principled and metrics-driven manner. These controls encompass [testing, deployment, traffic shaping, and version promotion functions](/concepts/experimentationstrategies/) and can be flexibly composed to automate [diverse use-cases](/code-samples/iter8-knative/canary-progressive/).
    2. The `Metric` object encapsulates the REST query that is used by Iter8 for retrieving a metric value from the metrics backend. Metrics are referenced in experiments.


!!! note "API Version"    
    This document describes version `v2alpha1` of the Iter8 API.

## Resources

### Experiment

??? info "Sample experiment"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      # target identifies the knative service under experimentation using its fully qualified name
      target: default/sample-app
      strategy:
        # this experiment will perform a canary test
        testingPattern: Canary
        actions:
          start: # run a sequence of tasks at the start of the experiment
          - library: knative
            task: init-experiment
          finish: # run the following sequence of tasks at the end of the experiment
          - library: common
            task: exec # promote the winning version
            with:
              cmd: kubectl
              args: 
              - "apply"
              - "-f"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
      criteria:
        # mean latency of version should be under 50 milliseconds
        # 95th percentile latency should be under 100 milliseconds
        # error rate should be under 1%
        objectives: 
        - metric: mean-latency
          upperLimit: 50
        - metric: 95th-percentile-tail-latency
          upperLimit: 100
        - metric: error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        # information about app versions used in this experiment
      baseline:
        name: current
        variables:
        # variables are used when querying metrics and when interpolating task inputs
        - name: revision
          value: sample-app-v1 
        - name: promote
          value: baseline
      candidates:
      - name: candidate
        variables:
        # variables are used when querying metrics and when interpolating task inputs
        - name: revision
          value: sample-app-v2
        - name: promote
          value: candidate 
    ```

#### Metadata
Standard Kubernetes [meta.v1/ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

#### Spec

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| target | string | Identifies the app under experimentation and determines which experiments can run concurrently. Experiments that have the same target value will not be scheduled concurrently but will be run sequentially in the order of their creation timestamps. Experiments whose target values differ from each other can be scheduled by Iter8 concurrently. | Yes |
| strategy | [Strategy](#strategy) | The experimentation strategy which specifies how app versions are tested, how traffic is shifted during experiment, and what tasks are executed at the start and end of the experiment. | Yes |
| criteria | [Criteria](#criteria) | Metrics used for evaluating versions along with acceptable limits for their values. | No |
| duration | [Duration](#duration) | Duration of the experiment. | No |
| versionInfo | [VersionInfo](#versioninfo) | App versions involved in the experiment. Every experiment involves a `baseline` version, and may involve zero or more `candidates`. | No |

#### Status

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| conditions | [][ExperimentCondition](#experimentcondition) | A set of conditions that express progress of an experiment. | No |
| initTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time the experiment is created. | No |
| startTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the first iteration of experiment begins  | No |
| endTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when an experiment has completed. | No |
| lastUpdateTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the status was most recently updated. | No |
| stage | string | Indicator of progress of an experiment. The stage is `Waiting` before an experiment executes its start action, `Initializing` while running the start action, `Running` while the experiment has begun its first iteration and is progressing, `Finishing` while any finish action is running and `Completed` when the experiment terminates. | No |
| currentWeightDistribution | [][WeightData](#weightdata) | Currently observed distribution of requests between app versions. | No |
| analysis | Analysis | Result of latest query to the Iter8 analytics service.  | No |
| recommendedBaseline | string | The version recommended for promotion. Although this field is populated by Iter8 even before the completion of the experiment, this field is intended to be used only on completion by the finish action. | No |
| message | string | User readable message. | No |

### Metric

Metrics are referenced within the `spec.criteria` field of the experiment. Metrics usage within experiments is described [here](/reference/metrics/using-metrics).

??? example "Sample metric"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha1
    kind: Metric
    metadata:
    name: request-count
    spec:
      params:
      - name: query
        value: sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$interval])) or on() vector(0)
      description: Number of requests
      type: counter
      provider: prometheus
    ```

#### Metadata
Standard Kubernetes [meta.v1/ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

#### Spec
| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| params | [][Param](#param) | List of name/value pairs corresponding to the name and value of the HTTP query parameters used by Iter8 when querying the metrics backend. Each name represents a parameter name; the corresponding value is a template, which will be instantiated by Iter8 at query time. For examples and more details, see [here](/reference/metrics/how-iter8-queries-metrics/).| No |
| description | string | Human-readable description of the metric. | No |
| units | string | Units of measurement. Units are used only for display purposes. | No |
| type | string | Metric type. Valid values are `counter` and `gauge`. Default value = `gauge`. | No |
| sampleSize | [MetricReference](#metricreference) | Reference to a metric that represents the number of data points over which the metric value is computed. This field applies only to `gauge` metrics. | No |
| provider | string | Type of the metrics database. Currently, `prometheus` is the only valid value. | No |

## Experiment field types

### Target

!!! abstract ""
    `spec.target` identifies the app under experimentation and determines which experiments can run concurrently.

!!! note "Target naming conventions"
    === "Knative"
        When experimenting with a single Knative service, the convention is to use the fully qualified name (namespace/name) of the Knative service as the target string. In the sample experiment above, the app under experimentation is the Knative service named `sample-app` under the `default` namespace. Hence, the target string is `default/sample-app`. 

!!! note "Concurrency policy"
    Experiments that have the same target value will **not** be scheduled concurrently but will be run sequentially in the order of their creation timestamps. Experiments whose target values differ from each other can be scheduled by Iter8 concurrently.

### Strategy

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| testingPattern | string | Determines the logic used to evaluate the app versions and determine the winner of the experiment. Iter8 supports two testing patterns, namely, `Canary` and `Conformance`. | Yes |
| deploymentPattern | string | Determines if and how traffic is shifted during an experiment. This field is relevant only for experiments using the `Canary` testing pattern. Iter8 supports two deployment patterns, namely, Progressive and FixedSplit. | No |
| actions | map[string][][TaskSpec](#taskspec) | An action is a sequence of tasks that can be executed by Iter8. spec.strategy.actions can be used to specify start and finish actions that will be run at the start and end of an experiment respectively. | No |

### TaskSpec

!!! abstract ""
    Specification of a task that will be executed as part of experiment actions. Task implementations are organized into libraries as documented [here](#task-implementations).

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| library | string | Name of library to which this task belongs. | Yes |
| task | string | Name of the task. Task names are unique within a library. | Yes |
| with | map[string][apiextensionsv1.JSON](https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON) | Inputs to the task. | No |

### Criteria

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| requestCount | string | Reference to the metric used to count the number of requests sent to app versions. | No |
| objectives | [Objective](#objective)[] | A list of metrics along with acceptable upper limits, lower limits, or both upper and lower limits for them. Iter8 will verify if app versions satisfy these objectives. | No |
| indicators | string[] | A list of metric references. Iter8 will collect and report the values of these metrics in addition to those referenced in the `objectives` section. | No |

!!! warning "" 
    **Note:** References to metric resource objects within experiment criteria should be in the `namespace/name` format or in the `name` format. If the `name` format is used (i.e., if only the name of the metric is specified), then Iter8 searches for the metric in the namespace of the experiment resource. If Iter8 cannot find the metric, then the reference is considered invalid and the experiment will terminate in a failure.

### Objective

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| metric | string | Reference to a metric resource. Also see [note on metric references](#criteria). | Yes |
| upperLimit | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | Upper limit on the metric value. If specified, for a version to satisfy this objective, its metric value needs to be below the limit. | No |
| lowerLimit | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | Lower limit on the metric value. If specified, for a version to satisfy this objective, its metric value needs to be above the limit. | No |    

### Duration

!!! abstract ""
    The duration of the experiment.

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| intervalSeconds | int32 | Duration of a single iteration of the experiment in seconds. Default value = 20 seconds. | No |
| maxIterations | int32 | Maximum number of iterations in the experiment. In case of failure, the experiment may be terminated earlier. Default value = 15. | No |

### VersionInfo

!!! abstract ""
    `spec.versionInfo` describes the app versions involved in the experiment. Every experiment involves a `baseline` version, and may involve zero or more `candidates`.

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| baseline | [VersionDetail](#versiondetail) | Details of the current or baseline version. | Yes |
| candidates | [][VersionDetail](#versiondetail) | Details of the candidate version or versions, if any. | No |

!!! note "Number of versions"
    `Conformance` experiments involve only a single version (baseline). Hence, in `Conformance` experiments, the `candidates` field of versionInfo must be omitted. A `Canary` experiment involves two versions, `baseline` and `candidate`. Hence, in `Canary` experiments, the `candidates` field must be a list of length one and must contain a single versionDetail object.[^1]

!!! note "Auto-creation of VersionInfo"
    Iter8 ships with helper tasks that can inspect an experiment resource with no or partially specified `spec.versionInfo`, automatically generate the remaining portion of `spec.versionInfo` and update the experiment with this information. See the [`init-experiment` task in the `knative` task library](#task-implementations) for an example.


### VersionDetail

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| name | string | Name of the version. | Yes |
| variables | [][Variable](#variable) | Variables are name-value pairs associated with a version. Metrics and tasks within experiment specs can contain strings with placeholders. Iter8 uses variables to interpolate these strings. | No |
| weightObjRef | [corev1.ObjectReference](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ObjectReference) | Reference to a Kubernetes resource and a field-path within the resource. Iter8 uses `weightObjRef` to get or set weight (traffic percentage) for the version. | No |

### Variable

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | name of the variable | Yes |
| value | string | value of the variable | Yes |

### ExperimentCondition

!!! abstract ""
    Conditions express aspects of the progress of an experiment. The `Completed` condition indicates whether or not an experiment has completed or not. The `Failed` condition indicates whether or not an experiment completed successfully or in failure. Finally, the `TargetAcquired` condition indicates that an experiment can proceed without interference from other experiments. Iter8 ensures that only one experiment has `TargetAcquired` set to `True` while `Completed` is set to `False`.

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| type | string | Type of condition. Valid types are `TargetAcquired`, `Completed` and `Failed`. | Yes |
| status | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ConditionStatus) | status of condition, one of `True`, `False`, or `Unknown`. | Yes |
| lastTransitionTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The last time any field in the condition was changed. | No |
| reason | string | A reason for the change in value. Reasons come from a set of reason | No |
| message | string | A user readable decription. | No |

### Analysis

!!! abstract ""
    Result of latest query to the Iter8 analytics service.

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| aggregatedMetrics | [AggregatedMetricsAnalysis](#aggregatedmetricsanalysis) | Most recently observed metric values for all metrics referenced in the experiment criteria. | No |
| winnerAssessment | [WinnerAssessmentAnalysis](#winnerassessmentanalysis) | Information about the `winner` of the experiment. | No |
| versionAssessments | [VersionAssessmentAnalysis](#versionassessmentanalysis) | For each version, a summary analysis identifying whether or not the version is satisfying the experiment criteria. | No |
| weights | [WeightsAnalysis](#weightanalysis) | Recommended weight distribution to be applied before the next iteration of the experiment. | No |

### VersionAssessmentAnalysis

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | map[string][]bool | map of version name to a list of boolean values, one for each objective specified in the experiment criteria, indicating whether not the objective is satisified. | No |


### AggregatedMetricsAnalysis

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | map[string][AggregatedMetricsData](#aggregatedmetricsdata) | Map from metric name to most recent data (from all versions) for the metric. | Yes |

### AggregatedMetricsData

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| max | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The maximum value observed for this metric accross all versions. | Yes |
| min | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The minimum value observed for this metric accross all versions. | Yes |
| data | map[string][AggregatedMetricsVersionData](#aggregatedmetricsversiondata) | A map from version name to the most recent aggregated metrics data for that version. | No |

### AggregatedMetricsVersionData

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| max | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The maximum value observed for this metric for this version over all observations. | No |
| min | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The minimum value observed for this metric for this version over all observations. | No |
| value | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The value. | No |
| sampleSize | int32 | The number of requests observed by this version. | No |


### WinnerAssessmentAnalysis

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | [WinnerAssessmentData](#winnerassessmentdata) | Details on whether or not a winner has been identified and which version if so. | No |

### WinnerAssessmentData

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| winnerFound | bool | Whether or not a winner has been identified. | Yes |
| winner | string | The name of the identified winner, if one has been found. | No |

### WeightAnalysis

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | [][WeightData](#weightdata) | List of version name/value pairs representing a recommended weight for each version | No |

### WeightData

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Version name | Yes |
| value | int32 | Percentage of traffic being sent to the version.  | Yes |

## Metric field types

### MetricReference

| Field name | Field Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| namespace | string | Namespace containing the referred metric. Defaults to the namespace of the referring metric. | No |
| name | string | Name of the referred metric. | Yes |

### Param

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Name of the HTTP query parameter. | Yes |
| value | string | Value of the HTTP query parameter. See [here](/reference/metrics/how-iter8-queries-metrics/) for documentation on how Iter8 interpolates params. | Yes |

## Tasks

Tasks are an extension mechanism for enhancing the behavior of Iter8 experiments and can be specified within the [spec.strategy.actions](#strategy) field of the experiment.

### Task implementations

Iter8 currently implements two tasks that help in setting up and finishing up experiments. These tasks are organized into the `knative` and `common` task libraries.

??? example "Sample experiment with start and finish actions with tasks"
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

#### `knative/init-experiment`

The `knative` task library provides the `init-experiment` task. Use this task as part of the `start` action when experimenting with a Knative service. This task will do the following.

1. Verify that the target Knative service resource specified in the experiment is available. The target string in the experiment must be formatted as `namespace/name` of the Knative service.[^2]

2. Verify that the target Knative service resource meets three conditions: `Ready`, `ConfigurationsReady` and `RoutesReady`.[^3]

3. Verify that `revision` information supplied for app versions in the experiment can be found in the Knative service. For example, the sample experiment above refers to two revisions, namely, `sample-app-v1` and `sample-app-v2`. The `init-experiment` task will inspect the `status.traffic` field of the target Knative service to verify that the revisions are found.

4. Add the `namespace` variable to the `spec.versionInfo` field in the experiment. The value of this variable is the namespace of the target Knative service.

5. Add `weightObjRef` clause within the `spec.versionInfo` field in the experiment.

??? info "`spec.versionInfo` before and after `init-experiment` is executed"
    === "Before"
        ``` yaml linenums="1"
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
        ``` yaml linenums="1"
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

#### `common/exec`

The `common` task library provides the `exec` task. Use this task to execute shell commands, in particular, the `kubectl`, `helm` and `kustomize` commands. Use the `exec` task as part of the `finish` action to promote the winning version at the end of an experiment. Use it as part of the `start` action to set up resources required for the experiment.

=== "kubectl"
    ``` yaml linenums="1"
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
    ``` yaml linenums="1"
    spec:
      strategy:
        actions:
          finish:
          - library: common
            task: exec
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
              - "--values=https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/{{ .promote }}-values.yaml" # values URL dynamically interpolated
    ```

=== "Kustomize"
    ``` yaml linenums="1"
    spec:
      strategy:
        actions:
          finish: # run the following sequence of tasks at the end of the experiment
          - library: common
            task: exec # promote the winning version using kustomize
            with:
              cmd: /bin/sh
              args:
              - "-c"
              - |
              kustomize build github.com/iter8-tools/iter8/samples/knative/canaryfixedsplit/{{ .name }}?ref=master | kubectl apply -f -
    ```

### Interpolation of task inputs

Inputs to tasks can container placeholders, or template variables which will be dynamically substituted when the task is executed by Iter8. Variable interpolation works as follows.

1. Iter8 will find the version recommended for promotion. This information is stored in the `status.recommendedBaseline` field of the experiment. The version recommended for promotion is the `winner`, if a `winner` has been found in the experiment. Otherwise, it is the baseline version supplied in the `spec.versionInfo` field of the experiment.

2. If the placeholder is `{{ .name }}`, Iter8 will substitute it with the name of the version recommended for promotion. Else, if it is any other variable, Iter8 will substitute it with the value of this corresponding variable for the version recommended for promotion. Note that variable values could have been supplied by the creator of the experiment, or by other tasks such as `init-experiment` that may be executed by Iter8 as part of the experiment.

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

### Task error handling
When a task exits with an error, it will result in the failure of the experiment to which it belongs.

[^1]: `A/B/n` experiments involve more than one candidate. Their description is coming soon.

[^2]: The `init-experiment` task will repeatedly attempt to find the target Knative service resource in the cluster over a period of 180 seconds. If it cannot find the service at the end of this period, it will exit with an error.

[^3]: The `init-experiment` task will repeatedly attempt to verify that the conditions are met over a period of 180 seconds. If it finds that the conditions are not met at the end of this period, it will exit with an error.

