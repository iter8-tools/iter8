---
date: 2020-08-05T12:00:00+00:00
menuTitle: A/B/N testing
title: Getting started with A/B/N testing
weight: 20
summary: Learn how to perform an A/B/N test
---

This tutorial shows how _iter8_ can be used to perform A/B/N testing on several versions of a service to select the one that maximizes a reward metric while also satisfiying any other requirements.

This tutorial has eight steps, which are meant to be tried in order.
You will learn:

- how to define application specific metrics;
- how to specify a reward metric; and
- how to execute an A/B/N experiment with iter8;

The content of this tutorial is captured in this video (COMING SOON).

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) distributed with [Istio](https://istio.io).
This application comprises 4 microservies: _productpage_, _details_, _reviews_, and _ratings_.
Of these, _productpage_ is a user-facing service while the others are backend services.

The version of the bookinfo _productpage_ service used in this tutorial has been modified from the original to allow the following behaviors (all configurable via environment variables when deploying):

- change the *color* of a phrase "*William Shakespeare's*" in the *Summary* line of the returned page;
- configurable delays in query response time (*delay_seconds* with probability *delay_probabilities*); and
- creation of a configurable number of "books sold" (uniformly distributed in the range *reward_min* to *reward_max*) that is exposed as a metric to be periodically collected by Prometheus

These changes enable us to visually distinguish between versions when using a browser and to configure the behavior with respect to metrics.
The source code for these changes is available [here](https://github.com/iter8-tools/bookinfoapp-productpage/tree/productpage-reward).

**Note** This rest of this tutorial assumes you have already installed _iter8_ (including Istio). If not, do so using the instructions [here](../installation/kubernetes/).

## Define New Metrics

Out of the box, iter8 comes with a set of predefined metrics.
For details of metrics definitions provided in iter8, see the [metrics reference](../../reference/metrics).

You can augment the default set of metrics by replacing `ConfigMap` *iter8config-metrics* (defined in the *iter8* namespace) with a new `ConfigMap`.

For demonstration purposes, we add two new metrics: a  *ratio* metric which measures how many requests take more than 500ms to process and an application specific *counter* metric that captures the number of books sold by our service.
We will use the latter metric as our reward metric -- our experiment will select the version that maximize the number of books sold subject to satisfying all other criteria.

To define the ratio metric, add the following to the `counter_metrics.yaml` field of the map:

```yaml
- name: le_500_ms_latency_request_count
  query_template: (sum(increase(istio_request_duration_milliseconds_bucket{le='500',job='envoy-stats',reporter='source'}[$interval])) by ($version_labels))
- name: le_inf_latency_request_count
  query_template: (sum(increase(istio_request_duration_milliseconds_bucket{le='+Inf',job='envoy-stats',reporter='source'}[$interval])
```

and the following to the ratio_metrics.yaml` value:

```yaml
- name: le_500_ms_latency_percentile
  numerator: le_500_ms_latency_request_count
  denominator: le_inf_latency_request_count
  preferred_direction: higher
  zero_to_one: true
```

To define the reward metric, ``, add the following to the `counter_metrics.yaml` field:

```yaml
- name: books_purchased_total
  query_template: sum(increase(number_of_books_purchased_total{}[$interval])) by ($version_labels)
```

We can do all of the above as follows:

```bash
kubectl --namespace iter8 apply -f {{< resourceAbsUrl path="tutorials/abn-tutorial/productpage-metrics.yaml" >}}
```

**Note** The above command assumes that you are using a new version of Istio (version 1.5 or greater) and not using it's "mixer" component. If the mixer is being used, use this [file]({{< resourceAbsUrl path="tutorials/abn-tutorial/productpage-metrics-telemetry-v1.yaml" >}} ) instead.

## Configure Prometheus

The *productpage* `Deployment` definition includes annotations that direct Prometheus to scrape the pods for metrics; in particular, the reward metric we defined.
The annotation is:

```yaml
prometheus.io/scrape: "true"
prometheus.io/path: /metrics
prometheus.io/port: "9080"
```

Unfortunately, the Prometheus server installed with Istio expects communication with the pod to be implemented using mTLS.
To avoid this, reconfigure Prometheus:

```bash
kubectl --namespace istio-system edit configmap/prometheus
```

Comment out six lines as follows:

```yaml
- job_name: 'kubernetes-pods'
  kubernetes_sd_configs:
  - role: pod
  relabel_configs:  # If first two labels are present, pod should be scraped  by the istio-secure job.
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  #- source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status]
  #  action: drop
  #  regex: (.+)
  #- source_labels: [__meta_kubernetes_pod_annotation_istio_mtls]
  #  action: drop
  #  regex: (true)
```

Then restart the prometheus pod:

```bash
kubectl --namespace istio-system delete pod $(kubectl --namespace istio-system get pod --selector='app=prometheus' -o jsonpath='{.items[0].metadata.name}')
```

You should only have to do this once.

## Deploy the Bookinfo application

To deploy the Bookinfo application, create a namespace configured to enable auto-injection of the Istio sidecar. You can use whatever namespace name you wish. By default, the namespace `bookinfo-iter8` is created.

```bash
export NAMESPACE=bookinfo-iter8
curl -s {{< resourceAbsUrl path="tutorials/namespace.yaml" >}} \
  | sed "s#bookinfo-iter8#$NAMESPACE#" \
  | kubectl apply -f -
```

Next, deploy the application:

```bash
kubectl --namespace $NAMESPACE apply -f {{< resourceAbsUrl path="tutorials/bookinfo-tutorial.yaml" >}}
```

You should see pods for each of the four microservices:

```bash
kubectl --namespace $NAMESPACE get pods
```

Note that we deployed version *v1* of the *productpage* microsevice; that is, *productpage-v1*.
Each pod should have two containers, since the Istio sidecar was injected into each.

## Expose the Bookinfo application

Expose the Bookinfo application by defining an Istio `Gateway`, `VirtualService` and `DestinationRule`:

```bash
kubectl --namespace $NAMESPACE apply -f {{< resourceAbsUrl path="tutorials/bookinfo-gateway.yaml" >}}
```

You can inspect the created resources:

```bash
kubectl --namespace $NAMESPACE get gateway,virtualservice,destinationrule
```

Note that the service has been associated with a fake host, `bookinfo.example.com` for demonstration purposes.

## Verify access to Bookinfo

To access the application, determine the ingress IP and port for the application.
You can do so by following steps 3 and 4 of the Istio instructions [here](https://istio.io/latest/docs/examples/bookinfo/#determine-the-ingress-ip-and-port) to set the environment variables `GATEWAY_URL`. You can then check if you can access the application with the following `curl` command:

```bash
curl --header 'Host: bookinfo.example.com' -o /dev/null -s -w "%{http_code}\n" "http://${GATEWAY_URL}/productpage"
```

If everything is working, the command above should return `200`.
Note that the curl command above sets the `Host` header to match the host we associated the VirtualService with (`bookinfo.example.com`).

**Note**: If you want to access the application from your browser, you will need to set this header using a browser plugin.

## Generate load

To simulate user requests, use a command such as the following:

```bash
watch -n 0.1 'curl --header "Host: bookinfo.example.com" -s "http://${GATEWAY_URL}/productpage" | grep -i "color:"'
```

This command requests the `productpage` microservice 10 times per second.
In turn, this causes about the same frequency of requests against the backend microservices.
We filter the response to see the color being used to display the text "*William Shakespeare's*" in the *Summary* line of the page.
The color varies between versions giving us a visual way to distinguish between them.
Initially it should be *red*.

## Create an A/B/N `Experiment`

We will now define an A/B/N experiment to compare versions *v2* and *v3* of the *productpage* application to the existing version, *v1*.
These versions are visually identical except for the color of the text "*William Shakespeare's*" that appears in the page *Summary*.
In version *v2* they are *gold* and in version *v3* they are *green*.

In addition to the visual difference, the *v2* version has a high number of books sold (it is greatest of the three versions) but it has a long response time.
The *v3* version has an intermediate number of books sold (compared to the other versions) and a response time comparable to the *v1* version.

To describe a A/B/N experiment, create an iter8 `Experiment` that identifies the original, or *baseline* version and the new, or *candidate* versions and some evaluation criteria including the identification of a *reward* metric.
For example:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: productpage-v2v3v4-abn
spec:
  service:
    name: productpage-v1
    candidates:
      - productpage-v2
      - productpage-v3
    hosts:
      - name: bookinfo-kubecon.example.com
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
kubectl --namespace $NAMESPACE apply -f {{< resourceAbsUrl path="tutorials/abn-tutorial/abn_productpage_v1v2v3.yaml" >}}
```

Inspection of the new experiment shows that it is paused because the specified candidate versions cannot be found in the cluster:

```bash
kubectl --namespace $NAMESPACE get experiment

NAME                   TYPE    HOSTS                                PHASE   WINNER FOUND   CURRENT BEST   STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Pause                                 TargetsError: Missing Candidate
```

Once the candidate versions are deployed, the experiment will start automatically.

## Deploy the candidate versions of the _productpage_ service

To deploy the *v2* and *v3* versions of the *productpage* microservice, execute:

```bash
kubectl --namespace $NAMESPACE apply -f {{< resourceAbsUrl path="tutorials/productpage-v2.yaml" >}} -f {{< resourceAbsUrl path="tutorials/productpage-v3.yaml" >}}
```

Once its corresponding pods have started, the `Experiment` will show that it is progressing:

```bash
kubectl --namespace $NAMESPACE get experiment

NAME                   TYPE    HOSTS                                PHASE         WINNER FOUND   CURRENT BEST     STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Progressing   false          productpage-v3   IterationUpdate: Iteration 3/20 completed
```

At approximately 20 second intervals, you should see the interation number change.
Over time, you should see a decision that one of the versions (it could be the baseline version) is identified as the *winner*.
To make this determination, iter8 will evaluate the specified criteria and apply advanced analytics to make a determination.
Based on intermediate evaluations, traffic will be adjusted between the versions.
iter8 will eventually identify that the best version is the candidate, `productpage-v3` and that it is confident that this choice will be the final choice (by indicating that a *winner* has been found:

```bash
kubectl --namespace $NAMESPACE get experiment

NAME                   TYPE    HOSTS                                PHASE         WINNER FOUND   CURRENT BEST     STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Progressing   true           productpage-v3   IterationUpdate: Iteration 7/20 completed
```

If iter8 is unable to determine a winner with confidence, the experiment will fail and the best choice will default to the baseline version.

When the experiment is finished (about 5 minutes), you will see that all traffic has been shifted to the winner, *productpage-v3*:

```bash
kubectl --namespace $NAMESPACE get experiment

NAME                   TYPE    HOSTS                                PHASE       WINNER FOUND   CURRENT BEST     STATUS
productpage-abn-test   A/B/N   [productpage bookinfo.example.com]   Completed   true           productpage-v3   ExperimentCompleted: Traffic To Winner
```

## Cleanup

To clean up, delete the namespace:

```bash
kubectl delete namespace $NAMESPACE
```

## Other things to try (before cleanup)

### Inspect progress using Grafana

Coming soon

### Inspect progress using Kiali

Coming soon

### Alter the duration of the experiment

The progress of an experiment can be impacted by `duration` and `trafficControl` parameters:

- `duration.interval` defines how long each test interval should be (*default: 30 seconds*)
- `duration.maxIterations` identifies what the maximum number of iterations there should be (*default: 100*)
- `trafficControl.maxIncrement` identifies the largest change (increment) that will be made in the percentage of traffic sent to a candidate (*default: 2 percent*)

The impact of the first two parameters on the duration of the experiment are clear.
Restricting the size of traffic shifts limits how quickly an experiment can come to a decision about a candidate.
