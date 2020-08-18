---
menuTitle: Canary Testing - OpenShift
title: Canary Testing
weight: 21
summary: Learn how to perform a canary release on Red Hat OpenShift
---

This tutorial shows how iter8 can be used to perform a canary release by gradually shifting traffic from one version of a microservice to another while evaluating the behavior of the new version.
Traffic is fully shifted only if the behavior of the candidate version meets specified acceptance criteria.

{{% notice info %}}
This tutorial is for use with Red Hat OpenShift. The corresponding tutorial for plain Kubernetes is [here]({{< ref "canary" >}}).
{{% /notice %}}

This tutorial has six steps, which are meant to be tried in order.
You will learn:

- how to perform a canary rollout with iter8; and
- how to define different success criteria for iter8 to analyze canary releases and determine success or failure;

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) distributed with [Istio](https://istio.io).
This application comprises 4 microservies: _productpage_, _details_, _reviews_, and _ratings_.
Of these, _productpage_ is a user-facing service while the others are backend services.

{{% notice info %}}
This rest of this tutorial assumes you have already installed _iter8_ (including Red Hat OpenShift Service Mesh). If not, do so using the instructions [here]({{< ref "red-hat" >}}).
{{% /notice %}}

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

Note that we deployed version *v2* of the *reviews* microsevice; that is, *reviews-v2*.
Each pod should have two containers, since the Istio sidecar was injected into each.

## Expose the Bookinfo application

Expose the Bookinfo application by defining an Istio `Gateway` and `VirtualService`. These will use the `route` defined for the istio ingress gateway:

```bash
export GATEWAY_URL=$(oc -n istio-system get route istio-ingressgateway -o jsonpath='{.spec.host}')
```

```bash
cat {{< resourceAbsUrl path="tutorials/bookinfo-gateway.yaml" >}} \
| sed "s#bookinfo.example.com#${GATEWAY_URL}#" \
| oc --namespace bookinfo-iter8 apply -f -
```

You can inspect the created resources:

```bash
oc --namespace bookinfo-iter8 get gateway,virtualservice
```

## Verify access to Bookinfo

You can then check if you can access the application with the following `curl` command:

```bash
curl -o /dev/null -s -w "%{http_code}\n" "http://${GATEWAY_URL}/productpage"
```

If everything is working, the command above should return `200`.

## Generate load

To simulate user requests, use a command such as the following:

```bash
watch -n 0.1 'curl -s "http://${GATEWAY_URL}/productpage" | grep -i "color=\""'
```

This command requests the `productpage` microservice 10 times per second.
In turn, this causes about the same frequency of requests against the backend microservice.
We filter the response to see the color being used to display the "star" rating of the application.
The color varies between versions giving us a visual way to distinguish between them.

## Create a canary `Experiment`

We will now define a canary experiment to rollout version *v3* of the *reviews* application.
These versions are visually identical except for the color of the stars that appear on the page.
In version *v3* they are *red*.
This can be seen in the inspected in the output of the above `watch` command.
As version *v3* is rolled out, you should see the color change.

To describe a canary rollout, create an iter8 `Experiment` that identifies the original, or *baseline* version and the new, or *candidate* version and some evaluation criteria.
For example:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: reviews-v3-rollout
spec:
  service:
    name: reviews
    baseline: reviews-v2
    candidates:
      - reviews-v3
  criteria:
    - metric: iter8_mean_latency
      threshold:
        type: absolute
        value: 200
  duration:
    maxIterations: 8
    interval: 15s
  trafficControl:
    maxIncrement: 20
```

In this example, the target of the experiment is the service `reviews`.
The baseline and candidate versions are specified using their `Deployment` names, `reviews-v2` and `reviews-v3`, respectively.
A single evaluation criteria is specified.
It requires that the measurements of the metric `iter8_mean_latency` should all return values less than `200` milliseconds.
The additional parameters control how long the experiment should run and how much traffic can be shifted to the new version in each interval. Details regarding these parameters are [here](#alter-the-duration-of-the-experiment).

The experiment can be created using the command:

```bash
oc --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/canary-tutorial/canary_reviews-v2_to_reviews-v3.yaml">}}
```

Inspection of the new experiment shows that it is paused because the specified candidate version cannot be found in the cluster:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                 TYPE     HOSTS       PHASE   WINNER FOUND   CURRENT BEST   STATUS
reviews-v3-rollout   Canary   [reviews]   Pause                                 TargetsError: Missing Candidate
```

Once the candidate version is deployed, the experiment will start automatically.

## Deploy the candidate version of the _reviews_ service

To deploy version *v3* of the *reviews* microservice, execute:

```bash
oc --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/reviews-v3.yaml" >}}
```

Once its corresponding pods have started, the `Experiment` will show that it is progressing:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                 TYPE     HOSTS       PHASE         WINNER FOUND   CURRENT BEST   STATUS
reviews-v3-rollout   Canary   [reviews]   Progressing   false          reviews-v3     IterationUpdate: Iteration 0/8 completed
```

At approximately 15 second intervals, you should see the interation number change. Traffic will gradually be shifted (in 20% increments) from version *v2* to version *v3*.
iter8 will quickly identify that the best version is the candidate, `reviews-v3` and that it is confident that this choice will be the final choice (by indicating that a *winner* has been found):

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                 TYPE     HOSTS       PHASE         WINNER FOUND   CURRENT BEST   STATUS
reviews-v3-rollout   Canary   [reviews]   Progressing   true           reviews-v3     IterationUpdate: Iteration 3/8 completed
```

When the experiment is finished (about 2 minutes), you will see that all traffic has been shifted to the winner, *reviews-v3*:

```bash
oc --namespace bookinfo-iter8 get experiment
```

```bash
NAME                 TYPE     HOSTS       PHASE       WINNER FOUND   CURRENT BEST   STATUS
reviews-v3-rollout   Canary   [reviews]   Completed   true           reviews-v3     ExperimentCompleted: Traffic To Winner
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

- `duration.interval` defines how long each test interval should be (*default: 30 seconds*)
- `duration.maxIterations` identifies what the maximum number of iterations there should be (*default: 100*)
- `trafficControl.maxIncrement` identifies the largest change (increment) that will be made in the percentage of traffic sent to a candidate (*default: 2 percent*)

The impact of the first two parameters on the duration of the experiment are clear.
Restricting the size of traffic shifts limits how quickly an experiment can come to a decision about a candidate.

### Try a version that fails the criteria

Version *v4* of the *reviews* service is a modification that returns after a 5 second delay.
If you try this version as a candidate, you should see the canary experiment reject it and choose the baseline version as the winner.

For your reference:

- A YAML for the deployment `reviews-v4` is: [{{< resourceAbsUrl path="tutorials/reviews-v4.yaml" >}}]({{< resourceAbsUrl path="tutorials/reviews-v4.yaml" >}})
- A YAML for an canary experiment from _reviews-v3_ to _reviews-v4_ is: [{{< resourceAbsUrl path="tutorials/canary-tutorial/canary_reviews-v3_to_reviews-v4.yaml" >}}]({{< resourceAbsUrl path="tutorials/canary-tutorial/canary_reviews-v3_to_reviews-v4.yaml" >}})

### Try a version which returns errors

Coming soon

### Try with a user-facing service

Coming soon
