## Importing iter8's Grafana dashboard template

Different versions of Istio telemetry (`v1` or `v2`) report application-level metrics differently. Similarly, since Kubernetes 1.16, cluster-level resource utilization metrics look slightly different on Prometheus. Therefore, depending on your combination of Istio telemetry and Kubernetes versions, you will need to populate Grafana with a specific iter8 Grafana dashboard template. 

Below are instructions for each of these combinations. Choose only the instructions for the combination that matches your environment.

### Istio _telemetry v1_ and _Kubernetes prior to 1.16_:

Execute the following two commands:

```bash
export DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/config/grafana/istio-telemetry-v1.json

curl -L -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/hack/grafana_install_dashboard.sh \
| /bin/bash -
```

### Istio _telemetry v2_ and _Kubernetes prior to 1.16_:

Execute the following two commands:

```bash
export DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/config/grafana/istio-telemetry-v2.json

curl -L -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/hack/grafana_install_dashboard.sh \
| /bin/bash -
```

### Istio _telemetry v1_ and _Kubernetes 1.16 and later_:

Execute the following two commands:

```bash
export DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/config/grafana/istio-telemetry-v1-k8s-16.json

curl -L -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/hack/grafana_install_dashboard.sh \
| /bin/bash -
```

### Istio _telemetry v2_ and _Kubernetes 1.16 and later_:

Execute the following two commands:

```bash
export DASHBOARD_DEFN=https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/config/grafana/istio-telemetry-v2-k8s-16.json

curl -L -s https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.1.0/hack/grafana_install_dashboard.sh \
| /bin/bash -
```
