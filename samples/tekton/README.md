# Iter8-in-Docker-in-Tekton

Iter8 experiments can be performed for remote applications. In particular, Iter8 can be run in a local cluster while the application/ML model resides in a remote K8s cluster. In these situations, it is desirable to run Iter8 within a local cluster like Kind, Minikube, or CodeReady Containers which itself is running within a Docker container.

[This Dockerfile](https://github.com/iter8-tools/iter8/blob/master/install/docker/README.md) creates an Iter8-in-Docker (`iter8/ind`) image for the above scenario.

***

> The samples in this folder show how to use the `iter8/ind` image within Tekton. 

***

## Running ind-in-Tekton

1. [Install Tekton in K8s and setup Tekton CLI](https://tekton.dev/docs/getting-started/#prerequisites)

2. Clone the Iter8 repo locally.
```shell
export ITER8=<root-of-your-local-Iter8-repo>/samples/tekton
```

3. Run your first ind-in-Tekton workflow
```shell
kubectl apply -f $ITER8/task-iter8.yaml
```

Start the above task with `tkn`.

```shell
tkn task start iter8
```

Tekton will now start running your Task. To see the logs of the last TaskRun, run the following `tkn` command:

```shell
tkn taskrun logs --last -f 
```

It will take a couple of minutes for your Task to complete. When it executes, it should show output that resembles the following:

```shell
[iter8] + /iter8/iter8.sh
[iter8] Hello from Iter8!
[iter8] Docker in Docker is running...
[iter8] Creating cluster "kind" ...
[iter8]  • Ensuring node image (kindest/node:v1.21.1) 🖼  ...
[iter8]  ✓ Ensuring node image (kindest/node:v1.21.1) 🖼
[iter8]  • Preparing nodes 📦   ...
[iter8]  ✓ Preparing nodes 📦 
[iter8]  • Writing configuration 📜  ...
[iter8]  ✓ Writing configuration 📜
[iter8]  • Starting control-plane 🕹️  ...
[iter8]  ✓ Starting control-plane 🕹️
[iter8]  • Installing CNI 🔌  ...
[iter8]  ✓ Installing CNI 🔌
[iter8]  • Installing StorageClass 💾  ...
[iter8]  ✓ Installing StorageClass 💾
[iter8]  • Waiting ≤ 5m0s for control-plane = Ready ⏳  ...
[iter8]  ✓ Waiting ≤ 5m0s for control-plane = Ready ⏳
[iter8]  • Ready after 29s 💚
[iter8] Set kubectl context to "kind-kind"
[iter8] You can now use your cluster with:
[iter8] 
[iter8] kubectl cluster-info --context kind-kind
[iter8] 
[iter8] Have a question, bug, or feature request? Let us know! https://iter8.tools 🙂
[iter8] Kubernetes control plane is running at https://127.0.0.1:33979
[iter8] CoreDNS is running at https://127.0.0.1:33979/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
[iter8] 
[iter8] To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
[iter8] Installing Iter8...
[iter8] namespace/iter8-system created
[iter8] customresourcedefinition.apiextensions.k8s.io/experiments.iter8.tools created
[iter8] customresourcedefinition.apiextensions.k8s.io/metrics.iter8.tools created
[iter8] serviceaccount/iter8-analytics created
[iter8] serviceaccount/iter8-controller created
[iter8] serviceaccount/iter8-handlers created
[iter8] role.rbac.authorization.k8s.io/iter8-leader-election-role created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-experiments created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-istio-dr created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-istio-vs created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-isvc-for-kfs created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-jobs created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-ksvc-for-kn created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-metrics created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-sdep-for-seldon created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-vs-for-kfs created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-vs-for-kn created
[iter8] clusterrole.rbac.authorization.k8s.io/iter8-vs-for-seldon created
[iter8] rolebinding.rbac.authorization.k8s.io/iter8-leader-election-rolebinding created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-experiments created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-istio-dr created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-istio-vs created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-isvc-for-kfs created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-jobs created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-ksvc-for-kn created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-metrics created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-sdep-for-seldon created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-vs-for-kfs created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-vs-for-kn created
[iter8] clusterrolebinding.rbac.authorization.k8s.io/iter8-vs-for-seldon created
[iter8] configmap/iter8-handlers-bt9hh9fmtm created
[iter8] service/iter8-analytics created
[iter8] deployment.apps/iter8-analytics created
[iter8] deployment.apps/iter8-controller-manager created
[iter8] timed out waiting for the condition on pods/iter8-analytics-7f7dc6d9c-skwcb
[iter8] timed out waiting for the condition on pods/iter8-controller-manager-6975994cfb-pvw8p
[iter8] Iter8 in Docker is running...
[iter8] All systems go.
```


