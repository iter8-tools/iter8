---
template: main.html
---

# Service Level Objectives (SLOs)

Service level objectives (SLOs) are metrics along with acceptable limits on their values. Iter8 will report how the app/ML model is performing with respect to these metrics and whether or not they satisfy the SLOs.

???+ example "Examples"
    * The 99th-percentile tail latency of the application should be under 50 msec.
    * The precision of the ML model version should be over 92%.
    * The (average) number of GPU cores consumed by a model should be under 5.0

SLOs are specified as part of the [`assess-app-versions` task](../tasks/assess.md).

