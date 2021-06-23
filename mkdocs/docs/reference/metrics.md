---
template: main.html
---

# Metric Resource

!!! abstract "Metric resource"
    Iter8 defines the **Metric** resource type, which encapsulates the REST query that is used to retrieve a metric value from the metrics provider. Metric resources are referenced in experiments.


!!! note "Version"
    This document describes version `v2alpha2` of Iter8's metric API.

Metrics usage is documented [here](../metrics/using-metrics.md). Iter8 provides a set of builtin metrics that are documented [here](../metrics/builtin.md). Creation of custom metrics is documented [here](../metrics/custom.md). It is also possible to mock metric values; mock metrics are documented [here](../metrics/mock.md).

??? example "Sample metric"
    ```yaml linenums="1"
    apiVersion: iter8.tools/v2alpha2
    kind: Metric
    metadata:
      name: request-count
    spec:
      params:
      - name: query
        value: |
          sum(increase(revision_app_request_latencies_count{revision_name='$revision'}[$elapsedTime])) or on() vector(0)
      description: Number of requests
      type: counter
      provider: prometheus
      jqExpression: ".data.result[0].value[1] | tonumber"
      urlTemplate: http://prometheus-operated.iter8-system:9090/api/v1/query      
    ```

#### Metadata
Standard Kubernetes [meta.v1/ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta) resource.

#### Spec
| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| description | string | Human readable description. This field is meant for informational purposes. | No |
| units | string | Units of measurement. This field is meant for informational purposes. | No |
| provider | string | Type of the metrics provider (example, `prometheus`, `newrelic`, `sysdig`, `elastic`, ...). The keyword `iter8` is reserved for Iter8 builtin metrics. | No |
| params | [][NamedValue](../experiment/#namedvalue) | List of name/value pairs corresponding to the name and value of the HTTP query parameters used by Iter8 when querying the metrics provider. Each name represents a parameter name; the corresponding value is a string template with placeholders; the placeholders will be dynamically substituted by Iter8 with values at query time. | No |
| body | string | String used to construct the JSON body of the HTTP request. Body may be templated, in which Iter8 will attempt to substitute placeholders in the template at query time using version information. | No |
| type | string | Metric type. Valid values are `Counter` and `Gauge`. Default value = `Gauge`. A `Counter` metric is one whose value never decreases over time. A `Gauge` metric is one whose value may increase or decrease over time. | No |
| method | string | HTTP method (verb) used in the HTTP request. Valid values are `GET` and `POST`. Default value = `GET`. | No |
| authType | string | Identifies the type of authentication used in the HTTP request. Valid values are `Basic`, `Bearer` and `APIKey` which correspond to HTTP authentication with these respective methods. | No |
| sampleSize | string | Reference to a metric that represents the number of data points over which the value of this metric is computed. This field applies only to `Gauge` metrics. References can be expressed in the form 'name' or 'namespace/name'. If just `name` is used, the implied namespace is the namespace of the referring metric. | No |
| secret | string | Reference to a secret that contains information used for authenticating with the metrics provider. In particular, Iter8 uses data in this secret to substitute placeholders in the HTTP headers and URL while querying the provider. References can be expressed in the form 'name' or 'namespace/name'. If just `name` is used, the implied namespace is the namespace where Iter8 is installed (which is `iter8-system` by default). | No |
| headerTemplates | [][NamedValue](../experiment/#namedvalue) | List of name/value pairs corresponding to the name and value of the HTTP request headers used by Iter8 when querying the metrics provider. Each name represents a header field name; the corresponding value is a string template with placeholders; the placeholders will be dynamically substituted by Iter8 with values at query time. Placeholder substitution is attempted only if `authType` and `secret` fields are present. | No |
| jqExpression | string | The [jq](https://stedolan.github.io/jq/) expression used by Iter8 to extract the metric value from the JSON response returned by the provider. This field may be omitted if the metric is [mocked](../../metrics/mock). | No |
| urlTemplate | string | Template for the metric provider's URL. Typically, urlTemplate is expected to be the actual URL without any placeholders. However, urlTemplate may be templated, in which case, Iter8 will attempt to substitute placeholders in the urlTemplate at query time using the `secret` referenced in the metric. Placeholder substitution will not be attempted if `secret` is not specified. This field may be omitted if the metric is [mocked](../../metrics/mock). | No |
| urlTemplate | string | Template for the metric provider's URL. Typically, urlTemplate is expected to be the actual URL without any placeholders. However, urlTemplate may be templated, in which case, Iter8 will attempt to substitute placeholders in the urlTemplate at query time using the `secret` referenced in the metric. Placeholder substitution will not be attempted if `secret` is not specified. This field may be omitted if the metric is [mocked](../../metrics/mock). | No |
| mock | [][NamedLevel](#named-level) | Information on how to mock this metric. The presence of this fields indicates that this metric will be mocked; the absence of this field indicates that this is a real metric. | No |

##### Named Level

| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| name | string | Name of the version. Version names should be unique. Version name should match the name of a version in the experiment's `versionInfo` section. If not, any value generated for the non-matching name will be ignored. | Yes |
| level | string | Level of a version. The semantics of level are as follows. If the metric is a counter, if level is `x`, and time elapsed since the start of the experiment is `y` seconds, then `xy` is the mocked metric value. This will keep increasing the metric value over time. If the metric is gauge, if level is `x`, the metric value is a random value with mean `x`. The expected value of the mocked metric is `x`; the actual value may increase or decrease over time. | Yes |
