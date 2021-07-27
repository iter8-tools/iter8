---
template: main.html
---

# `gitops/helmex-update`
The `github/helmex-update` task can be used to update Helm values file in a GitHub repo. This task requires the Helm values file to conform to the [Helmex schema](../helmex-schema.md). This task is intended to be included in the finish action of an experiment.

## Example
The following is an experiment snippet with a `gitops/helmex-update` task.

```yaml
...
spec:
  strategy:
    actions:
      finish:
      - task: gitops/helmex-update
        with:
          # GitHub repo containing the values.yaml file
          gitRepo: "https://github.com/ghuser/iter8.git"
          # Path to values.yaml file
          filePath: "samples/second-exp/values.yaml"
          # GitHub username
          username: "ghuser"
          # Branch modified by this task
          branch: "gitops-test"
          # Secret containing the personal access token needed for git push
          secretName: "my-secret"
          # Namespace containing the above secret
          secretNamespace: "default"
```

## Inputs
| Field name | Field type         | Description | Required |
| ----- | ------------ | ----------- | -------- |
| gitRepo | string | GitHub repo containing the `values.yaml` file. The repo needs to begin with the prefix `https://`. | Yes |
| filePath | string | Path to the Helm values file, relative to the root of this repo. | Yes |
| username | string | GitHub username. For organization account, this can also be an org name. | Yes |
| branch | string | Branch to be updated by this task. Default value is `main`. | No |
| secretName | string | This task requires [a personal access token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token) in order to modify the GitHub repo. `secretName` is the name of the Kubernetes secret which contains this token. Default value is `ghtoken`. | No |
| secretNamespace | string | Namespace where the above secret is located. Default value is the namespace of the experiment. | No |

**Note:** The task above expects to find a key named `token` within the secret's `Data` section; i.e., `secret.Data["token"]` needs to be the GitHub personal access token. In addition, the `iter8-handler` service account in the `iter8-system` namespace needs to be given read permissions using RBAC rules for this secret.

## Result
The [version recommended for promotion](../../concepts/buildingblocks.md#version-promotion) by Iter8 will be promoted as the new baseline in the `values.yaml` file. Suppose the `values.yaml` file in the GitHub repo is the same as the one in [this example](../helmex-schema.md#example).

=== "Baseline is promoted"
    Assuming baseline is recommended for promotion by Iter8, the new `values.file` in the GitHub repo after this task executes, will look as follows. Notice how the `dynamic` field differs between the two scenarios.

    ```yaml
    common:
      application: hello
      repo: "gcr.io/google-samples/hello-app"
      serviceType: ClusterIP
      servicePortInfo:
        port: 8080
      regularLabels:
        app.kubernetes.io/managed-by: Iter8
      selectorLabels:
        app.kubernetes.io/name: hello

    baseline:
      name: hello
      selectorLabels:
        app.kubernetes.io/track: baseline
      dynamic:
        id: "mn82l82"
        tag: "1.0"

    # even though there is an experiment section below, there will be
    # no Iter8 experiment in the cluster, since there is no candidate version
    experiment:
      time: 5s
      QPS: 8.0
      limitMeanLatency: 500.0
      limitErrorRate: 0.01 
      limit95thPercentileLatency: 1000.0
    ```    

=== "Candidate is promoted"
    Assuming candidate is recommended for promotion by Iter8, the new `values.file` in the GitHub repo after this task executes, will look as follows. Notice how the `dynamic` field differs between the two scenarios.

    ```yaml
    common:
      application: hello
      repo: "gcr.io/google-samples/hello-app"
      serviceType: ClusterIP
      servicePortInfo:
        port: 8080
      regularLabels:
        app.kubernetes.io/managed-by: Iter8
      selectorLabels:
        app.kubernetes.io/name: hello

    baseline:
      name: hello
      selectorLabels:
        app.kubernetes.io/track: baseline
      dynamic:
        id: "8s72oa"
        tag: "2.0"

    # even though there is an experiment section below, there will be
    # no Iter8 experiment in the cluster, since there is no candidate version
    experiment:
      time: 5s
      QPS: 8.0
      limitMeanLatency: 500.0
      limitErrorRate: 0.01 
      limit95thPercentileLatency: 1000.0
    ```    
