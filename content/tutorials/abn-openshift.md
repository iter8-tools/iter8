---
menuTitle: A/B/n Rollout - OpenShift
title: A/B/n Rollout
weight: 26
summary: Learn how to perform an A/B/n rollout on Red Hat OpenShift
---

This tutorial shows how iter8 can be used to perform A/B/n rollout of several versions of a service to select the one that maximizes a reward metric while also satisfiying any other requirements.

{{% notice info %}}
This tutorial is for use with Red Hat OpenShift. A corresponding tutorial for plain Kubernetes is [here]({{< ref "abn" >}}).
{{% /notice %}}

This tutorial has eight steps, which are meant to be tried in order.
You will learn:

- how to define application specific metrics;
- how to specify a reward metric; and
- how to execute an A/B/n experiment with iter8;

The content of this tutorial is captured in this video (COMING SOON).

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) distributed with [Istio](https://istio.io).
This application comprises 4 microservies: *productpage*, *details*, *reviews*, and *ratings*.
Of these, *productpage* is a user-facing service while the others are backend services.

The version of the bookinfo *productpage* service used in this tutorial has been modified from the original to allow the following behaviors (all configurable via environment variables when deploying):

- change the *color* of a phrase "*William Shakespeare's*" in the *Summary* line of the returned page;
- configurable delays in query response time (*delay_seconds* with probability *delay_probabilities*); and
- creation of a configurable number of "books sold" (uniformly distributed in the range *reward_min* to *reward_max*) that is exposed as a metric to be periodically collected by Prometheus

These changes enable us to visually distinguish between versions when using a browser and to configure the behavior with respect to metrics.
The source code for these changes is available [here](https://github.com/iter8-tools/bookinfoapp-productpage/tree/productpage-reward).

{{% notice info %}}
This rest of this tutorial assumes you have already installed iter8 (including the Red Hat OpenShift Service Mesh). If not, do so using the instructions [here]({{< ref "red-hat" >}}).
{{% /notice %}}

## Define New Metrics

Out of the box, iter8 comes with a set of predefined metrics.
For details of metrics definitions provided in iter8, see the [metrics reference]({{< ref "metrics" >}}).

You can augment the default set of metrics by replacing `ConfigMap` *iter8config-metrics* (defined in the *iter8* namespace) with a new `ConfigMap`.

For demonstration purposes, we add two new metrics: a *ratio* metric which measures how many requests take more than 500ms to process and an application specific *counter* metric that captures the number of books sold by our service.
We will use the latter metric as our reward metric -- our experiment will select the version that maximize the number of books sold subject to satisfying all other criteria.

To define the ratio metric, add the following to the `counter_metrics.yaml` field of the map:

```yaml
- name: le_500_ms_latency_request_count
  query_template: (sum(increase(istio_request_duration_seconds_bucket{le='0.5',reporter='source',job='istio-mesh'}[$interval])) by ($version_labels))
- name: le_inf_latency_request_count
  query_template: (sum(increase(istio_request_duration_seconds_bucket{le='+Inf',reporter='source',job='istio-mesh'}[$interval])) by ($version_labels))
```

and the following to the `ratio_metrics.yaml` value:

```yaml
- name: le_500_ms_latency_percentile
  numerator: le_500_ms_latency_request_count
  denominator: le_inf_latency_request_count
  preferred_direction: higher
  zero_to_one: true
```

To define the reward metric, `books_purchased_total`, add the following to the `counter_metrics.yaml` field:

```yaml
- name: books_purchased_total
  query_template: sum(increase(number_of_books_purchased_total{}[$interval])) by ($version_labels)
```

We can do all of the above as follows:

```bash
kubectl --namespace iter8 apply -f {{< resourceAbsUrl path="tutorials/abn-tutorial/productpage-metrics-telemetry-v1.yaml" >}}
```

{{% notice tip %}}
The above discussion and command assumes that you are using a version of the Service Mesh that does not have the Istio *mixer* component disabled. If the mixer is disabled, use [{{< resourceAbsUrl path="tutorials/abn-tutorial/productpage-metrics.yaml" >}}]({{< resourceAbsUrl path="tutorials/abn-tutorial/productpage-metrics.yaml" >}}) instead.
{{% /notice %}}

## Configure Prometheus

The *productpage* `Deployment` definition includes annotations that direct Prometheus to scrape the pods for metrics; in particular, the reward metric we defined.
The annotation is:

```yaml
prometheus.io/scrape: "true"
prometheus.io/path: /metrics
prometheus.io/port: "9080"
```

Unfortunately, the Prometheus server installed with the Red Hat OpenShift Service Mesh expects communication with the pod to be implemented using mTLS.
To avoid this, reconfigure Prometheus:

```bash
oc --namespace istio-system edit configmap/prometheus
```

Find the `scrape_configs` entry with `job_name: 'kubernetes-pods`.
Comment out the entry with a `source_label` of `__meta_kubernetes_pod_annotation_prometheus_io_scrape` if one exists.
In this example, the last three lines have been commented out:

```yaml
- job_name: 'kubernetes-pods'
  kubernetes_sd_configs:
  - role: pod
  relabel_configs:  # If first two labels are present, pod should be scraped  by the istio-secure job.
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  # Keep target if there's no sidecar or if prometheus.io/scheme is explicitly set to "http"
  #- source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status, __meta_kubernetes_pod_annotation_prometheus_io_scheme]
  #  action: keep
  #  regex: ((;.*)|(.*;http))
```

Then restart the prometheus pod:

```bash
oc --namespace istio-system delete pod $(oc --namespace istio-system get pod --selector='app=prometheus' -o jsonpath='{.items[0].metadata.name}')
```

You should only have to do this once.

## Deploy the Bookinfo application

To deploy the Bookinfo application, create a namespace configured to enable auto-injection of the Istio sidecar. You can use whatever namespace name you wish. By default, the namespace `bookinfo-iter8` is created.

```bash
oc apply -f {{< resourceAbsUrl path="tutorials/namespace.yaml" >}}
```

Next, deploy the application:

```bash
oc --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/bookinfo-tutorial.yaml" >}}
```

You should see pods for each of the four microservices:

```bash
oc --namespace bookinfo-iter8 get pods
```

Note that we deployed version *v1* of the *productpage* microsevice; that is, *productpage-v1*.
Each pod should have two containers, since the Istio sidecar was injected into each.

## Expose the Bookinfo application

Expose the Bookinfo application by defining a `Gateway`, `VirtualService` and `DestinationRule`.
These will use the `Route` defined for the Istio ingress gateway:

```bash
export GATEWAY_URL=$(oc -n istio-system get route istio-ingressgateway -o jsonpath='{.spec.host}')
```

```bash
curl -s {{< resourceAbsUrl path="tutorials/bookinfo-gateway.yaml" >}} \
| sed "s#bookinfo.example.com#${GATEWAY_URL}#" \
| oc --namespace bookinfo-iter8 apply -f -
```

You can inspect the created resources:

```bash
oc --namespace bookinfo-iter8 get gateway,virtualservice,destinationrule
```

## Verify access to Bookinfo

You can check access to the application with the following `curl` command:

```bash
curl -o /dev/null -s -w "%{http_code}\n" "http://${GATEWAY_URL}/productpage"
```

If everything is working, the command above should return `200`.

## Generate load

To simulate user requests, use a command such as the following:

```bash
watch -n 0.1 'curl -s "http://${GATEWAY_URL}/productpage" | grep -i "color:"'
```

This command requests the `productpage` microservice 10 times per second.
In turn, this causes about the same frequency of requests against the backend microservices.
We filter the response to see the color being used to display the text "*William Shakespeare's*" in the *Summary* line of the page.
The color varies between versions giving us a visual way to distinguish between them.
Initially it should be *red*.

## Create an A/B/n `Experiment`

We will now define an A/B/n experiment to compare versions *v2* and *v3* of the *productpage* application to the existing version, *v1*.
These versions are visually identical except for the color of the text "*William Shakespeare's*" that appears in the page *Summary*.
In version *v2* they are *gold* and in version *v3* they are *green*.

In addition to the visual difference, the *v2* version has a high number of books sold (it is greatest of the three versions) but it has a long response time.
The *v3* version has an intermediate number of books sold (compared to the other versions) and a response time comparable to the *v1* version.

To describe a A/B/n experiment, create an iter8 `Experiment` that identifies the original, or *baseline* version and the new, or *candidate* versions and some evaluation criteria including the identification of a *reward* metric.
For example:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: productpage-abn-test
spec:
  service:
    name: productpage
    baseline: productpage-v1
    candidates:
      - productpage-v2
      - productpage-v3
    hosts:
      - name: bookinfo.example.com
        gateway: bookinfo-gateway
  criteria:
    - metric: iter8_mean_latency
      threshold:
        type: relative
        value: 1.6
    - metric: iter8_error_rate
      threshold:
        type: absolute
        value: 0.05
    - metric: le_500_ms_latency_percentile
      threshold:
        type: absolute
        value: 0.95
    - metric: mean_books_purchased
      isReward: true
  duration:
    interval: 20s
    maxIterations: 20
  trafficControl:
    strategy: progressive
    maxIncrement: 10
```

In this example, the target of the experiment is the service `productpage`.
The baseline and candidate versions are specified using their `Deployment` names; `productpage-v1` for the baseline version and `productpage-v2` and `productpage-v3` for the candidate versions.
Three evaluation criteria are specified and a reward metric is identified.
The evaluation criteria ensure that:

- compared to the baseline, the candidates `iter8_mean_latency` should less than 1.6 times greater;
- less than 5% of queries should return an error; and
- 95% of latencies should be less than 500ms.

Additionally, the reward metric is `mean_books_purchased`.

The additional parameters control how long the experiment should run and how much traffic can be shifted to the new version in each interval. Details regarding these parameters are [here](#alter-the-duration-of-the-experiment).

The experiment can be created using the command:

```bash
curl -s {{< resourceAbsUrl path="tutorials/abn-tutorial/abn_productpage_v1v2v3.yaml" >}} \
| sed sed "s#bookinfo.example.com#${GATEWAY_URL}#" \
| oc --namespace bookinfo-iter8 apply -f -
```

Inspection of the new experiment shows that it is paused because the specified candidate versions cannot be found in the cluster:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                   TYPE    HOSTS                                PHASE   WINNER FOUND   CURRENT BEST   STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Pause                                 TargetsError: Missing Candidate
```

Once the candidate versions are deployed, the experiment will start automatically.

## Deploy the candidate versions of the _productpage_ service

To deploy the *v2* and *v3* versions of the *productpage* microservice, execute:

```bash
oc --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/productpage-v2.yaml" >}} -f {{< resourceAbsUrl path="tutorials/productpage-v3.yaml" >}}
```

Once its corresponding pods have started, the `Experiment` will show that it is progressing:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                   TYPE    HOSTS                                PHASE         WINNER FOUND   CURRENT BEST     STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Progressing   false                           IterationUpdate: Iteration 3/20 completed
```

At approximately 20 second intervals, you should see the interation number change.
Over time, you should see a decision that one of the versions (it could be the baseline version) is identified as the *winner*.
To make this determination, iter8 will evaluate the specified criteria and apply advanced analytics to make a determination.
Based on intermediate evaluations, traffic will be adjusted between the versions.
iter8 will eventually identify that the best version is the candidate, `productpage-v3` and that it is confident that this choice will be the final choice (by indicating that a *winner* has been found:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                   TYPE    HOSTS                                PHASE         WINNER FOUND   CURRENT BEST     STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Progressing   true           productpage-v3   IterationUpdate: Iteration 7/20 completed
```

If iter8 is unable to determine a winner with confidence, the experiment will fail and the best choice will default to the baseline version.

When the experiment is finished (about 5 minutes), you will see that all traffic has been shifted to the winner, *productpage-v3*:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                   TYPE    HOSTS                                PHASE       WINNER FOUND   CURRENT BEST     STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Completed   true           productpage-v3   ExperimentCompleted: Traffic To Winner
```

## Cleanup

To clean up, delete the namespace:

```bash
oc delete namespace bookinfo-iter8
```

## Other things to try (before cleanup)

### Inspect progress using Grafana

You can inspect the progress of your experiment using the sample *iter8 Metrics* dashboard. To install this dashboard, see [here]({{< ref "grafana" >}}).

### Inspect progress using Kiali

Coming soon

### Alter the duration of the experiment

The progress of an experiment can be impacted by `duration` and `trafficControl` parameters:

- `duration.maxIterations` defines the number of iterations in the experiment (*default: 100*)
- `duration.interval` defines the duration of each iteration (*default: 30 seconds*)
- `trafficControl.maxIncrement` identifies the largest change (increment) that will be made in the percentage of traffic sent to a candidate (*default: 2 percent*)

The impact of the first two parameters on the duration of the experiment are clear.
Restricting the size of traffic increment also influences how quickly an experiment can come to a decision about a candidate.
