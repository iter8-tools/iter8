---
template: main.html
---

# Useful Knative Annotations

This document discusses a few annotations that tune the behavior of Knative and are useful to incorporate in Knative services participating in Iter8 experiments. The first of these enables experimentation with cluster-local services, while the last two avoid cold start issues.

## Cluster-local (backend) services
Cluster-local or backend services are private services which are only available inside the cluster. Iter8 experiments with cluster-local Knative services are similar to any other Iter8 experiments. The following example from the [Knative documentation on cluster-local services](https://knative.dev/docs/serving/cluster-local-route/) shows how you can label a Knative service as cluster-local.

``` shell
kubectl label kservice ${KSVC_NAME} networking.knative.dev/visibility=cluster-local
```

## Scale boundaries
You can configure upper and lower bounds to control Knative's autoscaling behavior. The lower bound can be used to ensure that at least one replica is available for every version (Knative revision), and cold start issues do not interfere with version assessments during an experiment. The following example from the [Knative documentation on configuring scale boundaries](https://knative.dev/docs/serving/autoscaling/scale-bounds/#lower-bound) shows how you can control the minimum number of replicas each revision should have.

``` yaml linenums="1" hl_lines="10"
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
  namespace: default
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "3"
    spec:
      containers:
        - image: gcr.io/knative-samples/helloworld-go
```

## Pod retention period
The `scale-to-zero-pod-retention-period` annotation can be used to specify the minimum amount of time that the last pod will remain active after the Knative Autoscaler decides to scale pods to zero. Like the `minScale` annotation, this annotation is also useful for avoiding cold start issues during an experiment. For more information, see the [Knative documentation on configuring scale to zero](https://knative.dev/docs/serving/autoscaling/scale-to-zero/). The following example from the [Iter8 traffic segmentation tutorial](../../../tutorials/knative/traffic-segmentation/) shows how you can use this annotation in a service.

``` yaml linenums="1" hl_lines="11"
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: sample-app-v1
  namespace: default
spec:
  template:
    metadata:
      name: sample-app-v1-blue
      annotations:
        autoscaling.knative.dev/scaleToZeroPodRetentionPeriod: "10m"
    spec:
      containers:
      - image: gcr.io/knative-samples/knative-route-demo:blue 
        env:
        - name: T_VERSION
          value: "blue"
```

