url: http://url/query
provider: nan-prom
method: GET
metrics:

- name: metric-tonumber
  type: counter
  description: tonumber
  params:
  - name: query
    value: query-tonumber
  jqExpression: .value | tonumber

- name: metric-no-tonumber
  type: counter
  description: no-tonumber
  params:
  - name: query
    value: query-no-tonumber
  jqExpression: .value
