---
template: main.html
---

# `run`
The `run` task executes a bash script.

## Examples
Send a Slack notification.
```yaml
- run: |
    curl -d "text=Experiment is complete. New version is promoted." -d "channel=C123456" \
    -H "Authorization: Bearer xoxb-not-a-real-token-this-will-not-work" \
    -X POST https://slack.com/api/chat.postMessage  
```

Trigger a GitHub Actions workflow.
```yaml
- run: |
    echo xoxb-not-a-real-token-this-will-not-work > token.txt
    gh auth login --with-token < token.txt
    gh repo clone my-repo
    cd my-repo
    gh workflow run promote.yaml -R github.com/me/my-repo
```

Run a `kubectl` command.
```yaml
- run: |
    kubectl apply -f new-version-of-my-app.yaml -n my-app-namespace
```

Assess app versions. `If` SLOs are `not` satisfied by version numbered 1, rollback. This is an example of [conditional task execution](../conditional.md).
```yaml
- task: assess-app-versions
  ...
- if: not SLOsBy(1)
  run: |
    kubectl rollout undo deployment/my-app-deployment
```

## Temp dir

The script in `run` can have environment variables. One such pre-defined variable is `$TEMP_DIR` which points to the default directory to use for temporary files.

```yaml
- run: |
    cd $TEMP_DIR
    echo "hello" > world.txt
```

## Available commands

When running experiments on your local machine, any command that is available in your `PATH` can be used as part of the `run` task. When running experiments in Kubernetes, in addition to the `iter8` command, the Iter8 container also includes `kubectl`, `kustomize`, `helm`, `yq`, `git`, `curl`, and `gh`, all of which can be used as part of the `run` task.

```yaml
- run: |
    kustomize build hello/world/folder > manifest.yaml
    kubectl apply -f manifest.yaml
    helm upgrade my-app helm/chart --install
    yq -i a=b manifest.yaml
    git clone https://github.com/iter8-tools/iter8.git
    gh pr create
    curl https://iter8.tools -O $SCARCH_DIR/i.html
```
