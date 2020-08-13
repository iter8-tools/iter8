
---
menuTitle: Kiali
title: Integrating with Kiali
weight: 15
summary: Describes iter8's integrations with Kiali
---

Kiali is an observability console for Istio with service mesh configuration capabilities. It helps you to understand the structure of your service mesh by inferring the topology, and also provides the health of your mesh. Kiali provides detailed metrics, and a basic Grafana integration is available for advanced queries. Distributed tracing is provided by integrating Jaeger. For Detail Installation and Information about Kiali in general, please reference [Kiali.io](https://kiali.io)

## Enabling the iter8 Console in Kiali

### A. Install Kiali Using Operator

Kiali version 1.18.1 (and above) now provides the capabilities to observe iter8 experiment runtime behavior. To enable the Kiali iter8 extensions, which are disabled by default, you must update the Kiali CR using the Kiali Operator.

**Note**: Currently the iter8 extension for Kiali works only for [iter8 v0.2.1](https://www.iter8.tools), not the current version. To install this version of iter8, see [here](https://iter8.tools/getting-started/installation/kubernetes/).

To check if Kiali operator is installed, use:

```bash
kubectl get pods -n kiali-operator
```

To install the Kiali Operator, please follow the steps in [Advanced Install](https://kiali.io/documentation/latest/installation-guide/#_advanced_install_operator_only). And make sure the Kiali CR is created by using command:

```bash
kubectl get kialis.kiali.io kiali -n kiali-operator
```

If this is the new installation, you will be asked to choose an authentication strategy (login, anonymous, ldap, openshift, token and openid). Depending on the chosen strategy, the installation may prompt for additional information. Please reference [Kiali Login Options](https://kiali.io/documentation/latest/installation-guide/#_login_options) for details about authentication strategies.

### B. Enable iter8 in the Kiali Operator CR

1. Follow the step [Create or Edit the Kiali CR](https://kiali.io/documentation/latest/installation-guide/#_create_or_edit_the_kiali_cr) or use:

    ```bash
    kubectl edit kialis.kiali.io kiali -n kiali-operator
    ```

2. Enable `iter_8` extensions under `Spec`.

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

3. Restart the Kiali pods:

    ```bash
    kubectl -n istio-system delete pod $(kubectl -n istio-system get pod --selector='app=kiali' -o jsonpath='{.items[0].metadata.name}')
    ```

    You can inspect the pods using `kubectl get pods -n istio-system` to check if Kiali pod has restarted.

4. Install [iter8 v0.2.1](https://github.com/iter8-tools/docs/blob/v0.2.1/doc_files/iter8_install.md)

5. Start kiali - `istioctl dashboard kiali`

## iter8 Extension Features

1. Iter8 experiments page

     {{< figure src="/images/kiali-iter8-listing.png" title="iter8 main page" caption="iter8 main page lists all the experiments in available namespace(s).">}}

2. Experiment create page

   You can create new experiment from the Action pulldown on the right of the listing page.

    {{< figure src="/images/kiali-experiment-create-1.png" title="Experiment create page" caption="">}}

    {{< figure src="/images/kiali-experiment-create-2.png" title="Experiment create page" caption="">}}

3. Experiment detail page

    {{< figure src="/images/kiali-experiment-detail.png" title="Experiment detail page" caption="Click on the name of the experiment from the listing page will show the experiment detail page. In the detail page, user can `pause`, `resume`, `terminate with success` and `terminate with failure` from the action pulldown. User can also `delete` experiment.">}}

## Troubleshooting

1. Cannot find kiali cr in namespace kiali-operator.

    Try using this command to install and start operator

    ```bash
    bash <(curl -L https://kiali.io/getLatestKialiOperator) --accessible-namespaces '**' -oiv latest -kiv latest --operator-install-kiali true
    ```

2. If iter8 did not show up in Kiali, check the configmap using this command `kubectl edit configmap kiali -n istio-system`. And make the iter_8 has `enabled: true`. In order for the configmap to take effect, please delete the kiali pod in namespace istio-system.  Kiali operator will restart the kiali pod automatically.

3. Error message `Kiali has iter8 extension enabled but it is not detected in the cluster`

    Make sure iter8 is installed, Use `kubectl -n iter8 get pods` to check if both controller and analytics services are running.

4. Experiment(s) are missing in the iter8 main page.

    Make sure the namespace that contains the experiment is included in the Kiali accessible  namespace `accessible_namespaces:` definitions in the CR.
