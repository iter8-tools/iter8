# iter8 Experiment CRD

When iter8 is installed, a new Kubernetes CRD is added to your cluster. This CRD is named `Experiment`. Below we document iter8's Experiment CRD. For clarity, we break the documentation down into the CRD 3 sections: `spec`, `metrics`, and `status`.

## Experiment Spec

Following the Kubernetes model, the `spec` section specifies the details of the object and its desired state. The `spec` of an `Experiment` custom resource identifies the target service of a canary release or A/B test, the baseline deployment corresponding to the stable service version, the candidate deployment corresponding to the service version being assessed, etc. In the YAML representation below, we show sample values for the `spec` attributes and comments describing their meaning and whether or not they are optional.

```yaml
spec:
    # targetService specifies the reference to experiment targets
    targetService:

      # apiVersion of the target service (required) 
      # options:
      #   v1: indicates that the target service is a Kubernetes service 
      #   serving.knative.dev/v1alpha1: indicates that the target service is a Knative service
      apiVersion: v1

      # name of target service (required) 
      # identifies either a Kubernetes service or a Knative service
      name: reviews

      # the baseline and candidate versions of the target service (required)
      # for Kubernetes, these two components refer to names of deployments
      # for Knative, they are names of revisions
      baseline: reviews-v3
      candidate: reviews-v5

    # routingReference is a reference to an existing Istio VirtualService (optional)
    # this should be used only if an Istio VirtualService has already been defined for the target Kubernetes service
    routingReference:
      apiversion: networking.istio.io/v1alpha3
      kind: VirtualService
      name: reviews-external
    
    # analysis contains the parameters for configuring the analytics service
    analysis:
    
      # analyticsService specifies analytics service endpoint (optional)
      # default value is http://iter8-analytics:5555
      analyticsService: http://iter8-analytics.iter8

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

        # indicates whether or not the experiment must finish if this criterion is not satisfied (optional)
        # default is false
        stopOnFailure: false
  
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
      #   increment_without_check: increase traffic to candidate by trafficStepSize at each iteration without calling analytics
      # default is check_and_increment
      strategy: check_and_increment

      # the maximum traffic increment per iteration (optional)
      # default is 2.0
      trafficStepSize: 20
  
      # determines how the traffic must be split at the end of the experiment (optional)
      # options: 
      #   baseline: all traffic goes to the baseline version
      #   candidate: all traffic goes to the candidate version
      #   both: traffic is split across baseline and candidate
      # default is candidate
      onSuccess: candidate

    # a flag that allows the user to terminate an ongoing experiment (optional)
    # options:
    #   override_success: terminate the experiment indicating that the candidate succeeded
    #   override_failure: abort the experiment indicating that the candidate failed
    # default is the empty string
    assessment: ""
```

## Experiment Metrics 
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

## Experiment Status
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
