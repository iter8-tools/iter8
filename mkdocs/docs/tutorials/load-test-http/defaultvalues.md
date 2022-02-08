---
template: main.html
hide:
- toc
---

# Default Values for `load-test-http`
```yaml
### Defaults values for the load-test-http experiment chart.

### The documentation follows Helm recommendations described in the URL below.
### https://helm.sh/docs/chart_best_practices/values/#document-valuesyaml

##################################

### url   HTTP(S) URL where the app receives GET or POST requests.
### This field is required.
url: null

### headers   HTTP headers to be used in requests sent to the app.
### Specified as a map with key being header names, and values being header values.
headers: null

### numQueries    Number of requests sent to the app.
numQueries: 100

### duration    Duration for which requests are sent to the app.
### Value can be any Go duration string (https://pkg.go.dev/maze.io/x/duration#ParseDuration)
### This field is ignored if `numQueries` is specified.
duration: 20s

### qps   Number of requests per second sent to each version.
qps: 8.0

### connections   Number of parallel connections used to send requests.
connections: 4

### payloadStr    String data to be sent as payload. 
### If this field is specified, Iter8 will send HTTP POST requests.
### with this string as the payload.
### This field is ignored if `payloadURL` is specified.
payloadStr: null

### payloadUrl    URL of payload. If this field is specified, 
### Iter8 will send HTTP POST requests to versions with 
### data downloaded from this URL as the payload.
payloadURL: null

### contentType   The type of the payload. Indicated using the Content-Type HTTP header value. 
### This is intended to be used in conjunction with one of the `payload*` fields above. 
### If this field is specified, Iter8 will send HTTP POST requests to versions 
### with this content type header value. If payload is supplied, and this field is omitted, 
### it will be defaulted to "application/octet-stream".
contentType: null

### errorsAbove   Any HTTP response code above this value is considered an error.
### Default value is 400.
errorsAbove: 400

### SLOs    A map of service level objectives (SLOs) that the app needs to satisfy.
### Metrics collected during the load test are used to verify if the app satisfies SLOs.
### Each SLO has a key which is the metric name, 
### and a value which is the upper limit on the metric.
### Valid metric names are error-rate, error-count, latency-max, 
### latency-mean, latency-stddev, and latency-pX, where X is any latency percentile 
### (i.e., any float value between 0 and 100).
SLOs: null
```