---
template: overrides/main.html
---

# Experiment Resource Object

Fields in an iter8 experiment resource object `spec` are documented here. Unsupported fields are not documented here. For complete documentation, see the iter8 Experiment API [here](https://pkg.go.dev/github.com/iter8-tools/etc3@v0.1.13-pre/api/v2alpha1).

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
| variables | []Variable | A list of name/value pairs that can passed to action tasks and used to specify metrics queries | No |
| weightObjRef | [corev1.ObjectReference](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ObjectReference) | A reference to the field in a Kubernetes object that specfies the traffic sent to this version. | No |
