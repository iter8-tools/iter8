url: test-database.com/prometheus/api/v1/query
provider: test-request-body
method: GET
# Note: elapsedTimeSeconds is produced by Iter8
metrics:
- name: request-count
  type: counter
  description: |
    Number of requests
  body: |
    example request body
  params:
  - name: query
    value: |
      example query parameter
  jqExpression: .data.result[0].value[1] | tonumber