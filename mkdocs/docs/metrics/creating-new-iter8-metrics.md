---
template: main.html
---

# Creating New Iter8 Metrics

This document describes how an end-user can define new Iter8 metrics and (optionally) supply authentication information that may be required for querying the metrics provider. The samples provided in this document differ in the following aspects.

* Providers[^1]: Prometheus, NewRelic, Sysdig, or Elastic
* HTTP request authentication method: no authentication, basic auth, API keys, or bearer token
* HTTP request method: GET or POST
* Format of the JSON response returned by the provider
* The `jq` expression used by Iter8 to extract the metric value from the JSON response

## Defining metrics

> **Note:** Metrics are defined by the **end-user**.

=== "Prometheus"

    Prometheus does not support any authentication mechanism *out-of-the-box*. However,     Prometheus can be setup in conjunction with a reverse proxy, which in turn can support HTTP request authentication, as described [here](https://prometheus.io/docs/guides/basic-auth/).

    === "No Authentication"
        The following is an example of an Iter8 metric with Prometheus as the provider. This example assumes that Prometheus can be queried by Iter8 without any authentication.

        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Metric
        metadata:
          name: request-count
        spec:
          params:
          - name: query
            value: >-
              sum(increase(revision_app_request_latencies_count{service_name='${name}',${userfilter}}[${elapsedTime}s])) or on() vector(0)
          description: A Prometheus example
          type: Counter
          provider: prometheus
          jqExpression: ".data.result[0].value[1] | tonumber"
          urlTemplate: http://myprometheusservice.com/api/v1
        ```

    === "Basic auth"
        Suppose Prometheus, in conjunction with the reverse proxy, is set up to enforce basic auth with the following credentials:

        ```yaml
        username: produser
        password: t0p-secret
        ```

        You can then enable Iter8 to query this Prometheus instance as follows.

        1. **Create secret:** Create a Kubernetes secret containing the authentication information. In particular, this secret needs to have `username` and `password` keys in the `data` section.
        ```shell
        kubectl create secret generic promcredentials -n myns --from-literal=username=produser --from-literal=password=t0p-secret
        ```

        2. **Create RBAC rule:** Provide the required permissions for Iter8 to read this secret. The service account `iter8-analytics` in the `iter8-system` namespace will have permissions to read secrets in the `myns` namespace.
        ```shell
        kubectl create rolebinding iter8-cred --clusterrole=iter8-secret-reader-analytics --serviceaccount=iter8-system:iter8-analytics --namespace=myns
        ```

        3. **Define metric:** When defining the metric, ensure that the `authType` field is set to `Basic` and the appropriate `secret` is referenced.
        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Metric
        metadata:
          name: request-count
        spec:
          params:
          - name: query
            value: >-
              sum(increase(revision_app_request_latencies_count{service_name='${name}',${userfilter}}[${elapsedTime}s])) or on() vector(0)
          description: A Prometheus example
          type: Counter
          authType: Basic
          secret: myns/promcredentials
          provider: prometheus
          jqExpression: ".data.result[0].value[1] | tonumber"
          urlTemplate: https://my.secure.prometheus.service.com/api/v1
        ```

    ???+ hint "Brief explanation of the `request-count` metric"
        1. Prometheus enables metric queries using HTTP GET requests. `GET` is the default value for the `method` field of an Iter8 metric. This field is optional; it is omitted in the  definition of `request-count`, and defaulted to `GET`.
        2. Iter8 will query Prometheus during each iteration of the experiment. In each iteration, Iter8 will use `n` HTTP queries to fetch metric values for each version, where `n` is the number of versions in the experiment[^2].
        3. The HTTP query used by Iter8 contains a single query parameter named `query` as [required by Prometheus](https://prometheus.io/docs/prometheus/latest/querying/api/). The value of this parameter is derived by [substituting the placeholders](#placeholder-substitution) in the value string.
        4. The `urlTemplate` field provides the URL of the prometheus service.

=== "New Relic"
    New Relic uses API Keys to authenticate requests as documented [here](https://docs.newrelic.com/docs/apis/rest-api-v2/get-started/introduction-new-relic-rest-api-v2/). The API key may be directly specified within the metric, or supplied as part of a Kubernetes secret.

    === "API key embedded in metric"
        The following is an example of an Iter8 metric with Prometheus as the provider. In this example, `t0p-secret-api-key` is the New Relic API key.

        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Metric
        metadata:
          name: name-count
        spec:
          params:
          - name: nrql
            value: >-
              SELECT count(appName) FROM PageView WHERE revisionName='${revision}' SINCE ${elapsedTime} seconds ago
          description: A New Relic example
          type: Counter
          headerTemplates:
          - name: X-Query-Key
            value: t0p-secret-api-key
          provider: newrelic
          jqExpression: ".results[0].count | tonumber"
          urlTemplate: https://insights-api.newrelic.com/v1/accounts/my_account_id
        ```

    === "API key within K8s secret"
        Suppose your New Relic API key is `t0p-secret-api-key`; you wish to store this API key in a Kubernetes secret, and reference this secret in an Iter8 metric. You can do so as follows.

        1. **Create secret:** Create a Kubernetes secret containing the API key.
        ```shell
        kubectl create secret generic nrcredentials -n myns --from-literal=mykey=t0p-secret-api-key
        ```
        The above secret contains a data field named `mykey` whose value is the API key. The data field name (which can be any string of your choice) will be used in Step 3 below as a placeholder.

        2. **Create RBAC rule:** Provide the required permissions for Iter8 to read this secret. The service account `iter8-analytics` in the `iter8-system` namespace will have permissions to read secrets in the `myns` namespace.
        ```shell
        kubectl create rolebinding iter8-cred --clusterrole=iter8-secret-reader-analytics --serviceaccount=iter8-system:iter8-analytics --namespace=myns
        ```

        3. **Define metric:** When defining the metric, ensure that the `authType` field is set to `APIKey` and the appropriate `secret` is referenced. In the `headerTemplates` field, include `X-Query-Key` as the name of a header field (as [required by New Relic](https://docs.newrelic.com/docs/insights/event-data-sources/insights-api/query-insights-event-data-api/#create-request)). The value for this header field is a templated string. Iter8 will substitute the placeholder ${mykey} at query time, by looking up the referenced `secret` named `nrcredentials` in the `myns` namespace.

        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Metric
        metadata:
          name: name-count
        spec:
          params:
          - name: nrql
            value: >-
              SELECT count(appName) FROM PageView WHERE revisionName='${revision}' SINCE ${elapsedTime} seconds ago
          description: A New Relic example
          type: Counter
          authType: APIKey
          secret: myns/nrcredentials
          headerTemplates:
          - name: X-Query-Key
            value: ${mykey}
          provider: newrelic
          jqExpression: ".results[0].count | tonumber"
          urlTemplate: https://insights-api.newrelic.com/v1/accounts/my_account_id
        ```

    ???+ hint "Brief explanation of the `name-count` metric"
        1. New Relic enables metric queries using both HTTP GET and POST requests. `GET` is the default value for the `method` field of an Iter8 metric. This field is optional; it is omitted in the definition of `name-count`, and defaulted to `GET`.
        2. Iter8 will query New Relic during each iteration of the experiment. In each iteration, Iter8 will use `n` HTTP queries to fetch metric values for each version, where `n` is the number of versions in the experiment[^2].
        3. The HTTP query used by Iter8 contains a single query parameter named `nrql` as [required by New Relic](https://docs.newrelic.com/docs/insights/event-data-sources/insights-api/query-insights-event-data-api/). The value of this parameter is derived by [substituting the placeholders](#placeholder-substitution) in its value string.
        4. The `urlTemplate` field provides the URL of the New Relic service.

=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

## Placeholder substitution

> **Note:** This step is automated by **Iter8**.

Iter8 will substitute placeholders in the metric query based on the time elapsed since the start of the experiment, and information associated with each version in the experiment.

Suppose the [metrics defined above](#defining-metrics) are referenced within an experiment as follows. Further, suppose this experiment has started, Iter8 is about to do an iteration of this experiment, and the time elapsed since the start of the experiment is 600 seconds.

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

For the sample experiment above, Iter8 will use two HTTP(S) queries to fetch metric values, one for the baseline version, and another for the candidate version.

=== "Prometheus"

    For the baseline version, Iter8 will send an HTTP(S) request with a single parameter named `query` whose value equals:
    ```
    sum(increase(revision_app_request_latencies_count{service_name='current',usergroup!~"wakanda"}[600s])) or on() vector(0)
    ```

=== "New Relic"
    For the baseline version, Iter8 will send an HTTP(S) request with a single parameter named `nrql` whose value equals:
    ```
    SELECT count(appName) FROM PageView WHERE revisionName='sample-app-v1' SINCE 600 seconds ago
    ```
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

The placeholder `$elapsedTime` has been substituted with 600, which is the time elapsed since the start of the experiment. The other placeholders have been substituted based on information associated with the baseline version in the experiment. Iter8 builds and sends an HTTP request in a similar manner for the candidate version as well.

## JSON response format

> **Note:** This step is handled by the **metrics provider**.

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
    The format of the New Relic JSON response is [discussed here](https://docs.newrelic.com/docs/insights/event-data-sources/insights-api/query-insights-event-data-api/#example). A sample New Relic response is as follows.
    ```json linenums="1"
    {
      "results": [
        {
          "count": 80275388
        }
      ],
      "metadata": {
        "eventTypes": [
          "PageView"
        ],
        "eventType": "PageView",
        "openEnded": true,
        "beginTime": "2014-08-03T19:00:00Z",
        "endTime": "2017-01-18T23:18:41Z",
        "beginTimeMillis=": 1407092400000,
        "endTimeMillis": 1484781521198,
        "rawSince": "'2014-08-04 00:00:00+0500'",
        "rawUntil": "`now`",
        "rawCompareWith": "",
        "clippedTimeWindows": {
          "Browser": {
            "beginTimeMillis": 1483571921198,
            "endTimeMillis": 1484781521198,
            "retentionMillis": 1209600000
          }
        },
        "messages": [],
        "contents": [
          {
            "function": "count",
            "attribute": "appName",
            "simple": true
          }
        ]
      }
    }
    ```
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

## Processing the JSON response

> **Note:** This step is automated by **Iter8**.

Iter8 uses [jq](https://stedolan.github.io/jq/) to extract the metric value from the JSON response of the provider. The `jqExpression` used by Iter8 is supplied as part of the metric definition. When the `jqExpression` is applied to the JSON response, it is expected to yield a number.

=== "Prometheus"
    Consider the `jqExpression` defined in the [sample Prometheus metric](#defining-metrics). Let us apply it to the [sample JSON response from Prometheus](#json-response-format).
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
    Consider the `jqExpression` defined in the [sample New Relic metric](#defining-metrics). Let us apply it to the [sample JSON response from New Relic](#json-response-format).
    ```shell
    echo '{
      "results": [
        {
          "count": 80275388
        }
      ],
      "metadata": {
        "eventTypes": [
          "PageView"
        ],
        "eventType": "PageView",
        "openEnded": true,
        "beginTime": "2014-08-03T19:00:00Z",
        "endTime": "2017-01-18T23:18:41Z",
        "beginTimeMillis=": 1407092400000,
        "endTimeMillis": 1484781521198,
        "rawSince": "'2014-08-04 00:00:00+0500'",
        "rawUntil": "`now`",
        "rawCompareWith": "",
        "clippedTimeWindows": {
          "Browser": {
            "beginTimeMillis": 1483571921198,
            "endTimeMillis": 1484781521198,
            "retentionMillis": 1209600000
          }
        },
        "messages": [],
        "contents": [
          {
            "function": "count",
            "attribute": "appName",
            "simple": true
          }
        ]
      }
    }' | jq ".results[0].count | tonumber"
    ```
    Executing the above command results yields `80275388`, a number, as required by Iter8. 
    
=== "Sysdig"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](../images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](../../tutorials/knative/conformance/).

=== "Elastic"

> **Note:** The shell command above is for illustration only. Iter8 uses Python bindings for `jq` to evaluate the `jqExpression`.

## Error handling

> **Note:** This step is automated by **Iter8**.

Errors may occur during Iter8's metric queries due to a number of reasons (for example, due to an invalid `jqExpression` supplied within the metric). If Iter8 encounters errors during its attempt to retrieve metric values, Iter8 will mark the respective metric as unavailable.

[^1]: Iter8 can be used with any provider that can receive an HTTP request and respond with a JSON object containing the metrics information. Documentation requests and contributions (PRs) are welcome for providers not listed here.
[^2]: In a conformance experiment, `n = 1`. In canary and A/B experiments, `n = 2`. In A/B/n experiments, `n > 2`.