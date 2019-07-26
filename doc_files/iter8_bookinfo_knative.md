
# Automated canary releases with iter8 on Knative

This tutorial shows you how _iter8_ can be used to perform canary releases by gradually shifting traffic to a canary version of a Knative service. In the first part of the tutorial, we will walk you through a case where the canary version performs as expected and, therefore, takes over from the previous version at the end. In the second part, we will deal with a canary version that is not satisfactory, in which case _iter8_ will roll back to the previous version.

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) that is distributed with Istio. This application comprises 4 Knative services, namely, _productpage_, _details_, _reviews_, and _ratings_, as illustrated [here](https://istio.io/docs/examples/bookinfo/). Please, follow our instructions below to deploy the sample application as part of the tutorial.

# YAML files used in the tutorial

All Kubernetes YAML files you will need in this tutorial are in the _iter8-controller_ repository [here](https://github.com/iter8-tools/iter8-controller/tree/master/doc/tutorials/knative/bookinfo). Thus, the first thing you need to do is to clone this repository.

```bash
git clone git@github.com:iter8-tools/iter8-controller.git
```

## Part 1: Successful canary release: _reviews-v2_ to _reviews-v3_

### 1. Deploy the Bookinfo application

At this point, we assume that you have already followed the [instructions](iter8_install.md) to install _iter8_ on your Kubernetes cluster. The next step is to deploy the sample application we will use for the tutorial.

Then, let us create a `knative-bookinfo-iter8` namespace:

```bash
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/namespace.yaml
```

Next, let us deploy the Bookinfo application:

```bash
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/bookinfo-tutorial.yaml
```

You should see the following pods in the `knative-bookinfo-iter8` namespace. Make sure the Knative services readiness is "True".

```bash
$ kubectl get ksvc -n knative-bookinfo-iter8
NAME          URL                                                                                    LATESTCREATED       LATESTREADY         READY   REASON
details       http://details-knative-bookinfo-iter8.stable.us-south.containers.appdomain.cloud       details-rrz5c       details-rrz5c       True
productpage   http://productpage-knative-bookinfo-iter8.stable.us-south.containers.appdomain.cloud   productpage-68hfh   productpage-68hfh   True
ratings       http://ratings-knative-bookinfo-iter8.stable.us-south.containers.appdomain.cloud       ratings-mwcfk       ratings-mwcfk       True
reviews       http://reviews-knative-bookinfo-iter8.stable.us-south.containers.appdomain.cloud       reviews-rgx8x       reviews-rgx8x       True```
```
We have deployed "version 2" of the _reviews_ service, and version 1 of all others.

### 2. Access the Bookinfo application

To access the application, you need to determine the ingress IP and port for the application in your environment, as follows:

```sh
export IP_ADDRESS=$(kubectl get svc istio-ingressgateway --namespace istio-system --output 'jsonpath={.status.loadBalancer.ingress[0].ip}')
```

You can now check if you can access the application with the following command:

```bash
curl -H "Host: productpage.knative-bookinfo-iter8.svc.cluster.local" -o /dev/null -s -w "%{http_code}\n" "http://${IP_ADDRESS}/productpage"
```

If everything is working, the command above should show `200`. Note that the curl command above sets the host header to match the host associated to the Knative service. If you want to access the application from your browser, you will need to set this header using a browser's plugin of your choice or create an ingress.

### 3. Generate load to the application

Let us now generate load to the application, emulating requests coming from users. To do so, we recommend you run the command below on a separate terminal:

```bash
watch -n 0.1 'curl -H "Host: productpage.knative-bookinfo-iter8.svc.cluster.local" -Is "$IP_ADDRESS/productpage"'
```

This command will send 10 requests per second to the application. Note that the environment variable `IP_ADDRESS` must have been set as per step 2 above. Among other things, the command output should show an HTTP code of 200, as below:

```
HTTP/1.1 200 OK
content-type: text/html; charset=utf-8
content-length: 3728
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
  namespace: knative-bookinfo-iter8
spec:
  targetService:
    apiVersion: serving.knative.dev/v1alpha1
    name: reviews
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

The configuration above specifies the baseline and candidate versions in terms of Knative service revision name. The rollout is configured to last for 8 iterations (`maxIterations`) of `30s` (`interval`). At the end of each iteration, if the candidate version meets the specified success criteria, the traffic sent to it will increase by 20 percentage points (`trafficStepSize`) up to 80% (`maxTrafficPercentage`). At the end of the last iteration, if the success criteria are met, the candidate version will take over from the baseline.

In the example above, we specified only one success criterion. In particular, we stated that the mean latency exhibited by the candidate version should not exceed the threshold of 0.2 seconds. At the end of each iteration, _iter8-controller_ calls _iter8-analytics_, which in turn analyzes the metrics of interest (in this case, only mean latency) against the corresponding criteria. The number of data points analyzed during an experiment is cumulative, that is, it carries over from iteration to iteration.

The next step of this tutorial is to actually create the configuration above. To that end, you can either copy and paste the yaml above to a file and then run `kubectl apply -f ...` on it, or you can run the following command:

```bash
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/canary_reviews-v2_to_reviews-v3.yaml
```

You can verify that the `Experiment` object has been created as shown below:

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS                     BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   False       MissingCandidateRevision   reviews-v2   100            reviews-v3   0
```

As you can see, _iter8_ is reporting that 100% of the traffic is sent to the baseline version (_reviews-v2_) and that the candidate (_reviews-v3_) is missing. As soon as the controller sees the candidate version, it will start the rollout. Next, let us deploy the candidate version to trigger the canary rollout.

### 5. Deploy the canary version and start the rollout

As soon as we deploy _reviews-v3_, _iter8-controller_ will start the rollout. To deploy _reviews-v3_, you can run the following command:

```bash
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/reviews-v3.yaml
```

Now, if you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that 20% of the traffic is now being sent to _reviews-v3_:

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS        BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   False       Progressing   reviews-v2   80           reviews-v3   20
```

At about every 30s you should see the traffic shift towards _reviews-v3_ by 20 percentage points.

### 6. Check the Grafana dashboard

You can also check a Grafana dashboard specific to the `Experiment` object corresponding to the rollout you are running. The URL to the Grafana dashboard for the experiment is the value of the field `grafanaURL` under the object's `status`. One way to get the Grafana URL that you can paste to your browser is through the following command:

```bash
kubectl get experiment reviews-v3-rollout -o jsonpath='{.status.grafanaURL}' -n knative-bookinfo-iter8
```

Below is a screenshot of a portion of the Grafana dashboard showing the request rate and the mean latency for reviews-v2 and reviews-v3, right after the controller ended the experiment.

![Grafana Dashboard](../img/grafana_knative-reviews-v2-v3.png)

Note how the traffic shifted towards the canary during the experiment. You can also see that the canary's mean latency was way below the configured threshold of 0.2 seconds.

## Part 2: Canary release resulting in rollback: _reviews-v3_ to _reviews-v4_

At this point, you must have completed the part 1 of the tutorial successfully. You can confirm it as follows:

```bash
$ kubectl get experiment.iter8.tools reviews-v3-rollout -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS   BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   True                 reviews-v2   0            reviews-v3   100
```

The command above's output shows that _reviews-v3_ took over from _reviews-v2_ as part of the canary rollout performed before.

### 1. Canary rollout configuration

Now, let us set up a canary rollout for _reviews-v4_, using the following `Experiment` configuration:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
metadata:
  name: reviews-v4-rollout
  namespace: knative-bookinfo-iter8
spec:
  targetService:
    apiVersion: serving.knative.dev/v1alpha1
    name: reviews
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
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/canary_reviews-v3_to_reviews-v4.yaml
```

You can list all `Experiment` objects like so:

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS                            BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   True                                          reviews-v2   0            reviews-v3   100
reviews-v4-rollout   False       MissingCandidateRevision   reviews-v3   100          reviews-v4   0
```

The output above shows the new object you just created, for which the candidate deployment _reviews-v4_ is missing. Let us deploy _reviews-v4_ next so that the rollout can begin.

### 2. Deploy _reviews-v4_ and start the rollout

As you have already seen, as soon as we deploy the candidate version, _iter8-controller_ will start the rollout. This time, however, the candidate version (_reviews-v4_) has a performance issue preventing it from satisfying the success criteria in the experiment object. As a result, _iter8_ will roll back to the baseline version.

To deploy _reviews-v4_, run the following command:

```bash
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/reviews-v4.yaml
```

Now, if you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that 20% of the traffic is now being sent to _reviews-v4_.

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS        BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   True                      reviews-v2   0            reviews-v3   100
reviews-v4-rollout   False       Progressing   reviews-v3   80           reviews-v4   20
```

However, unlike the previous rollout, traffic will not shift towards the candidate _reviews-v4_ because it does not meet the success criteria due to a performance problem. At the end of the experiment, _iter8_ rolls back to the baseline (_reviews-v3_), as seen below:

```bash
$ kubectl get experiments.iter8.tools reviews-v4-rollout -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS                                     BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v4-rollout   True        ExperimentFailure                          reviews-v3   100          reviews-v4   0
```

You can also check that the `reviews` Knative service traffic all goes to _reviews-v3_:

```sh
$ kubectl get ksvc -n knative-bookinfo-iter8 reviews -o custom-columns=NAME:.metadata.name,BASELINE:.spec.traffic[0].revisionName,PERCENT:.spec.traffic[0].percent,CANDIDATE:.spec.traffic[1].revisionName,PERCENT:.spec.traffic[1].percent
```

### 3. Check the Grafana dashboard

As before, you can check the Grafana dashboard corresponding to the canary release of _reviews-v4_. To get the URL to the dashboard specific to this canary release, run the following command:

```bash
kubectl get experiments.iter8.tools reviews-v4-rollout -o jsonpath='{.status.grafanaURL}' -n knative-bookinfo-iter8
```

![Grafana Dashboard](../img/grafana_knative-reviews-v3-v4.png)

The dashboard screenshot above shows that the canary version (_reviews-v4_) consistently exhibits a high latency of 5 seconds, way above the threshold of 0.2 seconds specified in our success criterion, and way above the baseline version's latency.


## Part 3: Error-producing canary release: _reviews-v3_ to _reviews-v5_

At this point, you must have completed parts 1 and 2 of the tutorial successfully. You can confirm it as follows:

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS              BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   True                            reviews-v2   0            reviews-v3   100
reviews-v4-rollout   True        ExperimentFailure   reviews-v3   100          reviews-v4   0
```

The command above's output shows that _reviews-v3_ took over from _reviews-v2_ as part of the canary rollout performed before on part 1, and that it continues to be the current version after iter8 had determined that _reviews-v4_ was unsatisfactory.


### 1. Canary rollout configuration

Now, let us set up a canary rollout for _reviews-v5_, using the following `Experiment` configuration:

```yaml
apiVersion: iter8.tools/v1alpha1
kind: Experiment
metadata:
  name: reviews-v5-rollout
  namespace: knative-bookinfo-iter8
spec:
  targetService:
    apiVersion: serving.knative.dev/v1alpha1
    name: reviews
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
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/canary_reviews-v3_to_reviews-v5.yaml
```

### 2. Deploy _reviews-v5_ and start the rollout

As you already know, as soon as we deploy the candidate version, _iter8-controller_ will start the rollout. This time, the candidate version (_reviews-v5_) has a bug that causes it to return HTTP errors to its callers. As a result, _iter8_ will roll back to the baseline version based on the success criterion on the error-rate metric defined above.

To deploy _reviews-v5_, run the following command:

```bash
kubectl apply -f iter8-controller/doc/tutorials/knative/bookinfo/reviews-v5.yaml
```

If you check the state of the `Experiment` object corresponding to this rollout, you should see that the rollout is in progress, and that 20% of the traffic is now being sent to _reviews-v5_.

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS              BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   True                            reviews-v2   0            reviews-v3   100
reviews-v4-rollout   True        ExperimentFailure   reviews-v3   100          reviews-v4   0
reviews-v5-rollout   False       Progressing         reviews-v3   80           reviews-v5   20
```

Because _review-v5_ has an issue causing it to return HTTP errors, as per the success criteria we have specified the traffic will not shift towards it. Furthermore, because the error-rate success criteria indicated the need to stop on failure, without waiting for the entire duration of the experiment, iter8 will rollback to _reviews-v3_ quickly. You should see the following after several seconds:

```bash
$ kubectl get experiments.iter8.tools -n knative-bookinfo-iter8
NAME                 COMPLETED   STATUS              BASELINE     PERCENTAGE   CANDIDATE    PERCENTAGE
reviews-v3-rollout   True                            reviews-v2   0            reviews-v3   100
reviews-v4-rollout   True        ExperimentFailure   reviews-v3   100          reviews-v4   0
reviews-v5-rollout   True        ExperimentFailure   reviews-v3   100          reviews-v5   0
```

### 3. Check the Grafana dashboard

As before, you can check the Grafana dashboard corresponding to the canary release of _reviews-v5_. To get the URL to the dashboard specific to this canary release, run the following command:

```bash
kubectl get experiment.iter8.tools reviews-v5-rollout -o jsonpath='{.status.grafanaURL}' -n knative-bookinfo-iter8
```

![Grafana Dashboard](../img/grafana_knative-reviews-v3-v5-req-rate.png)
![Grafana Dashboard](../img/grafana_knative-reviews-v3-v5-error-rate.png)

The dashboard screenshots above show that traffic to the canary version (_reviews-v5_) is quickly interrupted. Also, while the _reviews-v5_ latency is way below the threshold of 0.2 seconds we defined in the latency success criterion, its error rate is 100%, i.e., it generates errors for every single request it processes. That does not meet the error-rate success criterion we defined, which specified that the canary's error rate must be within 2% of that of the baseline (_reviews-v3_) version. According to the dashboard, _reviews-v3_ produced no errors at all.
