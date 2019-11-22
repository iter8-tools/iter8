# Iter8 on Kubernetes and Istio

These instructions show you how to set up iter8 on Kubernetes with Istio or with Knative.

## Prerequisites

* Kubernetes v1.11 or newer.

If you want to use iter8 with Istio, we require:

* Istio v1.1.5 or newer (those are the versions of Istio we tested iter8 with).
* Your Istio installation must have at least the **istio-pilot** as well as **telemetry** and **Prometheus** enabled.

If you want to use iter8 with Knative, we require:

* [Knative 0.6](https://knative.dev/docs/install/) or newer.

## Install _iter8_ on Kubernetes

iter8 has two components, _iter8_analytics_ and _iter8_controller_. These can be installed as for use with Istio follows:

    kubectl apply \
        -f https://raw.githubusercontent.com/iter8-tools/iter8-analytics/master/install/kubernetes/iter8-analytics.yaml \
        -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/install/iter8-controller.yaml 

To install for use with knative, modify the first file:

    kubectl apply \
        -f https://raw.githubusercontent.com/iter8-tools/iter8-analytics/master/install/knative/iter8-analytics.yaml \
        -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/install/iter8-controller.yaml 


For more control over the installation, use the helm charts directly by cloning the projects:

    git clone git@github.com:iter8-tools/iter8-analytics.git
    git clone git@github.com:iter8-tools/iter8-controller.git

The helm charts are in:

    https://github.com/iter8-tools/iter8-analytics/tree/master/install/kubernetes/helm/iter8-analytics

and
    https://github.com/iter8-tools/iter8-controller/tree/master/install/helm/iter8-controller

**Note on Prometheus:** In order to make assessments on canary releases, _iter8_analytics_ needs to query metrics collected by Istio and stored on Prometheus. The default values for the Helm chart parameters (used in the command above) point _iter8_analytics_ to Prometheus at `http://prometheus.istio-system:9090`, which is the default internal Kubernetes URL of Prometheus installed as an Istio addon. If your Istio installation is shipping metrics to a different Prometheus installation, you need to configure the install parameter `iter8Config.metricsBackendURL` to be _Prometheus host:port_".

### Verify the installation

The command above should have created in the `iter8` namespace an additional pod and service. If you also installed _iter8-analytics_, you should see the following pods and services:

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

To enable users to see Prometheus metrics that pertain to their canary releases, iter8 provides a Grafana dashboard template. To take advantage of Grafana, you will need to import this dashboard template from the Grafana UI.

If you are using iter8 with Istio, you must import the following dashboard template file located in the _iter8-controller_ repository:

```
iter8-controller/config/grafana/istio.json
```

In a typical Istio installation, you can port-forward Grafana from Kubernetes to your localhost's port 3000 with the command below:

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

After running that command, you can access Grafana's UI at `http://localhost:3000` to import the dashboard template from your filesystem (`iter8-controller/config/grafana/istio.json`).

If you are using iter8 with Knative, the dashboard template file that you need to import in the _iter8-controller_ repository is the following:

```
iter8-controller/config/grafana/knative.json
```

## Uninstall _iter8_

If you want to uninstall all _iter8_ components from your Kubernetes cluster, first delete all instances of `Experiment` from all namespaces. Then you can delete iter8 by running the following command from the top directory of your copy of the **_iter8_controller_** repository:

    kubectl delete -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/master/install/iter8-controller.yaml

Note that this command will delete our CRD and wipe out the `iter8` namespace.
