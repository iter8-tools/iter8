---
template: main.html
---

# `common/readiness`
The `common/readiness` task can be used to verify that Kubernetes resources (for example, deployments, Knative services, or Istio virtual services) that are required for the experiment are available and ready. This task is intended to be included in the start action of an experiment.

## Example
The following is an experiment snippet with a `common/readiness` task.

```yaml
...
spec:
  strategy:
    actions:
      start:
      - task: common/readiness
        with:
          # verify that the following deployments exist
          objRefs:
          - kind: Deployment
            name: hello
            namespace: default 
            # verify that the deployment is available
            waitFor: condition=available
          - kind: Deployment
            name: hello-candidate
            namespace: default
            # verify that the deployment is available
            waitFor: condition=available
  ...
  # `common/readiness` task will also inspect the versionInfo section.
  # If resources are specified as part of weightObjRef fields within versionInfo, 
  # the task will verify the existence of these resources as well.
  versionInfo:
    baseline:
      name: stable
      weightObjRef:
        apiVersion: networking.istio.io/v1beta1
        kind: VirtualService
        namespace: default
        name: hello-routing
        fieldPath: .spec.http[0].route[0].weight
    candidates:
    - name: candidate
      weightObjRef:
        apiVersion: networking.istio.io/v1beta1
        kind: VirtualService
        namespace: default
        name: hello-routing
        fieldPath: .spec.http[0].route[1].weight
```

## Inputs
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| initialDelaySeconds | int | Verification will be attempted only after this initial delay. Default value is `5`. | No |
| numRetries | int | If the task cannot verify the existence/conditions of the resources after the first attempt, it will retry with further attempts. Total number of attempts = 1 + numRetries. Default value for `numRetries` is `12`. | No |
| intervalSeconds | int | Time interval between each attempt. Default value is `5`. | No |
| objRefs | [][ObjRef](#objref) | A list of Kubernetes object references along with any associated conditions which need to be verified. | No |

### ObjRef
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| kind | string | Kind of the Kubernetes resource. Specified in the TYPE[.VERSION][.GROUP] format used by the [`kubectl get` command](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#get) | Yes |
| namespace | string | Namespace of the Kubernetes resource. Default value is the namespace of the experiment resource. | No |
| name | string | Name of the Kubernetes resource. | Yes |
| waitFor | string | Any value that is accepted by the `--for` flag of the [`kubectl wait` command](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#wait). | No |


## Result
The task will succeed if all the specified resources are verified to exist (along with any associated conditions) within the specified number of attempts. Otherwise, the task will fail, resulting in the failure of the experiment.