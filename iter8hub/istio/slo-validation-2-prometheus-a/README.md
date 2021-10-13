# Use Case

SLO validation of a two versions of an application deployed to Istio.
Test with prometheus.

## Comments

This just explores the template of the experiment; no effort to convert to a configmap or to incluce RBAC rules is included.

See `slo-validation-1-prometheus` for some comments.

In this template, different versions are expressed using `.Values.versionA` and `.Values.versionB`. The template can probably be written to use either these or `.Values.version`; that is, for either 1 or 2 versions.  This would be more reusable but also be more complex.  It should also be possible to express versions with an array `.Values.versions` and write the template to use the specified number.

## Sample Use

Differs from builtin metrics use case by addtion of backend parameters.

```bash
helm upgrade -n $NAMESPACE slo-experiment $ITER8/samples/istio/slo-validation \
  --set limitMeanLatency=100.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=200.0 \
  --set versionA.name=hello \
  --set versionB.name=hello-candidate \
  --set prometheus.URL=http://prometheus.istio-system.cluster.local:9090 \
  --set winner.versionA=https://raw.githubusercontent.com/iter8-tools/iter8/v0.8/samples/istio/hello/hello.yaml \
  --set winner.versionB=https://raw.githubusercontent.com/iter8-tools/iter8/v0.8/samples/istio/hello/hello-candidate.yaml
  --install 
```
