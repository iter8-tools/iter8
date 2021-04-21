---
template: main.html
---

# Creating New Iter8 Metrics

This document describes how an end-user can define new Iter8 metrics and supply (optional) authentication information needed by Iter8 for querying the metrics provider. The samples provided in this document differ in the following aspects.

* Providers[^1]: Prometheus, NewRelic, Sysdig, or Elastic
* HTTP request method: GET or POST
* HTTP request authentication method: no authentication, basic auth, bearer token, or API keys
* Format of the JSON response returned by the provider
* The `jq` expression used by Iter8 to extract the metric value from the JSON response

## Defining metrics and supplying authentication information

Metrics are defined by end-users, and referenced within experiment specifications. Authentication info (optional) is also supplied by end-users.

=== "Prometheus (no auth)"
    The following is an example of an Iter8 metric with Prometheus as the provider. This example does not involve authentication.
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
    spec:
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{service_name='${name}',${userfilter}}[${elapsedTime}s])) or on() vector(0)
      description: Number of requests
      type: Counter
      provider: prometheus
      jqExpression: ".data.result[0].value[1] | tonumber"
      urlTemplate: http://myprometheusservice.com/api/v1
    ```

    ???+ hint "Brief explanation of the `request-count` metric"
        1. Prometheus enables metric queries using HTTP GET requests. The `request-count` metric defined above uses the HTTP GET method; `GET` is the default value for the `method` field of a metric.
        2. Iter8 will query Prometheus during each iteration of the experiment. In each iteration, Iter8 will use `n` HTTP queries to fetch metric values for each version, where `n` is the number of versions in the experiment[^2].
        3. The HTTP query used by Iter8 contains a single query parameter named `query` (Line 7) as [required by Prometheus](https://prometheus.io/docs/prometheus/latest/querying/api/). The value of this parameter is derived by [substituting the placeholders](#placeholder-substitution) in the value string (Line 9).
        4. The `urlTemplate` field provides the URL of the prometheus service (Line 14).

=== "Prometheus + Basic auth"
    Prometheus can be setup in conjunction with a reverse proxy, which in turn can support HTTP request authentication, as described [here](https://prometheus.io/docs/guides/basic-auth/). Suppose the proxy layer is set up to enforce basic auth with the following credentials:

    ```yaml
    username: produser
    password: t0p-secret
    ```

    You can enable Iter8 to query this Prometheus instance as follows.

    1. **Create secret:** Create a Kubernetes secret containing the authentication information. In particular, this secret needs to have `username` and `password` keys in the `data` section.
    ```shell
    kubectl create secret generic promcredentials -n myns --from-literal=username=produser --from-literal=password=t0p-secret
    ```

    2. **Create RBAC rule:** Provide the required permissions for Iter8 to read this secret. The service account `iter8-analytics` in the `iter8-system` namespace will have permissions to read secrets in the `myns` namespace.
    ```shell
    kubectl create rolebinding iter8-cred --clusterrole=iter8-secret-reader-analytics --serviceaccount=iter8-system:iter8-analytics --namespace=myns
    ```

    3. **Define metric with basic auth:** When defining the metric, ensure that the `authType` field is set to `Basic` and the appropriate secret is referenced.
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
    spec:
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{service_name='${name}',${userfilter}}[${elapsedTime}s])) or on() vector(0)
      description: Number of requests
      type: Counter
      authType: Basic
      secret: myns/promcredentials
      provider: prometheus
      jqExpression: ".data.result[0].value[1] | tonumber"
      urlTemplate: http://myprometheusservice.com/api/v1
    ```

    The `request-count` metric defined above is an enhancement of the `request-count` metric defined in the *Prometheus (no auth)* tab. For a brief explanation of this metric, click on that tab.


=== "New Relic"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

## Placeholder substitution during Iter8's HTTP request

Iter8 will substitute placeholders in the metric query based on the time elapsed since the start of the experiment, and information associated with each version in the experiment.

