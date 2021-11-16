---
template: main.html
---

# Conditional Execution
The execution of a task within an experiment can be made conditional based on an `if` clause. The task is executed if and only if the clause evaluates to true, and is skipped otherwise. A few illustrative examples of conditional execution are shown below.

## Illustrative Examples

Assess app versions. `If` SLOs are `not` satisfied, send a Slack notification.
```yaml
- task: assess-app-versions
  ...
- if: not SLOs()
  run: |
    curl -d "text=SLOs are not satisfied by your app" -d "channel=C123456" \
    -H "Authorization: Bearer xoxb-not-a-real-token-this-will-not-work" \
    -X POST https://slack.com/api/chat.postMessage  
```

Assess app versions. `If` SLOs are satisfied, trigger a GitHub Actions workflow.
```yaml
- task: assess-app-versions
  ...
- if: SLOs()
  run: |
    echo xoxb-not-a-real-token-this-will-not-work > token.txt
    gh auth login --with-token < token.txt
    gh repo clone my-repo
    cd my-repo
    gh workflow run promote.yaml -R github.com/me/my-repo
```

Assess app versions. `If` SLOs are satisfied by version numbered 1, promote it.
```yaml
- task: assess-app-versions
  ...
- if: SLOsBy(1)
  run: |
    kubectl apply -f new-version-of-my-app.yaml -n my-app-namespace
```

Assess app versions. `If` SLOs are `not` satisfied by version numbered 1, rollback.
```yaml
- task: assess-app-versions
  ...
- if: not SLOsBy(1)
  run: |
    kubectl rollout undo deployment/my-app-deployment
```

## Syntax

The following conditions are supported within the `if` clause.

* `SLOs()` -- this returns true if all the app versions satisfy SLOs.
* `SLOsBy(i)` -- this returns true if the [version with number](versionnumbering.md) `i` satisfies SLOs.

Negations, logical `and`, and `or` are also supported.

* `not SLOs()`
* `not SLOsBy(i)`
* `SLOsBy(0) and not SLOsBy(1)`
* `SLOsBy(0) or SLOsBy(1)`

