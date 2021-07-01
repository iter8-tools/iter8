---
template: main.html
---

# Tasks

Tasks are an extension mechanism for enhancing the behavior of Iter8 experiments and can be specified within the [spec.strategy.actions](experiment.md#strategy) field of the experiment.

## Task Libraries

Tasks are grouped into libraries. The following task libraries are available.

- `common` [library](tasks/common.md#common-tasks)
    * Tasks that have wide applicability such as executing shell commands.
- `metrics` [library](tasks/metrics.md#metrics-tasks)
    * Task for collecting builtin metrics.
    * The above task can also be used to generate requests for app/ML model versions without collecting builtin metrics.
- `notification` [library](tasks/notification.md#notification-tasks)
    * Task for sending a Slack notification.

## Dynamic Variable Substitution

Inputs to tasks may contain placeholders, or template variables, which will be dynamically substituted when the task is executed by Iter8. For example, in the task:

```bash
- task: common/bash # promote the winning version      
  with:
    script: |
        kubectl apply -f {{ .promoteManifest }}
```

`{{ .promotionManifest}}` is a placeholder.

Placeholders are specified using the Go language specification for data-driven [templates](https://golang.org/pkg/html/template/). In particular, placeholders are specified between double curly braces.

Iter8 supports placeholders for:

- Values of variables of the version recommended for promotion. To specify such placeholders, use the name of the variable as defined in the [`versionInfo` section](experiment.md#versioninfo) of the experiment definition. For example, in the above example, `{{ .promotionManifest }}` is a placeholder for the value of the variable with the name `promotionManifest` of the version Iter8 recommends for promotion (see [`.status.versionRecommendedForPromotion`](experiment.md#status)).

- Values defined in the experiment itself. To specify such placeholders, use the prefix `.this`. For example, `{{ .this.metadata.name }}` is a placeholder for the name of the experiment.

If Iter8 cannot evaluate a placeholder expression, a blank value ("") will be substituted. This may occur, for example, in an start tasks when Iter8 has not yet defined a version recommended for promotion.
