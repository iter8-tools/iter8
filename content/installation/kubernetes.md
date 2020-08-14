---
menuTitle: Kubernetes
title: Install on Kubernetes
weight: 10
summary: Install on Kubernetes and Istio
---

These instructions show you how to set up iter8 on Kubernetes with Istio.

## Prerequisites

* Kubernetes v1.11 or newer.
* Istio v1.1.5 and newer.
* Your Istio installation must have at least the **istio-pilot** as well as **telemetry** and **Prometheus** enabled.

## Install iter8 on Kubernetes

iter8 has two components, _iter8_analytics_ and _iter8_controller_. To install them, follow the instructions below. For additional considerations when installing iter8 on Red Hat OpenShift, check out [these instructions](../red-hat/).

### Quick installation

To install iter8 with the default settings, you can run the following install script:

```bash
curl -L -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/v1.0.0-preview/install/install.sh \
| /bin/bash -
```

### Customized installation via Helm charts

In case you need to customize the installation of iter8, use the Helm charts listed below:

* _iter8-analytics_: [https://github.com/iter8-tools/iter8-analytics/releases/download/v1.0.0-preview/iter8-analytics.tgz](https://github.com/iter8-tools/iter8-analytics/releases/download/v1.0.0-preview/iter8-analytics.tgz)

* _iter8-controller_: [https://github.com/iter8-tools/iter8-controller/releases/download/v1.0.0-preview/iter8-controller.tgz](https://github.com/iter8-tools/iter8-controller/releases/download/v1.0.0-preview/iter8-controller.tgz)

**Note on Prometheus:** In order to make assessments, _iter8-analytics_ needs to query metrics collected by Istio and stored on Prometheus. The default values for the helm chart parameters (used in the quick installation) point _iter8-analytics_ to the Prometheus server at `http://prometheus.istio-system:9090` (the default internal Kubernetes URL of Prometheus installed as an Istio addon) without specifying any need for authentication. If your Istio installation is shipping metrics to a different Prometheus service, or if you need to configure authentication to access Prometheus, you need to set appropriate _iter8-analytics_ Helm chart parameters. Look in the section `metricsBackend` of the Helm chart's `values.yaml` file for details.

**Note on Istio Telemetry:** When deploying _iter8-controller_ using helm, make sure to set the parameter `istioTelemetry` to conform with your environment. Possible values are `v1` or `v2`. Use `v1` if the Istio mixer is not disabled. You can determine whether or not the mixer is disabled using this command:

```bash
kubectl -n $ISTIO_NAMESPACE get cm istio -o json | jq .data.mesh | grep -o 'disableMixerHttpReports: [A-Za-z]\+' | cut -d ' ' -f2
```

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

## Uninstall _iter8_

If you want to uninstall all _iter8_ components from your Kubernetes cluster, first delete all instances of `Experiment` from all namespaces. Then, you can delete iter8 by running the following command:

```bash
kubectl delete -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v1.0.0-preview/install/iter8-controller.yaml
```
