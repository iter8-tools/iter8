---
template: main.html
---

# Iter8 Hub

Iter8 hub is a GitHub location containing Iter8 experiment folders. The public Iter8 hub is located at `github.com/iter8-tools/iter8//mkdocs/docs/hub`. The `iter8 hub` command can be used to download experiment folders from the hub.

```shell
# download load-test experiment folder from the public Iter8 Hub
iter8 hub -e load-test
```

## Custom Iter8 hub

It is easy to create and use your own custom Iter8 hubs.

```shell
# Suppose you forked github.com/iter8-tools/iter8 under 
# the GitHub account named $GHUSER,
# created a branch called 'ml', and pushed an Iter8 experiment folder 
# called 'tensorflow' under the path 'mkdocs/docs/hub'. 
# Anyone with read access to this repo can download `tensorflow` as follows.

export ITER8HUB=github.com/$GHUSER/iter8.git?ref=ml//hub
iter8 hub -e tensorflow
```

See the [iter8 hub](../commands/iter8_hub.md) command for the syntax of the `ITER8HUB` environment variable used above.