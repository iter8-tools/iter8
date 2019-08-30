# Iter8 Experiment API
Iter8 Experiment CRD includes 3 parts -- Spec, Metrics and Status. 

## Spec
Spec allows user to specify details of the experiment.
```yaml
spec:
    # targetService specifies the reference to experiment targets
    targetService:
      # apiVersion of the target service
      # required; options:
      # v1: target service is a kubernetes service 
      # serving.knative.dev/v1alpha1: tarfet service is a knative service
      apiVersion: v1
      # name of target service
      name: reviews
      # The baseline and candidate for comparison
      # required; For Kubernetes, these two components refer to names of deployments
      # For Knative, they are names of revisions
      baseline: reviews-v3
      candidate: reviews-v5
    # RoutingReference provides references to routing rules
    # optional; now only a single istio virtualservice can be supported
    routingReference:
      apiversion: networking.istio.io/v1alpha3
      kind: VirtualService
      name: reviews-external
    # analysis contains the parameters for configuring the analytics service
    analysis:
      # analyticsService specifies analytics service endpoint
      # optional; Default is http://iter8-analytics:5555
      analyticsService: http://iter8-analytics.iter8
      # endpoint to Grafana Dashboard
      # optional; Default is http://localhost:3000
      grafanaEndpoint: http://localhost:3000
      # successCriteria contains the list of criteria for assessing the candidate version
      # optional; If the list is empty, the controller will progress without contacting the analytics service
      successCriteria:
      # metricName: Name of the metric to which the criterion applies.
      # This refers to the definitions in the Metrics section.
      # requried; Default options: 
      # iter8_latency: mean latency of the service 
      # iter8_error_rate: mean error rate (~5** HTTP Status codes) of the service
      # iter8_error_count: total error count (~5** HTTP Status codes) of the service
      - metricName: iter8_latency
        # Minimum number of data points required to make a decision based on this criterion; 
        # optional; Default is 10.
        sampleSize: 5
        # The value to check for toleranceType
        tolerance: 0.2
        # required; Options: 
        # delta: compares the candidate against the baseline version with respect to the metric;
        # threshold: checks the candidate with respect to the metric
        toleranceType: threshold
        # Indicates whether or not the experiment must finish if this criterion is not satisfied; 
        # optional; Default is false.
        stopOnFailure: false
    # trafficControl controls the behavior of the controller
    trafficControl:
      # time before the next increment.
      # optional; Default is 1mn
      interval: 30s
      # Maximum number of iterations for this experiment. 
      # optional; Default is 100.
      maxIterations: 6
      # the maximum traffic ratio to send to the candidate. 
      # optional; Default is 50
      maxTrafficPercentage: 80
      # strategy is the strategy used for experiment. 
      # Options:
      # check_and_increment: get decision on traffic increament from analytics 
      # increment_without_check: increase traffic each intervalwithout calling analytics
      # optional; Default is check_and_increment
      strategy: check_and_increment
      # the traffic increment per interval.
      # optional; Default is 2.0
      trafficStepSize: 20
      # Determines how the traffic must be split at the end of the experiment; 
      # optional; options: 
      # baseline: all traffic goes to the baseline version; 
      # candidate: all traffic goes to the candidate version;
      # both: traffic is split across baseline and candidate.
      # Default to candidate.
      onSuccess: candidate
    # a flag set to terminate experiment externally with action
    # optional; Default is "".
    # Options:
    # override_success: terminate experiment with condition specified in onSuccess
    # override_failure: terminate experiment with failure condition
    assessment: ""
```

## Metrics 
Metrics are stored as a map from metric name to metric definition (`query_template`, `sample_size_template`, `type`).   
They are read from _`iter8_metrics`_ configmap in runtime.  
Only metrics referenced in `.spec.analysis.successCriteria` will be cached in runtime object.  
Here shows a example of how a metric is stored in an experiment object.  
```yaml
metrics:
  iter8_latency:
      query_template: (sum(increase(istio_request_duration_seconds_sum{source_workload_namespace!='knative-serving',reporter='source'}[$interval]$offset_str))
        by ($entity_labels)) / (sum(increase(istio_request_duration_seconds_count{source_workload_namespace!='knative-serving',reporter='source'}[$interval]$offset_str))
        by ($entity_labels))
      sample_size_template: sum(increase(istio_requests_total{source_workload_namespace!='knative-serving',reporter='source'}[$interval]$offset_str))
        by ($entity_labels)
      type: Performance
```

## Status
Iter8 status includes the runtime details of an experiment.
```yaml
  status:
    # the last analysis state
    analysisState: {}
    # assessment returned from the analytics service
    assessment:
      conclusions:
      - Experiment started
    # A list of boolean conditions describing the status of experiment
    # For each condition, if the status is "False", the reason field will give detailed explanations
    # lastTransistionTime records the time when the last condition change is triggered
    # When a condition is not set, its status will be "Unknown"
    conditions:
    # AnalyticsServiceNormal is "True" when the controller can get interpretable response from the anaytics server
    - lastTransitionTime: "2019-08-26T14:13:08Z"
      status: "True"
      type: AnalyticsServiceNormal
    # ExperimentCompleted tells whether the experiment is completed or not
    - lastTransitionTime: "2019-08-26T14:13:39Z"
      status: "True"
      type: ExperimentCompleted
    # ExperimentSucceeded indicates whether the experiment is succeeded or not when it's completed
    - lastTransitionTime: "2019-08-26T14:13:39Z"
      reason: 'Aborted, Traffic: AllToBaseline.'
      status: "False"
      type: ExperimentSucceeded
    # MetricsSynced states whether the required metrics have been synced from configmap into the Metrics section
    - lastTransitionTime: "2019-08-26T14:12:53Z"
      status: "True"
      type: MetricsSynced
    # Ready records the status of any last-updated condition
    - lastTransitionTime: "2019-08-26T14:13:39Z"
      reason: 'Aborted, Traffic: AllToBaseline.'
      status: "False"
      type: Ready
    # TargetsProvided is "True" when all components in the targetService section are detected by the controller; otherwise, missing elements will be shown in the reason field 
    - lastTransitionTime: "2019-08-26T14:13:08Z"
      status: "True"
      type: TargetsProvided
    # the iteration that experiment is 
    currentIteration: 1
    # Unix timestamp in milliseconds
    startTimestamp: "1566828773475"
    endTimestamp: "1566828819018"
    # The url to grafana dashboard
    grafanaURL: http://localhost:3000/d/eXPEaNnZz/iter8-application-metrics?var-namespace=bookinfo-iter8&var-service=reviews&var-baseline=reviews-v3&var-candidate=reviews-v5&from=1566828773475&to=1566828819018
    # the time when last iteration is completed
    lastIncrementTime: "2019-08-26T14:13:08Z"
    # This is the message to be shown in the STATUS of kubectl printer, which shows the abstract of the experiment status
    message: 'Aborted, Traffic: AllToBaseline.'
    # the phase of the experiment; 
    # values could be: Initializing, Progressing, Pause, Succeeded, Failed
    phase: Failed
    # the percentage of traffic to baseline and candidate
    trafficSplitPercentage:
      baseline: 100
      candidate: 0

  ```
