---
template: main.html
---

# Tasks

Tasks are an extension mechanism for enhancing the behavior of Iter8 experiments and can be specified within the [spec.strategy.actions](../experiment/#strategy) field of the experiment.

Tasks are grouped into libraries. The following task libraries are available.

- `common` [library](common/#common-tasks)
    * Task for executing a shell command.
- `metrics` [library](metrics/#metrics-tasks)
    * Task for collecting builtin metrics.
    * The above task can also be used to generate requests for app/ML model versions without collecting builtin metrics.
- `notification` [library](notification/#notification-tasks)
    * Task for sending a slack notification.
