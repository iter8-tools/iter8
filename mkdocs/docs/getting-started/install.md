---
template: main.html
title: Installation
---

# Installation

## Iter8

Install Iter8 in your Kubernetes cluster as follows. This installation requires [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

```shell
export TAG=v0.4.3
curl -s https://raw.githubusercontent.com/iter8-tools/iter8-install/main/install.sh | bash
```

## iter8ctl
The `iter8ctl` client facilitates real-time observability of Iter8 experiments. Install `iter8ctl` on your local machine as follows. This installation requires Go 1.13+.

```shell
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl
```