---
template: main.html
---

# Request Payload

!!! tip "Send various types of content as request payload"
    HTTP services with POST endpoints may accept payloads as part of requests. Send various types of content as payload during the load test.

***

???+ note "Before you begin"
    1. [Install Iter8](../../getting-started/install.md).

## 1. Run sample app
Run the [httpbin](https://httpbin.org) sample app from a separate terminal. We will load test this app in this example.
```shell
docker run -p 80:80 kennethreitz/httpbin
```
You may also use [Podman](https://podman.io) or other alternatives to Docker in the above command.


## 2. Download experiment chart
```shell
iter8 hub -e load-test-http
cd load-test-http
```

## 3. Run experiment
Iter8 enables you to send any type of content as payload during the load test of HTTP POST endpoints, either by specifying the payload as a string (`payloadStr`), or by specifying a URL for Iter8 to fetch the payload from (`payloadURL`). You can also specify the HTTP Content Type header (`contentType`).

### Payload examples

=== "string"
    Supply a string as payload. When `payloadStr` is set, content type is set to `application/octet-stream` by default.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123" \
              --set SLOs.error-rate=0 \
              --set SLOs.latency-mean=50 \
              --set SLOs.latency-p90=100 \
              --set SLOs.latency-p'97\.5'=200
    ```

=== "string"
    Supply a string as payload. Set content type to `text/plain`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadStr="abc123" \
              --set contentType="text/plain" \
              --set SLOs.error-rate=0 \
              --set SLOs.latency-mean=50 \
              --set SLOs.latency-p90=100 \
              --set SLOs.latency-p'97\.5'=200
    ```

=== "JSON from URL"
    Fetch JSON content from a URL. Use this JSON as payload. Set content type to `application/json`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadURL=https://data.police.uk/api/crimes-street-dates \
              --set contentType="application/json" \
              --set SLOs.error-rate=0 \
              --set SLOs.latency-mean=50 \
              --set SLOs.latency-p90=100 \
              --set SLOs.latency-p'97\.5'=200
    ```

=== "Image from URL"
    Fetch jpeg image from a URL. Use this image as payload. Set content type to `image/jpeg`.
    ```shell
    iter8 run --set url=http://127.0.0.1/post \
              --set payloadURL=https://cdn.pixabay.com/photo/2021/09/08/17/58/poppy-6607526_1280.jpg \
              --set contentType="image/jpeg" \
              --set SLOs.error-rate=0 \
              --set SLOs.latency-mean=50 \
              --set SLOs.latency-p90=100 \
              --set SLOs.latency-p'97\.5'=200
    ```
