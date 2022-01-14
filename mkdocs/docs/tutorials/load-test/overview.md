---
template: main.html
---

# Load Testing with SLOs: Overview
 
> Load testing experiments generate requests for HTTP and gRPC services, collect built-in latency and error-related metrics and validate SLOs. 

***

## Examples

[Load test an HTTP service and validate SLOs](../../getting-started/your-first-experiment.md).

: Use an Iter8 experiment to load test an HTTP service and validate latency and error-related service level objectives (SLOs).

[Control the request generation process during a load test](requests.md).
: While running a load test, you can set the total number of requests/the duration of the load test, the number of requests sent per second, and the number of parallel connections used to send requests. This provides fine-grained control over the request generation process.

[HTTP POST endpoint accepting payload](payload.md).
: While load testing an HTTP service with a POST endpoint, you may send any type of content as payload during the load test.

[Specify error codes, latency percentiles and SLOs](percentilesandslos.md).
: While running a load test, you can specify the range of HTTP status codes that are considered as errors, the latency percentile values that are computed and reported, and the SLOs that are evaluated.

## Community examples

These samples are contributed and maintained by members of the Iter8 community.

!!! tip "Dear Iter8 community" 

    Community examples may become outdated. If you find that something is not working, lend a helping hand and fix it in a PR. More examples are very welcome. Please submit a PR for yours.

### Knative

[Load test a Knative HTTP service](community/knative/loadtest.md)
: Use an Iter8 experiment to load test a Knative HTTP service and validate latency and error-related service level objectives (SLOs).