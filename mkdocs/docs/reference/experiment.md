---
template: overrides/main.html
---

# Experiment Resource Object

Fields in an iter8 experiment resource object are documented here. Unsupported fields, or those reserved for future, are not documented here. For complete documentation, see the iter8 Experiment API [here](https://pkg.go.dev/github.com/iter8-tools/etc3@v0.1.13-pre/api/v2alpha1).

## ExperimentSpec

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| target | string | Identifies the resources involved in an experiment. The meaning depends on the domain but typically identifies the resources participating in the experiment. Two experiments with the same target cannot run concurrently; those with different targets can | Yes |
| strategy | [Strategy](#strategy) | Strategy used for experimentation. | Yes |
| criteria | [Criteria](#criteria) | Criteria used to evaluate versions. | No |
| duration | [Duration](#duration) | Duration of the experiment. | No |
| versionInfo | [VersionInfo](#versioninfo) | Details about the application versions participating in the experiment. | No |

## Strategy

Defines the behavior of an experiment.

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| testingPattern | string | Type of iter8 experiment. Currently, `Canary` and `Conformance` tests are supported. | Yes |
| deploymentPattern | string | The method by which traffic is shifted between versions in an experiment. Currently, these are `Progressive` (default) and `FixedSplit`. A Progressive pattern shifts traffic between versions while a FixedSplit pattern leaves traffic as it is. | No |
| actions | map[string][][TaskSpec](#taskspec) | Sequence of tasks to be called before an experiment begins or after an experiment completes. | No |
| handlers | [Handlers](#handlers) | (Deprecated) External methods that should be called before and after an experiment. `actions` are the preferred mechanism for this. | No |

## TaskSpec

Identifies the implementation of a task to be run by reference to a library implemented in [https://github.com/iter8-tools/handler](https://github.com/iter8-tools/handler), documented [here](https://pkg.go.dev/github.com/iter8-tools/handler).

 Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| library | string | Name of library containing the implementation of the task to be run. | Yes |
| task | string | Name of task to run | Yes |
| with | map[string][apiextensionsv1.JSON](https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON) | Inputs to the task. | No |

## Handlers

Do we want to document this?

## Criteria

> Note: References to metric resource objects within experiment criteria can be in the `namespace/name` format or in the `name` format. If the `name` format is used (i.e., if only the name of the metric is specified), then iter8 first searches for the metric in the namespace of the experiment resource object followed by the `iter8-system` namespace. If iter8 cannot find the metric in either of these namespaces, then the reference is considered in valid and the will terminate in a failure.

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| requestCount | string | The name of the metric to be used count the number of requests seen by a version. May be defaulted by the configuration of iter8. | no |
| objectives | [Objective](#objective)[] | A list of objectives. Satisfying all objectives in an experiment is a necessary condition for a version to be declared a `winner`. | No |
| indicators | string[] | A list of metrics that, during the experiment, for each version, metric values are recorded by iter8 in the experiment status section. | No |

## Objective

An objective identifies the range of acceptable values for a metric. A version with values in the specified range are considered to be passing.

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| metric | string | Reference to a metric resource object in the `namespace/name` format or in the `name` format.  | Yes |
| upperLimit | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | Upper limit on the metric value. If specified, for a version to satisfy this objective, its metric value needs to be below the limit. | No |
| lowerLimit | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | Lower limit on the metric value. If specified, for a version to satisfy this objective, its metric value needs to be above the limit. | No |

## Duration

The duration of an experiment expressed as two integer fields: the number of iterations in the experiment and the time interval in seconds between successive iterations.

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| intervalSeconds | int32 | Duration of a single iteration of the experiment in seconds. Default value = 20 seconds. | No |
| maxIterations | int32 | Maximum number of iterations in the experiment. In case of failure, the experiment may be terminated earlier. Default value = 15. | No |

## VersionInfo

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| baseline | [VersionDetail](#versiondetail) | Details of the current or baseline version. | Yes |
| candidates | [][VersionDetail](#versiondetail) | Details of the candidate version or versions, if any. | No |

## VersionDetail

| Field | Type | Description | Required |
| ----- | ---- | ----------- | -------- |
| name | string | Name of the version. | Yes |
| variables | [][Variable0(#variable) | A list of name/value pairs that can passed to action tasks and used to specify metrics queries | No |
| weightObjRef | [corev1.ObjectReference](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ObjectReference) | A reference to the field in a Kubernetes object that specfies the traffic sent to this version. | No |

## Variable

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | name of a variable | No |
| value | string | value that should be substituted or name | No |

## ExperimentStatus

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| conditions | [][ExperimentCondition](#experimentcondition) | A set of conditions that express progress through an experiment. | No |
| initTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time the experiment is created. | No |
| startTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when an experiment begins running (after any start actions have completed)  | No |
| endTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when an experiment has completed. | No |
| lastUpdateTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the status was most recently updated. | No |
| stage | string | Indicator of progress of an experiment. The stage is `Waiting` before an experiment executes its start actions, `Initializing` while running the start actions, `Running` while the experiment is progressing, `Finishing` while any finish actions are running and `Completed` when the experiment terminates. | No |
| currentWeightDistribution | [][WeightData](#weightdata) | Currently observed distribution of requests. | No |
| analysis | Analysis | Result of latest query to the iter8 analytics service.  | No |
| recommendedBaseline | string | The version recommended as the version that should replace the baseline version when the experiment completes. | No |
| message | string | User readable message. | No |

## ExperimentCondition

Conditions express aspects of the progress of an experiment. The `Completed` condition indicates whether or not an experiment has completed or not. The `Failed` condition indicates whether or not an experiment completed successfully or in failure. Finally, the `TargetAcquired` condition indicates that an experiment can proceed without interference from other experiments. iter8 ensures that only one experiment has `TargetAcquired` set to `True` while `Completed` is set to `False`.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| type | string | Type of condition. Valid types are `TargetAcquired`, `Completed` and `Failed`. | Yes |
| status | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ConditionStatus) | status of condition, one of `True`, `False`, or `Unknown`. | Yes |
| lastTransitionTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The last time any field in the condition was changed. | No |
| reason | string | A reason for the change in value. Reasons come from a set of reason | No |
| message | string | A user readable decription. | No |

## WeightData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Version name | Yes |
| value | int32 | Percentage of traffic being sent to the version.  | Yes |

## Analysis

Result of latest query to the iter8 analytics service.
Queries may, but are not required to, return results along 4 dimensions.

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| aggregatedMetrics | [AggregatedMetricsAnalysis](#aggregatedmetricsanalysis) | Latest metrics for each required criteria. | No |
| winnerAssessment | WinnerAssessmentAnalysis(#winnerassessmentanalysis) | If identified, the recommended winning version. | No |
| versionAssessments | VersionAssessmentAnalysis(#versionassessmentanalysis) | For each version, a summary analysis identifying whether or no the version is satisfying the experiment criteria. | No |
| weights | WeightsAnalysis(#weightanalysis) | Recommended weight distributuion for next iteration of the experiment. | No |

## AggregatedMetricsAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data, usually a URL. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | map[string][AggregatedMetricsData](#aggregatedmetricsdata) | Map from metric name to most recent data (from all versions) for the metric. | Yes |

## WinnerAssessmentAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data, usually a URL. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | [WinnerAssessmentData](#winnerassessmentdata) | Details on whether or not a winner has been identified and which version if so. | No |

## VersionAssessmentAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data, usually a URL. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | map[string][]bool | map of version name to a list of boolean values, one for each objective specified in the experiment criteria, indicating whether not not the objective is satisified for not. | No |

## WeightAnalysis

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data, usually a URL. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | User readable message. | No |
| data | [][WeightData](#weightdata) | List of version name/value pairs representing a recommended weight for each version | No |

## WinnerAssessmentData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| winnerFound | bool | Whether or not a winner has been identified. | Yes |
| winner | string | The name of the identified winner, if one has been found. | No |

## AggregatedMetricsData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| max | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The maximum value observed for this metric accross all versions. | Yes |
| min | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The minimum value observed for this metric accross all versions. | Yes |
| data | map[string][AggregatedMetricsVersionData](#aggregatedmetricsversiondata) | A map from version name to the most recent aggregated metrics data for that version. | No |

## AggregatedMetricsVersionData

| Field | Type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| max | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The maximum value observed for this metric for this version over all observations. | No |
| min | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The minimum value observed for this metric for this version over all observations. | No |
| value | [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) | The value. | No |
| sampleSize | int32 | The number of requests observed by this version. | No |
