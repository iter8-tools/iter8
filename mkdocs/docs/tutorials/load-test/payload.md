---
template: main.html
---

# Payload

!!! tip "Send various types of content as payload"
    HTTP services with POST endpoints may accept payloads. Send various types of content as payload during the load test.

***

???+ note "Before you begin"
    1. [Install Iter8](../../getting-started/install.md).
    2. Complete the [quick start tutorial](../../getting-started/your-first-experiment.md).

## 1. Run sample app
Run the [httpbin](https://httpbin.org) sample app from a separate terminal. We will load test this app in this example.
```shell
docker run -p 80:80 kennethreitz/httpbin
```
You may also use [Podman](https://podman.io) or other alternatives to Docker in the above command.


## 2. Download experiment chart
```shell
iter8 hub -e load-test
cd load-test
```

## 3. Run experiment
We will load test and validate the HTTP service whose URL is http://127.0.0.1/post. 

Iter8 enables you to send any type of content as payload during the load test, either by specifying the payload as a string (`payloadStr`), or by specifying a URL for Iter8 to fetch the payload from (`payloadURL`).

### Payload examples

=== "string (application/octet-stream)"
    Supply a string as payload. This sets the content type to `application/octet-stream`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123" \
              --set SLOs.error-rate=0 \
              --set SLOs.mean-latency=50 \
              --set SLOs.p90=100 \
              --set SLOs.p'97\.5'=200
    ```

=== "string (text/plain)"
    Supply a string as payload and explicitly set the content type to `text/plain`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123" \
              --set contentType="text/plain" \
              --set SLOs.error-rate=0 \
              --set SLOs.mean-latency=50 \
              --set SLOs.p90=100 \
              --set SLOs.p'97\.5'=200
    ```

=== "URL (application/json)"
    Fetch JSON content from a payload URL. Use this JSON as payload and explicitly set the content type to `application/json`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadURL=https://httpbin.org/stream/1 \
              --set contentType="application/json" \
              --set SLOs.error-rate=0 \
              --set SLOs.mean-latency=50 \
              --set SLOs.p90=100 \
              --set SLOs.p'97\.5'=200
    ```

=== "URL (image/jpeg)"
    Fetch jpeg image from a payload URL. Use this image as payload and explicitly set the content type to `image/jpeg`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadURL=https://cdn.pixabay.com/photo/2021/09/08/17/58/poppy-6607526_1280.jpg \
              --set contentType="image/jpeg" \
              --set SLOs.error-rate=0 \
              --set SLOs.mean-latency=50 \
              --set SLOs.p90=100 \
              --set SLOs.p'97\.5'=200
    ```

***

Assert experiment outcomes and view reports as described in the [quick start tutorial](../../getting-started/your-first-experiment.md).