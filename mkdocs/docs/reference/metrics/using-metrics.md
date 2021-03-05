---
template: overrides/main.html
---

# Using Metrics in Experiments

!!! abstract ""
    List metrics available in your cluster using the `kubectl get metrics.iter8.tools` command. Use metrics in experiments by referencing them in `spec.criteria` field.

## Listing metrics
Iter8 metrics are Kubernetes resources which means you can list them using `kubectl get`.

``` shell
kubectl get metrics.iter8.tools --all-namespaces
```
```shell
NAMESPACE      NAME                           TYPE      DESCRIPTION
iter8-system   95th-percentile-tail-latency   gauge     95th percentile tail latency
iter8-system   error-count                    counter   Number of error responses
iter8-system   error-rate                     gauge     Fraction of requests with error responses
iter8-system   mean-latency                   gauge     Mean latency
iter8-system   request-count                  counter   Number of requests
```

## Referencing metrics

References to metrics in the `spec.criteria` field of an experiment must be in the `namespace/name` format.

??? example "Sample experiment illustrating the use of metrics"
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
        - metric: iter8-system/mean-latency
          upperLimit: 50
        # 95th percentile latency should be under 100 milliseconds
        - metric: iter8-system/95th-percentile-tail-latency
          upperLimit: 100
        # error rate should be under 1%
        - metric: iter8-system/error-rate
          upperLimit: "0.01"
      indicators:
      # report values for the following metrics in addition those in spec.criteria.objectives
      - iter8-system/99th-percentile-tail-latency
      - iter8-system/90th-percentile-tail-latency
      - iter8-system/75th-percentile-tail-latency
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

## Observing metric values

During an experiment, Iter8 reports the metric values observed for each version. [Use `iter8ctl`](http://localhost:8000/concepts/observability/) to observe these metric values in realtime.