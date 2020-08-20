---
menuTitle: Performance Validation
title: Performance Validation
weight: 30
summary: Learn how to validate performance of a service
---

This tutorial shows how iter8 can be used to validate the performance of a service; that is, that it satisfies some performance criteria.

This tutorial has six steps, which are meant to be tried in order.
You will learn:

- how to perform a canary rollout with iter8; and
- how to define different success criteria for iter8 to analyze canary releases and determine success or failure.

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) distributed with [Istio](https://istio.io).
This application comprises 4 microservies: _productpage_, _details_, _reviews_, and _ratings_.
Of these, _productpage_ is a user-facing service while the others are backend services.

{{% notice info %}}
This rest of this tutorial assumes you have already installed iter8 (including Istio). If not, do so using the instructions [here]({{< ref "kubernetes" >}}).
{{% /notice %}}

## Deploy the Bookinfo application

To deploy the Bookinfo application, create a namespace configured to enable auto-injection of the Istio sidecar. You can use whatever namespace name you wish. By default, the namespace `bookinfo-iter8` is created.

```bash
kubectl apply -f {{< resourceAbsUrl path="tutorials/namespace.yaml" >}}
```

Next, deploy the application:

```bash
kubectl --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/bookinfo-tutorial.yaml" >}}
```

You should see pods for each of the four microservices:

```bash
kubectl --namespace bookinfo-iter8 get pods
```

Note that we deployed version *v2* of the *reviews* microsevice; that is, *reviews-v2*.
Each pod should have two containers, since the Istio sidecar was injected into each.

## Expose the Bookinfo application

Expose the Bookinfo application by defining an Istio `Gateway` and `VirtualService`:

```bash
kubectl --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/bookinfo-gateway.yaml" >}}
```

You can inspect the created resources:

```bash
kubectl --namespace bookinfo-iter8 get gateway,virtualservice
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

{{% notice tip %}}
If you want to access the application from your browser, you will need to set this header using a browser plugin.
{{% /notice %}}

## Generate load

To simulate user requests, use a command such as the following:

```bash
watch -n 0.1 'curl --header "Host: bookinfo.example.com" -s "http://${GATEWAY_URL}/productpage" | grep -i "color=\""'
```

This command requests the `productpage` microservice 10 times per second.
In turn, this causes about the same frequency of requests against the backend microservice.
We filter the response to see the color being used to display the "star" rating of the application.
The color varies between versions giving us a visual way to distinguish between them.

## Create a performance `Experiment`

We will now define a performance experiment to validate that the deployed version *v2* of the *reviews* application meets some performance criteria.

To validate service performance, create an iter8 `Experiment` where  the service is both the `baseline` and the single `candidate`:
For example:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: performance-reviews-v2
spec:
  service:
    name: reviews
    baseline: reviews-v2
    candidates:
      - reviews-v2
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

In this example, we will test the performance of the service `reviews`.
A single evaluation criteria is specified.
It requires that the measurements of the metric `iter8_mean_latency` should all return values less than `200` milliseconds.
The additional parameters control how long the experiment should run.
Details regarding these parameters are [here](#alter-the-duration-of-the-experiment).

The experiment can be created using the command:

```bash
kubectl --namespace bookinfo-iter8 apply -f {{< resourceAbsUrl path="tutorials/performance-tutorial/performance-validation_reviews-v2.yaml">}}
```

{{% notice warning %}}
Iter8 will configure `VirtualService` to send all of the traffic to the version under test. It will not be restored to the original version (which is not specified). Use with caution.
{{% /notice %}}

Since the version to be tested is already running, the experiment should begin immediately. Inspection of the `Experiment`  will show that it is progressing:

```bash
kubectl --namespace bookinfo-iter8 get experiment
```

```bash
NAME                     TYPE     HOSTS       PHASE       WINNER FOUND   CURRENT BEST   STATUS
performance-reviews-v2   Canary   [reviews]   Progressing   false                         IterationUpdate: Iteration 0/8 completed
```

At approximately 15 second intervals, you should see the interation number change.
Initially, the experiment will indicate that no *winner* has been found.
This means that iter8 is still uncertain that the service satisfies the specified criteria.
When it gains this confidence, it will indicate that it has found a *winner*.

```bash
kubectl --namespace bookinfo-iter8 get experiment
```

```bash
NAME                     TYPE     HOSTS       PHASE         WINNER FOUND   CURRENT BEST   STATUS
performance-reviews-v2   Canary   [reviews]   Progressing   true           reviews-v2     IterationUpdate: Iteration 3/8 completed
```

When the experiment is finished (about 2 minutes), you will see that the winner is *reviews-v2*:

```bash
kubectl --namespace bookinfo-iter8 get experiment
```

```bash
NAME                     TYPE     HOSTS       PHASE       WINNER FOUND   CURRENT BEST   STATUS
performance-reviews-v2   Canary   [reviews]   Completed   true           reviews-v2     ExperimentCompleted: Traffic To Winner
```

If no winner was found, we would conclude that the version of the service we tested does not satisiy the performance criteria.

## Cleanup

To clean up, delete the namespace:

```bash
kubectl delete namespace bookinfo-iter8
```

## Other things to try (before cleanup)

### Try a version that fails the criteria

Version *v4* of the *reviews* service is a modification that returns after a 5 second delay.
If you try this version as a candidate, you should see the performance experiment reject it and choose the baseline version as the winner.

For your reference:

- A YAML for the deployment `reviews-v4` is: [{{< resourceAbsUrl path="tutorials/reviews-v4.yaml" >}}]({{< resourceAbsUrl path="tutorials/reviews-v4.yaml" >}})
- A YAML for performance experiment is: [{{< resourceAbsUrl path="tutorials/performance-tutorial/performance-validation_reviews-v4.yaml" >}}]({{< resourceAbsUrl path="tutorials/performance-tutorial/performance-validation_reviews-v4.yaml" >}})
