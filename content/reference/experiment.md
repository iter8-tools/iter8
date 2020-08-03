---
date: 2016-04-09T16:50:16+02:00
menuTitle: Experiment
title: Iter8's experiment CRD
weight: 20
summary: A new Kubernetes CRD that compares baseline and candidate deployments
---

When iter8 is installed, a new Kubernetes CRD is added to your cluster. Our CRD kind and current API version are as follows:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
```

Below we document iter8's Experiment CRD. For clarity, we break the documentation down into the CRD's 4 sections: `spec`, `action`, `metrics`, and `status`.

## `Experiment spec`

Following the Kubernetes model, the `spec` section specifies the details of the object and its desired state. The `spec` of an `Experiment` custom resource identifies the target service of a candidate release or A/B test, the baseline deployment corresponding to the stable service version, the candidate deployment corresponding to the service version being assessed, etc. In the YAML representation below, we show sample values for the `spec` attributes and comments describing their meaning and whether or not they are optional.

```yaml
spec:
    # targetService specifies the reference to experiment targets
    targetService:
      # kind of baseline/candidate
      # options: Service, Deployment
      # default is Deployment
      kind: Deployment

      # name of a kubernetes service which receives actual traffic to the application
      # it's required when baseline/candidate are specified as deployments
      name: reviews

      # hosts specifies how baseline/candidate can be accessed from outside the cluster
      # Each entry contains the name of a host and the gateway(istio) associated with it
      # This is optional and only applies to services directly accessible from outside the K8s cluster
      hosts:
      - name: reviews.com
        gateway: bookinfo-gateway

      # Name of the baseline and candidate versions (required)
      # If kind (above) is Deployment, baseline and candidate are references to K8s deployments
      # If kind (above) is Service, baseline and candidate are references to K8s services
      baseline: reviews-v3
      candidate: reviews-v5

      # port of the kubernetes service(.spec.targetService.name) that receives traffic
      # When there is only one port listening on the service, this is optional
      # If baseline/candidate are services, they should share the same port number
      port: 9080

    # analysis contains the parameters for configuring the analytics service
    analysis:

      # analyticsService specifies analytics service endpoint (optional)
      # default value is http://iter8-analytics.iter8:8080
      analyticsService: http://iter8-analytics.iter8:8080

      # endpoint to Grafana dashboard (optional)
      # default is http://localhost:3000
      grafanaEndpoint: http://localhost:3000

      # successCriteria is a list of criteria for assessing the candidate version (optional)
      # if the list is empty, the controller will not rely on the analytics service
      successCriteria:

      # metricName: name of the metric to which this criterion applies (required)
      # the name should match the name of an iter8 metric or that of a user-defined custom metric
      # names of metrics supported by iter8 out of the box:
      #   iter8_latency: mean latency of the service
      #   iter8_error_rate: mean error rate (~5** HTTP Status codes) of the service
      #   iter8_error_count: total error count (~5** HTTP Status codes) of the service
      - metricName: iter8_latency

        # minimum number of data points required to make a decision based on this criterion (optional)
        # default is 10
        # Used by the check_and_increment and epsilon_greedy algorithms
        # Ignored by other algorithms
        sampleSize: 100

        # the metric value for the candidate version defining this success criterion (required)
        # it can be an absolute threshold or one relative to the baseline version, depending on the
        # attribute toleranceType described next
        tolerance: 0.2

        # indicates if the tolerance value above should be interpreted as an absolute threshold or
        # a threshold relative to the baseline (required)
        # options:
        #   threshold: the metric value for the candidate must be below the tolerance value above
        #   delta: the tolerance value above indicates the percentage within which the candidate metric value can deviate
        # from the baseline metric value
        toleranceType: threshold

        # The range of possible metric values (optional)
        # Used by bayesian routing algorithms if available.
        # Ignored by other algorithms.
        min_max:
          # The minimum possible value for the metric
          min: 0.0

          # The maximum possible value for the metric
          max: 1.0

        # indicates whether or not the experiment must finish if this criterion is not satisfied (optional)
        # default is false
        stopOnFailure: false

      # reward is an optional field that can be used when an a/b testing is conducted
      # When both versions satisfy all the success criteria, the one with higher reward value wins the comparison
      # This is effective when a bayesian routing strategy is specified in trafficControl (posterior_bayesian_routing or optimistic_bayesian_routing)
      reward:
        # the metric whose value is treated as reward (required)
        - metricName: iter8_latency

        # The range of possible metric values (optional)
        min_max:
          # The minimum possible value for the metric
          min: 0.0

          # The maximum possible value for the metric
          max: 1.0

    # trafficControl controls the experiment durarion and how the controller should change the traffic split
    trafficControl:

      # frequency with which the controller calls the analytics service
      # it corresponds to the duration of each "iteration" of the experiment
      interval: 30s

      # maximum number of iterations for this experiment (optional)
      # the duration of an experiment is defined by maxIterations * internal
      # default is 100
      maxIterations: 6

      # the maximum traffic percentage to send to the candidate during an experiment (optional)
      # default is 50
      maxTrafficPercentage: 80

      # strategy used to analyze the candidate and shift the traffic (optional)
      # except for the strategy increment_without_check, the analytics service is called
      # at each iteration and responds with the appropriate traffic split which the controller honors
      # options:
      #   check_and_increment
      #   epsilon_greedy
      #   posterior_bayesian_routing
      #   optimistic_bayesian_routing
      #   increment_without_check: increase traffic to candidate by trafficStepSize at each iteration without calling analytics
      # default is check_and_increment
      strategy: check_and_increment

      # the maximum traffic increment per iteration (optional)
      # default is 2.0
      # Used by check_and_increment algorithm
      # Ignored by other algorithms
      trafficStepSize: 20

      # The required confidence in the recommended traffic split (optional)
      # default is 0.95
      # Used by bayesian routing algorithms
      # Ignored by other algorithms
      confidence: 0.9

      # determines how the traffic must be split at the end of the experiment (optional)
      # options:
      #   baseline: all traffic goes to the baseline version
      #   candidate: all traffic goes to the candidate version
      #   both: traffic is split across baseline and candidate
      # default is candidate
      onSuccess: candidate

    # indicates whether or not iter8 should perform a clean-up action at the end of the experiment (optional)
    # if no action is specified, nothing is done to clean up at the end
    # if used, the currently supported actions are:
    #   delete: at the end of the experiment, the version that ends up with no traffic (if any) is deleted
    cleanup:
