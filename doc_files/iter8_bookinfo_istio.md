# Canary releases with iter8 on Kubernetes and Istio

This tutorial shows you how _iter8_ can be used to perform canary releases by gradually shifting traffic to a canary version of a microservice. In the first part of the tutorial, we will walk you through a case where the canary version performs as expected and, therefore, takes over from the previous version at the end. In the second part, we will deal with a canary version that is not satisfactory, in which case _iter8_ will roll back to the previous version.

The tutorial is based on the [Bookinfo sample application](https://istio.io/docs/examples/bookinfo/) that is distributed with Istio. This application comprises 4 microservices, namely, _productpage_, _details_, _reviews_, and _ratings_, as illustrated [here](https://istio.io/docs/examples/bookinfo/). Please, follow our instructions below to deploy the sample application as part of the tutorial.

## Part 1: Successful canary release: _reviews-v2_ to _reviews-v3_

### 1. Deploy the Bookinfo application

At this point, we assume that you have already followed the [instructions](istio_install.md) to install _iter8_ on your Kubernetes cluster. The next step is to deploy the sample application we will use for the tutorial.

First, let us create a `bookinfo-iter8` namespace configured to enable auto-injection of the Istio sidecar:

```bash
kubectl apply -f https://raw.github.ibm.com/istio-research/iter8-controller/master/doc/tutorials/istio/bookinfo/namespace.yaml?token=AAAROHqyPLzp4h4FWozSZdHNcRkz2sGCks5dE9sMwA%3D%3D
```

Next, let us deploy the Bookinfo application:

```bash
kubectl apply -n bookinfo-iter8 -f https://raw.github.ibm.com/istio-research/iter8-controller/master/doc/tutorials/istio/bookinfo/bookinfo-tutorial.yaml?token=AAAROPJZY04WTinFDpmJohfu0K28lxPFks5dE-rRwA%3D%3D
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
kubectl apply -n bookinfo-iter8 -f https://raw.github.ibm.com/istio-research/iter8-controller/master/doc/tutorials/istio/bookinfo/bookinfo-gateway.yaml?token=AAAROOeOINW85InWPx9RWpYjulM23RJYks5dE-43wA%3D%3D
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

To access the application, you need to determine the ingress IP and port for the application in your environment. You can do so by following steps 3 and 4 of the Istio instructions [here](https://istio.io/docs/examples/bookinfo/#determining-the-ingress-ip-and-port) to set the environment variables `INGRESS_HOST`, `INGRESS_PORT`, and `GATEWAY_URL`, which will capture the correct IP address and port for your environment. Once you have done so, you can check if you can access the application with the following command:

```bash
curl -H "Host: bookinfo.sample.dev" -o /dev/null -s -w "%{http_code}\n" "http://${GATEWAY_URL}/productpage"
```

If everything is working, the command above should show `200`. Note that the curl sets the host header to match the host we associated the VirtualService with (`bookinfo.sample.dev`). If you want to access the application from your browser, you will need to set this header using the browser plugin of your choice.

### 3. Perform a canary release of the _reviews_ service

At this point, Bookinfo is using version 2 of the _reviews_ service (_reviews-v2_). Let us now use _iter8_ to perform a canary rollout of version 3 of this service (_reviews-v3_).

First, we need to instruct _iter8_ that we are about to perform this canary rollout.