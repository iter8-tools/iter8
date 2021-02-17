---
template: overrides/main.html
---

# Canary Experiment
Follow these instructions to perform **zero-downtime progressive and fixed-split metrics-driven canary rollout of a Knative application**.

!!! example "Prepare for this experiment"

    If you executed the `cleanup` step from your previous experiment and left the cluster intact, you are all set. If not, follow instructions from the [quick start tutorial] (https : //...) to install Knative and iter8 on your Kubernetes cluster, and install iter8ctl locally.

!!! note "Change directory"

    ```shell
    cd path-to-your-git-cloned-iter8-folder/iter8mkdocs/docs/code-samples/iter8-knative/canary
    ```

## Create Knative service
```shell
kubectl apply -f service.yaml
```

??? info "Inside the service"
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod
    nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor
    massa, nec semper lorem quam in massa.

    ``` python
    def bubble_sort(items):
        for i in range(len(items)):
            for j in range(len(items) - 1 - i):
                if items[j] > items[j + 1]:
                    items[j], items[j + 1] = items[j + 1], items[j]
    ```

    Nunc eu odio eleifend, blandit leo a, volutpat sapien. Phasellus posuere in
    sem ut cursus. Nullam sit amet tincidunt ipsum, sit amet elementum turpis.
    Etiam ipsum quam, mattis in purus vitae, lacinia fermentum enim.

## Generate traffic
```shell
kubectl apply -f fortio.yaml
```

## Create iter8 experiment

=== "progressive"
    ```shell
    kubectl apply -f experiment.yaml
    ```
    ??? info "Inside the progressive canary experiment"
        Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod
        nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor
        massa, nec semper lorem quam in massa.

        ``` python
        def bubble_sort(items):
            for i in range(len(items)):
                for j in range(len(items) - 1 - i):
                    if items[j] > items[j + 1]:
                        items[j], items[j + 1] = items[j + 1], items[j]
        ```

        Nunc eu odio eleifend, blandit leo a, volutpat sapien. Phasellus posuere in
        sem ut cursus. Nullam sit amet tincidunt ipsum, sit amet elementum turpis.
        Etiam ipsum quam, mattis in purus vitae, lacinia fermentum enim.

=== "fixed-split"
    ```shell
    kubectl apply -f experiment.yaml
    ```
    ??? info "Inside the fixed-split canary experiment"
        Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod
        nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor
        massa, nec semper lorem quam in massa.

        ``` python
        def bubble_sort(items):
            for i in range(len(items)):
                for j in range(len(items) - 1 - i):
                    if items[j] > items[j + 1]:
                        items[j], items[j + 1] = items[j + 1], items[j]
        ```

        Nunc eu odio eleifend, blandit leo a, volutpat sapien. Phasellus posuere in
        sem ut cursus. Nullam sit amet tincidunt ipsum, sit amet elementum turpis.
        Etiam ipsum quam, mattis in purus vitae, lacinia fermentum enim.

## Observe experiment in realtime

To observe the experiment in realtime, follow the instructions from [quick start] ( https:// ... ). When the experiment completes, you will see the state of the experiment change to `Completed` in the `iter8ctl` output.

## Cleanup
```shell
kubectl delete -f fortio.yaml
kubectl delete -f experiment.yaml
kubectl delete -f service.yaml
```

??? info "Understanding what happened"
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod
    nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor
    massa, nec semper lorem quam in massa.

    ``` python
    def bubble_sort(items):
        for i in range(len(items)):
            for j in range(len(items) - 1 - i):
                if items[j] > items[j + 1]:
                    items[j], items[j + 1] = items[j + 1], items[j]
    ```

    Nunc eu odio eleifend, blandit leo a, volutpat sapien. Phasellus posuere in
    sem ut cursus. Nullam sit amet tincidunt ipsum, sit amet elementum turpis.
    Etiam ipsum quam, mattis in purus vitae, lacinia fermentum enim.