```

## `Experiment action`: user-provided action

The user can interfere with an ongoing experiment by setting the value of an `action` attribute. Iter8 currently supports 4 user actions: `pause`, `resume`, `override_success`, and `override_failure`. Pause and resume are self-explanatory. The action `override_success` causes the experiment to immediately terminate with a success status, whereas `override_failure` causes the experiment to terminate with a failure status.

```yaml
# user-provided input (optional)
# options:
#   pause: pause the experiment
#   resume: resume a paused experiment
#   override_success: terminate the experiment indicating that the candidate succeeded
#   override_failure: abort the experiment indicating that the candidate failed
action: ""
```

## `Experiment metrics`

Information about all Prometheus metrics known to iter8 are stored in a Kubernetes `ConfigMap` named _`iter8config-metrics`_. When iter8 is installed, that `ConfigMap` is populated with information on the 3 metrics that iter8 supports out of the box, namely: `iter8_latency`, `iter8_error_rate`, and `iter8_error_count`. Users can add their own custom metrics.

When an `Experiment` custom resource is created, the iter8 controller will check the metric names referenced by `.spec.analysis.successCriteria`, look them up in the `ConfigMap`, retrieve the information about them from the `ConfigMap`, and store that information in the `metrics` section of the newly created `Experiment` object. The information about a metric allows the iter8 analytics service to query Prometheus to retrieve metric values for the baseline version and candidate versions of the service . Below we show an example of how a metric is stored in an `Experiment` object.

```yaml
metrics:
  iter8_latency:
    absent_value: None
    is_counter: false
    query_template: (sum(increase(istio_request_duration_seconds_sum{job='istio-mesh',reporter='source'}[$interval]$offset_str))
      by ($entity_labels)) / (sum(increase(istio_request_duration_seconds_count{job='istio-mesh',reporter='source'}[$interval]$offset_str))
      by ($entity_labels))
    sample_size_template: sum(increase(istio_requests_total{job='istio-mesh',reporter='source'}[$interval]$offset_str))
      by ($entity_labels)
