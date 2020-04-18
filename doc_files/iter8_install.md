# Iter8 on Kubernetes and Istio

These instructions show you how to set up iter8 on Kubernetes with Istio or with Knative.

## Prerequisites

* Kubernetes v1.11 or newer.

If you want to use iter8 with Istio, we require:

* Istio v1.1.5 and newer.
* Your Istio installation must have at least the **istio-pilot** as well as **telemetry** and **Prometheus** enabled.

If you want to use iter8 with Knative, we require:

* [Knative 0.6](https://knative.dev/docs/install/) or newer.

## Install iter8 on Kubernetes

### Quick installation

iter8 has two components, _iter8_analytics_ and _iter8_controller_. These can be installed for use with Istio as follows:

```bash
kubectl apply \
    -f https://raw.githubusercontent.com/iter8-tools/iter8-analytics/v0.0.1/install/kubernetes/iter8-analytics.yaml \
    -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.0.1/install/iter8-controller.yaml
```

To install for use with knative, modify the first file:

```bash
kubectl apply \
    -f https://raw.githubusercontent.com/iter8-tools/iter8-analytics/master/v0.0.1/knative/iter8-analytics.yaml \
    -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/v0.0.1/iter8-controller.yaml
```

### Customized installation

In case you need to change the default configuration options used in the quick installation, use the helm charts directly by cloning the projects:

```bash
git clone git@github.com:iter8-tools/iter8-analytics.git
git clone git@github.com:iter8-tools/iter8-controller.git
```

The _iter8-analytics_ helm chart is [here](https://github.com/iter8-tools/iter8-analytics/tree/master/install/kubernetes/helm/iter8-analytics), and the _iter8-controller_ helm chart is [here](https://github.com/iter8-tools/iter8-controller/tree/master/install/helm/iter8-controller).

**Note on Prometheus:** In order to make assessments on canary releases, _iter8_analytics_ needs to query metrics collected by Istio and stored on Prometheus. The default values for the helm chart parameters (used in the quick installation) point _iter8_analytics_ to Prometheus at `http://prometheus.istio-system:9090`, which is the default internal Kubernetes URL of Prometheus installed as an Istio addon. If your Istio installation is shipping metrics to a different Prometheus installation, you need to set the _iter8-analytics_ helm chart parameter `iter8Config.metricsBackendURL` to your Prometheus `host:port`.

### Verify the installation

After installing _iter8-analytics_ and _iter8-controller_, you should see the following pods and services in the newly created `iter8` namespace:

```bash
$ kubectl get pods -n iter8
NAME                                  READY   STATUS    RESTARTS   AGE
controller-manager-5f54bb4b88-drr8s   1/1     Running   0          4s
iter8-analytics-5c5758ccf9-p575b      1/1     Running   0          61s
```

```bash
$ kubectl get svc -n iter8
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
controller-manager-service   ClusterIP   172.21.62.217   <none>        443/TCP   20s
iter8-analytics              ClusterIP   172.21.106.44   <none>        80/TCP    76s
```

### Import iter8's Grafana dashboard

To enable users to see Prometheus metrics that pertain to their canary releases, iter8 provides a Grafana dashboard template. To take advantage of Grafana, you will need to import this template. To do so, first make sure you can access Grafana. In a typical Istio installation, you can port-forward Grafana from Kubernetes to your localhost's port 3000 with the command below:

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

After running that command, you can access Grafana's UI at `http://localhost:3000`.

To import iter8's dashboard template for Istio, execute the following two commands:

```bash
export DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/config/grafana/istio.json

curl -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/hack/grafana_install_dashboard.sh \
| /bin/bash -
```

If you are using iter8 with Knative, use these two commands instead:

```bash
export DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/config/grafana/knative.json

curl -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/hack/grafana_install_dashboard.sh \
| /bin/bash -
```

## Uninstall _iter8_

If you want to uninstall all _iter8_ components from your Kubernetes cluster, first delete all instances of `Experiment` from all namespaces. Then you can delete iter8 by running the following command:

```bash
kubectl delete -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/install/iter8-controller.yaml
```

Note that this command will delete the `Experiment` CRD and wipe out the `iter8` namespace, but it will not remove the iter8 Grafana dashboard if created.
