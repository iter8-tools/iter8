---
template: main.html
---

# Fixed Split Deployment

!!! tip "Scenario: FixedSplit deployment"

    [FixedSplit deployment](../../../concepts/buildingblocks/#deployment-patterns), as the name indicates, is meant for scenarios where you do **not** want Iter8 to shift traffic between versions during the experiment. In this tutorial, you will:

    * Modify the [quick start tutorial](../../../getting-started/quick-start) to use FixedSplit instead of Progressive deployment.

    The modified A/B testing experiment with FixedSplit deployment pattern is depicted below.

    ![Canary](../../images/fixedsplitab-exp.png)

???+ warning "Before you begin... "

    This tutorial is available for the following K8s stacks.

    [Istio](#before-you-begin){ .md-button }
    [KFServing](#before-you-begin){ .md-button }
    [Knative](#before-you-begin){ .md-button }

    Please choose the same K8s stack consistently throughout this tutorial. If you wish to switch K8s stacks between tutorials, start from a clean K8s cluster, so that your cluster is correctly setup.
    
## Steps 1 to 3
    
Please follow steps 1 through 3 of the [quick start tutorial](../../../getting-started/quick-start/#1-create-kubernetes-cluster).

## 4. Create versions and initialize traffic split
=== "Istio"

    ```shell
    kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/fixed-split/bookinfo-app.yaml
    kubectl apply -n bookinfo-iter8 -f $ITER8/samples/istio/quickstart/productpage-v2.yaml
    kubectl wait -n bookinfo-iter8 --for=condition=Ready pods --all
    ```

    ??? info "Virtual service with traffic split"
        ```yaml linenums="1"
        apiVersion: networking.istio.io/v1alpha3
        kind: VirtualService
        metadata:
          name: bookinfo
        spec:
          gateways:
          - mesh
          - bookinfo-gateway
          hosts:
          - productpage
          - "bookinfo.example.com"
          http:
          - match:
            - uri:
                exact: /productpage
            - uri:
                prefix: /static
            - uri:
                exact: /login
            - uri:
                exact: /logout
            - uri:
                prefix: /api/v1/products
            route:
            - destination:
                host: productpage
                port:
                  number: 9080
                subset: productpage-v1
              weight: 60
            - destination:
                host: productpage
                port:
                  number: 9080
                subset: productpage-v2
              weight: 40
        ```

=== "KFServing"

    ```shell
    kubectl apply -f $ITER8/samples/kfserving/quickstart/baseline.yaml
    kubectl apply -f $ITER8/samples/kfserving/quickstart/candidate.yaml
    kubectl apply -f $ITER8/samples/kfserving/fixed-split/routing-rule.yaml
    kubectl wait --for=condition=Ready isvc/flowers -n ns-baseline
    kubectl wait --for=condition=Ready isvc/flowers -n ns-candidate
    ```

    ??? info "Virtual service with traffic split"
        ```yaml linenums="1"
        apiVersion: networking.istio.io/v1alpha3
        kind: VirtualService
        metadata:
          name: routing-rule
          namespace: default
        spec:
          gateways:
          - knative-serving/knative-ingress-gateway
          hosts:
          - example.com
          http:
          - route:
            - destination:
                host: flowers-predictor-default.ns-baseline.svc.cluster.local
              headers:
                request:
                  set:
                    Host: flowers-predictor-default.ns-baseline
                response:
                  set:
                    version: flowers-v1
              weight: 60
            - destination:
                host: flowers-predictor-default.ns-candidate.svc.cluster.local
              headers:
                request:
                  set:
                    Host: flowers-predictor-default.ns-candidate
                response:
                  set:
                    version: flowers-v2
              weight: 40
        ```

=== "Knative"

    ```shell
    kubectl apply -f $ITER8/samples/knative/quickstart/baseline.yaml
    kubectl apply -f $ITER8/samples/knative/fixed-split/experimentalservice.yaml
    kubectl wait --for=condition=Ready ksvc/sample-app
    ```

    ??? info "Knative service with traffic split"
        ```yaml linenums="1"
        apiVersion: serving.knative.dev/v1
        kind: Service
        metadata:
          name: sample-app
          namespace: default
        spec:
          template:
            metadata:
              name: sample-app-v2
            spec:
              containers:
              - image: gcr.io/knative-samples/knative-route-demo:green 
                env:
                - name: T_VERSION
                  value: "green"
          traffic:
          - tag: current
            revisionName: sample-app-v1
            percent: 60
          - tag: candidate
            latestRevision: true
            percent: 40
        ```

## 5. Generate requests
Please follow [Step 5 of the quick start tutorial](../../../getting-started/quick-start/#5-generate-requests).

## 6. Define metrics
Please follow [Step 6 of the quick start tutorial](../../../getting-started/quick-start/#6-define-metrics).

## 7. Launch experiment
=== "Istio"

    ```shell
    kubectl apply -f $ITER8/samples/istio/fixed-split/experiment.yaml
    ```

    ??? info "Look inside experiment.yaml"
        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Experiment
        metadata:
          name: fixedsplit-exp
        spec:
          # target identifies the service under experimentation using its fully qualified name
          target: bookinfo-iter8/productpage
          strategy:
            # this experiment will perform an A/B test
            testingPattern: A/B
            # this experiment will not shift traffic during iterations
            deploymentPattern: FixedSplit
            actions:
              # when the experiment completes, promote the winning version using kubectl apply
              finish:
              - task: common/exec
                with:
                  cmd: /bin/bash
                  args: [ "-c", "kubectl -n bookinfo-iter8 apply -f {{ .promote }}" ]
          criteria:
            rewards:
            # (business) reward metric to optimize in this experiment
            - metric: iter8-istio/user-engagement 
              preferredDirection: High
            objectives: # used for validating versions
            - metric: iter8-istio/mean-latency
              upperLimit: 300
            - metric: iter8-istio/error-rate
              upperLimit: "0.01"
            requestCount: iter8-istio/request-count
          duration: # product of fields determines length of the experiment
            intervalSeconds: 10
            iterationsPerLoop: 10
          versionInfo:
            # information about the app versions used in this experiment
            baseline:
              name: productpage-v1
              variables:
              - name: namespace # used by final action if this version is the winner
                value: bookinfo-iter8
              - name: promote # used by final action if this version is the winner
                value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/istio/quickstart/vs-for-v1.yaml
            candidates:
            - name: productpage-v2
              variables:
              - name: namespace # used by final action if this version is the winner
                value: bookinfo-iter8
              - name: promote # used by final action if this version is the winner
                value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/istio/quickstart/vs-for-v2.yaml
        ```

=== "KFServing"

    ```shell
    kubectl apply -f $ITER8/samples/kfserving/fixed-split/experiment.yaml
    ```

    ??? info "Look inside experiment.yaml"
        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Experiment
        metadata:
          name: fixedsplit-exp
        spec:
          target: flowers
          strategy:
            testingPattern: A/B
            deploymentPattern: FixedSplit
            actions:
              # when the experiment completes, promote the winning version using kubectl apply
              finish:
              - task: common/exec
                with:
                  cmd: /bin/bash
                  args: [ "-c", "kubectl apply -f {{ .promote }}" ]
          criteria:
            requestCount: iter8-kfserving/request-count
            rewards: # Business rewards
            - metric: iter8-kfserving/user-engagement
              preferredDirection: High # maximize user engagement
            objectives:
            - metric: iter8-kfserving/mean-latency
              upperLimit: 1500
            - metric: iter8-kfserving/95th-percentile-tail-latency
              upperLimit: 2000
            - metric: iter8-kfserving/error-rate
              upperLimit: "0.01"
          duration:
            intervalSeconds: 10
            iterationsPerLoop: 25
          versionInfo:
            # information about model versions used in this experiment
            baseline:
              name: flowers-v1
              variables:
              - name: ns
                value: ns-baseline
              - name: promote
                value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v1.yaml
            candidates:
            - name: flowers-v2
              variables:
              - name: ns
                value: ns-candidate
              - name: promote
                value: https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/kfserving/quickstart/promote-v2.yaml
        ```

=== "Knative"

    ```shell
    kubectl apply -f $ITER8/samples/knative/fixed-split/experiment.yaml
    ```

    ??? info "Look inside experiment.yaml"
        ```yaml linenums="1"
        apiVersion: iter8.tools/v2alpha2
        kind: Experiment
        metadata:
          name: fixedsplit-exp
        spec:
          # target identifies the knative service under experimentation using its fully qualified name
          target: default/sample-app
          strategy:
            testingPattern: A/B
            deploymentPattern: FixedSplit
            actions:
              finish: # run the following sequence of tasks at the end of the experiment
              - task: common/exec # promote the winning version      
                with:
                  cmd: /bin/sh
                  args:
                  - "-c"
                  - |
                    kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml
          criteria:
            requestCount: iter8-knative/request-count
            rewards: # Business rewards
            - metric: iter8-knative/user-engagement
              preferredDirection: High # maximize user engagement
            objectives: 
            - metric: iter8-knative/mean-latency
              upperLimit: 50
            - metric: iter8-knative/95th-percentile-tail-latency
              upperLimit: 100
            - metric: iter8-knative/error-rate
              upperLimit: "0.01"
          duration:
            intervalSeconds: 10
            iterationsPerLoop: 10
          versionInfo:
            # information about app versions used in this experiment
            baseline:
              name: sample-app-v1
              variables:
              - name: promote
                value: baseline
            candidates:
            - name: sample-app-v2
              variables:
              - name: promote
                value: candidate
        ```

The process automated by Iter8 during this experiment is depicted below.

![Iter8 automation](../../images/fixedsplit-iter8-process.png)

## 8. Observe experiment
Please follow [Step 8 of the quick start tutorial](../../../getting-started/quick-start/#8-observe-experiment) to observe the experiment in realtime. Note that the experiment in this tutorial uses a different name from the quick start one. Replace the experiment name `quickstart-exp` with `fixedsplit-exp` in your commands. You can also observe traffic by suitably modifying the commands for observing traffic.


???+ info "Understanding what happened"
    1. You created two versions of your app/ML model.
    2. You generated requests for your app/ML model versions. At the start of the experiment, 60% of the requests are sent to the baseline and 40% to the candidate.
    3. You created an Iter8 experiment with A/B testing pattern and FixedSplit deployment pattern. In each iteration, Iter8 observed the latency and error-rate metrics collected by Prometheus, and the user-engagement metric from New Relic; Iter8 verified that the candidate satisfied all objectives, verified that the candidate improved over the baseline in terms of user-engagement, identified candidate as the winner, and finally promoted the candidate.

## 9. Cleanup
=== "Istio"
    ```shell
    kubectl delete -f $ITER8/samples/istio/fixed-split/experiment.yaml
    kubectl delete -f $ITER8/samples/istio/quickstart/fortio.yaml
    kubectl delete ns bookinfo-iter8
    ```

=== "KFServing"
    ```shell
    kubectl delete -f $ITER8/samples/kfserving/fixed-split/experiment.yaml
    kubectl delete -f $ITER8/samples/kfserving/fixed-split/routing-rule.yaml
    kubectl delete -f $ITER8/samples/kfserving/quickstart/candidate.yaml
    kubectl delete -f $ITER8/samples/kfserving/quickstart/baseline.yaml
    ```

=== "Knative"
    ```shell
    kubectl delete -f $ITER8/samples/knative/fixed-split/experiment.yaml
    kubectl delete -f $ITER8/samples/knative/quickstart/fortio.yaml
    kubectl apply -f $ITER8/samples/knative/fixed-split/experimentalservice.yaml
    ```
