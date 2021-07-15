---
template: main.html
---

# `common/http-request`
The `common/http-request` task can be used to send an HTTP request either as a form of notification or to trigger an action. For example, this task can be used to trigger a GitHub Action or a Tekton pipeline.

## Example
The following is an experiment snippet with a `common/http-request` task.

```yaml
...
spec:
  strategy:
    actions:
      finish:
      - task: common/http-request
        with:
          url: https://api.github.com/repos/ORG/REPO/actions/workflows/ACTION.yaml/dispatches
          body: |
            {
              "ref":"master", 
              "inputs":{
                "name": "{{.this.metadata.name}}",
                "home":"{{.this.metadata.namespace}}"
              }
            }
          secret: default/github-token
          authType: Bearer
          headers:
          - name: Accept
            value: application/vnd.github.v3+json
  ...
```

## Inputs
| Field name | Field type | Description | Required |
| ----- | ---- | ----------- | -------- |
| url | string | URL to which request is to be made. May contain placeholders that will be subsituted at runtime. | Yes |
| method | string | HTTP request method to use; either `POST` or `GET`. Default value is `POST`. | No |
| authtype | string | Type of authentication to use. Valid values are `Basic`, `Bearer` and `APIKey`. If not set, no authentication is used. | No |
| secret | string | Name of a secret (in form of `namespace/name`). Values are used to dynamically substitute placeholders. | No |
| headers | [][NamedValue](../../experiment/#namedvalue) | A list of name-value pairs that are converted into headers as part of the request. Values may contain placeholders that will be subsituted at runtime.| No |
| body | string | The body of the request to be sent. May contain placeholders that will be subsituted at runtime. | No |

The `url`, `headers` and `body` may all contain placeholders that will be substituted at runtime. 
A placeholder prefixed by `.secret` will use value from the secret. 
A placeholder prefixed by `.this` will use a value from the experiment.

If `authtype` is:

  - `Basic`: the task expects fields named `username` and `password` in the `secret`.
  - `Bearer`: the expects a field named `token` in the `secret`.
  - `APIKey`: the task expects the `header` field to specify any needed information. Placeholders can be used to explicitly refer to any required values from the `secret`.


## Result
The task will create and send an HTTP request. If the request returns an error, the task will fail.