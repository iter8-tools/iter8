---
menuTitle: Experiment
title: Experiment CRD
weight: 20
summary: Introduction to iter8 experiment
---

The experiment resource object specifies the configuration of an experiment and records the status of the experiment during its progress and after its completion.
***

## Experiment Spec

Configuration of an iter8 experiment.

Field | Type | Description | Required
------|------|-------------|---------
*service* | Service | Configuration that identifies the microservice and its versions that are involved in this experiment | yes
*trafficControl* | TrafficControl | Configuration that affect how application traffic is split across different versions of the service during and after the experiment | no
*cleanup* | boolean | Boolean field indicating if routing rules set up by iter8 during the experiment should be deleted after the experiment. Default value: `false`  | no
*analyticsEndoint* | HTTP URL | URL of the *iter8-analytics* service. Default value: http://iter8-analytics.iter8:8080 | no
*criteria* | Criterion[] | A list of criteria which defines the winner in this experiment | no
*duration* | Duration | Fields that affect how long the iter8 experiment will last  | no
*manualOverride* | ManualOverride | User actions that override the current status of an experiment  | no

An example of experiment spec is as follows. This experiment spec rolls out a new version of *reviews* (*reviews-v2* candidate deployment), if it has a mean latency of at most *250* milli seconds. Otherwise, it rolls back to the baseline version (*reviews-v1* deployment).

```yaml
# current version of iter8 experiment CRD is v1alpha2
apiVersion: iter8.tools/v1alpha2 
kind: Experiment
metadata:
  name: reviews-experiment
spec:
  service:
    kind: Deployment
    name: reviews
    baseline: reviews-v1
    candidates:
    - reviews-v2
  criteria:
  - metric: iter8_latency
    threshold:
      type: absolute
      value: 250
```

