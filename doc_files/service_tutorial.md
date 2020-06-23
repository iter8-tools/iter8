# Automated Canary Rollout Using Services

In iter8 the versions of a service being compared can be specified using deployment names or using service names. Other [tutorials](iter8_bookinfo_istio.md) showed how to specify different versions using Kubernetes deployment names. In this tutorial, we learn how to do a canary rollout of an application when different versions are indicated by different Kubernetes service names.

In this tutorial, we again consider the user facing service _productpage_ of the bookinfo application and we we learn how to create an iter8 `Experiment` that specifies the baseline and candidate versions using Kubernetes services. The scenario we consier is here:

![Example Application Deployment Using Services](../img/service_deployment.png)

In this example, the application _productpage.example.com_ can be routed, via an Istio `Gateway` and `VirtualService`, to the Kubernetes services. Iter8 can be used to automate a the rollout including the creation of the Istio `VirtualService`.

## Step 1: Deploy the bookinfo Application

Create a new namespace: $NAMESPACE. We use the name `bookinfo-serivce`.

```bash
export NAMESPACE=bookinfo-service
kubectl create ns $NAMESPACE
kubectl label ns $NAMESPACE istio-injection=enabled
```

Deploy the bookinfo application to a new namespace. In particular, we create the service _productpage-v1_ to access the _productpage_ application.

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.0/doc/tutorials/istio/bookinfo/bookinfo-tutorial.yaml -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.0/doc/tutorials/istio/bookinfo/service/productpage-v1.yaml
```

Create an Istio gateway for the external host `productpage.example.com`:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.0/doc/tutorials/istio/bookinfo/service/bookinfo-gateway.yaml
```

******
Note that at this point, the application is not actually accessible because no Istio `VirtualService` is defined. This tutorial shows how iter8 will create the `VirtualService` for us. This is a use case that might apply, for example, when we do not specify all the Istio components in an application such as when using `helm`.

## Step 2: Create an iter8 Canary Experiment

We can now create an iter8 `Experiment` specifying using the Kubernetes services. The `targetService` portion of the `Experiment` look like this:

```yaml
targetService:
    kind: Service
    baseline: productpage-v1
    candidate: productpage-v2
    hosts:
      - name: productpage.example.com
        gateway: productpage-service
```

It identifies the type of the baseline and candidate as services using `kind: Service`. Finally, it identifies the external host name and the Istio `Gateway` already configured to route traffic.

You can create the `Experiment` using:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.0/doc/tutorials/istio/bookinfo/service/canary_productpage-v1_to_productpage-v2.yaml
```

You can verify that the `Experiment` has been created:

```bash
kubectl -n $NAMESPACE get experiment productpage-v2-rollout
NAME                     PHASE   STATUS                               BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Pause   TargetsNotFound: Missing Candidate   productpage-v1   100          productpage-v2   0
```

The experiment is paused since only the baseline version can be identified. When the candidate version is detected, the experiment will automatically begin execution.

## Step 3: Generate load

As in earlier tutorials, emulate requests coming from users using `curl`:

```bash
watch -x -n 0.1 curl -Is -H 'Host: productpage.example.com' "http://${GATEWAY_URL}/productpage"
```

## Step 4: Deploy the candidate version _productpage-v2_

To start the rollout of the new version of the product page application, deploy the new version:

```bash
kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.0/doc/tutorials/istio/bookinfo/productpage-v2.yaml -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2.0/doc/tutorials/istio/bookinfo/service/productpage-v2.yaml
```

You can verify the experiment has started:

```bash
kubectl -n $NAMESPACE get experiment productpage-v2-rollout
NAME                     PHASE         STATUS                                 BASELINE         PERCENTAGE   CANDIDATE        PERCENTAGE
productpage-v2-rollout   Progressing   IterationUpdate: Iteration 1 Started   productpage-v1   80           productpage-v2   20
```

You should also see a  `VirtualService` has been created to route traffic between the versions of the application:

```bash
kubectl -n $NAMESPACE get virtualservice
NAME                                       GATEWAYS                HOSTS                       AGE
productpage.example.com.iter8-experiment   [productpage-service]   [productpage.example.com]   20m
```

If  you look at the details (using `-o yaml`), you will see the route information contains an entry like this:

```yaml
- destination:
    host: productpage-v1.bookinfo-service.svc.cluster.local
  weight: 80
- destination:
    host: productpage-v2.bookinfo-service.svc.cluster.local
  weight: 20
```

As the canary rollout progresses, you should see traffic shift from the baseline to the candidate version until all of the traffic is being sent to the new version.

## Step 5: Cleanup

You can cleanup by deleting the namespace:

```bash
kubectl delete ns $NAMESPACE
```
