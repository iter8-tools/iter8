---
template: overrides/main.html
---

# Conceptual Building Blocks

We introduce the conceptual building blocks of an Iter8 experiment. The concepts defined in this document are used throughout Iter8 documentation.

## Version
**Version** is a variant of your app/ML model.

## Validation

A version is considered **validated** if it satisfies the given set of service-level objectives or SLOs. An example of an SLO could be as follows: the 99th-percentile tail latency of the version should be under 50 msec.

## Release

**Release** is the process by which a new version becomes responsible for serving production traffic.

## Experiment

Iter8 defines a new Kubernetes resource called **experiment** that automates releases, validation, and experiments with your app/ML model versions based on metrics. Iter8 experiments may be run in a production cluster or dev/test/staging clusters.

## Winner

The **winner** is the best version among all the versions involved in an experiment.

## Baseline

Every Iter8 experiment involves a version called the **baseline**. Typically, baseline is the latest stable version of your app/ML model that has been released.

## Candidate

**Candidate** is a version, other than the baseline, that participates in an experiment.

## Version recommended for promotion

When two or more versions participate in an experiment, Iter8 **recommends a version for promotion**; if the experiment yielded a winner, then the version recommended for promotion is the winner; otherwise, the version recommended for promotion is the baseline.

## Metrics

Metric backends like Prometheus, New Relic, Sysdig and Elastic collect metrics for deployed versions and serve them through REST APIs. Iter8 defines a new Kubernetes resource called **Metric** that makes it easy to use metrics in experiments from any RESTful metrics backend.

## Objectives

**Objectives** correspond to service-level objectives or SLOs. In Iter8 experiments, objectives are specified as metrics along with acceptable limits on their values. Iter8 will report how versions are performing with respect to these metrics and whether or not they satisfy the objectives.

## Indicators

**Indicators** correspond to service-level indicators or SLIs. In Iter8 experiments, indicators are specified as a list metrics. Iter8 will report how versions are performing with respect to these metrics.

## Testing pattern

**Testing pattern** defines the number of versions involved in the experiment (1, 2, or more), and determines how the winner is identified. Iter8 supports **canary** and **conformance** testing patterns.

=== "Canary"
    Canary testing involves two versions, the baseline and a candidate. If the candidate is validated (i.e., it satisfies objectives specified in the experiment), then candidate is the winner; else, if baseline satisfies objectives, then baseline is the winner; else, there is no winner.

    ![Canary](/assets/images/canary-progressive-kubectl.png)

    !!! tip ""
        Try a [canary experiment](/getting-started/quick-start/with-knative/).

=== "Conformance"
    Conformance testing involves a single version, a baseline. If it is validated (i.e., it satisfies objectives) then baseline is the winner; else, there is no winner.

    ![Conformance](/assets/images/conformance.png)

    !!! tip ""
        Try a [conformance experiment](/tutorials/knative/conformance/).

## Deployment pattern

**Deployment pattern** determines how traffic is split between versions. Iter8 supports **progressive** and **fixed-split** deployment patterns.

=== "Progressive"
    Progressive deployment incrementally shifts traffic towards the winner over multiple iterations.

    ![Canary](/assets/images/canary-progressive-helm.png)

    !!! tip ""
        Try a [progressive deployment experiment](/tutorials/knative/canary-progressive/).

=== "Fixed-split"
    Fixed-split deployment does not shift traffic between versions.

    ![Canary](/assets/images/canary-fixedsplit-kustomize.png)

    !!! tip ""
        Try a [fixed-split deployment experiment](/tutorials/knative/canary-fixedsplit/).

## Traffic shaping

**Traffic shaping** refers to features such as **traffic mirroring/shadowing** and **traffic segmentation** that provide fine-grained controls over how traffic is routed to and from app versions. 

Iter8 enables you to take total advantage of all the traffic shaping features available in the service mesh, ingress technology, or networking layer present in your Kubernetes cluster.

=== "Traffic mirroring/shadowing"
    **Traffic mirroring** or **shadowing** enables experimenting with a *dark* launched version with zero-impact on end-users. Mirrored traffic is a replica of the real user requests[^1] that is routed to the dark version. Metrics are collected and evaluated for the dark version, but responses from the dark version are ignored.

    ![Canary](/assets/images/mirroring.png)

    !!! tip ""
        Try an experiment with [traffic mirroring/shadowing](/tutorials/knative/mirroring/).

=== "Traffic segmentation"
    **Traffic segmentation** is the ability to carve out a specific segment of the traffic to be used in an experiment, leaving the rest of the traffic unaffected by the experiment. Service meshes and ingress controllers often provide the ability to route requests dynamically to different versions based on request attributes such as user identity, URI, IP address prefixes, or origin. Iter8 can leverage this functionality in experiments to control the segment of the traffic that will participate in the experiment. For example, in the canary experiment depicted below, requests from the country `Wakanda` may be routed to baseline or candidate; requests that are not from `Wakanda` will not participate in the experiment and are routed only to the baseline.

    ![Canary](/assets/images/request-routing.png)

    !!! tip ""
        Try an experiment with [traffic segmentation](/tutorials/knative/traffic-segmentation/).


## Version promotion

Iter8 can optionally **promote a version** at the end of an experiment, based on the [version recommended for promotion](#version-recommended-for-promotion). As part of the version promotion task, Iter8 can configure Kubernetes resources by installing or upgrading Helm charts, building and applying Kustomize resources, or using the `kubectl` CLI to apply YAML/JSON resource manifests and perform other cleanup actions such as resource deletion.

=== "Helm charts"
    An experiment that uses `helm` for version promotion is illustrated below.

    ![Canary](/assets/images/canary-progressive-helm.png)

    !!! tip ""
        Try an [experiment that uses Helm charts](/tutorials/knative/canary-progressive/).

=== "Kustomize resources"
    An experiment that uses `kustomize` for version promotion is illustrated below.

    ![Canary](/assets/images/canary-fixedsplit-kustomize.png)

    !!! tip ""
        Try an [experiment that uses Kustomize resources](/tutorials/knative/canary-fixedsplit/).

=== "Plain YAML/JSON manifests"
    An experiment that uses plain YAML/JSON manifests and the `kubectl` CLI for version promotion is illustrated below.

    ![Canary](/assets/images/canary-progressive-kubectl.png)

    !!! tip ""
        Try an [experiment that uses plain YAML/JSON manifests](/getting-started/quick-start/with-knative/).

[^1]: It is possible to mirror only a certain percentage of the requests instead of all requests.