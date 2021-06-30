---
template: main.html
---

# Dynamic placeholder substitution

Inputs to tasks may contain placeholders, or template variables, which will be dynamically substituted when the task is executed by Iter8. For example, in the task:

```bash
- task: common/bash # promote the winning version      
  with:
    script: |
        kubectl apply -f {{ .promoteManifest }}
```

`{{ .promotionManifest}}` is a placeholder.

Placeholders are specified using the Go language specification for data-driven [templates](https://golang.org/pkg/html/template/). In particular, placeholders are specified between double curly braces.

Iter8 supporsts placeholders for:

- Values of variables of the version recommended for promotion. To specify such placeholders, use the name of the variable as defined in the [`versionInfo` section](../../experiment/#versioninfo) of the experiment definition. For example, in the above example, `{{ .promotionManifest }}` is a placeholder for the value of the variable with the name `promotionManifest` of the version Iter8 recommends for promotion (see [`.status.versionRecommendedForPromotion`](../../experiment/#status)).

- Values defined in the experiment itself. To specify such placeholders, use the prefix `.this`. For example, `{{ .this.metadata.name }}` is a placeholder for the name of the experiment.

If Iter8 cannot evaluate a placeholder expression, a blank value ("") will be substituted. This may occur, for example, in an start tasks when Iter8 has not yet defined a version recommended for promotion.
