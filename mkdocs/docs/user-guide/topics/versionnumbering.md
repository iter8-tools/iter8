---
template: main.html
---

# Version Numbering

Iter8 can observe metrics for multiple versions of an app during an experiment. The number of versions equals the length of the `versionInfo` input field of the following task.

* [`gen-load-and-collect-metrics`](../tasks/collect.md)

If there are `n` versions, they are numbered `0, ..., n-1`.