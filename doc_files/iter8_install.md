# Iter8 on Kubernetes and Istio

These instructions show you how to set up iter8 on Kubernetes with Istio.

## Prerequisites

* Kubernetes v1.11 or newer.
* Istio v1.1.5 and newer.
* Your Istio installation must have at least the **istio-pilot** as well as **telemetry** and **Prometheus** enabled.

## Install iter8 on Kubernetes

iter8 has two components, _iter8_analytics_ and _iter8_controller_. To install them, follow these instructions.

### Quick installation (latest release)

To install the latest iter8 release with the default settings, you can apply the default yaml files for _iter8-analytics_ and _iter8-controller_ by running the following command:

```bash
kubectl apply \
    -f https://github.com/iter8-tools/iter8-analytics/releases/latest/download/iter8-analytics.yaml \
    -f https://github.com/iter8-tools/iter8-controller/releases/latest/download/iter8-controller.yaml
```

### Customized installation via Helm charts (latest release)

In case you need to customize the installation of iter8's latest release, use the Helm charts listed below:

* _iter8-analytics_: [https://github.com/iter8-tools/iter8-analytics/releases/latest/download/iter8-analytics-helm-chart.tar](https://github.com/iter8-tools/iter8-analytics/releases/latest/download/iter8-analytics-helm-chart.tar)

* _iter8-controller_: [https://github.com/iter8-tools/iter8-controller/releases/latest/download/iter8-controller-helm-chart.tar](https://github.com/iter8-tools/iter8-controller/releases/latest/download/iter8-controller-helm-chart.tar)

### Installing an older release

In case you need to install an old iter8 release, please refer to its corresponding documentation in the link below. In the URL below, you need to replace the string `<release>` with the string corresponding to your desired release. For example, a valid release string would be `v0.0.1`.

```
https://github.com/iter8-tools/docs/tree/<release>
```

Note that the URL above points to the GitHub branch and tag corresponding to your desired release in the documentation repository.

**Note on Prometheus:** In order to make assessments, _iter8_analytics_ needs to query metrics collected by Istio and stored on Prometheus. The default values for the helm chart parameters (used in the quick installation) point _iter8_analytics_ to Prometheus at `http://prometheus.istio-system:9090`, which is the default internal Kubernetes URL of Prometheus installed as an Istio addon. If your Istio installation is shipping metrics to a different Prometheus installation, you need to set the _iter8-analytics_ Helm chart parameter `iter8Config.metricsBackendURL` to your Prometheus `host:port`.

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

To enable users to see Prometheus metrics that pertain to their canary releases or A/B tests, iter8 provides a Grafana dashboard template. To take advantage of Grafana, you will need to import this template. To do so, first make sure you can access Grafana. In a typical Istio installation, you can port-forward Grafana from Kubernetes to your localhost's port 3000 with the command below:

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

After running that command, you can access Grafana's UI at `http://localhost:3000`.

Depending on the version of Istio telemetry (`v1` or `v2`) and Kubernetes (prior to 1.16 and 1.16+) you are using, you will need to import a different Grafana dashboard. Follow [these instructions](grafana.md) to import the appropriate dashboard template.

## Uninstall _iter8_

If you want to uninstall all _iter8_ components from your Kubernetes cluster, first delete all instances of `Experiment` from all namespaces. Then, you can delete iter8 by running the following command, adjusting the URL based on the release you are uninstalling.

```bash
kubectl delete -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/<release>/install/iter8-controller.yaml
```

In the URL above, replace the string `<release>` with the string for your desired release. For example, a valid release string is `v0.0.1`.

Note that this command will delete the `Experiment` CRD and wipe out the `iter8` namespace, but it will not remove the iter8 Grafana dashboard if created.
