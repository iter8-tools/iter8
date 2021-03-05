---
template: overrides/main.html
---

# Experiment Resource Object

!!! abstract ""
    Iter8 defines a Kubernetes custom resource kind called `Experiment` to automate metrics-driven experiments, progressive delivery, and rollout of Kubernetes and OpenShift apps. This document describes Experiment version `v2alpha1`. Experiment resource objects  are reconciled by Iter8's `etc3` controller. For documentation on etc3 and the Go client for `Experiment` API, see [here](https://pkg.go.dev/github.com/iter8-tools/etc3@v0.1.14/).

## ExperimentSpec

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| target | string | Identifies the app under experimentation and determines which experiments can run concurrently. Experiments that have the same target value will not be scheduled concurrently but will be run sequentially in the order of their creation timestamps. Experiments whose target values differ from each other can be scheduled by Iter8 concurrently. | Yes |
| strategy | [Strategy](#strategy) | The experimentation strategy which specifies how app versions are tested, how traffic is shifted during experiment, and what tasks are executed at the start and end of the experiment. | Yes |
| criteria | [Criteria](#criteria) | Metrics used for evaluating versions along with acceptable limits for their values. | No |
| duration | [Duration](#duration) | Duration of the experiment. | No |
| versionInfo | [VersionInfo](#versioninfo) | App versions involved in the experiment. Every experiment involves a `baseline` version, and may involve zero or more `candidates`. | No |

### Strategy

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| testingPattern | string | Determines the logic used to evaluate the app versions and determine the winner of the experiment. Iter8 supports two testing patterns, namely, `Canary` and `Conformance`. | Yes |
| deploymentPattern | string | Determines if and how traffic is shifted during an experiment. This field is relevant only for experiments using the `Canary` testing pattern. Iter8 supports two deployment patterns, namely, Progressive and FixedSplit. | No |
| actions | map[string][][TaskSpec](#taskspec) | An action is a sequence of tasks that can be executed by Iter8. spec.strategy.actions can be used to specify start and finish actions that will be run at the start and end of an experiment respectively. | No |

#### TaskSpec

!!! abstract ""
    Specification of a task that will be executed as part of experiment actions. Tasks are organized into libraries as documented [here](http://localhost:8000/usage/experiment/actions/#tasks). Tasks and task libraries are implemented by Iter8's [handler repo](https://github.com/iter8-tools/handler) which is documented [here](https://pkg.go.dev/github.com/iter8-tools/handler).

 Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| library | string | Name of library to which this task belongs. | Yes |
| task | string | Name of the task. Task names are unique within a library. | Yes |
| with | map[string][apiextensionsv1.JSON](https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON) | Inputs to the task. | No |

### Criteria

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| requestCount | string | Reference to the metric used to count the number of requests sent to app versions. | No |
| objectives | [Objective](#objective)[] | A list of metrics along with acceptable upper limits, lower limits, or both upper and lower limits for them. Iter8 will verify if app versions satisfy these objectives. | No |
| indicators | string[] | A list of metric references. Iter8 will collect and report the values of these metrics in addition to those referenced in the `objectives` section. | No |

!!! warning "" 
    **Note:** References to metric resource objects within experiment criteria should be in the `namespace/name` format or in the `name` format. If the `name` format is used (i.e., if only the name of the metric is specified), then Iter8 searches for the metric in the namespace of the experiment resource. If Iter8 cannot find the metric, then the reference is considered invalid and the experiment will terminate in a failure.

#### Objective

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| metric | string | Reference to a metric resource. Also see [note on metric references](#criteria). | Yes |
| upperLimit | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | Upper limit on the metric value. If specified, for a version to satisfy this objective, its metric value needs to be below the limit. | No |
| lowerLimit | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | Lower limit on the metric value. If specified, for a version to satisfy this objective, its metric value needs to be above the limit. | No |    

### Duration

!!! abstract ""
    The duration of the experiment.

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| intervalSeconds | int32 | Duration of a single iteration of the experiment in seconds. Default value = 20 seconds. | No |
| maxIterations | int32 | Maximum number of iterations in the experiment. In case of failure, the experiment may be terminated earlier. Default value = 15. | No |

### VersionInfo

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| baseline | [VersionDetail](#versiondetail) | Details of the current or baseline version. | Yes |
| candidates | [][VersionDetail](#versiondetail) | Details of the candidate version or versions, if any. | No |

!!! note ""
    `Conformance` experiments involve only a single version (baseline). Hence, in `Conformance` experiments, the `candidates` field of versionInfo must be omitted. A `Canary` experiment involves two versions, `baseline` and `candidate`. Hence, in `Canary` experiments, the `candidates` field must be a list of length one and must contain a single versionDetail object.

#### VersionDetail

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| name | string | Name of the version. | Yes |
| variables | [][Variable](#variable) | Variables are name-value pairs associated with a version. Metrics and tasks within experiment specs can contain strings with placeholders. Iter8 uses variables to interpolate these strings. | No |
| weightObjRef | [corev1.ObjectReference](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ObjectReference) | Reference to a Kubernetes resource and a field-path within the resource. Iter8 uses `weightObjRef` to get or set weight (traffic percentage) for the version. | No |

##### Variable

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | name of the variable | Yes |
| value | string | value of the variable | Yes |


## ExperimentStatus

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| conditions | [][ExperimentCondition](#experimentcondition) | A set of conditions that express progress through an experiment. | No |
| initTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time the experiment is created. | No |
| startTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the first iteration of experiment begins  | No |
| endTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when an experiment has completed. | No |
| lastUpdateTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the status was most recently updated. | No |
| stage | string | Indicator of progress of an experiment. The stage is `Waiting` before an experiment executes its start action, `Initializing` while running the start action, `Running` while the experiment has begun its first iteration and is progressing, `Finishing` while any finish action is running and `Completed` when the experiment terminates. | No |
| currentWeightDistribution | [][WeightData](#weightdata) | Currently observed distribution of requests between app versions. | No |
| analysis | Analysis | Result of latest query to the Iter8 analytics service.  | No |
| recommendedBaseline | string | The version recommended for promotion. Although this field is populated by Iter8 even before the completion of the experiment, this field is intended to be used only on completion by the finish action. | No |
| message | string | User readable message. | No |

### ExperimentCondition

!!! abstract ""
    Conditions express aspects of the progress of an experiment. The `Completed` condition indicates whether or not an experiment has completed or not. The `Failed` condition indicates whether or not an experiment completed successfully or in failure. Finally, the `TargetAcquired` condition indicates that an experiment can proceed without interference from other experiments. Iter8 ensures that only one experiment has `TargetAcquired` set to `True` while `Completed` is set to `False`.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| type | string | Type of condition. Valid types are `TargetAcquired`, `Completed` and `Failed`. | Yes |
| status | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ConditionStatus) | status of condition, one of `True`, `False`, or `Unknown`. | Yes |
| lastTransitionTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The last time any field in the condition was changed. | No |
| reason | string | A reason for the change in value. Reasons come from a set of reason | No |
| message | string | A user readable decription. | No |

### Analysis

!!! abstract ""
    Result of latest query to the Iter8 analytics service.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| aggregatedMetrics | [AggregatedMetricsAnalysis](#aggregatedmetricsanalysis) | Most recently observed metric values for all metrics referenced in the experiment criteria. | No |
| winnerAssessment | [WinnerAssessmentAnalysis](#winnerassessmentanalysis) | Information about the `winner` of the experiment. | No |
| versionAssessments | [VersionAssessmentAnalysis](#versionassessmentanalysis) | For each version, a summary analysis identifying whether or not the version is satisfying the experiment criteria. | No |
| weights | [WeightsAnalysis](#weightanalysis) | Recommended weight distribution to be applied before the next iteration of the experiment. | No |

##### VersionAssessmentAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | map[string][]bool | map of version name to a list of boolean values, one for each objective specified in the experiment criteria, indicating whether not the objective is satisified. | No |


#### AggregatedMetricsAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | map[string][AggregatedMetricsData](#aggregatedmetricsdata) | Map from metric name to most recent data (from all versions) for the metric. | Yes |

##### AggregatedMetricsData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| max | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The maximum value observed for this metric accross all versions. | Yes |
| min | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The minimum value observed for this metric accross all versions. | Yes |
| data | map[string][AggregatedMetricsVersionData](#aggregatedmetricsversiondata) | A map from version name to the most recent aggregated metrics data for that version. | No |

###### AggregatedMetricsVersionData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| max | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The maximum value observed for this metric for this version over all observations. | No |
| min | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The minimum value observed for this metric for this version over all observations. | No |
| value | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The value. | No |
| sampleSize | int32 | The number of requests observed by this version. | No |


#### WinnerAssessmentAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | [WinnerAssessmentData](#winnerassessmentdata) | Details on whether or not a winner has been identified and which version if so. | No |

##### WinnerAssessmentData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| winnerFound | bool | Whether or not a winner has been identified. | Yes |
| winner | string | The name of the identified winner, if one has been found. | No |

#### WeightAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | [][WeightData](#weightdata) | List of version name/value pairs representing a recommended weight for each version | No |

##### WeightData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Version name | Yes |
| value | int32 | Percentage of traffic being sent to the version.  | Yes |

