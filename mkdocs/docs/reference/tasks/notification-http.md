---
template: main.html
---

# `notification/http`
The `notification/http` task can be used to send an HTTP request either as a form of notification or to trigger an action. For example, this task can be used to trigger a GitHub Action or a Tekton pipeline.

## Example
The following is an experiment snippet with a `notification/http` task that triggers a GitHub action that takes two inputs. Here the inputs are set to the name and namespace of the experiment.

```yaml
...
spec:
  strategy:
    actions:
      finish:
      - task: notification/http
        with:
          url: https://api.github.com/repos/ORG/REPO/actions/workflows/ACTION.yaml/dispatches
          body: |
            {
              "ref":"master", 
              "inputs":{
                "name": "@<.this.metadata.name>@",
                "home":"@<.this.metadata.namespace>@"
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
| authtype | string | Type of authorization to use. Valid values are `Basic`, `Bearer` and `APIKey`. If not set, no authorization is used. | No |
| secret | string | Name of a secret (in form of `namespace/name`). Values are used to dynamically substitute placeholders. | No |
| headers | [][NamedValue](../../experiment/#namedvalue) | A list of name-value pairs that are converted into headers as part of the request. Values may contain placeholders that will be subsituted at runtime.| No |
| body | string | The body of the request to be sent. May contain placeholders that will be subsituted at runtime. | No |
| ignoreFailure | bool | A flag indicating whether or not to ignore failures. If failures are not ignored, they cause the experiment to fail. Default is `true`. | No |

The `url`, `headers` and `body` may all contain placeholders that will be substituted at runtime. 
A placeholder prefixed by `.secret` will use value from the secret. 
A placeholder prefixed by `.this` will use a value from the experiment.
Placeholders in this task use `@<` and `>@` as the left and right delimiters respectively.

If not set, the default body will be of the following form:

```json
{
  "summary": {
    "winnerFound": true,
    "winner": "candidate",
    "versionRecommededForPromotion": "candidate",
    "lastRecommendedWeights": [
      { "name": "candiate"
        "weight": 95},
      { "name": "baseline"
        "weight": 5}
    ]
  },
  "experiment": <JSON representation of the experiment object>
}
```

In the default body, the `winner` will be set only if `winnerFound` is `true`. The `versionRecommededForPromotion` field will be omitted in start actions but will be included thereafter.
The weights are the last weights recommended by the analytics engine. Note that this may not match the current weights.

No authoriziation is provided if `authtype` is not set. If it is set, the behavior is as follows:

  - `Basic`: the task expects fields named `username` and `password` in the `secret`. These will be used to add an appropriate `Authorization` header to the request.
  - `Bearer`: the expects a field named `token` in the `secret`. This will be used to construct a suitable `Authorization` header to the request.
  - `APIKey`: the task expects the `header` field to explicitly specify any needed authorization headers. Placeholders can be used to explicitly refer to any required values from the `secret`.

By default, a `Content-type` header with value `application/json` is included in the request. This can be replaced by specifying a different value. For example to set it by `text/plain` by:

```yaml
...
  headers:
  - name: Content-type
    value: text/plain
...
```

## Result
The task will create and send an HTTP request.
