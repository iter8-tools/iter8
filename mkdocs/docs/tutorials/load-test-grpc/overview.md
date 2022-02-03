---
template: main.html
---

# Overview

!!! tip "Load Testing and SLO Validation for gRPC Services"
    Iter8's gRPC load testing and SLO validation experiments can generate requests for gRPC services, collect built-in latency and error-related metrics, and validate service-level objectives (SLOs). This experiment can work with all four kinds of gRPC service methods, namely, unary, server-streaming, client-streaming, and bidirectional streaming.

    **Use-case:** Continuous delivery (CD) of gRPC services is a motivating use-case for this experiment. If the gRPC service satisfies the SLOs specified in the experiment, it may be safely rolled out (for example, from a test environment to a production environment).
    
    This experiment is illustrated in the figure below.

    ![HTTP load test with SLOs](images/grpc-overview.png)

***

## Examples

[Load test a unary gRPC service and validate SLOs](unary.md).

: Use an Iter8 experiment to load test a unary gRPC service and validate latency and error-related service level objectives (SLOs).

## Community examples

These samples are contributed and maintained by members of the Iter8 community.

!!! tip "Dear Iter8 community" 

    Community examples may become outdated. If you find that something is not working, lend a helping hand and fix it in a PR. More examples are always welcome.

***

### Knative

[Load test a Knative gRPC service](community/knative/loadtest.md)
: Use an Iter8 experiment to load test a Knative gRPC service and validate latency and error-related service level objectives (SLOs).