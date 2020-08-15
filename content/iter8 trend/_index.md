---
menuTitle: iter8-trend
title: Getting Started with iter8-trend
weight: 40
---

iter8-trend collects metrics of past experiments and summarizes them to show
trends that could expose performance problems creeping up over time, which might
not be obvious if one only compares consecutive versions during canary testing.
This is invaluable to developers when tracking down a performance problem and
pin-pointing it to a particular version that was deployed in the past.

This is an optional component to [iter8](http://github.com/iter8-tools) and
cannot run standalone. It should be installed either as part of iter8
installation process or separately after iter8 is installed.

The following short video introduces iter8-trend:

{{< youtube Sh_4vMcmh6A >}}

## Installation
```
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-trend/master/install/kubernetes/iter8-trend.yaml
```

## Visualization
iter8-trend implements a Prometheus scrape target, so summarized metric data can
be collected by Prometheus and visualized in Grafana. To enable Prometheus to
scrape iter8-trend, you need to add a new scrape target to Prometheus
configuration, e.g., in Istio, you do the following:
```
kubectl -n istio-system edit configmap prometheus
```

In the list of jobs, copy and paste the following at the bottom of the job list:

```
    - job_name: 'iter8_trend'
      static_configs:
      - targets: ['iter8-trend.iter8:8888']
```

and then restart the Prometheus pod for the change to take effect:

```
kubectl -n istio-system delete pod prometheus-xxx-yyy
```

Then, we use `port-forward` to make Grafana available on `localhost:3000`:
```
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

Finally, we import iter8-trend dashboard in Grafana.
```
curl -Ls https://raw.githubusercontent.com/iter8-tools/iter8-trend/master/grafana/install.sh \
| DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-trend/master/grafana/iter8-trend.json /bin/bash -
```

## Uninstall
```
kubectl delete -f https://raw.githubusercontent.com/iter8-tools/iter8-trend/master/install/kubernetes/iter8-trend.yaml
```

