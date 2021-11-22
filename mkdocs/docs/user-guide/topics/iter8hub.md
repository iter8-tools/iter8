---
template: main.html
---

# Iter8 Hub

Iter8 hub is a Git folder which hosts Iter8 experiment folders underneath it. The public Iter8 hub is located at `github.com/iter8-tools/iter8//mkdocs/docs/hub`.

Experiment folders from the hub can be downloaded using the `iter8 hub` command.

```shell
# download load-test experiment folder from the public Iter8 Hub
iter8 hub -e load-test
```

## Custom Iter8 hub

It is easy to create and use your own custom Iter8 hub.

```shell
# Suppose you forked github.com/iter8-tools/iter8 under 
# the GitHub org or account named $GHUSER,
# created a branch called 'ml', and pushed a new experiment folder 
# called 'tensorflow' under the path 'mkdocs/docs/hub'. 
# It can now be downloaded as follows.

export ITER8HUB=github.com/$GHUSER/iter8.git?ref=ml//mkdocs/docs/hub/
iter8 hub -e tensorflow
```

See the [iter8 hub](../commands/iter8_hub.md) command for the syntax of the `ITER8HUB` environment variable used above.