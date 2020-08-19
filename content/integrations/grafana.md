---
menuTitle: Grafana
title: Grafana
weight: 20
summary: Describes how to use Grafana with iter8
---

iter8 provides a sample Grafana dashboard that can be used to visualize it's default metrics.
The dashboard can be used to follow the progress of an experiment as it executes.

## Import the Sample *iter8 Metrics* Dashboard

To import the sample iter8 dashboard use `kubectl port-forward` to forward a local port to the Grafana port in your Kubernetes cluster:

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

Verify that you can access grafana by accessing it using a browser: http://localhost:3000

Install the iter8 dashboard using the provided install command:

```bash
curl -L -s https://raw.githubusercontent.com/iter8-tools/iter8/v1.0.0-rc1/integrations/grafana/install_dashboard.sh \
| /bin/bash -
```

Verify the installation of the dashboard named *iter8 Metrics* by using a browser to navigate to it.

**Note**: The sample dashboard has been tested on Kubernetes 1.16 and greater with Istio versions 1.4 through 1.6.

## Using the Sample *iter8 Metrics* Dashboard

The iter8 Metrics dashboard shows a number of metrics related to request rate, latency and error.
To see those relavant to a particular experiment, use the drop down menus at the top of the dashboard.
You can select the target namespace and service.
You can then select a baseline version and one or more candidate versions.
This input will be used as filters to display the graphs. Here is an example for a canary experiment:

![Grafana Dashboard]({{< resourceAbsUrl path="images/grafana_reviews-v2-v3.png" >}})

In this example, the request rate diagram shows traffic shifting from one version to another over the course of the experiment.

## Delete the *iter8 Metrics* Dashboard

To remove the dashboard, use the grafana dashboard to *Manage* dashboards, select it and `Delete`.
