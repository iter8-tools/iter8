# Installing iter8 on Red Hat OpenShift

## Prerequisites

We recommend using the _Red Hat OpenShift Service Mesh_. This can be installed using the Red Hat OpenShift Service Mesh Operator. For details, see: <https://docs.openshift.com/container-platform/4.3/service_mesh/service_mesh_install/installing-ossm.html>.

Installing the Service Mesh involves installing the Elasticsearch, Jaeger, Kiali and Red Hat OpenShift Service Mesh Operators, creating and managing a `ServiceMeshControlPlane` resource to deploy the control plane, and creating a `ServiceMeshMemberRoll` resource to specify the namespaces associated with the Red Hat OpenShift Service Mesh.

## Installing iter8

By default, iter8 uses the Prometheus service installed as part of the Red Hat OpenShift Service Mesh for the metrics used to assess the quality of different versions of a service. The Red Hat OpenShift Service Mesh configures the Prometheus service to require authentication. To configure iter8 to authenticate with Prometheus, some additional steps are needed.

### Install the iter8 analytics service

Download and untar the [helm chart](https://github.com/iter8-tools/iter8-analytics/releases/download/v0.2/iter8-analytics-helm-char.tar) for the iter8-analytics service. The following options can be used to generate the needed yaml:

```bash
REPO=iter8/iter8-analytics
PROMETHEUS_SERVICE='https://prometheus.istio-system:9090'
PROMETHEUS_USERNAME='internal'
PROMETHEUS_PASSWORD=<FILL IN>
helm template install/kubernetes/helm/iter8-analytics \
    --name iter8-analytics \
    --set image.repository=${REPO} \
    --set image.tag=v0.2 \
    --set iter8Config.authentication.type=basic \
    --set iter8Config.authentication.username=${PROMETHEUS_USERNAME} \
    --set iter8Config.authentication.password=${PROMETHEUS_PASSWORD} \
    --set iter8Config.authentication.insecure_skip_verify=true \
    --set iter8Config.metricsBackendURL=${PROMETHEUS_SERVICE} \
| kubectl -n iter8 apply -f -
```

The password, to be used can be found in the secret `htpasswd` in the namespace where Istio is installed. For example, the following might work to identify it:

```bash
PROMETHEUS_PASSWORD=$(kubectl -n istio-system get secret htpasswd -o jsonpath='{.data.rawPassword}' | base64 --decode)
```

### Install the iter8 controller

The [quick install instructions](https://github.com/iter8-tools/docs/blob/v0.2/doc_files/iter8_install.md#quick-installation) can be used to install the iter8 controller. The Service Mesh currently uses Istio telemetry version `v1`:

```bash
kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v0.2/install/iter8-controller.yaml
```

## Target Services

The Red Hat OpenShift Service Mesh is restricted to the set of namespaces defined in the `ServiceMeshMemberRoll` resource. In particular, if you will be trying the tutorials, add the namespace `bookinfo-iter8` to the `ServiceMeshMemberRoll`.

Istio relies a sidecar injected into each pod to provide its capabilities. Istio provides several ways this sidecar can be [injected](https://istio.io/docs/setup/additional-setup/sidecar-injection/). Red Hat recommends the use of the annotation `sidecar.istio.io/inject: "true"` in the deployment yaml. Examples can be found in the yaml for the tutorial: <https://github.com/iter8-tools/iter8-controller/blob/v0.2/doc/tutorials/istio/bookinfo/bookinfo-tutorial.yaml>
