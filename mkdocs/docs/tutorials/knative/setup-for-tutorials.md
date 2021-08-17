---
template: main.html
---

# Setup For Tutorials

For production installation of Knative, refer to the [official Knative instructions](https://knative.dev/docs/install/). Iter8 can work with any Knative networking layer. For simplicity, we recommend Kourier as the Knative networking layer for Iter8 tutorials. You can install Knative-serving in your cluster with Kourier networking as follows.

```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
$ITER8/samples/knative/quickstart/platform-setup.sh
```
