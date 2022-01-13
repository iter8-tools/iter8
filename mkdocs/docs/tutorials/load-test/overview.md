---
template: main.html
---

# Load Testing with SLOs

<!-- ![Load testing with SLOs]({% include ".icons/material/atom.svg" %}){ align=left } -->
 
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