For other examples of experiment spec objects, refer to the [canary release](../../tutorials/canary/#create-a-canary-experiment) and [A/B/n rollout](../../tutorials/abn/#create-an-abn-experiment) tutorials.


***

### Service

Configuration that identifies the microservice and its versions that are involved in this experiment

Field | Type | Description | Required
------|------|-------------|---------
*kind* | Enum: {*Deployment, Service*} | Enum which identifies whether service versions are implemented as `Deployment`s or as `Service`s. Default value: `Deployment`. | yes
*name* | string | Name of the service whose versions are being compared in the experiment | yes
*namespace* | string | Namespace to which the service, whose versions are being compared in the experiment, belongs to. | no
*baseline* | string | Name of the baseline version. If `kind == Deployment`, then this is the name of a deployment. Else, if `kind == Service`, then this is the name of a service. | yes
*candidates* | string[] | A list of names of candidate versions. If `kind == Deployment`, then these are names of candidate deployments. Else, if `kind == Service`, these are names of candidate services. | no
*port* | integer | Port number where the service listens. | no
*hosts* | Host[] | List of external hosts and gateways associated with this service and defined in Istio Gateway. | no

#### Host

External host and gateway that is associated with a service within iter8 experiment. Refer to host and gateway [documentation for Istio virtual services](https://istio.io/latest/docs/reference/config/networking/virtual-service/).

Field | Type | Description | Required
------|------|-------------|---------
*name* | string | The destination host to which traffic is being sent. This could be a DNS name with wildcard prefix or an IP address. | yes
*gateway* | string | The name of gateway to which this host is attached. | yes

An example of the `service` subsection of an experiment object is as follows. Observe that versions correspond to services and not deployments in this example.

```yaml
service:
  kind: Service
  name: reviews
  namespace: test
  baseline: reviews-v1
  candidates:
  - reviews-v2
  - reviews-v3
  port: 9080
  hosts:
  - name: "reviews.com"
    gateway: reviews-service
```

***

### Traffic Control

Configuration that affect how application traffic is split across different versions of the service during and after the experiment.

Field | Type | Description | Required
------|------|-------------|---------
*strategy* | Enum: {*progressive, top_2, uniform*} | Enum which identifies the algorithm used for shifting traffic during an experiment (refer to [Algorithms](../algorithms) for in-depth descriptions of iter8's algorithms). Default value: `progressive`. | no
*onTermination* | Enum: {to_winner,to_baseline,keep_last} | Enum which determines the traffic split behavior after the termination of the experiment. Setting `to_winner` ensures that, if a winning version is found at the end of the experiment, all traffic will flow to this version after the experiment terminates. Setting `to_baseline` will ensure that all traffic will flow to the baseline version, after the experiment terminates. Setting `keep_last` will ensure that the traffic split used during the final iteration of the experiment continues even after the experiment has terminated. Default value: `to_winner`. | no
*match* | [HTTPMatchRequest clause of Istio virtual service](https://istio.io/latest/docs/reference/config/networking/virtual-service/#HTTPMatchRequest) | Specifies the portion of traffic which can be routed to candidates during the experiment. Traffic that does not match this clause will be sent to baseline and never to a candidate during an experiment. By default, if this field is left unspecified, all traffic is used for an experiment (i.e., match all). | no
*maxIncrement* | integer | Specifies the maximum percentage by which traffic routed to a candidate can increase during a single iteration of the experiment. Default value: 2 (percent) | no
*routerID* | string | Refers to the id of router used to handle traffic for the experiment. Default value: first entry of effective host. | no

An example of the `trafficControl` subsection of an experiment object is as follows.

```yaml
trafficControl:
  strategy: progressive
  onTermination: to_winner
  match:
    http:
     - uri:
         prefix: "/wpcatalog"
  maxIncrement: 20
  routerID: reviews-router
```

***

### Criterion
When the `criteria` field is non-empty in an experiment (this is the usual case), each version featured in an experiment is evaluated with respect to one or more criteria. This section describes the anatomy of a single criterion.

Field | Type | Description | Required
------|------|-------------|---------
*metric* | string | The metric used in this criterion. Metrics can be iter8's out-of-the-box metrics or custom metrics. See [metrics documentation](../metrics) for more details. Iter8 computes and reports a variety of assessments that describe how each version is performing with respect to this metric. | yes
*threshold* | Threshold | An optional threshold for this metric. Iter8 computes and reports a variety of assessments that describe how each version is performing with respect to this threshold.  | no
*isReward* | boolean | This field indicates if the metric used in this criterion is a reward metric. When a metric is marked as reward metric, the winning version in an experiment is one which optimizes the reward while satisfying all thresholds at the same time. Only ratio metrics can be designated as a reward. Default value: `false` | no

#### Threshold

Threshold specified for a metric within a criterion.

Field | Type | Description | Required
------|------|-------------|---------
*value* | float | Threshold value.  | yes
*type* | Enum: {*absolute*, *relative*} | When the threshold type is `absolute`, the threshold value indicates an absolute limit on the value of the metric. When the threshold type is `relative`, the threshold value indicates a multiplier relative to the baseline. For example, if the metric is *iter8_latency*, and if threshold is `absolute` and value is 250, a candidate is said to satisfy this threshold if its mean latency is within 250 milli seconds; otherwise, if threshold is `relative` and value is 1.6, a candidate is said to satisfy this threshold if its mean latency is within 1.6 times that of the baseline version's mean latency. Relative thresholds can only be used with [ratio metrics](../metrics/#ratio-metrics). The interpretation of threshold depends on the [preferred direction](../metrics/#extending-iter8s-metrics) of the metric. If the preferred direction is `lower`, then the threshold value represents a desired upper limit. If the preferred direction is `higher`, then the threshold value represents a desired lower limit. | yes

An example of the `criteria` subsection of an experiment object is as follows.

```yaml
criteria:
    - metric: iter8_mean_latency
      threshold:
        type: relative
        value: 1.6
    - metric: iter8_error_rate
      threshold:
        type: absolute
        value: 0.05
    - metric: le_500_ms_latency_percentile
      threshold:
        type: absolute
        value: 0.95
    - metric: mean_books_purchased
      isReward: true
```


*** 

### Duration

Configuration affecting the duration of the experiment.

Field | Type | Description | Required
------|------|-------------|---------
*interval* | string | Length of an iteration in the experiment. Values for this field should be valid [go duration strings](https://golang.org/pkg/time/#ParseDuration) (e.g., 30s, 1m, 1h, 1h10m10s). Default value: 30s  | no
*maxIterations* | integer | Number of iterations in an experiment. Default value: 100  | no

An example of the `duration` subsection of an experiment object is as follows.

```yaml
duration:
  interval: 20s
  maxIterations: 10
```

***

### Manual Override

Manual / out-of-band actions that override the current execution of the experiment.

Field | Type | Description | Required
------|------|-------------|---------
*action* | Enum: {*pause, resume, terminate*} | This field enables manual / out-of-band intervention during the course of an experiment. Execution of the experiment will be paused, or resumed from a previously paused state, or terminated respectively depending upon whether the value of this field is `pause`, `resume` or `terminate`. | yes
*trafficSplit* | Object | Traffic split between different versions of the experiment which will take effect if `action == terminate`. | no

An example of the `manualOverride` subsection of an experiment object is as follows.

```yaml
manualOverride:
  action: terminate
  trafficSplit:
    reviews-v2:80
    reviews-v3:20
```

***

<!-- ```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  # name of experiment
  name: reviews-experiment
  namespace: test
``` -->

<!-- ```yaml
spec:

  # service section contains infomation on the target service that experiment will test on
  # required
  service:

    # the kind of versions
    # optional; options: {Deployment, Service}; default is Deployment
    kind: Service

    # name of internal service that directs traffic to experiment versions
    # optional when version kind is Service
    # required when version kind is Deployment
    name: reviews

    # namespace of the service
    # optional; default is the same as experiment namespace
    namespace: test

    # name of baseline version
    # required
    baseline: reviews-v1

    # list of names of candidate versions
    # required
    candidates:
    - reviews-v2
    - reviews-v3

    # port number of service listening on
    # optional;
    port: 9080

    # list of external hosts and gateways associated (defined in Istio Gateway)
    # optional;
    hosts:
    - name: "reviews.com"
      gateway: reviews-service

  # this section gives instructions on traffic management for this experiment
  trafficControl:

    # the strategy used to shift traffic
    # optional; options: {progressive, top_2, uniform}: default is progressive
    strategy: progressive

    # determines traffic split status at the end of experiment
    # optional; options: {to_winner,to_baseline,keep_last}; default is to_winner
    onTermination: to_winner

    # Istio matching clauses used to restrict traffic to service
    match:
      http:
       - uri:
           prefix: "/wpcatalog"

    # upperlimit of traffic increment for a version in one iteration
    # optional; default is 100
    maxIncrement: 20

    # this id refers to the id of router used to handle traffic for the experiment
    # optional; default is the first entry of effictive host
    routerID: reviews-router


  # Whether routing rules should be deleted at the end of experiment
  # optional; default is false
  cleanup: false

  # endpoint of analytics service
  # optional; default is http://iter8-analytics.iter8:8080
  analyticsEndpoint: http://iter8-analytics.iter8:8080

  # the list of criteria that defines success of versions
  # optio
  criteria:
  - metric: iter8_latency
    threshold:
      type: relative
      value: 3000

  # length and number of intervals to re-evaluate the assessment
  duration:
    interval: 20s
    maxIterations: 1

  # user actions to override the current status of the experiment
  manualOverride:

    # options: {pause, resume, terminate}
    # required
    action:

    # Traffic split status specification
    # Applied to action terminate only
    # example:
    #   reviews-v2:80
    #   reviews-v3:20
    # optional
    trafficSplit:
``` -->

## Experiment Status

Iter8 records a variety of status information such as assessments about the various service versions and current traffic split across the versions during the course of an experiment and after its completion. An example of the status section within an experiment object is as follows.

```yaml
status:
  # assessment from analytics on testing versions
  assessment:

    # assessment details for baseline version
    baseline:

      # generated uuid for the version in this experiment
      id: baseline

      # name of version
      name: reviews-v1

      # total requests that have been received by the version
      request_count: 10

      # recommended traffic weight to this version
      weight: 0

      # probability of being the best version
      win_probability: 0

    # list of candidate assessments
    candidates:
    # same format as baseline assessment
    - ...

    # assessment for winner version
    winner:

      # id of current best version
      current_best_version: candidate-1

      # name of the current best version
      name: reviews-v3

      # indicates whether the current best version is winner or not
      winning_version_found: true

  # A list of conditions reflecting status of the experiment from the controller
  conditions:

    # the last time when this condition is updated
  - lastTransitionTime: "2020-08-13T17:26:37Z"

    # a human-readable message explaining the status of this condition
    message: ""

    # the reason of updating this condition
    reason: SyncMetricsSucceeded

    # status of the condition; value can be "True", "False" or "Unknown"
    status: "True"

    # type of condition
    # MetricsSynced indicates whether metrics referenced in the criteria has been read in experiment or not
    type: MetricsSynced
  - ...
    # TargetsProvided indicates existence of all target objects in the cluster
    type: TargetsProvided
  - ...
    # indicate completeness of experiment
    type: ExperimentCompleted
  - ...
    # condition on the analytics server
    type: AnalyticsServiceNormal
  - ...
    # condition on routing rules readiness
    type: RoutingRulesReady

  # the current iteration experiment is going
  currentIteration: 1

  # list of hosts that will direct traffic to service
  effectiveHosts:
  - reviews

  # type of this experiment
  # Canary: Canary rollout(only one candidate with no reward criteria)
  # A/B: A/B testing(with one candidate with reward criteria)
  # A/B/N: A/B/N testing(with more than one candidate)
  experimentType: A/B/N

  # timestamp when experiment is initialized
  initTimestamp: "2020-08-13T17:26:37Z"

  # the timestamp when last iteration updates
  lastUpdateTime: "2020-08-13T17:26:38Z"

  # the latest message on condition of the experiment
  message: 'ExperimentCompleted: Traffic To Winner'

  # the phase of
  phase: Completed

  # the timestamp when experiment starts
  # detection of baseline version kicks off the experiment
  startTimestamp: "2020-08-13T17:26:37Z"

  # timestamp when experiment ends
  endTimestamp: "2020-08-13T17:26:38Z"
```
