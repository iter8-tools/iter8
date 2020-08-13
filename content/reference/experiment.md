---
menuTitle: Experiment
title: Iter8's experiment CRD
weight: 20
summary: Introduction to iter8 experiment
---

The *Experiment* CRD(Custom Resource Definition) contains 2 sections: _spec_ and _stauts_. _spec_ provides you the schema to configure your test while _status_ reflects runtime assesment details about the experiment. You can find the crd yaml here[link].

Let's go through a sample Experiment CR:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  # name of experiment
  # each experiment can only 
  name: reviews-experiment
  namespace: test
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
    # name of baseline version
    # required
    baseline: reviews-v1
    # list of names of candidate versions
    # required
    candidates:
    - reviews-v2
    - reviews-v3
  # this section gives instructions on traffic management for this experiment
  trafficControl:
    # this id refers to the id of router used to handle traffic for the experiment
	  # optional; default is the first entry of effictive host
    routerID: reviews-router
    # the strategy used to shift traffic
	  # optional; options are {progressive, top_2, uniform}: default is progressive
    strategy: progressive
    # determines traffic split status at the end of experiment
    # 
    onTermination: 

    match: 

    maxIncrement: 
  # 
  analyticsEndpoint: http://127.0.0.1:57105
  criteria:
  - metric: iter8_latency
    threshold:
      type: relative
      value: 3000
  duration:
    interval: 1s
    maxIterations: 1
  metrics:
    counter_metrics:
    - name: iter8_request_count
      query_template: sum(increase(istio_requests_total{reporter='source',job='envoy-stats'}[$interval]))
        by ($version_labels)
    - name: iter8_total_latency
      query_template: sum(increase(istio_request_duration_milliseconds_sum{reporter='source',job='envoy-stats'}[$interval]))
        by ($version_labels)
    - name: iter8_error_count
      preferred_direction: lower
      query_template: sum(increase(istio_requests_total{response_code=~'5..',reporter='source',job='envoy-stats'}[$interval]))
        by ($version_labels)
    ratio_metrics:
    - denominator: iter8_request_count
      name: iter8_mean_latency
      numerator: iter8_total_latency
      preferred_direction: lower
    - denominator: iter8_request_count
      name: iter8_error_rate
      numerator: iter8_error_count
      preferred_direction: lower
      zero_to_one: true
status:
  analysisState: {}
  assessment:
    baseline:
      id: baseline
      name: reviews-v1
      request_count: 10
      weight: 0
      win_probability: 0
    candidates:
    - id: candidate-0
      name: reviews-v2
      request_count: 10
      weight: 0
      win_probability: 0
    - id: candidate-1
      name: reviews-v3
      request_count: 10
      weight: 100
      win_probability: 100
    winner:
      current_best_version: candidate-1
      name: reviews-v3
      winning_version_found: true
  conditions:
  - lastTransitionTime: "2020-08-13T17:26:37Z"
    message: ""
    reason: SyncMetricsSucceeded
    status: "True"
    type: MetricsSynced
  - lastTransitionTime: "2020-08-13T17:26:37Z"
    message: ""
    reason: TargetsFound
    status: "True"
    type: TargetsProvided
  - lastTransitionTime: "2020-08-13T17:26:38Z"
    message: Traffic To Winner
    reason: ExperimentCompleted
    status: "True"
    type: ExperimentCompleted
  - lastTransitionTime: "2020-08-13T17:26:38Z"
    message: ""
    reason: AnalyticsServiceRunning
    status: "True"
    type: AnalyticsServiceNormal
  - lastTransitionTime: "2020-08-13T17:26:37Z"
    message: ""
    reason: RoutingRulesReady
    status: "True"
    type: RoutingRulesReady
  currentIteration: 1
  effectiveHosts:
  - reviews
  endTimestamp: "2020-08-13T17:26:38Z"
  experimentType: A/B/N
  grafanaURL: localhost:3000/d/eXPEaNnZz/iter8-application-metrics?var-namespace=test&var-service=reviews&var-baseline=reviews-v1&var-candidate=reviews-v2,reviews-v3&from=1597339597000&to=1597339598992
  initTimestamp: "2020-08-13T17:26:37Z"
  lastUpdateTime: "2020-08-13T17:26:38Z"
  message: 'ExperimentCompleted: Traffic To Winner'
  phase: Completed
  startTimestamp: "2020-08-13T17:26:37Z"

```