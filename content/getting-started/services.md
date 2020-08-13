---
menuTitle: Services
title: Automated Canary Rollout Using Services
weight: 30
summary: Perform a canary rollout when different versions have different service names
---

In iter8 the versions of a service being compared can be specified using deployment names or using service names. Other [tutorials](../deployments/) showed how to specify different versions using Kubernetes deployment names. In this tutorial, we learn how to do a canary rollout of an application when different versions are indicated by different Kubernetes service names.

In this tutorial, we again consider the user facing service _productpage_ of the bookinfo application and we learn how to create an iter8 `Experiment` that specifies the baseline and candidate versions using Kubernetes services. The scenario we consider is here:

![Example Application Deployment Using Services]({{< resourceAbsUrl path="images/service_deployment.png" >}})

In this example, the application _productpage.example.com_ can be routed, via an Istio `Gateway` and `VirtualService`, to the Kubernetes services. Iter8 can be used to automate the rollout including the creation of the Istio `VirtualService`.

## Step 1: Deploy the bookinfo Application

Create a new namespace: $NAMESPACE. We use the name `bookinfo-serivce`.

```bash
export NAMESPACE=bookinfo-service
kubectl create ns $NAMESPACE
kubectl label ns $NAMESPACE istio-injection=enabled
```

Deploy the bookinfo application to a new namespace. In particular, we create the service _productpage-v1_ to access the _productpage_ application.

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/bookinfo-tutorial.yaml -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/service/productpage-v1.yaml
```

## Step 2: Configure Traffic to the Application

Create an Istio gateway for the external host `productpage.example.com`:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/service/bookinfo-gateway.yaml
```

At this point, the application is not actually accessible to users because no `VirtualService` has been defined. Rather than manually define it, this tutorial shows how iter8 can be used to create it for us. To do so, define an iter8 canary experiment from the current version of the application to itself. Clearly, this will succeed, and, as a side effect, a `VirtualService` will be created.

In the experiment, the `targetService` will look like:

```yaml
targetService:
    kind: Service
    baseline: productpage-v1
    candidate: productpage-v1
    hosts:
      - name: productpage.example.com
        gateway: productpage-service
```

It identifies the type of the baseline and candidate as services using `kind: Service`. The baseline and candidate names are the same. Further, it identifies the external host name and the Istio `Gateway` already configured.

To optimize the bootstrapping process, we can eliminate all of the `successCriteria`. This has the further benefit of eliminating the need for user traffic. We can also alter the  `trafficControl` options to reduce the time and number of iterations required.

You can apply an optimized `Experiment` using:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/service/bootstrap-productpage.yaml
```

You can verify that the `Experiment` has been created and finishes quickly:

```bash
kubectl -n $NAMESPACE get experiment productpage-bootstrap
NAME                     PHASE       STATUS                                              BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-bootstrap    Completed   ExperimentSucceeded: Last Iteration Was Completed   productpage-v1   0          productpage-v1   100
```

You can also verify that a virtual service has been created:

```bash
kubectl -n $NAMESPACE get virtualservice
NAME                                       GATEWAYS                HOSTS                       AGE
productpage.example.com.iter8-experiment   [productpage-service]   [productpage.example.com]   20m
```

This approach may seem unintuitive. However, we illustrate it here because this bootstrapping issue often arises when automating the use of canary rollouts.

## Step 3: Create an iter8 Canary Experiment

We can now create a canary `Experiment` from version `productpage-v1` to `productpage-v2`. The following command will create the `Experiment`:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/service/canary_productpage-v1_to_productpage-v2.yaml
```

You can verify that the `Experiment` has been created:

```bash
kubectl -n $NAMESPACE get experiment productpage-v2-rollout
NAME                     PHASE   STATUS                               BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Pause   TargetsNotFound: Missing Candidate   productpage-v1   100   productpage-v2   0
```

The experiment is paused since only the baseline version can be identified. When the candidate version is detected, the experiment will automatically begin execution.

## Step 4: Generate load

As in earlier tutorials, emulate requests coming from users using `curl`:

```bash
watch -x -n 0.1 curl -Is -H 'Host: productpage.example.com' "http://${GATEWAY_URL}/productpage"
```

## Step 5: Deploy the candidate version _productpage-v2_

To start the rollout of the new version of the productpage application, deploy the new version:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/productpage-v2.yaml -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.1/doc/tutorials/istio/bookinfo/service/productpage-v2.yaml
```

You can verify the experiment has started:

```bash
kubectl -n $NAMESPACE get experiment productpage-v2-rollout
NAME                     PHASE         STATUS                                 BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Progressing   IterationUpdate: Iteration 1 Started   productpage-v1   80           productpage-v2   20
```

You can also verify that the  `VirtualService` created by the bootstrap step has been reused:

```bash
kubectl -n $NAMESPACE get virtualservice
NAME                                       GATEWAYS                HOSTS                       AGE
productpage.example.com.iter8-experiment   [productpage-service]   [productpage.example.com]   20m
```

As the canary rollout progresses, you should see traffic shift from the baseline to the candidate version until all of the traffic is being sent to the new version.

## Cleanup

You can cleanup by deleting the namespace:

```bash
kubectl delete ns $NAMESPACE
```
