metadata:
  name:      myName
  namespace: myNamespace
spec:
# task 1: generate HTTP requests for application URL
# collect Iter8's built-in HTTP latency and error-related metrics
- task: http
  with:
    duration: 2s
    errorRanges:
    - lower: 500
    url: {{ .URL }}
