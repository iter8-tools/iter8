
---
date: 2020-08-06T12:00:00+00:00
menuTitle: iter8-kui
title: Integrating Iter8 with KUI
weight: 20
summary: Describes iter8's integrations with KUI
---

# Integrating Iter8 with KUI


[KUI](https://kui.tools) combines the power of familiar CLIs with visualizations in high-impact areas. Kui enables you to manipulate complex JSON and YAML data models, integrate disparate tooling, and provides quick access to aggregate views of operational data.
We are integrating our work with KUI to enable user intuitive and interactive options for the users running iter8. KUI has allowed us to implement Humand-In-The-Loop iter8 experiments and easily modifiable and customizable metrics as you will see below.


## Installation

The following steps builds the plugin directly from the repo.

```sh
$ git clone https://github.com/iter8-tools/iter8-kui.git
$ cd kui/
$ npm ci
```

To run the KUI Terminal, run the following command:

```sh
$ npm start
```


## Currently available commands

We have implemented the following commands that can be used once the KUI terminal is up and running:
1. `iter8 metrics`: This command opens a KUI sidecar where the user can perform CRUD operations on the iter8 metric configmap. Specifically, users can add, edit, delete and restore metrics on the KUI sidecar that is opened. A sample image of the output is as follows:

  [Iter8 KUI Metric]({{< resourceAbsUrl path="images/iter8-kui-metric.png" >}})

  Delete and restore operations can be performed on the same page by the _trashcan_ icon. Users can add a Counter or a Ratio Metric by clicking on the _+_ icon adjacent to the metric titles. This opens up a form as follows:

  [Iter8 KUI Add Metric]({{< resourceAbsUrl path="images/iter8-kui-add-metric.png" >}})

  Once the form is filled, the user can create the new metric and see it listed in the original page.

  To edit any of the currently available metrics, users can click on the _edit_ icon for that metric. This also opens a form that is filled with the values currently held by that metric as follows. Note that standard iter8 metrics such as _iter8_mean_latency_, _iter8_error_count_, etc cannot be edited as they come out of the box with iter8.

  [Iter8 KUI Edit Metric]({{< resourceAbsUrl path="images/iter8-kui-edit-metric.png" >}})

2. `iter8 create experiment`: This command also opens a KUI sidecar and is used to create Human-In-The-Loop experiments with iter8. This command opens a sidecar with two tabs- one for creating the experiment and one for viewing the decision and metrics for the experiment from _iter8-analytics_. The sidecar options are interactive and can be experimented with by the user.


3. `iter8 about`: _Coming soon_
