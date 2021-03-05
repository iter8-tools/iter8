---
template: overrides/main.html
hide:
- toc
---

# Generating Requests Externally

All Iter8 tutorials use [`fortio`](https://github.com/fortio/fortio) or `curl` based Kubernetes jobs that run inside the cluster to simulate user requests during experiments. You can try other variations of these tutorials where requests are generated outside the cluster.

- The Knative service setup of [quick start](/getting-started/quick-start/with-knative/), [Canary + Progressive + Helm](/code-samples/iter8-knative/canary-progressive/), [Canary + FixedSplit + Kustomize](/code-samples/iter8-knative/canary-fixedsplit/), and [Conformance](/code-samples/iter8-knative/conformance/) tutorials is similar to that of this [Knative tutorial](https://knative.dev/docs/serving/samples/traffic-splitting/index.html#traffic-splitting).

- The Istio virtual service setup of [Conformance + Mirroring](/code-samples/iter8-knative/mirroring/) and [Canary + Request routing](/code-samples/iter8-knative/requestrouting/) tutorials is similar to that of this [Knative tutorial](https://knative.dev/docs/serving/samples/knative-routing-go/index.html#access-the-services).

Refer to the above mentioned Knative tutorials for generating requests for your applications from outside the cluster.

