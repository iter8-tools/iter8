---
template: main.html
---

# Metric Resource

!!! abstract "Metric resource"
    Iter8 defines the **Metric** resource type, which encapsulates the REST query that is used to retrieve a metric value from the metrics provider. Metric resources are referenced in experiments.


!!! note "Version"
    This document describes version `v2alpha2` of the Iter8 metric API type.

## Resources

### Experiment

??? info "Sample experiment"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
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
          - task: knative/init-experiment
          finish: # run the following sequence of tasks at the end of the experiment
          - task: common/exec # promote the winning version
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
        - metric: iter8-knative/mean-latency
          upperLimit: 50
        - metric: iter8-knative/95th-percentile-tail-latency
          upperLimit: 100
        - metric: iter8-knative/error-rate
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
| target | string | Identifies the app under experimentation and determines which experiments can run concurrently. Experiments that have the same target value will not be scheduled concurrently but will be run sequentially in the order of their creation timestamps. Experiments whose target values differ from each other can be scheduled by Iter8 concurrently. It is good practice to follow [target naming conventions](#target-naming-conventions). | Yes |
| strategy | [Strategy](#strategy) | The experimentation strategy which specifies how app versions are tested, how traffic is shifted during experiment, and what tasks are executed at the start and end of the experiment. | Yes |
| criteria | [Criteria](#criteria) | Criteria used for evaluating versions. This section includes (business) rewards, service-level objectives (SLOs) and indicators (SLIs). | No |
| duration | [Duration](#duration) | Duration of the experiment. | No |
| versionInfo | [VersionInfo](#versioninfo) | Versions involved in the experiment. Every experiment involves a baseline version, and may involve zero or more candidates. | No |

#### Status

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| conditions | [][ExperimentCondition](#experimentcondition) | A set of conditions that express progress of an experiment. | No |
| initTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time the experiment is created. | No |
| startTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the first iteration of experiment begins  | No |
| lastUpdateTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the status was most recently updated. | No |
| stage | string | Indicator of the progress of an experiment. The stage is `Waiting` before an experiment executes its start action, `Initializing` while running the start action, `Running` while the experiment has begun its first iteration and is progressing, `Finishing` while any finish action is running and `Completed` when the experiment terminates. | No |
| completedIterations | int32 | Number of completed iterations of the experiment. This is undefined until the experiment reaches the `Running` stage. | No |
| currentWeightDistribution | [][WeightData](#weightdata) | Currently observed split of traffic between versions. Expressed as percentage. | No |
| analysis | Analysis | Result of latest query to the Iter8 analytics service.  | No |
| versionRecommendedForPromotion | string | The version recommended for promotion. This field is initially populated by Iter8 as the baseline version and continuously updated during the course of the experiment to match the winner. The value of this field is typically used by finish actions to promote a version at the end of an experiment. | No |
| metrics | [][MetricInfo](#metricinfo) | A list of metrics referenced in the criteria section of this experiment. | No |
| message | string | Human readable message. | No |

### Metric

Metrics are referenced within the criteria field of the experiment spec. Metrics usage within experiments is described [here](../../../metrics/using-metrics).

??? example "Sample metric"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
    spec:
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$elapsedTime])) or on() vector(0)
      description: Number of requests
      type: counter
      provider: prometheus
      jqExpression: ".data.result[0].value[1] | tonumber"
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query      
    ```

#### Metadata
Standard Kubernetes [meta.v1/ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

#### Spec
| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| description | string | Human readable description. This field is meant for informational purposes. | No |
| units | string | Units of measurement. This field is meant for informational purposes. | No |
| provider | string | Type of the metrics provider. This field is meant for informational purposes. | No |
| params | [][NamedValue](#namedvalue) | List of name/value pairs corresponding to the name and value of the HTTP query parameters used by Iter8 when querying the metrics provider. Each name represents a parameter name; the corresponding value is a string template with placeholders; the placeholders will be dynamically substituted by Iter8 with values at query time. | No |
| body | string | String used to construct the JSON body of the HTTP request. Body may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using version information. | No |
| type | string | Metric type. Valid values are `Counter` and `Gauge`. Default value = `Gauge`. A `Counter` metric is one whose value never decreases over time. A `Gauge` metric is one whose value may increase or decrease over time. | No |
| method | string | HTTP method (verb) used in the HTTP request. Valid values are `GET` and `POST`. Default value = `GET`. | No |
| authType | string | Identifies the type of authentication used in the HTTP request. Valid values are `Basic`, `Bearer` and `APIKey` which correspond to HTTP authentication with these respective methods. | No |
| sampleSize | string | Reference to a metric that represents the number of data points over which the value of this metric is computed. This field applies only to `Gauge` metrics. References can be expressed in the form 'name' or 'namespace/name'. If just `name` is used, the implied namespace is the namespace of the referring metric. | No |
| secret | string | Reference to a secret that contains information used for authenticating with the metrics provider. In particular, Iter8 uses data in this secret to substitute placeholders in the HTTP headers and URL while querying the provider. References can be expressed in the form 'name' or 'namespace/name'. If just `name` is used, the implied namespace is the namespace where Iter8 is installed (which is `iter8-system` by default). | No |
| headerTemplates | [][NamedValue](#namedvalue) | List of name/value pairs corresponding to the name and value of the HTTP request headers used by Iter8 when querying the metrics provider. Each name represents a header field name; the corresponding value is a string template with placeholders; the placeholders will be dynamically substituted by Iter8 with values at query time. Placeholder substitution is attempted only if `authType` and `secret` fields are present. | No |
| jqExpression | string | The [jq](https://stedolan.github.io/jq/) expression used by Iter8 to extract the metric value from the JSON response returned by the provider. | Yes |
| urlTemplate | string | Template for the metric provider's URL. Typically, urlTemplate is expected to be the actual URL without any placeholders. However, urlTemplate may be templated, in which case, Iter8 will attempt to substitute placeholders in the urlTemplate at query time using the `secret` referenced in the metric. Placeholder substitution will not be attempted if `secret` is not specified. | Yes |

## Experiment field types

### Strategy

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| testingPattern | string | Determines the logic used to evaluate the app versions and determine the winner of the experiment. Iter8 supports two testing patterns, namely, `Canary` and `Conformance`. | Yes |
| deploymentPattern | string | Determines if and how traffic is shifted during an experiment. This field is relevant only for experiments using the `Canary` testing pattern. Iter8 supports two deployment patterns, namely, `Progressive` and `FixedSplit`. | No |
| actions | map[string][][TaskSpec](#taskspec) | An action is a sequence of tasks that can be executed by Iter8. `spec.strategy.actions` can be used to specify start and finish actions that will be run at the start and end of an experiment respectively. | No |

### TaskSpec

!!! abstract ""
    Specification of a task that will be executed as part of experiment actions. Task implementations are organized into libraries as documented [here](#task-implementations).

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| task | string | Name of the task. Task names express both the library and the task within the library in the format 'library/task' . | Yes |
| with | map[string][apiextensionsv1.JSON](https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON) | Inputs to the task. | No |

### Criteria

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| requestCount | string | Reference to the metric used to count the number of requests sent to app versions. | No |
| rewards | [Reward](#reward)[] | A list of metrics along with their preferred directions. Currently, this list needs to be of size one. This field can only be used in experiments with A/B and A/B/n testing patterns. | No |
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

### Reward

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| metric | string | Reference to a metric resource. Also see [note on metric references](#criteria). | Yes |
| preferredDirection | string | Indicates if higher values or lower values of this metric are preferable. `High` and `Low` are the two permissible values for this string. | Yes |

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
| variables | [][NamedValue](#namedvalue) | Variables are name-value pairs associated with a version. Metrics and tasks within experiment specs can contain strings with placeholders. Iter8 uses variables to substitute placeholders in these strings. | No |
| weightObjRef | [corev1.ObjectReference](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ObjectReference) | Reference to a Kubernetes resource and a field-path within the resource. Iter8 uses `weightObjRef` to get or set weight (traffic percentage) for the version. | No |


### MetricInfo

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| name | string | Identifies an Iter8 metric using the [`namespace/name` or `name` format](#criteria). | Yes |
| metric | [][Metric](#metric) | Iter8 metric object referenced by name. | No |

### ExperimentCondition

!!! abstract ""
    Conditions express aspects of the progress of an experiment. The `Completed` condition indicates whether or not an experiment has completed. The `Failed` condition indicates whether or not an experiment completed successfully or in failure. The `TargetAcquired` condition indicates that an experiment has acquired the target and is now scheduled to run. At any point in time, for any given target, Iter8 ensures that at most one experiment has the conditions `TargetAcquired` set to `True` and `Completed` set to `False`.

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| type | string | Type of condition. Valid types are `TargetAcquired`, `Completed` and `Failed`. | Yes |
| status | [corev1.ConditionStatus](https://pkg.go.dev/k8s.io/api@v0.20.0/core/v1#ConditionStatus) | status of condition, one of `True`, `False`, or `Unknown`. | Yes |
| lastTransitionTime | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The last time any field in the condition was changed. | No |
| reason | string | A reason for the change in value. | No |
| message | string | Human readable decription. | No |

### Analysis

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
| message | string | Human readable message. | No |
| data | map[string][]bool | map of version name to a list of boolean values, one for each objective specified in the experiment criteria, indicating whether not the objective is satisified. | No |


### AggregatedMetricsAnalysis

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| provenance | string | Source of the data. Currently, Iter8 analytics service URL is the only value for this field. | Yes |
| timestamp | [metav1.Time](https://pkg.go.dev/k8s.io/apimachinery@v0.20.2/pkg/apis/meta/v1#Time) | The time when the analysis took place. | Yes |
| message | string | Human readable message. | No |
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
| message | string | Human readable message. | No |
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
| message | string | Human readable message. | No |
| data | [][WeightData](#weightdata) | List of version name/value pairs representing a recommended weight for each version | No |

### WeightData

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Version name | Yes |
| value | int32 | Percentage of traffic being sent to the version.  | Yes |

## Common field types

### NamedValue

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Name of a variable. | Yes |
| value | string | Value of a variable. | Yes |


[^1]: `A/B/n` experiments involve more than one candidate. Their description is coming soon.