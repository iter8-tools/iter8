# Automated canary releases with iter8 on Kubernetes and Istio

This tutorial shows you how _iter8_ can be used to perform canary releases by gradually shifting traffic to a canary version of a microservice.

This tutorial has 4 parts, which are supposed to be tried in order. **Here you will learn:**

- how to perform a canary rollout with _iter8_;
- how to set different success criteria for _iter8_ to analyze canary releases and determine success or failure; and
- how _iter8_ can be used for canary releases of both internal and edge services.

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) that is distributed with Istio. This application comprises 4 microservices, namely, _productpage_, _details_, _reviews_, and _ratings_, as illustrated [here](https://istio.io/docs/examples/bookinfo/). Please, follow our instructions below to deploy the sample application as part of the tutorial.

## YAML files used in the tutorial

All Kubernetes YAML files you will need in this tutorial are in the _iter8-controller_ repository [here](https://github.com/iter8-tools/iter8-controller/tree/master/doc/tutorials/istio/bookinfo). Thus, the first thing you need to do is to clone this repository.

```bash
git clone git@github.com:iter8-tools/iter8-controller.git
```

## Part 1: Successful canary release: _reviews-v2_ to _reviews-v3_

### 1. Deploy the Bookinfo application

At this point, we assume that you have already followed the [instructions](iter8_install.md) to install _iter8_ on your Kubernetes cluster. The next step is to deploy the sample application we will use for the tutorial.

First, let us create a `bookinfo-iter8` namespace configured to enable auto-injection of the Istio sidecar:

```bash
kubectl apply -f iter8-controller/doc/tutorials/istio/bookinfo/namespace.yaml
```

Next, let us deploy the Bookinfo application:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/bookinfo-tutorial.yaml
```

You should see the following pods in the `bookinfo-iter8` namespace. Make sure the pods' status is "Running." Also, note that there should be 2 containers in each pod, since the Istio sidecar was injected.

```bash
$ kubectl get pods -n bookinfo-iter8
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-68c7c8666d-m78qx       2/2     Running   0          64s
productpage-v1-7979869ff9-fln6g   2/2     Running   0          63s
ratings-v1-8558d4458d-rwthl       2/2     Running   0          64s
reviews-v2-df64b6df9-ffb42        2/2     Running   0          63s
```

We have deployed "version 2" of the _reviews_ microservice, and version 1 of all others.

Let us now expose the edge _productpage_ service by creating an Istio Gateway for it.

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/bookinfo-gateway.yaml
```

You should now see the Istio Gateway and VirtualService for _productpage_, as below:

```bash
$ kubectl get gateway -n bookinfo-iter8
NAME               AGE
bookinfo-gateway   22s
```

```bash
$ kubectl get vs -n bookinfo-iter8
NAME       GATEWAYS             HOSTS                   AGE
bookinfo   [bookinfo-gateway]   [bookinfo.sample.dev]   27s
```

As you can see above, we have associated Bookinfo's edge service with a fake host, namely, `bookinfo.sample.dev`.

### 2. Access the Bookinfo application

To access the application, you need to determine the ingress IP and port for the application in your environment. You can do so by following steps 3 and 4 of the Istio instructions [here](https://istio.io/docs/examples/bookinfo/#determine-the-ingress-ip-and-port) to set the environment variables `INGRESS_HOST`, `INGRESS_PORT`, and `GATEWAY_URL`, which will capture the correct IP address and port for your environment. Once you have done so, you can check if you can access the application with the following command:

```bash
curl -H "Host: bookinfo.sample.dev" -o /dev/null -s -w "%{http_code}\n" "http://${GATEWAY_URL}/productpage"
```

If everything is working, the command above should show `200`. Note that the curl command above sets the host header to match the host we associated the VirtualService with (`bookinfo.sample.dev`). If you want to access the application from your browser, you will need to set this header using a browser's plugin of your choice.

### 3. Generate load to the application

Let us now generate load to the application, emulating requests coming from users. To do so, we recommend you run the command below on a separate terminal:

```bash
watch -n 0.1 'curl -H "Host: bookinfo.sample.dev" -Is "$GATEWAY_URL/productpage"'
```

This command will send 10 requests per second to the application. Note that the environment variable `GATEWAY_URL` must have been set as per step 2 above. Among other things, the command output should show an HTTP code of 200, as below:

```
HTTP/1.1 200 OK
content-type: text/html; charset=utf-8
content-length: 5719
server: istio-envoy
(...)
```

### 4. Configure a canary rollout for the _reviews_ service

At this point, Bookinfo is using version 2 of the _reviews_ service (_reviews-v2_). Let us now use _iter8_ to automate the canary rollout of version 3 of this service (_reviews-v3_).

First, we need to tell _iter8_ that we are about to perform this canary rollout. To that end, we create an `Experiment` configuration specifying the rollout details. In this tutorial, let us use the following `Experiment` configuration:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
metadata:
  name: reviews-v3-rollout
spec:
  targetService:
    name: reviews
    apiVersion: v1
    baseline: reviews-v2
    candidate: reviews-v3
  trafficControl:
    strategy: check_and_increment
    interval: 30s
    trafficStepSize: 20
    maxIterations: 8
    maxTrafficPercentage: 80
  analysis:
    analyticsService: "http://iter8-analytics.iter8"
    successCriteria:
      - metricName: iter8_latency
        toleranceType: threshold
        tolerance: 0.2
        sampleSize: 5
```

The configuration above specifies the baseline and candidate versions in terms of Kubernetes deployment names. The rollout is configured to last for 8 iterations (`maxIterations`) of `30s` (`interval`). At the end of each iteration, if the candidate version meets the specified success criteria, the traffic sent to it will increase by 20 percentage points (`trafficStepSize`) up to 80% (`maxTrafficPercentage`). At the end of the last iteration, if the success criteria are met, the candidate version will take over from the baseline.

In the example above, we specified only one success criterion. In particular, we stated that the mean latency exhibited by the candidate version should not exceed the threshold of 0.2 seconds. At the end of each iteration, _iter8-controller_ calls _iter8-analytics_, which in turn analyzes the metrics of interest (in this case, only mean latency) against the corresponding criteria. The number of data points analyzed during an experiment is cumulative, that is, it carries over from iteration to iteration.

The next step of this tutorial is to actually create the configuration above. To that end, you can either copy and paste the yaml above to a file and then run `kubectl apply -n bookinfo-iter8 -f` on it, or you can run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/canary_reviews-v2_to_reviews-v3.yaml
```

You can verify that the `Experiment` object has been created as shown below:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE   STATUS                       BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Pause   MissingCandidateDeployment   reviews-v2   100          reviews-v3   0
```

As you can see, _iter8_ is reporting that 100% of the traffic is sent to the baseline version (_reviews-v2_) and that the candidate (_reviews-v3_) is missing. As soon as the controller sees the candidate version, it will start the rollout. Next, let us deploy the candidate version to trigger the canary rollout.

### 5. Deploy the canary version and start the rollout

As soon as we deploy _reviews-v3_, _iter8-controller_ will start the rollout. To deploy _reviews-v3_, you can run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/reviews-v3.yaml
```

Now, if you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that 20% of the traffic is now being sent to _reviews-v3_:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE         STATUS                  BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Progressing   Iteration 1 Completed   reviews-v2   80           reviews-v3   20
```

At about every 30s you should see the traffic shift towards _reviews-v3_ by 20 percentage points.

### 6. Check the Grafana dashboard

You can also check a Grafana dashboard specific to the `Experiment` object corresponding to the rollout you are running. The URL to the Grafana dashboard for the experiment is the value of the field `grafanaURL` under the object's `status`. One way to get the Grafana URL that you can paste to your browser is through the following command:

```bash
kubectl get experiment reviews-v3-rollout -o jsonpath='{.status.grafanaURL}' -n bookinfo-iter8
```

By default, the base URL given by iter8 to Grafana is `http://localhost:3000`. In a typical Istio installation, you can port-forward your Grafana from Kubernetes to your localhost's port 3000 with the following command:

```
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

Below is a screenshot of a portion of the Grafana dashboard showing the request rate and the mean latency for reviews-v2 and reviews-v3, right after the controller ended the experiment.

![Grafana Dashboard](../img/grafana_reviews-v2-v3.png)

Note how the traffic shifted towards the canary during the experiment. You can also see that the canary's mean latency was way below the configured threshold of 0.2 seconds.

## Part 2: High-latency canary release: _reviews-v3_ to _reviews-v4_

At this point, you must have completed the part 1 of the tutorial successfully. You can confirm it as follows:

```bash
$ kubectl get experiment reviews-v3-rollout -n bookinfo-iter8
NAME                 PHASE       STATUS                                           BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Succeeded   AllSuccessCriteriaMet, Traffic: AllToCandidate   reviews-v2   0            reviews-v3   100
```

The command above's output shows that _reviews-v3_ took over from _reviews-v2_ as part of the canary rollout performed before.

### 1. Canary rollout configuration

Now, let us set up a canary rollout for _reviews-v4_, using the following `Experiment` configuration:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
metadata:
  name: reviews-v4-rollout
spec:
  targetService:
    name: reviews
    apiVersion: v1
    baseline: reviews-v3
    candidate: reviews-v4
  trafficControl:
    strategy: check_and_increment
    interval: 30s
    trafficStepSize: 20
    maxIterations: 6
    maxTrafficPercentage: 80
  analysis:
    analyticsService: "http://iter8-analytics.iter8"
    successCriteria:
      - metricName: iter8_latency
        toleranceType: threshold
        tolerance: 0.2
        sampleSize: 5
```

The configuration above is pretty much the same we used in part 1, except that now the baseline version is _reviews-v3_ and the candidate is _reviews-v4_.

To create the above `Experiment` object, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/canary_reviews-v3_to_reviews-v4.yaml
```

You can list all `Experiment` objects like so:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE       STATUS                                           BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Succeeded   AllSuccessCriteriaMet, Traffic: AllToCandidate   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Pause       MissingCandidateDeployment                       reviews-v3   100          reviews-v4   0
```

The output above shows the new object you just created, for which the candidate deployment _reviews-v4_ is missing. Let us deploy _reviews-v4_ next so that the rollout can begin.

### 2. Deploy _reviews-v4_ and start the rollout

As you have already seen, as soon as we deploy the candidate version, _iter8-controller_ will start the rollout. This time, however, the candidate version (_reviews-v4_) has a performance issue preventing it from satisfying the success criteria in the experiment object. As a result, _iter8_ will roll back to the baseline version.

To deploy _reviews-v4_, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/reviews-v4.yaml
```

Now, if you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that 20% of the traffic is now being sent to _reviews-v4_.

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE         STATUS                                           BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Succeeded     AllSuccessCriteriaMet, Traffic: AllToCandidate   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Progressing   Iteration 1 Completed                            reviews-v3   80           reviews-v4   20
```

However, unlike the previous rollout, traffic will not shift towards the candidate _reviews-v4_ because it does not meet the success criteria due to a performance problem. At the end of the experiment, _iter8_ rolls back to the baseline (_reviews-v3_), as seen below:

```bash
$ kubectl get experiment reviews-v4-rollout -n bookinfo-iter8
NAME                 PHASE    STATUS                                             BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v4-rollout   Failed   NotAllSuccessCriteriaMet, Traffic: AllToBaseline   reviews-v3   100          reviews-v4   0
```

### 3. Check the Grafana dashboard

As before, you can check the Grafana dashboard corresponding to the canary release of _reviews-v4_. To get the URL to the dashboard specific to this canary release, run the following command:

```bash
kubectl get experiment reviews-v4-rollout -o jsonpath='{.status.grafanaURL}' -n bookinfo-iter8
```

![Grafana Dashboard](../img/grafana_reviews-v3-v4.png)

The dashboard screenshot above shows that the canary version (_reviews-v4_) consistently exhibits a high latency of 5 seconds, way above the threshold of 0.2 seconds specified in our success criterion, and way above the baseline version's latency.


## Part 3: Error-producing canary release: _reviews-v3_ to _reviews-v5_

At this point, you must have completed parts 1 and 2 of the tutorial successfully. You can confirm it as follows:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE       STATUS                                             BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Succeeded   AllSuccessCriteriaMet, Traffic: AllToCandidate     reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Failed      NotAllSuccessCriteriaMet, Traffic: AllToBaseline   reviews-v3   100          reviews-v4   0
```

The command above's output shows that _reviews-v3_ took over from _reviews-v2_ as part of the canary rollout performed before on part 1, and that it continues to be the current version after iter8 had determined that _reviews-v4_ was unsatisfactory.

### 1. Canary rollout configuration

Now, let us set up a canary rollout for _reviews-v5_, using the following `Experiment` configuration:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
metadata:
  name: reviews-v5-rollout
spec:
  targetService:
    name: reviews
    apiVersion: v1
    baseline: reviews-v3
    candidate: reviews-v5
  trafficControl:
    strategy: check_and_increment
    interval: 30s
    trafficStepSize: 20
    maxIterations: 6
    maxTrafficPercentage: 80
  analysis:
    analyticsService: "http://iter8-analytics.iter8"
    successCriteria:
      - metricName: iter8_latency
        toleranceType: threshold
        tolerance: 0.2
        sampleSize: 5
      - metricName: iter8_error_rate
        toleranceType: delta
        tolerance: 0.02
        sampleSize: 10
        stopOnFailure: true
```

The configuration above differs from the previous ones as follows. We added a second success criterion on the error-rate metric so that the canary version (_reviews-v5_) not only must have a mean latency below 0.2 seconds, but it also needs to have an error rate that cannot exceed the baseline error rate by more than 2%. That comparative analysis on a metric is specified as a `delta` tolerance type. Furthermore, the second success criterion sets the flag `stopOnFailure`, which means iter8 will roll back to the baseline as soon as the error rate criterion is violated and the minimum number of 10 data points is collected (`sampleSize = 10`).

To create the above `Experiment` object, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/canary_reviews-v3_to_reviews-v5.yaml
```

### 2. Deploy _reviews-v5_ and start the rollout

As you already know, as soon as we deploy the candidate version, _iter8-controller_ will start the rollout. This time, the candidate version (_reviews-v5_) has a bug that causes it to return HTTP errors to its callers. As a result, _iter8_ will roll back to the baseline version based on the success criterion on the error-rate metric defined above.

To deploy _reviews-v5_, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/reviews-v5.yaml
```

If you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that 20% of the traffic is now being sent to _reviews-v5_.

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE         STATUS                                             BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Succeeded     AllSuccessCriteriaMet, Traffic: AllToCandidate     reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Failed        NotAllSuccessCriteriaMet, Traffic: AllToBaseline   reviews-v3   100          reviews-v4   0
reviews-v5-rollout   Progressing   Iteration 1 Completed                              reviews-v3   80           reviews-v5   20
```

Because _review-v5_ has an issue causing it to return HTTP errors, as per the success criteria we have specified the traffic will not shift towards it. Furthermore, because the error-rate success criteria indicated the need to stop on failure, without waiting for the entire duration of the experiment, iter8 will rollback to _reviews-v3_ quickly. You should see the following after several seconds:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE       STATUS                                             BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Succeeded   AllSuccessCriteriaMet, Traffic: AllToCandidate     reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Failed      NotAllSuccessCriteriaMet, Traffic: AllToBaseline   reviews-v3   100          reviews-v4   0
reviews-v5-rollout   Failed      Aborted, Traffic: AllToBaseline.                   reviews-v3   100          reviews-v5   0
```

### 3. Check the Grafana dashboard

As before, you can check the Grafana dashboard corresponding to the canary release of _reviews-v5_. To get the URL to the dashboard specific to this canary release, run the following command:

```bash
kubectl get experiment reviews-v5-rollout -o jsonpath='{.status.grafanaURL}' -n bookinfo-iter8
```

![Grafana Dashboard](../img/grafana_reviews-v3-v5-req-rate.png)
![Grafana Dashboard](../img/grafana_reviews-v3-v5-error-rate.png)

The dashboard screenshots above show that traffic to the canary version (_reviews-v5_) is quickly interrupted. Also, while the _reviews-v5_ latency is way below the threshold of 0.2 seconds we defined in the latency success criterion, its error rate is 100%, i.e., it generates errors for every single request it processes. That does not meet the error-rate success criterion we defined, which specified that the canary's error rate must be within 2% of that of the baseline (_reviews-v3_) version. According to the dashboard, _reviews-v3_ produced no errors at all.

## Part 4: Canary release for an edge service

Up to now, we have demonstrated rolling out a new version of an internal service. In this part of the tutorial we will show you how to use _iter8_ to perform a canary analysis for an edge service. By edge service we mean one that is exposed to users and services outside the Kubernetes cluster where it runs. In the case of the Bookinfo sample application we use in the tutorial, the _productpage_ service is the edge service.

### Edge service exposed using Kubernetes Ingress

If you expose your edge service using [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/), you do not need anything special. The `Experiment` object you will need to create will be similar to the ones you saw in the previous parts of this tutorial.

### Edge service exposed using Istio's VirtualService and Gateway

An edge service can also be exposed using Istio's VirtualService and Gateway. To remind you, after we deployed Bookinfo [in Part 1 of the tutorial](#part-1-successful-canary-release-reviews-v2-to-reviews-v3), we exposed the _productpage_ service by creating an Istio Gateway and Virtual Service. The VirtualService defines the mapping from an external hostname to an internal service, and binds that to a specific gateway.

We defined _productpage_'s VirtualService and Gateway before using the file `iter8-controller/doc/tutorials/istio/bookinfo/bookinfo-gateway.yaml`, which looks like this:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: bookinfo-gateway
spec:
  selector:
    istio: ingressgateway # use istio default controller
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "bookinfo.sample.dev"
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: bookinfo
spec:
  hosts:
  - "bookinfo.sample.dev"
  gateways:
  - bookinfo-gateway
  http:
  - match:
    - uri:
        exact: /productpage
    - uri:
        exact: /login
    - uri:
        exact: /logout
    - uri:
        prefix: /api/v1/products
    route:
    - destination:
        host: productpage
        port:
          number: 9080                                   
```

Above, the VirtualService for _productpage_ is named `bookinfo`.

### 1. Configure a canary rollout for the _productpage_ service

To perform a canary rollout of a service that has been exposed using Istio's VirtualService and Gateway, _iter8_ needs to be pointed to the existing VirtualService object. For rolling out version 2 of the _productpage_ service, we will create an `Experiment` object with the specification below:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
metadata:
  name: productpage-v2-rollout
spec:
  routingReference:
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    name: bookinfo
  targetService:
    name: productpage
    apiVersion: v1
    baseline: productpage-v1
    candidate: productpage-v2
  trafficControl:
    strategy: check_and_increment
    interval: 30s
    trafficStepSize: 20
    maxIterations: 6
    maxTrafficPercentage: 80
  analysis:
    analyticsService: "http://iter8-analytics.iter8"
    successCriteria:
      - metricName: iter8_latency
        toleranceType: threshold
        tolerance: 3.0
        sampleSize: 5
```

If you look carefully at the definition above, you will notice a reference to the Istio VirtualService named _bookinfo_. This reference will instruct _iter8_ to manipulate that existing VirtualService for the purposes of traffic management.

Let us now create the `Experiment` object above by running the following command:

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/canary_productpage-v1_to_productpage-v2.yaml
```

You can verify that the `Experiment` object has been created:

```bash
$ kubectl get experiment productpage-v2-rollout -n bookinfo-iter8
NAME                     PHASE   STATUS                       BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Pause   MissingCandidateDeployment   productpage-v1   100          productpage-v2   0
```

### 2. Deploy _productpage-v2_ and start the rollout

To start the rollout let us deploy the candidate version (_productpage-v2_).

```bash
kubectl apply -n bookinfo-iter8 -f iter8-controller/doc/tutorials/istio/bookinfo/productpage-v2.yaml
```

You can verify that experiment has started:

```bash
$ kubectl get experiment productpage-v2-rollout -n bookinfo-iter8
NAME                     PHASE         STATUS                  BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Progressing   Iteration 1 Completed   productpage-v1   80           productpage-v2   20
```

Now, if you inspect the `bookinfo` VirtualService, you should see a change in the `route` section reflecting the current traffic split.

```bash
kubectl get vs bookinfo -n bookinfo-iter8 -o yaml
```

If you look at the spec of that VirtualService, you will see something like this:

```yaml
  spec:
    gateways:
    - bookinfo-gateway
    hosts:
    - bookinfo.sample.dev
    http:
    - match:
      - uri:
          exact: /productpage
      - uri:
          exact: /login
      - uri:
          exact: /logout
      - uri:
          prefix: /api/v1/products
      route:
      - destination:
          host: productpage
          port:
            number: 9080
          subset: baseline
        weight: 80
      - destination:
          host: productpage
          port:
            number: 9080
          subset: candidate
        weight: 20
```

As the canary rollout progresses, you should see the traffic shifting to the candidate version (_productpage-v2_).

Of course, you can check the Grafana dashboard as before.
