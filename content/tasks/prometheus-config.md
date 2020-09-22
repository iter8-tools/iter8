---
menuTitle: Prometheus Configuration
title: Prometheus Configuration
weight: 20
summary: Learn how to configure Prometheus to scrape application metrics
---

Istio provides deployment annotations that direct Prometheus to scrape application pods for metrics. You might use this to scrape application specific metrics such as the reward metric used in the [A/B/n tutoria](({{< ref "abn" >}})).

The annotations are:

```yaml
prometheus.io/scrape: "true"
prometheus.io/path: /metrics
prometheus.io/port: "9080"
```

Unfortunately, in some instances of Istio, the Prometheus server installed with Istio expects communication with the pod to be implemented using mTLS. We have observed this in versioms prior to Istio 1.7.0. To avoid this, you can reconfigure Prometheus as follows:

```bash
kubectl --namespace istio-system edit configmap/prometheus
```

Find the `scrape_configs` entry with `job_name: 'kubernetes-pods`.
Comment out the entry with a `source_label` of `__meta_kubernetes_pod_annotation_sidecar_istio_io_status` if one exists.
In this example, the last three lines have been commented out:

```yaml
- job_name: 'kubernetes-pods'
  kubernetes_sd_configs:
  - role: pod
  relabel_configs:  # If first two labels are present, pod should be scraped  by the istio-secure job.
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  #- source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status]
  #  action: drop
  #  regex: (.+)
```

Then restart the prometheus pod if any changes were made:

```bash
kubectl --namespace istio-system delete pod $(kubectl --namespace istio-system get pod --selector='app=prometheus' -o jsonpath='{.items[0].metadata.name}')
```

You should only have to do this once.
