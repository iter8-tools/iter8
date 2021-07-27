---
template: main.html
---

# Setup For Istio Tutorials

## Clone **iter8** repository
```shell
git clone https://github.com/iter8-tools/iter8.git
cd iter8
export ITER8=$(pwd)
```

## Install Istio

For production installation of Istio, refer to the [official Istio instructions](https://istio.io/latest/docs/setup/getting-started/). For exercising Iter8 tutorials, install Istio as follows.

If not already cloned, [clone the iter8 repositiory](#clone-iter8-repository).

```shell
$ITER8/samples/istio/quickstart/istio-setup.sh
```

## Install Argo CD

If not already cloned, [clone the iter8 repositiory](#clone-iter8-repository).

```shell
$ITER8/samples/istio/gitops/argocd-setup.sh
```

The output from the install script will provide instructions on how to access the Argo CD UI to setup your Argo CD app. Take those steps now. After logging in, you should see Argo CD showing no application is currently installed.

## Create GitHub Token

Login to [GitHub](https://github.com). From the upper right corner of the page, go to Settings > Developer settings > Personal access token > Generate new token. Make sure the token is granted access for `repo` and `workflow`.

Save the generated token to a Kubernetes secret as follows:

```shell
kubectl create secret generic iter8-token --from-literal=token=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Finally, give Iter8 permission to read the secret:

```shell
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/tasks/rbac/read-secrets.yaml
```