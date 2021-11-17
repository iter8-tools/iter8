---
template: main.html
---

# Load Testing

Load testing experiments can generate requests for HTTP and gRPC services, collect built-in latency and error related metrics and validate SLOs. The [`gen-load-and-collect-metrics`](../user-guide/tasks/collect.md) task generates load and collects built-in metrics. The [`assess-app-versions`](../user-guide/tasks/assess.md) task validates SLOs.

***

## v0.8 Examples

* Load testing and SLO validation of an app with an [HTTP GET API endpoint](../getting-started/your-first-experiment.md).

***

## v0.7 Examples

* Load testing and SLO validation of an app that receives a payload through an [HTTP POST API endpoint](https://iter8.tools/0.7/tutorials/deployments/slo-validation-payload/).

* Load testing and SLO validation of a [Knative app managed by Helm](https://iter8.tools/0.7/tutorials/knative/slovalidation-helmex/).