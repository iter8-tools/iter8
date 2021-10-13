# Use Case

SLO validation of a single version of an application deployed to Istio.
Test with built-in metrics.

## Comments

This just explores the template of the experiment; no effort to convert to a configmap or to incluce RBAC rules is included.

There is nothing special due to Istio; the same helm chart that is used for any other SLO validation with built-in metrics can be used.

Note, however, that it is slightly more general than other SLO valiation charts we have developed in the past. It enables use of a header in the request.

```yaml
          {{- if .Values.version.headers }}
          headers:
{{ toYaml .Values.version.headers | indent 12 }}
          {{ end }}
```

## Sample Use

```bash
helm upgrade -n $NAMESPACE slo-experiment $ITER8/samples/istio/slo-validation \
  --set limitMeanLatency=100.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=200.0 \
  --set version.name=hello \
  --set version.URL=http://istio-ingressgateway.istio-system.cluster.local:80/hello \
  --set version.headers.host=hello.example.com \
  --install 
```
