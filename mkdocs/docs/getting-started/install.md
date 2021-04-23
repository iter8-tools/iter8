---
template: main.html
title: Installation
---

# Installation

## Iter8

Install Iter8 in your Kubernetes cluster as follows. This installation requires [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

```shell
export TAG=v0.4.1
curl -s https://raw.githubusercontent.com/iter8-tools/iter8-install/main/install.sh | bash
```

## iter8ctl
`iter8ctl` is a client-side CLI that facilitates real-time observability of Iter8 experiments. Install `iter8ctl` locally as follows. This installation requires Go 1.13+.

```shell
GO111MODULE=on GOBIN=/usr/local/bin go get github.com/iter8-tools/iter8ctl
```