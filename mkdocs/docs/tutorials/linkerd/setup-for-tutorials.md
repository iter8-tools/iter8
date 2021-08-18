---
template: main.html
---

# Setup For Tutorials

## Install Linkerd
For production installation of Knative, refer to the [official Linkerd instructions](https://linkerd.io/2.10/tasks/install/). For exercising Iter8 tutorials, install Linkerd as follows.


* **Clone Iter8 repo**
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
```

* **Install Linkerd**
```shell
$ITER8/samples/linkerd/quickstart/platformsetup.sh
```

* **Enable Linkerd in the default namespace**
```shell
kubectl annotate namespace default linkerd.io/inject=enabled
```