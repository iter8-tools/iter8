---
template: main.html
---

# Add a K8s Stack / Service Mesh / Ingress

Performing Iter8 experiments requires RBAC rules, which are contained in [this Kustomize folder](https://github.com/iter8-tools/iter8/tree/master/install/core/rbac/stacks) and are installed as part of the Iter8 installation.

Enable Iter8 experiments over a new K8s stack by extending these RBAC rules.

## Step 0: Fork Iter8
Fork the [Iter8 GitHub repo](https://github.com/iter8-tools/iter8). Locally clone your forked repo.

For the rest of this document, `$ITER8` will refer to the root of your local Iter8 repo.

## Step 1: Edit `kustomization.yaml`
```shell
cd $ITER8/install/core/rbac/stacks
```

Edit `kustomization.yaml` to add your K8s stack. At the time of writing, it contains the following stacks:
```yaml
resources:
- iter8-knative
- iter8-istio
- iter8-kfserving
# -iter8-<your stack> # add your stack here
```

## Step 2: Create subfolder
```shell
mkdir iter8-<your stack>
cp iter8-kfserving/kustomization.yaml iter8-<your stack>/kustomization.yaml
```

## Step 3: Create RBAC rules
```shell
cd iter8-<your stack>
```

=== "Foo & Istio virtual service example"
    Suppose Iter8 experiments on your stack involves manipulation of two types of resources:

    1. The `foo` resource belonging to the API group `bar.my.org`.
    2. The Istio virtual service resource.

    > **Note:** `Foo` and `bar` are merely placeholders. It can be replaced by any standard K8s resource type like `deployment` or `service`, or a custom resource type, as required.
    
    Create RBAC rules that will enable Iter8 to manipulate `foo` resources and Istio virtual service resources during experiments. You can do so by creating `roles.yaml` and `rolebindings.yaml` files as follows.


    **roles.yaml**
    ```yaml
    # This cluster role enables manipulation of foo resources
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: foo-for-<your stack>
    rules:
    - apiGroups:
      - bar.my.org
      resources:
      - foo
      verbs:
      - get
      - list
      - patch
      - update
      - create
      - delete
      - watch
    ---
    # This cluster role enables manipulation of Istio virtual services
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: vs-for-<your stack>
    rules:
    - apiGroups:
      - networking.istio.io
      resources:
      - virtualservices
      verbs:
      - get
      - list
      - patch
      - update
      - create
      - delete
      - watch
    ```    

    **rolebindings.yaml**
    ```yaml
    # This cluster role binding enables Iter8 controller and task runner to manipulate 
    # foo resources in any namespace
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: foo-for-<your stack>
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: foo-for-<your stack>
    subjects:
    - kind: ServiceAccount
      name: controller
    - kind: ServiceAccount
      name: handlers
    ---
    # This role binding enables Iter8 controller and handler to manipulate 
    # Istio virtual services in any namespace
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: vs-for-<your stack>
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: vs-for-<your stack>
    subjects:
    - kind: ServiceAccount
      name: controller
    - kind: ServiceAccount
      name: handlers
    ```

    You can also refer to the [KFServing](https://github.com/iter8-tools/iter8/tree/master/install/core/rbac/stacks/iter8-kfserving), [Knative](https://github.com/iter8-tools/iter8/tree/master/install/core/rbac/stacks/iter8-knative), and [Istio](https://github.com/iter8-tools/iter8/tree/master/install/core/rbac/stacks/iter8-istio) examples.

## Step 4: Update RBAC rules
Update the [RBAC rules that are applied to the Kubernetes cluster](../../getting-started/install/#rbac-rules) as part of the Iter8 installation.

## Step 5: Submit PR
[Sign your commit](../overview/#sign-your-commits) and submit your pull request to the Iter8 repo.