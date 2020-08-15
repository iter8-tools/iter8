
---
menuTitle: Kiali
title: Kiali
weight: 15
summary: Describes iter8's integrations with Kiali
---

Kiali is an observability console for Istio with service mesh configuration capabilities. It helps you to understand the structure of your service mesh by inferring the topology, and also provides the health of your mesh. Kiali provides detailed metrics, and a basic Grafana integration is available for advanced queries. Distributed tracing is provided by integrating Jaeger. For Detail Installation and Information about Kiali in general, please reference [Kiali.io](https://kiali.io)

## Enabling the iter8 Console in Kiali

Currently the iter8 extension for Kiali works only for [iter8 v0.2.1](https://www.iter8.tools), not the current version. To install this version of iter8, see [here](https://iter8.tools/getting-started/installation/kubernetes/).

### Install Kiali Using Operator

Kiali version 1.18.1 (and above) provides the capability to observe iter8 experiment runtime behavior. To enable the iter8 extensions for Kiali, which are disabled by default, you must update the Kiali CR using the Kiali operator.

To check if Kiali operator is installed, use:

```bash
kubectl --namespace kiali-operator get pods
```

To install the Kiali operator, follow the steps in [Install Kiali]( https://kiali.io/documentation/latest/installation-guide/#_install_kiali_latest). You can verify that the Kiali CR is created by using command:

```bash
kubectl  --namespace kiali-operator get kialis.kiali.io kiali
```

If this is the new installation, you will be asked to choose an authentication strategy (login, anonymous, ldap, openshift, token or openid). Depending on the chosen strategy, the installation process may prompt for additional information. Please see [Kiali Login Options](https://kiali.io/documentation/latest/installation-guide/#_login_options) for details about authentication strategies.

### Enable iter8 in the Kiali Operator CR

Follow the step [Create or Edit the Kiali CR](https://kiali.io/documentation/latest/installation-guide/#_create_or_edit_the_kiali_cr) or use:

```bash
kubectl --namespace kiali-operator edit kialis.kiali.io kiali
```

Find the `iter_8` key under `spec.extensions` and set `enabled` to `true`. The relevant portion of the CR is:

```
# Kiali enabled integration with Iter8 project.
# If this extension is enabled, Kiali will communicate with Iter8 controller allowing to manage Experiments and review results.
# Additional documentation https://iter8.tools/
#    ---
#    iter_8:
#
# Flag to indicate if iter8 extension is enabled in Kiali
#      ---
#      enabled: false
#
extensions:
  iter_8:
    enabled: true
```

Restart the Kiali pods:

```bash
kubectl --namespace istio-system delete pod $(kubectl --namespace istio-system get pod --selector='app=kiali' -o jsonpath='{.items[0].metadata.name}')
```

You can check if the pod has successfully restarted by inspecting the pods:

```bash
kubectl --namespace istio-system get pods
```

Install iter8 v0.2.1. See [install instructions](https://github.com/iter8-tools/docs/blob/v0.2.1/doc_files/iter8_install.md)

Start kiali using: 

```bash
istioctl dashboard kiali
```

## Features of the iter8 Extension for Kiali

### Experiments Overview

{{< figure src="/images/kiali-iter8-listing.png" title="iter8 main page" caption="iter8 main page lists all the experiments in available namespace(s).">}}

### Create Experiment

You can create new experiment from the Action pulldown on the right of the listing page.

{{< figure src="/images/kiali-experiment-create-1.png" title="Experiment creation">}}

{{< figure src="/images/kiali-experiment-create-2.png" title="Experiment creation -- additional configuration options">}}

### Experiment Detail

{{< figure src="/images/kiali-experiment-detail.png" title="Experiment detail page" caption="Click on the name of the experiment from the listing page will show the experiment detail page. In the detail page, user can `pause`, `resume`, `terminate with success` and `terminate with failure` from the action pulldown. User can also `delete` experiment.">}}

## Troubleshooting Guide

**Issue**: Cannot find the kiali CR in namespace `kiali-operator`.

Try using this command to install and start operator

```bash
bash <(curl -L https://kiali.io/getLatestKialiOperator) --accessible-namespaces '**' -oiv latest -kiv latest --operator-install-kiali true
```

---

**Issue**: The iter8 extension is not visible in Kiali

Check the configmap `kiali` using this command:

 ```bash
 kubectl  --namespace istio-systemedit configmap kiali
 ```

Ensure that `spec.extensions.iter_8.enabled` is set to `true`. To ensure that this configuration has taken effect, restart the kiali pod:

```bash
kubectl --namespace istio-system delete pod $(kubectl --namespace istio-system get pod --selector='app=kiali' -o jsonpath='{.items[0].metadata.name}')
```

---

**Issue**: Error message `Kiali has iter8 extension enabled but it is not detected in the cluster`

Make sure iter8 is installed, check that both iter80-controller and iter8-analytics are functioning:

```bash
kubectl --namespace iter8 get pods
```

---

**Issue**: Experiment(s) are missing in the iter8 main page

Make sure the namespace that contains the experiment is included in the Kiali accessible  namespace `accessible_namespaces:` definitions in the CR.
