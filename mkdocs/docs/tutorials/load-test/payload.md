---
template: main.html
---

# HTTP POST with Payload

!!! tip "Send a payload as part of the requests sent during the load-test"
    While load testing an HTTP service with a POST endpoint, you may send any type of content as payload during the load test. This tutorial shows how.

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
The target URL for our load test is http://127.0.0.1/post, which implements an HTTP POST endpoint. You can send any type of content as payload during the load test, either by setting it as a string, or by fetching it from a payload URL.

### Examples

=== "string (application/octet-stream)"
    Supply a string as payload. This sets the content type to `application/octet-stream` during the load test.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123"
    ```

=== "string (text/plain)"
    Supply a string as payload and explicitly set the content type to `text/plain` during the load test.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123" \
              --set contentType="text/plain"
    ```

=== "URL (application/json)"
    Fetch JSON content from a payload URL. Use this JSON as payload and explicitly set the content type to `application/json` during the load test.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadURL=https://json-generator.com/ \
              --set contentType="application/json"
    ```

=== "URL (image/jpeg)"
    Fetch jpeg image from a payload URL. Use this image as payload and explicitly set the content type to `image/jpeg` during the load test.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadURL=https://cdn.pixabay.com/photo/2021/09/08/17/58/poppy-6607526_1280.jpg \
              --set contentType="image/jpeg"
    ```

You can assert experiment outcomes and generate report as described in the [quick start tutorial](../../getting-started/your-first-experiment.md).