# Iter8 library chart

> This chart provides reusable templates that are used by Iter8 experiment charts. Experiments can be composed by using this chart as a dependency, and by including templates available in this library chart.

**Templates for Iter8 tasks:** 

- `task.http`: The `gen-load-and-collect-metrics-http` task
- `task.grpc`: The `gen-load-and-collect-metrics-grpc` task
- `task.assess`: The `assess-app-versions` task

**Templates for Kubernetes experiments:**

- `k.job`: Kubernetes experiment job
- `k.spec.secret`: Kubernetes secret containing experiment spec
- `k.spec.role`: Role for Kubernetes spec secret
- `k.spec.rolebinding`: Role binding for Kubernetes spec secret
- `k.result.role`: Role for Kubernetes result secret
- `k.result.rolebinding`: Role binding for Kubernetes result secret
