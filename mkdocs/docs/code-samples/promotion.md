---
template: overrides/main.html
hide:
- toc
---

# Promote `winner` using Helm, Kustomize, YAML/JSON artifacts

!!! abstract ""
    At the end of an experiment, iter8 can `promote` the `winner` by configuring Kubernetes resources. Iter8 can use Helm Charts, Kustomize resources, or explicitly supplied YAML/JSON artifacts for this task.

=== "Helm"
    [This tutorial](/code-samples/iter8-knative/canary-progressive/) illustrates the use of `helm upgrade` command within a `Canary` experiment to promote the `winner`. Expand the callout below to see experiment details.

    ??? info "Experiment with `winner` promotion using `helm upgrade`"
        ``` yaml linenums="1" hl_lines="22 23 24 25 26 27 28 29 30 31 32 33 34 35 36"
        apiVersion: iter8.tools/v2alpha1
        kind: Experiment
        metadata:
          name: canary-progressive
        spec:
          # target identifies the knative service under experimentation using its fully qualified name
          target: default/sample-app
          strategy:
            # this experiment will perform a canary test
            testingPattern: Canary
            deploymentPattern: Progressive
            weights: # fine-tune traffic increments to candidate
              # candidate weight will not exceed 75 in any iteration
              maxCandidateWeight: 75
              # candidate weight will not increase by more than 20 in a single iteration
              maxCandidateWeightIncrement: 20
            actions:
              start: # run the following sequence of tasks at the start of the experiment
              - library: knative
                task: init-experiment
              finish: # run the following sequence of tasks at the end of the experiment
              - library: common
                task: exec # promote the winning version using Helm upgrade
                with:
                  cmd: helm
                  args:
                  - "upgrade"
                  - "--install"
                  - "--repo"
                  # repo url
                  - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/helm-repo" 
                  - "sample-app" # release name
                  - "--namespace=iter8-system" # helm secrets/release namespace
                  - "sample-app" # chart name
                  # values URL is dynamically interpolated
                  - "--values=https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/canaryprogressive/{{ .promote }}-values.yaml"
          criteria:
            # mean latency of version should be under 50 milliseconds
            # 95th percentile latency should be under 100 milliseconds
            # error rate should be under 1%
            objectives: 
            - metric: mean-latency
              upperLimit: 50
            - metric: 95th-percentile-tail-latency
              upperLimit: 100
            - metric: error-rate
              upperLimit: "0.01"
          duration:
            intervalSeconds: 10
            iterationsPerLoop: 7
          versionInfo:
            # information about app versions used in this experiment
            baseline:
              name: current
              variables:
              - name: revision
                value: sample-app-v1 
              - name: promote
                value: baseline
            candidates:
            - name: candidate
              variables:
              - name: revision
                value: sample-app-v2
              - name: promote
                value: candidate 
        ```

=== "Kustomize"
    [This tutorial](/code-samples/iter8-knative/canary-fixedsplit/) illustrates the use of `kustomize build` followed by the `kubectl apply` within a `Canary` experiment to promote the `winner`. Expand the callout below to see experiment details.

    ??? info "Experiment with `winner` promotion using `kustomize build`"
        ``` yaml linenums="1" hl_lines="17 18 19 20 21 22 23 24 25 26 27"
        apiVersion: iter8.tools/v2alpha1
        kind: Experiment
        metadata:
          name: canary-fixedsplit
        spec:
          # target identifies the knative service under experimentation using its fully qualified name
          target: default/sample-app
          strategy:
            # this experiment will perform a canary test
            testingPattern: Canary
            deploymentPattern: FixedSplit
            actions:
              start: # run the following sequence of tasks at the start of the experiment
              - library: knative
                task: init-experiment
              finish: # run the following sequence of tasks at the end of the experiment
              - library: common
                task: exec # promote the winning version using kustomize
                with:
                  cmd: /bin/sh
                  args:
                  - "-c"
                  # {{ .name }} is dynamically interpolated based on the winner
                  - |
                  kustomize build \
                  github.com/iter8-tools/iter8/samples/knative/canaryfixedsplit/{{ .name }}?ref=master \
                  | kubectl apply -f -
          criteria:
            # mean latency of version should be under 50 milliseconds
            # 95th percentile latency should be under 100 milliseconds
            # error rate should be under 1%
            objectives: 
            - metric: mean-latency
              upperLimit: 50
            - metric: 95th-percentile-tail-latency
              upperLimit: 100
            - metric: error-rate
              upperLimit: "0.01"
          duration:
            intervalSeconds: 20
            iterationsPerLoop: 12
          versionInfo:
            # information about app versions used in this experiment
            baseline:
              name: baseline
              variables:
              - name: revision
                value: sample-app-v1 
            candidates:
            - name: candidate
              variables:
              - name: revision
                value: sample-app-v2
        ```

=== "YAML/JSON files"
    [This tutorial](/getting-started/quick-start/with-knative/) illustrates the use of the `kubectl apply` using YAML-files within a `Canary` experiment to promote the `winner`. Expand the callout below to see experiment details.

    ??? info "Experiment with `winner` promotion using YAML files and `kubectl apply`"
        ``` yaml linenums="1" hl_lines="17 18 19 20 21 22 23 24 25"
        apiVersion: iter8.tools/v2alpha1
        kind: Experiment
        metadata:
          name: quickstart-exp
        spec:
          # target identifies the knative service under experimentation using its fully qualified name
          target: default/sample-app
          strategy:
            # this experiment will perform a canary test
            testingPattern: Canary
            deploymentPattern: Progressive
            actions:
              start: # run the following sequence of tasks at the start of the experiment
              - library: knative
                task: init-experiment
              finish: # run the following sequence of tasks at the end of the experiment
              - library: common
                task: exec # promote the winning version
                with:
                  cmd: kubectl
                  args: 
                  - "apply"
                  - "-f"
                  # {{ .promote }} is dynamically interpolated based on winner
                  - "https://raw.githubusercontent.com/iter8-tools/iter8/master/samples/knative/quickstart/{{ .promote }}.yaml"
          criteria:
            # mean latency of version should be under 50 milliseconds
            # 95th percentile latency should be under 100 milliseconds
            # error rate should be under 1%
            objectives: 
            - metric: mean-latency
              upperLimit: 50
            - metric: 95th-percentile-tail-latency
              upperLimit: 100
            - metric: error-rate
              upperLimit: "0.01"
          duration:
            intervalSeconds: 10
            iterationsPerLoop: 10
          versionInfo:
            # information about app versions used in this experiment
            baseline:
              name: current
              variables:
              - name: revision
                value: sample-app-v1 
              - name: promote
                value: baseline
            candidates:
            - name: candidate
              variables:
              - name: revision
                value: sample-app-v2
              - name: promote
                value: candidate 
        ```

