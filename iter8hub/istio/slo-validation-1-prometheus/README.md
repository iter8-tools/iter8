# Use Case

SLO validation of a single version of an application deployed to Istio.
Test with metrics supplied by prometheus.

## Comments

This just explores the template of the experiment; no effort to convert to a configmap or to incluce RBAC rules is included.

### task `metrics/load-backend@v1`

This example uses a task `metrics/load-backend@v1` to "load" a metrics backend. This backend defines, for example, a set of metrics. Iter8 might provide sample defaults for Istio/Prometheus and Linkerd/Prometheus or the user might provide a set of their own.

This requires the location of a manifest for the backends.  An example is below.
Optionally, an `authType` can be specified (default is *Default*). When specified, a secret is further required.

I've included the optional parameter `name`. It allows the user to specify the name by which the backend will be known. In this way it can be whatever makes sense to the user.  Of course a default would still be necessary but this avoids the need to know the default value when defining an experiment.

If multiple backends were being used, they could loaded by using multiple `load-backend` tasks.

### task `metrics/analytics@v1`

The key challenge I see here is for the user to understand what paramters are needed. These are defined in the backend. By making the backend something that is loaded we may have reverted to the same problem we had with Metrics originally.

### Other Comments

The `run` task sleeps for `.Values.metricsCollecTime`. The same parameter might be used in `metrics/collect` when a timed collection is used. This unifies the concepts.

An alternative to `load-backend` would be to have an analytics engine load the backends. This is more analogous to my thinking for `metrics/collect-builtin`. However, I feel like this leaves the analytics task trying to do too much. We should think about tasks as doing 1 thing well. Not doing many things. Perhaps we need to split the function for `metrics/collect` as well. In the same vein, I wonder if we should further split `metrics/collect` into a collection only task and a `metrics/analyze` task similar to the one used here. The purpose for this split is so that we can use a common expression.

## Sample Use

Differs from builtin metrics use case by addtion of backend parameters.

```bash
helm upgrade -n $NAMESPACE slo-experiment $ITER8/samples/istio/slo-validation \
  --set limitMeanLatency=100.0 \
  --set limitErrorRate=0.0 \
  --set limit95thPercentileLatency=200.0 \
  --set version.name=hello \
  --set prometheus.URL=http://prometheus.istio-system.cluster.local:9090 \
  --install 
```
