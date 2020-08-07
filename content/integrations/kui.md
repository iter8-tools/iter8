
---
date: 2020-08-06T12:00:00+00:00
menuTitle: Kui
title: Integrating Iter8 with KUI
weight: 10
summary: Describes iter8's integrations with KUI
---

[Kui](https://kui.tools) combines the power of familiar CLIs with interactive visualizations to provide an elegant and intuitive experience for users of kubernetes.

The iter8-KUI integration aims to leverage these features for iter8. Using the iter8 plugin on KUI, you can run Human-In-The-Loop iter8 experiments and easily modify and customize experiment metrics.

## Installation

Follow these steps to build the iter8-KUI plugin:

```sh
git clone https://github.com/iter8-tools/iter8-kui.git
cd kui/
npm ci
```

To run the KUI Terminal, use:

```sh
npm start
```

## Currently available commands

You can use the following commands once the KUI terminal is up and running:
1. `iter8 metrics`: This command opens a KUI sidecar where the you can perform CRUD operations on the iter8 metric configmap. Specifically, you can add, edit, delete and restore metrics on the KUI sidecar that is opened. A sample image of the output is as follows:

![Iter8 Kui metrics]({{< resourceAbsUrl path="images/iter8-kui-metric.png" >}})

Delete and restore operations can be performed on the same page using the _trashcan_ icon. You can add a Counter or a Ratio Metric by clicking on the _+_ icon adjacent to the metric titles. This opens up a form as follows:

![Iter8 Kui add metric]({{< resourceAbsUrl path="images/iter8-kui-add-metric.png" >}})

Once the form is filled, you can create the new metric and see it listed in the original page.

To edit any of the currently available metrics, you can click on the _edit_ icon for that metric. This also opens a form that is pre-filled with the values currently held by that metric as in the following image. Note that standard iter8 metrics such as _iter8_mean_latency_, _iter8_error_count_, etc cannot be edited as they come out-of-the-box with iter8.

![Iter8 Kui edit metric]({{< resourceAbsUrl path="images/iter8-kui-edit-metric.png" >}})

2. `iter8 create experiment`: This command also opens a KUI sidecar and is used to create Human-In-The-Loop experiments with iter8. This command opens a sidecar with two tabs- one for creating the experiment and one for viewing the decision and metrics for the experiment from _iter8-analytics_. The sidecar options are interactive and can be experimented with according to your preferences.

To run this command, iter8 requires you to export a URL to access the _iter8_analytics_ service, as an environment variable. To do this, you may have to expose the iter8 analytics service to a NodePort first:

```sh
$ kubectl expose svc iter8-analytics -n iter8 --name=iter8-analytics-np --type=NodePort
$ export ITER8_ANALYTICS_URL='<insert-iter8-analytics-url>'
```

3. `iter8 about`: _Coming soon_
