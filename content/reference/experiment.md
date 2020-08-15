---
menuTitle: Experiment
title: Experiment CRD
weight: 20
summary: Introduction to iter8 experiment
---

The *Experiment* CRD(Custom Resource Definition) contains 2 sections: `spec` and `stauts`. `spec` provides you the schema to configure your test while `status` reflects runtime assesment details about the experiment. You can find the CRD YAML [here](https://github.com/iter8-tools/iter8/blob/master/install/helm/iter8-controller/templates/crds/v1alpha2/iter8.tools_experiments.yaml).

Let's go through a sample Experiment CR to understand fields in each section:

## apiVersion/Kind/Metadata

Current version is v1alpha2.

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  # name of experiment
  name: reviews-experiment
  namespace: test
```

## Spec

```yaml
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

    # this id refers to the id of router used to handle traffic for the experiment
    # optional; default is the first entry of effictive host
    routerID: reviews-router

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
```

## Status

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

  # url to access grafana dashboard showing metrics about this experiment
  grafanaURL: localhost:3000/d/eXPEaNnZz/iter8-application-metrics?var-namespace=test&var-service=reviews&var-baseline=reviews-v1&var-candidate=reviews-v2,reviews-v3&from=1597339597000&to=1597339598992

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
