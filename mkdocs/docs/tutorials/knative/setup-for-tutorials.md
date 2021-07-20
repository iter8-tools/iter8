---
template: main.html
---

# Setup For Tutorials

## Install Knative
For production installation of Knative, refer to the [official Knative instructions](https://knative.dev/docs/install/). For exercising Iter8 tutorials, install Knative as follows.


* **Clone Iter8 repo**
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
```

* **Install Knative serving**
Unless Istio is especially mentioned as a requirement, we recommend Kourier as the Knative networking layer for Iter8 tutorials.

    === "Kourier"

        ```shell
        $ITER8/samples/knative/quickstart/platform-setup.sh kourier
        ```

    === "Istio"

        ```shell
        $ITER8/samples/knative/quickstart/platform-setup.sh istio
        ```
