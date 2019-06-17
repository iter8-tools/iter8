# Iter8 on Kubernetes and Istio

These instructions show you how to set up iter8 on Kubernetes with Istio.

## Prerequisites

* Kubernetes v1.11 or newer; and
* Istio v1.1.5 and newer (those are the versions of Istio we tested iter8 with).

Your Istio installation must have at least the **istio-pilot** as well as **telemetry** and **Prometheus** enabled.

## Install _iter8_ on Kubernetes

Below are instructions to run the two _iter8_ components (_iter8_analytics_ and _iter8_controller_) on Kubernetes.

### Setting up _iter8-analytics_

#### Step 1. Clone the GitHub repository

```bash
git clone git@github.ibm.com:istio-research/iter8.git
```
#### Step 2. Run _iter8-analytics_ using our Helm chart

We have a Helm chart to make it easy to set up _iter8-analytics_. Thus, make sure you have the Helm client installed on your computer. If not, follow [these instructions](https://helm.sh/docs/using_helm/#installing-the-helm-client) to install it. 

Assuming you have the Kubernetes CLI `kubectl` pointing at your desired Kubernetes cluster (v1.11 or newer), with Istio installed (as per the prerequisites above), all you need to do to deploy _iter8_analytics_ to your cluster is to run the following command from the top directory of the iter8-analytics repository:

```bash
helm template install/kubernetes/helm/iter8-analytics --name iter8-analytics | kubectl apply  -f -
```

**Note on Prometheus:** In order to make assessments on canary releases, _iter8_analytics_ needs to query metrics collected by Istio and stored on Prometheus. The default values for the Helm chart parameters point _iter8_analytics_ to Prometheus at `http://prometheus.istio-system:9090`, which is the default internal Kubernetes URL of Prometheus installed as an Istio addon. If your Istio installation is shipping metrics to a different Prometheus URL, you need to run the following command instead:

```bash
helm template install/kubernetes/helm/iter8-analytics --name iter8-analytics --set iter8Config.metricsBackendURL="Your Prometheus URL"| kubectl apply  -f -
```

#### Step 3. Verify the installation

The command above should have created the `iter8` namespace, where you should see a running pod and a service as below:

```
$ kubectl get pods -n iter8
NAME                               READY   STATUS    RESTARTS   AGE
iter8-analytics-865c754857-bqxzt   1/1     Running   0          9s
```

```
$ kubectl get svc -n iter8
NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
iter8-analytics   ClusterIP   172.21.26.55   <none>        80/TCP    11s
```

### Setting up _iter8-controller_

#### Step 1. Clone the GitHub repository

```bash
git clone git@github.ibm.com:istio-research/iter8-controller.git
```

#### Step 2. Run _iter8-controller_
