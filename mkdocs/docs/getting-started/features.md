---
template: overrides/main.html
hide:
- toc
---

# Features at a Glance

- **Kubernetes stacks:** support for Knative, KFServing[^1] and Istio[^2] based apps
- **Testing patterns:** canary and conformance testing of app versions
- **Deployment patterns:** progressive traffic shifting and fixing traffic split during experiments
- **Advanced traffic shaping:** shadow deployments and request routing
- **Version promotion:** automatically roll-forward to the winning version or rollback to the baseline based on the outcome of the experiment
- **App config tools:** Iter8 can with Helm, Kustomize, and plain YAML/JSON app manifests
- **Metrics backend:** Prometheus

[^1]: An initial version of Iter8 for KFServing is available [here](https://github.com/iter8-tools/iter8-kfserving). An updated version is coming soon.
[^2]: An earlier version of Iter8 for Istio is available [here](https://github.com/iter8-tools/iter8). An updated version is coming soon.
