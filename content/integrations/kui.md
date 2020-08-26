---
menuTitle: Kui
title: Kui
weight: 10
summary: Describes iter8's integrations with Kui
---

[Kui](https://kui.tools) combines the power of familiar CLIs with interactive visualizations to provide an elegant and intuitive experience for users of Kubernetes.

The iter8-Kui integration aims to leverage these features for iter8. Using the iter8 plugin on Kui, you can run Human-In-The-Loop iter8 experiments and easily modify and customize experiment metrics.

## Installation

Follow these steps to build and run Kui:

```sh
git clone https://github.com/IBM/kui
cd kui/
npm ci
```

To run the Kui Terminal, use:

```sh
npm start
```

To install iter8, refer to the [iter8 installation guidelines]({{< ref "kubernetes" >}}).

## Currently available commands

You can use the following commands once the Kui terminal is up and iter8 has been installed:

#### `iter8 metrics`

This command opens a Kui sidecar where the you can perform CRUD operations on the iter8 metric configmap. Specifically, you can add, edit, delete and restore metrics on the Kui sidecar that is opened. A sample image of the output is as follows:

![iter8 Kui metrics]({{< resourceAbsUrl path="images/iter8-kui-metric.png" >}})

Delete and restore operations can be performed on the same page using the _trashcan_ icon. You can add a Counter or a Ratio Metric by clicking on the _+_ icon adjacent to the metric titles. This opens up a form as follows:

![iter8 Kui add metric]({{< resourceAbsUrl path="images/iter8-kui-add-metric.png" >}})

Once the form is filled, you can create the new metric and see it listed in the original page.

To edit any of the currently available metrics, you can click on the _edit_ icon for that metric. This also opens a form that is pre-filled with the values currently held by that metric as in the following image. Note that standard iter8 metrics such as *iter8_mean_latency*, *iter8_error_count*, etc cannot be edited as they come out-of-the-box with iter8.

![iter8 Kui edit metric]({{< resourceAbsUrl path="images/iter8-kui-edit-metric.png" >}})

#### `iter8 create experiment`

This command also opens a Kui sidecar and is used to create Human-In-The-Loop experiments with iter8. It opens two tabs- one for creating the experiment and one for viewing the decision and metrics for the experiment from *iter8-analytics*. The sidecar options are interactive and can be experimented with according to your preferences.

_Video coming soon._

To run this command, iter8 requires you to export a URL to access the *iter8-analytics* service, as an environment variable. To do this, you may have to expose the iter8 analytics service to a `NodePort` first:

```sh
kubectl expose svc iter8-analytics -n iter8 --name=iter8-analytics-np --type=NodePort
export ITER8_ANALYTICS_URL='<insert-iter8-analytics-url>'
```

#### `iter8 about`

This command gives you an overview of the components of iter8, a list of commands available and also directs you to the documentation website, Github repository and Slack channel.

![iter8 Kui about]({{< resourceAbsUrl path="images/iter8-kui-about.png" >}})

#### `iter8 config verify`

This command verifies if _iter8-analytics_ and _iter8-controller_ service is currently installed in the user's environment. it returns a _true_ or _false_ accordingly.