Suppose the [metrics defined above](#metric-and-auth-samples) are referenced within an experiment as follows. Further, suppose this experiment has started, Iter8 is about to do an iteration of this experiment, and the time elapsed since the start of the experiment is 600 seconds.

??? abstract "Look inside sample experiment"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Experiment
    metadata:
      name: sample-exp
    spec:
      target: default/sample-app
      strategy:
        testingPattern: A/B
      criteria:
        requestCount: namespace-of-metric/request-count
        objectives:
        - metric: namespace-of-metric/mean-latency
          upperLimit: 50
        - metric: namespace-of-metric/95th-percentile-tail-latency
          upperLimit: 100
        - metric: namespace-of-metric/error-rate
          upperLimit: "0.01"
      duration:
        intervalSeconds: 10
        iterationsPerLoop: 10
      versionInfo:
        baseline:
          name: current
          variables:
          - name: revision
            value: sample-app-v1 
          - name: userfilter
            value: 'usergroup!~"wakanda"'
        candidates:
        - name: candidate
          variables:
          - name: revision
            value: sample-app-v2
          - name: userfilter
            value: 'usergroup=~"wakanda"'
    ```

=== "Prometheus"
    For the sample experiment above, Iter8 will use two HTTP queries to fetch metric values, one for the baseline version, and another for the candidate version.

    1.  For the baseline version, Iter8 will send an HTTP request with a single parameter named `query` whose value equals:
    ```
    sum(increase(revision_app_request_latencies_count{service_name='current',usergroup!~"wakanda"}[600s])) or on() vector(0)
    ```
    The placeholder `$elapsedTime` has been substituted with 600, which is the time elapsed since the start of the experiment. The other placeholders have been substituted based on information associated with the baseline version in the experiment.
    
    2.  For the candidate version, Iter8 will send an HTTP request with a single parameter named `query` whose value equals:
    ```
    sum(increase(revision_app_request_latencies_count{service_name='candidate',usergroup=~"wakanda"}[600s])) or on() vector(0)
    ```
    The placeholder `$elapsedTime` has been substituted with 600, which is the time elapsed since the start of the experiment. The other placeholders have been substituted based on information associated with the candidate version in the experiment.

=== "New Relic"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

## JSON response format

The metrics provider is expected to respond to Iter8's HTTP request with a JSON object. The format of this JSON object is defined by the provider.

=== "Prometheus"
    The format of the Prometheus JSON response is [defined here](https://prometheus.io/docs/prometheus/latest/querying/api/#format-overview). A sample Prometheus response is as follows.
    ```json linenums="1"
    {
      "status": "success",
      "data": {
        "resultType": "vector",
        "result": [
          {
            "value": [1556823494.744, "21.7639"]
          }
        ]
      }
    }    
    ```

=== "New Relic"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

## `jqExpression` used by Iter8 for extracting metric value

Iter8 uses [jq](https://stedolan.github.io/jq/) to extract the metric value from the JSON response of the provider. The `jqExpression` used by Iter8 is supplied as part of the metric definition. When the `jqExpression` is applied to the JSON response, it is expected to yield a number.

=== "Prometheus"
    ```shell
    echo '{
      "status": "success",
      "data": {
        "resultType": "vector",
        "result": [
          {
            "value": [1556823494.744, "21.7639"]
          }
        ]
      }
    }' | jq ".data.result[0].value[1] | tonumber"
    ```
    Executing the above command results yields `21.7639`, a number, as required by Iter8. 
    
=== "New Relic"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

**Note:** The shell command above is for illustration only. Iter8 uses Python bindings for `jq` to evaluate the `jqExpression`.

## Error handling
Errors may occur during Iter8's metric queries due to a number of reasons (for example, due to an invalid `jqExpression` supplied within the metric). If Iter8 encounters errors during its attempt to retrieve metric values, Iter8 will mark the respective metric as unavailable.

[^1]: Iter8 can be used with any provider that can receive an HTTP request and respond with a JSON object containing the metrics information. The providers described here are the currently documented ones.
[^2]: In a conformance experiment, `n = 1`. In canary and A/B experiments, `n = 2`. In A/B/n experiments, `n > 2`.