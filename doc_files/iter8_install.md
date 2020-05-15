# Iter8 on Kubernetes and Istio

These instructions show you how to set up iter8 on Kubernetes with Istio.

## Prerequisites

* Kubernetes v1.11 or newer.
* Istio v1.1.5 and newer.
* Your Istio installation must have at least the **istio-pilot** as well as **telemetry** and **Prometheus** enabled.

## Install iter8 on Kubernetes

iter8 has two components, _iter8_analytics_ and _iter8_controller_. To install them, follow these instructions.

### Quick installation

To install iter8 with the default settings, you can apply the default yaml files for _iter8-analytics_ and _iter8-controller_.

To install _iter8-analytics_, run the following command:

```bash
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-analytics/v0.1.0/install/kubernetes/iter8-analytics.yaml
```

To install _iter8-controller_, you need to choose the iter8 yaml file corresponding to the version of Istio telemetry (`v1` or `v2`) you are using. Before Istio 1.5, only Istio telemetry `v1` existed. If that is your Istio telemetry version, run the command below:

```bash
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/install/iter8-controller.yaml
```

Alternatively, if you are using Istio telemetry `v2`, which became available since Istio 1.5 (https://istio.io/docs/reference/config/telemetry/), run the command below:

```bash
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/install/iter8-controller-telemetry-v2.yaml
```

### Customized installation via Helm charts

In case you need to customize the installation of iter8, use the Helm charts listed below:

* _iter8-analytics_: [ https://github.com/iter8-tools/iter8-analytics/releases/download/v0.1.0/iter8-analytics-helm-chart.tar](https://github.com/iter8-tools/iter8-analytics/releases/download/v0.1.0/iter8-analytics-helm-chart.tar)

* _iter8-controller_: [https://github.com/iter8-tools/iter8-controller/releases/download/v0.1.0/iter8-controller-helm-chart.tar](https://github.com/iter8-tools/iter8-controller/releases/download/v0.1.0/iter8-controller-helm-chart.tar)

**Note on Prometheus:** In order to make assessments, _iter8_analytics_ needs to query metrics collected by Istio and stored on Prometheus. The default values for the helm chart parameters (used in the quick installation) point _iter8_analytics_ to Prometheus at `http://prometheus.istio-system:9090` (the default internal Kubernetes URL of Prometheus installed as an Istio addon) without specifying the need for authentication. If your Istio installation is shipping metrics to a different Prometheus installation, or if you need to configure authentication to access Prometheus, you need to set appropriate _iter8-analytics_ Helm chart parameters. Look for the Prometheus-related parameters in the _iter8-analytics_ Helm chart's `values.yaml` file.

**Note on Istio Telemetry:** Make sure to set the parameter `istioTelemetry` in the Helm chart to conform with your environment. Possible values are `v1` or `v2`.

### Verify the installation

After installing _iter8-analytics_ and _iter8-controller_, you should see the following pods and services in the newly created `iter8` namespace:

```bash
$ kubectl get pods -n iter8
NAME                                  READY   STATUS    RESTARTS   AGE
iter8-controller-5f54bb4b88-drr8s     1/1     Running   0          4s
iter8-analytics-5c5758ccf9-p575b      1/1     Running   0          61s
```

```bash
$ kubectl get svc -n iter8
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
iter8-controller         ClusterIP   172.21.62.217   <none>        443/TCP   20s
iter8-analytics          ClusterIP   172.21.106.44   <none>        80/TCP    76s
```

### Import iter8's Grafana dashboard

To enable users to see Prometheus metrics that pertain to their canary releases or A/B tests, iter8 provides a Grafana dashboard template. To take advantage of Grafana, you will need to import this template. To do so, first make sure you can access Grafana. In a typical Istio installation, you can port-forward Grafana from Kubernetes to your localhost's port 3000 with the command below:

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

After running that command, you can access Grafana's UI at `http://localhost:3000`.

Depending on the version of Istio telemetry (`v1` or `v2`) and Kubernetes (prior to 1.16 and 1.16+) you are using, you will need to import a different Grafana dashboard. Follow [these instructions](grafana.md) to import the appropriate dashboard template.

## Uninstall _iter8_

If you want to uninstall all _iter8_ components from your Kubernetes cluster, first delete all instances of `Experiment` from all namespaces. Then, you can delete iter8 by running the following command:

```bash
kubectl delete -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/install/iter8-controller.yaml
```

Note that this command will delete the `Experiment` CRD and wipe out the `iter8` namespace, but it will not remove the iter8 Grafana dashboard if created.
