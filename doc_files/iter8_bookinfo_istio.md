# Automated canary releases with iter8 on Kubernetes and Istio

This tutorial shows you how _iter8_ can be used to perform canary releases by gradually shifting traffic to a canary version of a microservice.

This tutorial has 5 parts, which are designed to be tried in order. **Here you will learn:**

- how to perform a canary rollout with _iter8_;
- how to set different success criteria for _iter8_ to analyze canary releases and determine success or failure;
- how to have _iter8_ immediately stop an experiment as soon as a criterion is not met;
- how to use your own custom metrics in success criteria for canary analyses; and
- how _iter8_ can be used for canary releases of both internal and user-facing services.

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/)  distributed with Istio. This application comprises 4 microservices: _productpage_, _details_, _reviews_, and _ratings_, as illustrated [here](https://istio.io/docs/examples/bookinfo/). Please, follow our instructions below to deploy the sample application as part of the tutorial.

## YAML files used in the tutorial

All of the Kubernetes YAML files you will need in this tutorial are in the [_iter8-controller_ repository](https://github.com/iter8-tools/iter8-controller/tree/master/doc/tutorials/istio/bookinfo).

## Part 1: Successful canary release: _reviews-v2_ to _reviews-v3_

### 1. Deploy the Bookinfo application

At this point, we assume that you have already followed the [instructions](iter8_install.md) to install _iter8_ on your Kubernetes cluster. This step is to deploy the sample application we will use for the tutorial.

First, create a `bookinfo-iter8` namespace configured to enable auto-injection of the Istio sidecar:

```bash
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo/namespace.yaml
```

Note: If your cluster is a Red Hat OpenShift cluster with the Red Hat OpenShift Service Mesh, add the namespace to the `ServiceMeshMemberRoll` defined in the Service Mesh namespace (default is `istio-system`).

Next, deploy the Bookinfo application:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo/bookinfo-tutorial.yaml
```

You should see the following four pods in the `bookinfo-iter8` namespace. Make sure the pods' status is **Running**. Also, note that there should be 2 containers in each pod (**2/2**), since the Istio sidecar was injected.

```bash
$ kubectl get pods -n bookinfo-iter8
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-68c7c8666d-m78qx       2/2     Running   0          64s
productpage-v1-7979869ff9-fln6g   2/2     Running   0          63s
ratings-v1-8558d4458d-rwthl       2/2     Running   0          64s
reviews-v2-df64b6df9-ffb42        2/2     Running   0          63s
```

We have deployed "version 2" of the _reviews_ microservice, and version 1 of the others.

Let us now expose the edge service, _productpage_, by creating an Istio Gateway for it:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo/bookinfo-gateway.yaml
```

You should now see the Istio `Gateway` and `VirtualService` for _productpage_, as below:

```bash
$ kubectl get gateway -n bookinfo-iter8
NAME               AGE
bookinfo-gateway   22s
```

```bash
$ kubectl get vs -n bookinfo-iter8
NAME       GATEWAYS             HOSTS                   AGE
bookinfo   [bookinfo-gateway]   [bookinfo.example.com]   27s
```

As you can see above, we have, for demonstration purposes, associated Bookinfo's edge service with a fake host, namely, `bookinfo.example.com`.

**Issue:** Red Hat Openshift

### 2. Access the Bookinfo application

To access the application, you need to determine the ingress IP and port for the application in your environment. You can do so by following steps 3 and 4 of the Istio instructions [here](https://istio.io/docs/examples/bookinfo/#determine-the-ingress-ip-and-port) to set the environment variables `INGRESS_HOST`, `INGRESS_PORT`, and `GATEWAY_URL`, which will capture the correct IP address and port for your environment.

Note: If using Red Hat OpenShift Service Mesh, set `GATEWAY_URL` to the name of the route associated with the istio-ingressgateway:

**Red Hat**: get the command

Once you have defined `GATEWAY_URL`, you can check if you can access to the application with the following command:

```bash
curl -H "Host: bookinfo.example.com" -o /dev/null -s -w "%{http_code}\n" "http://${GATEWAY_URL}/productpage"
```

If everything is working, the command above should show `200`. Note that the curl command above sets the host header to match the host we associated the VirtualService with (`bookinfo.example.com`). If you want to access the application from your browser, you will need to set this header using a browser's plugin of your choice.

### 3. Generate load to the application

Let us now generate load to the application, emulating requests coming from users. To do so, we recommend you run the command below on a separate terminal:

```bash
watch -n 0.1 'curl -H "Host: bookinfo.example.com" -Is "http://${GATEWAY_URL}/productpage"'
```

This command will send up to 10 requests per second to the application (it will be fewer if the response time is large). Among other things, the command output should show an HTTP code of 200, as below:

```bash
HTTP/1.1 200 OK
content-type: text/html; charset=utf-8
content-length: 5719
server: istio-envoy
(...)
```

### 4. Configure a canary rollout for the _reviews_ service

At this point, Bookinfo is using version 2 of the _reviews_ service (_reviews-v2_). Let us now use _iter8_ to automate a canary rollout of version 3 of this service (_reviews-v3_).

First, we need to tell _iter8_ that we are about to perform this canary rollout. To that end, we create an `Experiment` resource specifying the rollout details. In this tutorial, let us use the [definition](https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/canary_reviews-v2_to_reviews-v3.yaml):

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: reviews-v3-rollout
spec:
  service:
    name: reviews
    apiVersion: v1
    baseline: reviews-v2
    candidates: [ "reviews-v3" ]
  criteria:
    - metric: iter8_mean_latenecy
      threshold:
        type: absolute
        value: 200
  duration:
    interval: 30s
    maxIterations: 8
  trafficControl:
    maxIncrement: 20
  cleanup: true
  analyticsSerivceURL: http://iter8-analytics:8080
```

The configuration above specifies the baseline and candidate versions in terms of Kubernetes deployment names. The rollout is configured to last for 8 iterations (`duration.maxIterations`) each of `30s` (`duration.interval`). At the end of each iteration, if the candidate version meets the specified success `criteria`, the traffic sent to it will increase by at most 20 percentage points (`trafficControl.maxIncrement`). At the end of the last iteration, if the success criteria are met, the candidate version will take over from the baseline. At the end of the experiment, the unused deployment will be deleted (`cleanup` is `true`).

**Question**: do we have a max amount?

In this exanple, we specified a single success critera. In particular, we stated that the mean latency exhibited by the candidate version should not exceed the threshold of 0.2 seconds. At the end of each iteration, _iter8_ analyzes the metrics relevant to the success criteria (in this case, only mean latency) against the corresponding criteria. The number of data points analyzed during an experiment is cumulative, that is, it carries over from iteration to iteration.

The next step of this tutorial is to actually create the configuration above. To that end, you can either copy and paste the yaml above to a file and then run `kubectl apply -n bookinfo-iter8 -f` on it, or you can run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/canary_reviews-v2_to_reviews-v3.yaml
```

You can verify that the `Experiment` resource has been created as shown below:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE   STATUS                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Pause   TargetsNotFound: Missing Candidate   reviews-v2   100          reviews-v3   0
```

Because the candidate version has not been deployed, _iter8_ ensures all traffic is routed to the baseline version and pauses the experiment. As soon as _iter8_ detects the candidate version, it will start the rollout. Next, let us deploy the candidate version to trigger the canary rollout.

### 5. Deploy the canary version and start the rollout

As soon as we deploy _reviews-v3_, _iter8-controller_ will start the rollout. You can deploy _reviews-v3_, with the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/reviews-v3.yaml
```

If you check the state of the `Experiment` resource we created earlier, you should see that the rollout is in progress, and that some of the traffic is being sent to _reviews-v3_:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE         STATUS                                 BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Progressing   IterationUpdate: Iteration 1 Started   reviews-v2   80           reviews-v3   20
```

At about every 30s you should see additional traffic shift towards _reviews-v3_.

### 6. Inspect the metrics via Granfana dashboard

You can inspect the metrics for each version using the Grafana dashboard. For your convenience, an experiment specific URL is defined in the `grafanaURL` field of under the experiment `status`. You can retrieve the Grafana URL using the following command:

```bash
kubectl get experiment reviews-v3-rollout -o jsonpath='{.status.grafanaURL}' -n bookinfo-iter8
```

By default, the base URL given by iter8 to Grafana is `http://localhost:3000`. In a typical Istio installation, you can port-forward your Grafana from Kubernetes to your localhost's port 3000 with the following command:

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

Below is a screenshot of a portion of the Grafana dashboard showing the request rate and the mean latency for reviews-v2 and reviews-v3, right after the controller ended the experiment.

![Grafana Dashboard](../img/grafana_reviews-v2-v3.png)

Note how the traffic shifted towards the canary during the experiment. You can also see that the canary's mean latency was way below the configured threshold of 0.2 seconds.

## Part 2: High-latency canary release: _reviews-v3_ to _reviews-v4_

At this point, you must have completed the part 1 of the tutorial successfully. You can confirm it as follows:

```bash
$ kubectl get experiment reviews-v3-rollout -n bookinfo-iter8
NAME                 PHASE         STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed     ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
```

The command above's output shows that _reviews-v3_ took over from _reviews-v2_ as part of the canary rollout performed before.

You should also see that _reviews-v2_ has been deleted because the experiment completed successfully, all the traffic was shifted to version _reviews-v3_ and the field `cleanup` was set to `true`:

```bash
$ kubectl get pods -n bookinfo-iter8
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-68c7c8666d-m78qx       2/2     Running   0          1h
productpage-v1-7979869ff9-fln6g   2/2     Running   0          1h
ratings-v1-8558d4458d-rwthl       2/2     Running   0          1h
reviews-v3-df64b6df9-ffb42        2/2     Running   0          30m
```

### 1. Canary rollout configuration

Define an `Experiment` for the canary rollout of _reviews-v4_, using the following configuration:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: reviews-v4-rollout
spec:
  service:
    name: reviews
    apiVersion: v1
    baseline: reviews-v3
    candidates: [ "reviews-v4" ]
  criteria:
    - metric: iter8_mean_latenecy
      threshold:
        type: absolute
        value: 200
  duration:
    interval: 30s
    maxIterations: 8
  trafficControl:
    maxIncrement: 20
  cleanup: true
  analyticsSerivceURL: http://iter8-analytics:8080
```

The configuration above is pretty much the same we used in part 1, except that now the baseline version is _reviews-v3_ and the candidate is _reviews-v4_.

To create the above `Experiment` object, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/canary_reviews-v3_to_reviews-v4.yaml
```

You can list all `Experiment` objects like so:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE       STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed   ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Pause       TargetsNotFound: Missing Candidate                   reviews-v3   100          reviews-v4   0
```

The output above shows the new object you just created, for which the candidate deployment _reviews-v4_ is missing. Let us deploy _reviews-v4_ next so that the rollout can begin.

### 2. Deploy _reviews-v4_ and start the rollout

As you have already seen, as soon as we deploy the candidate version, _iter8-controller_ will start the rollout. This time, however, the candidate version (_reviews-v4_) has a performance issue preventing it from satisfying the success criteria in the experiment object. As a result, _iter8_ will roll back to the baseline version.

To deploy _reviews-v4_, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/reviews-v4.yaml
```

Now, if you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that some traffic is now being sent to _reviews-v4_.

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE         STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed     ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Progressing   IterationUpdate: Iteration 1 Started                 reviews-v3   80           reviews-v4   20
```

Unlike the previous rollout, traffic will not shift towards the candidate _reviews-v4_ because it does not meet the success criteria due to a performance problem. At the end of the experiment, _iter8_ rolls back to the baseline (_reviews-v3_), as seen below:

```bash
$ kubectl get experiment reviews-v4-rollout -n bookinfo-iter8
NAME                 PHASE       STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed   ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Completed   ExperimentFailed: Not All Success Criteria Met       reviews-v3   100          reviews-v4   0
```

Because the canary experiment failed, all traffic has been restored to _reviews-v3_. The deployment _reviews-v4_ has been deleted because the `cleanup` field is set to `false` in the experiment. This can be seen by inspecting the pods:

```bash
$ kubectl get pods -n bookinfo-iter8
NAME                              READY   STATUS    RESTARTS   AGE
details-v1-68c7c8666d-m78qx       2/2     Running   0          1h30m
productpage-v1-7979869ff9-fln6g   2/2     Running   0          1h30m
ratings-v1-8558d4458d-rwthl       2/2     Running   0          1h30m
reviews-v3-df64b6df9-ffb42        2/2     Running   0          1h
```

### 3. Inspect the metrics via Granfana dashboard

As in the first experiment, you can inspect the metrics using Grafana. As before, to get the URL to the dashboard specific to this canary release, you can inspect the experiment status:

```bash
kubectl get experiment reviews-v4-rollout -o jsonpath='{.status.grafanaURL}' -n bookinfo-iter8
```

![Grafana Dashboard](../img/grafana_reviews-v3-v4.png)

The dashboard screenshot above shows that the canary version (_reviews-v4_) consistently exhibits a high latency of 5 seconds, way above the threshold of 0.2 seconds specified in our success criterion, and way above the baseline version's latency.

## Part 3: Error-producing canary release: _reviews-v3_ to _reviews-v5_

At this point, you must have completed parts 1 and 2 of the tutorial successfully. You can confirm it as follows:

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE       STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed   ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Completed   ExperimentFailed: Not All Success Criteria Met       reviews-v3   100          reviews-v4   0
```

The command above's output shows that _reviews-v3_ took over from _reviews-v2_ as part of the canary rollout performed before on part 1, and that it continues to be the current version after iter8 had determined that _reviews-v4_ was unsatisfactory.

### 1. Canary rollout configuration

Now, let us set up a canary rollout for _reviews-v5_, using the following `Experiment` configuration:

```yaml
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: reviews-v5-rollout
spec:
  service:
    name: reviews
    apiVersion: v1
    baseline: reviews-v3
    candidates: [ "reviews-v5" ]
  criteria:
    - metric: iter8_mean_latenecy
      threshold:
        type: absolute
        value: 200
    - metric: iter8_error_count
      threshold:
        type: relative
        value: 0.02
        cutoffTrafficOnViolation: true
  duration:
    interval: 30s
    maxIterations: 8
  trafficControl:
    maxIncrement: 20
  cleanup: true
  analyticsSerivceURL: http://iter8-analytics:8080
```

The configuration above differs from the previous ones as follows. We added a second success criterion on the error-rate metric so that the canary version (_reviews-v5_) not only must have a mean latency below 0.2 seconds, but it also needs to have an error rate that cannot exceed the baseline error rate by more than 2%. That comparative analysis on a metric is specified as a `relative` criteria type. Furthermore, the second success criterion sets the flag `cutoffTrafficOnViolation`, which means iter8 will immediately cut off all traffic to the candidate if the criteria fails.

To create the above `Experiment` object, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/canary_reviews-v3_to_reviews-v5.yaml
```

### 2. Deploy _reviews-v5_ and start the rollout

As you already know, as soon as we deploy the candidate version, _iter8-controller_ will start the rollout. This time, the candidate version (_reviews-v5_) has a bug that causes it to return HTTP errors to its callers. As a result, _iter8_ will roll back to the baseline version based on the success criterion on the error-rate metric defined above.

To deploy _reviews-v5_, run the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/reviews-v5.yaml
```

If you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that some of the traffic is now being sent to _reviews-v5_.

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE         STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed     ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Completed     ExperimentFailed: Not All Success Criteria Met       reviews-v3   100          reviews-v4   0
reviews-v5-rollout   Progressing   IterationUpdate: Iteration 1 Started                 reviews-v3   80           reviews-v5   20
```

Because _review-v5_ has an issue causing it to return HTTP errors, as per the success criteria we have specified the traffic will not shift towards it. Furthermore, because the error-rate success criteria indicated the need to cut off traffic on violation, without waiting for the entire duration of the experiment, iter8 will quickly redirect all traffic to _reviews-v3_. It will remain this way for the remainder of the experiment.

```bash
$ kubectl get experiments -n bookinfo-iter8
NAME                 PHASE       STATUS                                               BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   Completed   ExperimentSucceeded: All Success Criteria Were Met   reviews-v2   0            reviews-v3   100
reviews-v4-rollout   Completed   ExperimentFailed: Not All Success Criteria Met       reviews-v3   100          reviews-v4   0
reviews-v5-rollout   Completed   ExperimentFailed: Aborted                            reviews-v3   100          reviews-v5   0
```

### 3. Check the Grafana dashboard

As before, you can check the Grafana dashboard corresponding to the canary release of _reviews-v5_. To get the URL to the dashboard specific to this canary release, run the following command:

```bash
kubectl get experiment reviews-v5-rollout -o jsonpath='{.status.grafanaURL}' -n bookinfo-iter8
```

![Grafana Dashboard](../img/grafana_reviews-v3-v5-req-rate.png)
![Grafana Dashboard](../img/grafana_reviews-v3-v5-error-rate.png)

The dashboard screenshots above show that traffic to the canary version (_reviews-v5_) is quickly interrupted. Also, while the _reviews-v5_ latency is way below the threshold of 0.2 seconds we defined in the latency success criterion, its error rate is 100%, i.e., it generates errors for every single request it processes. That does not meet the error-rate success criterion we defined, which specified that the canary's error rate must be within 2% of that of the baseline (_reviews-v3_) version. According to the dashboard, _reviews-v3_ produced no errors at all.

## Part 4: Canary release of a user-facing service

Up to now, we have demonstrated rolling out a new version of an internal service. In this part of the tutorial we will show you how to use _iter8_ to perform a canary analysis for a user-facing service. By that we mean a service that is exposed to users and services outside the Kubernetes cluster where it runs. In the case of the Bookinfo sample application we use in the tutorial, the _productpage_ service is user facing.

### User-facing service exposed using Kubernetes Ingress

If you expose your service using [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/), you do not need anything special. The `Experiment` object you will need to create will be similar to the ones you saw in the previous parts of this tutorial.

### User-facing service exposed using Istio's VirtualService and Gateway

A service can also be exposed using Istio's VirtualService and Gateway. To remind you, after we deployed Bookinfo [in Part 1 of the tutorial](#part-1-successful-canary-release-reviews-v2-to-reviews-v3), we exposed the _productpage_ service by creating an Istio Gateway and Virtual Service. The VirtualService defines the mapping from an external hostname to an internal service, and binds that to a specific gateway.

We defined _productpage_'s VirtualService and Gateway earlier using the file `iter8-controller/doc/tutorials/istio/bookinfo/bookinfo-gateway.yaml`, which looks like this:

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
apiVersion: iter8.tools/v1alpha2
kind: Experiment
metadata:
  name: productpage-v2-rollout
spec:
  service:
    name: productpage
    apiVersion: v1
    baseline: productpage-v1
    candidates: [ "productpage-v2" ]
  routingReference:
    apiVersion: networking.istio.io/v1alpha3
    kind: VirtualService
    name: bookinfo
  criteria:
    - metric: iter8_mean_latenecy
      threshold:
        type: absolute
        value: 300
  duration:
    interval: 30s
    maxIterations: 8
  trafficControl:
    maxIncrement: 20
  cleanup: true
  analyticsSerivceURL: http://iter8-analytics:8080
```

Note the reference to the existing Istio `VirtualService` named _bookinfo_. This reference will instruct _iter8_ to manipulate that existing VirtualService for the purposes of traffic management.

Let us now create the `Experiment` resource above by running the following command:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/canary_productpage-v1_to_productpage-v2.yaml
```

You can verify that the `Experiment` resource has been created:

```bash
$ kubectl get experiment productpage-v2-rollout -n bookinfo-iter8
NAME                     PHASE   STATUS                               BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Pause   TargetsNotFound: Missing Candidate   productpage-v1   100          productpage-v2   0
```

### 2. Deploy _productpage-v2_ and start the rollout

To start the rollout let us deploy the candidate version (_productpage-v2_).

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/doc/tutorials/istio/bookinfo-v1.0/productpage-v2.yaml
```

You can verify that experiment has started:

```bash
$ kubectl get experiment productpage-v2-rollout -n bookinfo-iter8
NAME                     PHASE         STATUS                                 BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Progressing   IterationUpdate: Iteration 1 Started   productpage-v1   80           productpage-v2   20
```

At this point, if you inspect the `bookinfo` VirtualService, you should see a change in the `route` section reflecting the current traffic split.

```bash
kubectl get vs bookinfo -n bookinfo-iter8 -o yaml | yq r - spec.http[0].route
```

If you look at the spec of that VirtualService, the route section will look something like:

```yaml
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

## Cleanup

You can cleanup by deleting the namespace:

```bash
kubectl delete namespace bookinfo-iter8
```
