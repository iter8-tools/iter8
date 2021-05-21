---
template: main.html
---

# Notification Tasks

## `notification/slack`

### Overview

The `notification/slack` task posts a slack message summarizing the experiment.

### Arguments

| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| channel | string | Name of the slack channel to which messages should be posted. | Yes |
| secret | string | Identifies a secret containing a `token` to be used for authentication.  Expressed as `namespace/name`. If `namespace` is not specified, the namespace of the experiment is used. | Yes |

### Requirements

#### Slack API Token

An API token allowing posting messages to the desired slack channel is needed. It should be stored in a secret as described below.

1. Create a new slack app https://api.slack.com/apps?new_app=1. Select `from scratch`, set a name (ex., `iter8`), and select the workspace that contains the channel(s) where you want to send messages. Click the `Create App` button.
2. Select `Permissions`
3. Under `Scopes`, add Bot Token scopes `channels:read` and `chat:write`.
4. Click the `Install to Workspace` button.
5. Select the `Install your app` and click on the `Install to workspace` button.
6. Copy the token that is generated, add it to a secret. For example, to create the secret `slack-secret    ` in the default namespace:

    ```shell
    kubectl create secret generic slack-secret --from-literal=token=<slack token>
    ```

#### Permission to read secret with Slack token

The iter8 task handler needs permission to read the identified secret. This RBAC Role and RoleBinding may be used: [`read-secrets.yaml`](https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/tasks/rbac/read-secrets.yaml)

??? info "Inspect role and rolebinding"
    ```yaml linenums="1"
        # This role enables reading of secrets
        apiVersion: rbac.authorization.k8s.io/v1
        kind: ClusterRole
        metadata:
        name: iter8-secret-reader
        rules:
        - apiGroups:
        - ""
        resources:
        - secrets
        verbs: ["get", "list"]
        ---
        # This role binding enables Iter8 handler to read secrets in the default namespace.
        # To change the namespace apply to the target namespace
        apiVersion: rbac.authorization.k8s.io/v1
        kind: RoleBinding
        metadata:
        name: iter8-secret-reader-handler
        roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: ClusterRole
        name: iter8-secret-reader
        subjects:
        - kind: ServiceAccount
        name: iter8-handlers
        namespace: iter8-system

    ```

#### Target Slack Channel

A slack channel is identified by an id. To find the id, open the slack channel in a web browser. The channel id is the portion of the URL of the form: `CXXXXXXXX`

### Result

A slack message describing the experiment will be posted to the specified channel. It will look something like this:

![Sample slack notificiation](../../images/slack-notification.png)

If a winner has been identified, this will also be included.

### Examples

The following task in the start action of an experiment creates a notification that will be posted to the slack channel with id `channel` during the execution of the start action.

```yaml
- start:
  task: notification/slack
    with:
    - channel: channel
    - secret: ns/slack-token
```

<!-- -->
## `notification/webhook`

### Overview

### Arguments

### Requirements

### Result

### Examples
<!--  -->
