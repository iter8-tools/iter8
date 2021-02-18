---
template: overrides/main.html
---

# Features at a Glance

- **Kubernetes stacks:** use iter8 with Knative, KFServing[^1] and Istio[^2]
- **Testing patterns:** canary and conformance testing
- **Deployment patterns:** progressive traffic shifting and fixing traffic split during experiments
- **Advanced traffic shaping:** shadow deployments, request routing, and user stickiness
- **Version promotion:** automatically roll-forward to the winning version or rollback to the baseline based on the outcome of the experiment
- **App config tools:** YAML/JSON app manifests, Helm, Kustomize
- **Metrics backend:** Prometheus

[^1]: An initial version of iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.
