---
template: main.html
---

# Tasks
Tasks are an extension mechanism for enhancing the behavior of Iter8 experiments and can be specified within the [spec.strategy.actions](../experiment.md#strategy) field of the experiment. Tasks are grouped into libraries, namely, `common`, `metrics` and `notification`. They are referenced in Iter8 experiments using the `libraryName/taskName` format. The following tasks are available for use in Iter8 experiments.

- [`common/readiness`](common-readiness.md): Check if K8s resources needed for an experiment are available and/or ready.
- [`common/bash`](common-bash.md): Execute a bash script.
- [`common/exec`](common-readiness.md) (deprecated): Execute a shell command.
- [`metrics/collect`](metrics-collect.md): Generate GET/POST requests for the application versions and collect latency and error metrics. This task supports Iter8's builtin metrics feature.
- [`notification/slack`](notification-slack.md): Send a slack notification with a summary of the experiment.

## Actions
Tasks specified within start, finish, or loop actions with be executed by Iter8 at the start of an experiment, end of an experiment, or periodically during each loop of the experiment respectively. See [spec.strategy.actions](../experiment.md#strategy) for details about experiment actions.