```

## `Experiment status`

Following the Kubernetes model, the `status` section contains all relevant runtime details pertaining to the `Experiment` custom resource. In the YAML representation below, we show sample values for the `status` attributes and comments describing their meaning.

```yaml
  status:
    # the last analysis state
    analysisState: {}

    # assessment returned from the analytics service
    assessment:
      conclusions:
      - The experiment needs to be aborted
      - All success criteria were not met

    # list of boolean conditions describing the status of the experiment
    # for each condition, if the status is "False", the reason field will give detailed explanations
    # lastTransitionTime records the time when the last change happened to the corresponding condition
    # when a condition is not set, its status will be "Unknown"
    conditions:

    # AnalyticsServiceNormal is "True" when the controller can get an interpretable response from the analytics service
    - lastTransitionTime: "2019-12-20T05:38:37Z"
      status: "True"
      type: AnalyticsServiceNormal

    # ExperimentCompleted tells whether the experiment is completed or not
    - lastTransitionTime: "2019-12-20T05:39:37Z"
      status: "True"
      type: ExperimentCompleted

    # ExperimentSucceeded indicates whether the experiment succeeded or not when it is completed
    - lastTransitionTime: "2019-12-20T05:39:37Z"
      message: Aborted
      reason: ExperimentFailed
      status: "False"
      type: ExperimentSucceeded

    # MetricsSynced states whether the referenced metrics have been retrieved from the ConfigMap and stored in the metrics section
    - lastTransitionTime: "2019-12-20T05:38:22Z"
      status: "True"
      type: MetricsSynced

    # Ready records the status of the latest-updated condition
    - lastTransitionTime: "2019-12-20T05:39:37Z"
      message: Aborted
      reason: ExperimentFailed
      status: "False"
      type: Ready

    # RoutingRulesReady indicates whether the routing rules are successfully created/updated
    - lastTransitionTime: "2019-12-20T05:38:22Z"
      tatus: "True"
      type: RoutingRulesReady

    # TargetsProvided is "True" when both the baseline and the candidate versions of the targetService are detected by the controller; otherwise, missing elements will be shown in the reason field
    - lastTransitionTime: "2019-12-20T05:38:37Z"
      status: "True"
      type: TargetsProvided

    # the current experiment's iteration
    currentIteration: 2

    # Unix timestamp in nanoseconds when the experiment is created
    createTimestamp: 1576820317351000

    # Unix timestamp in nanoseconds when the experiment started
    startTimestamp: 1576820317351000

    # Unix timestamp in nanoseconds when the experiment finished
    endTimestamp: 1576820377696000

    # The url to he Grafana dashboard pertaining to this experiment
    grafanaURL: http://localhost:3000/d/eXPEaNnZz/iter8-application-metrics?var-namespace=bookinfo-iter8&var-service=reviews&var-baseline=reviews-v3&var-candidate=reviews-v5&from=1576820317351&to=1576820377696

    # the time when the previous iteration was completed
    lastIncrementTime: "2019-12-20T05:39:07Z"

    # this is the message to be shown in the STATUS column for the `kubectl` printer, which summarizes the experiment situation
    message: 'ExperimentFailed: Aborted'

    # the experiment's current phase
    # values could be: Progressing, Pause, Completed
    phase: Completed

    # the current traffic split
    trafficSplitPercentage:
      baseline: 100
      candidate: 0
  ```
