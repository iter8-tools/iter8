# API Testing

This chart installs an [Iter8](https://iter8.tools) experiment for API Testing. Once installed, the experiment generates requests for an HTTP API endpoint, collects responses, and ensures that latency and error related service level objetives (SLOs) are satisfied.

This experiment is ideally suited for dev/test/staging environments for testing apps before their release. It can be used to test both HTTP GET and POST API endpoints.

## Values

This chart supports the following values.

| Field       | Description | Required | Default |
| ----------- | ----------- | -------- | ------- |
| Header      | Title       | No | Foo |
| Paragraph   | Text        | No | Bar |

## Usage

Once installed, you can describe the results of the experiment using the [Iter8 plugin for `kubectl`](https://iter8.tools/latest/kubectl-plugin) as follows:

```console
kubectl iter8 describe -n <release namespace>
```

## More Info

Please see https://iter8.tools