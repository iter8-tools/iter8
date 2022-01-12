---
template: main.html
---

# HTTP POST with Payload

!!! tip "Send a payload as part of the requests sent during the load-test"
    HTTP services may implement an HTTP ... 

???+ note "Before you begin"
    1. [Install Iter8](../../getting-started/install.md).
    2. You may find it useful to try the [quick start tutorial](../../getting-started/your-first-experiment.md) first.

## 1. Run sample app
Run the [httpbin](https://httpbin.org) sample app from a separate terminal.
```shell
docker run -p 80:80 kennethreitz/httpbin
```
In the above command, you may also use [Podman](https://podman.io) or other alternatives to Docker.


## 2. Download experiment chart
```shell
iter8 hub -e load-test
cd load-test
```

## 3. Run experiment
The target URL for our load test is http://127.0.0.1/post, which implements an HTTP POST endpoint. There are various ways to send payload as part of this load test.

=== "text/plain (string)"
    Supply plain text as payload in the form of a string.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123" \
              --set contentType="text/plain"
    ```

You can assert experiment outcomes and generate report as described in the [quick start tutorial](../../getting-started/your-first-experiment.md).