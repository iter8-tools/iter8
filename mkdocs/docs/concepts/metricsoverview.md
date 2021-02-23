---
template: overrides/main.html
---

# Metrics Overview

> **iter8** defines a Kubernetes CRD called **metric**. A metric resource encapsulates the REST query that is used for retrieving a metric value from the metrics backend.

??? example "Sample metric"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Metric
    metadata:
    name: request-count
    spec:
      params:
      - name: query
        value: sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$interval])) or on() vector(0)
      description: Number of requests
      type: counter
      provider: prometheus
    ```

Metrics are referenced within the `spec.criteria` stanza of the experiment. Metrics usage within experiments is described [here](/usage/metrics/using-metrics).

??? example "Sample experiment illustrating metrics usage"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Experiment
    metadata:
      name: quickstart-exp
    spec:
      # `sample-app` Knative service in `default` namespace is the target of this experiment
      target: default/sample-app
      # information about app versions participating in this experiment
      versionInfo:         
        # every experiment has a baseline version
        # we will name it `current`
        baseline: 
          name: current
          variables:
          # `revision` variable is used for fetching metrics from Prometheus
          - name: revision 
            value: sample-app-v1 
          # `promote` variable is used by the finish task
          - name: promote
            value: baseline
        # candidate version(s) of the app
        # there is a single candidate in this experiment 
        # we will name it `candidate`
        candidates: 
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2
          - name: promote
            value: candidate 
      criteria:
        objectives: 
        # mean latency should be under 50 milliseconds
        - metric: mean-latency
          upperLimit: 50
        # 95th percentile latency should be under 100 milliseconds
        - metric: 95th-percentile-tail-latency
          upperLimit: 100
        # error rate should be under 1%
        - metric: error-rate
          upperLimit: "0.01"
        indicators:
        # report values for the following metrics in addition those in spec.criteria.objectives
        - 99th-percentile-tail-latency
        - 90th-percentile-tail-latency
        - 75th-percentile-tail-latency
      strategy:
        # canary testing => candidate `wins` if it satisfies objectives
        testingPattern: Canary
        # progressively shift traffic to candidate, assuming it satisfies objectives
        deploymentPattern: Progressive
        actions:
          # run tasks under the `start` action at the start of an experiment   
          start:
          # the following task verifies that the `sample-app` Knative service in the `default` namespace is available and ready
          # it then updates the experiment resource with information needed to shift traffic between app versions
          - library: knative
            task: init-experiment
          # run tasks under the `finish` action at the end of an experiment   
          finish:
          # promote an app version
          # `https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/candidate.yaml` will be applied if candidate satisfies objectives
          # `https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/baseline.yaml` will be applied if candidate fails to satisfy objectives
          - library: common
            task: exec # promote the winning version
            with:
              cmd: kubectl
              args:
              - "apply"
              - "-f"
              - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
      duration: # 12 iterations, 20 seconds each
        intervalSeconds: 20
        iterationsPerLoop: 12
    ```

## Metric spec in-brief
A brief explanation of the key stanzas in a metric spec is given below.

### spec.params
`spec.params` is a list of name-value pairs containing the HTTP params iter8 needs to use when it issues a REST query to the metrics database for this metric. The value string can be templated; iter8 will substitute the placeholders in the value string using version variables. This process is described [here](/usage/metrics/how-iter8-queries-metrics).

### spec.description
`spec.description` is a human-readable description of the metric.

### spec.type
An iter8 metric can be of type `counter` or `gauge`. The value of a `counter` metric is non-decreasing over time. The value of a `gauge` metric can increase or decrease over time. 

??? example "Sample counter metric"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Metric
    metadata:
    name: request-count
    spec:
    params:
    - name: query
      value: sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$interval])) or on() vector(0)
    description: Number of requests
    type: counter
    provider: prometheus
    ```

??? example "Sample gauge metric"
    ```yaml
    apiVersion: iter8.tools/v2alpha1
    kind: Metric
    metadata:
    name: mean-latency
    spec:
    description: Mean latency
    units: milliseconds
    params:
    - name: query
      value: (sum(increase(revision_app_request_latencies_sum{revision_name='$revision'}[$interval]))or on() vector(0)) / (sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$interval])) or on() vector(0))
    type: gauge
    sampleSize: 
      name: request-count
    provider: prometheus
    ```

### spec.provider
`spec.provider` denotes the type of the metric database that provides this metric. Currently, `prometheus` is the only supported value for this field. Support for other providers are planned as part of the [roadmap](/roadmap). Details about the Prometheus database URL used for metric queries are [here](/usage/metrics/metric-databases).

### spec.units
`spec.units` denotes the unit of measurement for the metric. Some metrics such as `request_count` in the above sample may not have units.

### spec.sampleSize
`spec.sampleSize` denotes the number of data points over which the metric is computed. This field applies only to `gauge` metrics. This field is described here [here](/usage/metrics/custom-metrics).

## Custom metrics
Creation of custom counter and gauge metric is described [here](/usage/metrics/custom-metrics).
