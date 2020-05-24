## Kiali

Kiali is an observability console for Istio with service mesh configuration capabilities. It helps you to understand the structure of your service mesh by inferring the topology, and also provides the health of your mesh. Kiali provides detailed metrics, and a basic Grafana integration is available for advanced queries. Distributed tracing is provided by integrating Jaeger. For Detail Installation and Information about Kiali in general, please reference [Kiali.io](https://kiali.io)

### Enable Iter8 console in Kiali

Kiali version 1.18.1 (and above) now provides the capabilities to observe Iter8 experiment runtime behavior. To enable Kiali Iter8 extensions, user needs to update the Kiali CR. The Iter8 extension is not enabled by default, to enable Iter8 extensions, the Kiali Operator needs to be installed. Please follow the steps in [Advanced Install](https://kiali.io/documentation/getting-started/#_advanced_install_operator_only).

For example: 
  ```
  bash <(curl -L https://kiali.io/getLatestKialiOperator) --accessible-namespaces '**' -oiv latest -kiv latest --operator-install-kiali true
  ```
  will install the latest operator and Kiali from stable master.

  Make sure the Kiali CR is created by `kubectl get kialis.kiali.io kiali -n kiali-operator`.


1. Follow the step [Create or Edit the Kiali CR](https://kiali.io/documentation/getting-started/#_create_or_edit_the_kiali_cr) or use ``kubectl edit kialis.kiali.io kiali -n kiali-operator`

2. Enable `iter_8` extensions under Spec.

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

3. Make sure Kiali is restarted. Using `kubectl get pods -n istio-system` to check if Kiali pod is restarting

4. Install [Iter8](https://github.com/iter8-tools/docs/blob/v0.1.0/doc_files/iter8_install.md)

5. Start kiali - `istioctl dashboard kiali`

### Troubleshooting

1. Cannot find kiali cr in namespace kiali-operator.

    Try using this command to install and start operator
    ```
    bash <(curl -L https://kiali.io/getLatestKialiOperator) --accessible-namespaces '**' -oiv latest -kiv latest --operator-install-kiali true
    ```
2. If Iter8 did not show up in Kiali, check the configmap using this command `kubectl edit configmap kiali -n istio-system`. And make the iter_8 has `enabled: true`. In order for the configmap to take effect, please delete the kiali pod in namespace istio-system.  Kiali operator will restart the kiali pod automatically. 

3. Error message `Kiali has Iter8 extension enabled but it is not detected in the cluster`

    Make sure Iter8 is installed, Use kubectl pods -n iter8` to check if both controller and anlytics services are running

4. Experiment(s) are missing in the Iter8 main page.

    Make sure the namespace that contains the experiment is included in the Kiali accessible  namespace `accessible_namespaces:` definitions in the CR.
### Iter8 Feature

1. Iter8 experiments page

    Iter8 main page lists all the experiments in available namespace(s).

    <img src=../img/kiali-iter8-listing.png width=95%>

2. Experiment create page

   User can create new experiment from the Action pulldown on the right of the listing page. 

   <img src=../img/kiali-experiment-create-1.png width=95%>
   <img src=../img/kiali-experiment-create-2.png width=95%>

3. Experiment detail page

   Click on the name of the experiment from the listing page will show the experiment detail page. In the detail page, user can `pause`, `resume`, `terminate with success` and `terminate with failure` from the action pulldown. User can also `delete` experiment.

    <img src=../img/kiali-experiment-detail.png width=95%>