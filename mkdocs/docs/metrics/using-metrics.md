---
template: overrides/main.html
---

# Using Metrics in Experiments

!!! tip "Iter8 metrics API"    
    Iter8 defines a new Kubernetes resource called Metric that makes it easy to use metrics in experiments from RESTful metric backends like Prometheus, New Relic, Sysdig and Elastic.

    List metrics available in your cluster using the `kubectl get metrics.iter8.tools` command. Use metrics in experiments by referencing them in experiment criteria.

## Listing metrics
Iter8 metrics are Kubernetes resources which means you can list them using `kubectl get`.

``` shell
kubectl get metrics.iter8.tools --all-namespaces
```

```shell
NAMESPACE       NAME                           TYPE      DESCRIPTION
iter8-knative   95th-percentile-tail-latency   gauge     95th percentile tail latency
iter8-knative   error-count                    counter   Number of error responses
iter8-knative   error-rate                     gauge     Fraction of requests with error responses
iter8-knative   mean-latency                   gauge     Mean latency
iter8-knative   request-count                  counter   Number of requests
```

## Referencing metrics

Use metrics in experiments by referencing them in criteria section. Reference metrics using the `namespace/name` or `name` [format](/reference/apispec/#criteria).

??? example "Sample experiment illustrating the use of metrics"
    ```yaml
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
        deploymentPattern: Progressive
        actions:
          start: # run the following sequence of tasks at the start of the experiment
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
        requestCount: iter8-knative/request-count
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
          - name: revision
            value: sample-app-v1 
          - name: promote
            value: baseline
        candidates:
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2
          - name: promote
            value: candidate 
    ```

## Observing metric values

During an experiment, Iter8 reports the metric values observed for each version. Use `iter8ctl` to observe these metric values in realtime. See [here](/getting-started/quick-start/with-knative/#7-observe-experiment) for an example.