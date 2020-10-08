---
menuTitle: Reusing VirtualServices
title: Reusing VirtualServices
weight: 30
summary: Learn how to resuse an existing VirtualService
---
To route a percentage of traffic between the baseline and candidate versions of a service, iter8 creates and modifies an Istio `VirtualSystem` and a set of `DestinationRules` (one for each of the baseline and canidate versions).
In some cases, a `VirtualSystem` may already exist.
Edge services will typically require a `VirtualService` to exists to route traffic to the service.
For example, when the bookinfo application (used in the [Canary]({{< ref "tutorials/canary.md" >}}) and [A/B/n]({{< ref "tutorials/abn.md" >}}) tutorials) is deployed, a `VirtualService` is created to route traffic to the *productpage* microservice.

Iter8 can reuse an existing `VirtualService` if it can identify the apporopriate one to use.
In general, iter8 cannot do so without assistance; to identify it to iter8, add the following labels to the `VirtualSystem`:

```yaml
iter8-tools/router: IDENTIFIER
iter8-tools/role: stable
```

If *IDENTIFIER* is of the form *\<service>.\<namespace>.svc.cluster.local*, nothing further is needed.
If it is not, add it to the `Experiment`:

```yaml
  networking:
    id: IDENTIFIER
```

For an example, see the [`VirtualService`]({{< resourceAbsUrl path="tutorials/bookinfo-gateway.yaml" >}}) and [`Experiment`]({{< resourceAbsUrl path="tutorials/abn-tutorial/abn_productpage_v1v2v3.yaml" >}}) used in the [A/B/n]({{< ref "tutorials/abn.md" >}}) tutorial.
