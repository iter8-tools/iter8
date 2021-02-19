---
template: overrides/main.html
---

# spec.strategy.actions

An action is a set of tasks that can be run by iter8. `spec.strategy.actions` is an object that can be used to specify `start` and `finish` actions that will be run at the start and end of an experiment respectively.

The sample experiment above consists of start and finish actions. The start action consists of a single task, namely `init-experiment`. This task verifies that the targeted Knative service is available and ready, and populates the experiment resource with details about the app versions such as JSON paths used for specifying traffic percentages. The finish action consists of a single task, namely `exec`. This task applies the Knative service manifest corresponding to the version to be promoted at the end of the experiment. Assuming that the candidate satisfies the experiment objectives, its manifest will be applied.

??? note "Links to in-depth description and code samples"
    1. In-depth description of deployment patterns is [here](aspects/actions.md).
    2. Code samples...
