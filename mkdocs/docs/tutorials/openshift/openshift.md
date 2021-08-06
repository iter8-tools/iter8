---
template: main.html
---

# CI/CD GitOps on Openshift

!!! tip "Openshift"
    Openshift shares many core components with Kubernetes and provides additional features that are catering specifically for enterprise usage. In such an environment, CI/CD pipelines are often automated to allow developers to quickly deliver new code to the production environment with certain level of assurance. In this tutorial, we assume the reader is already using Openshift, managing CI/CD pipelines with Openshift Pipeline (Tekton) and Openshift GitOps (ArgoCD), and has a GitOps setup to allow newly built images to be deployed into the cluster. We will describe how such users could integrate Iter8 into their existing pipeline in a minimally intrusive way to allow their applications to be progressively rolled out. 

This tutorial assumes a basic understanding of Iter8. See, for example, the Istio [quick start tutorial](../istio/quick-start.md).

## 1. Install Iter8

Installing Iter8 on Openshift is slightly different from [installing it on K8s](../../getting-started/install.md). After Iter8 is installed, an extra step one needs to perform is:

```shell
oc adm policy add-scc-to-group anyuid system:serviceaccounts:istio-system
```

## 2. Openshift Metrics

Openshift comes with a built-in Prometheus server, which is authenticated with a sidecar proxy. To allow Iter8 to retrieve metrics from it, one needs to provide a mean to authenticate. Iter8 currently supports basic authentication and bearer token, described in [Iter8 metrics](../../metrics/custom.md).

An example `request count` metric looks like:

```yaml
apiVersion: iter8.tools/v2alpha2
kind: Metric
metadata:
  labels:
    creator: iter8
  name: request-count
  namespace: default
spec:
  description: Number of requests
  jqExpression: .data.result[0].value[1] | tonumber
  params:
  - name: query
    value: |
      sum(increase(haproxy_server_http_responses_total{exported_service='$name',exported_namespace='$namespace'}[${elapsedTime}s])) or vector(0)
  provider: prometheus
  type: Counter
  urlTemplate: https://prometheus-operated.openshift-monitoring:9091/api/v1/query
  authType: Bearer
  secret: default/promsecret
  headerTemplates:
  - name: Authorization
    value: Bearer ${token}

```

## 3. Augment the CI Pipeline with Iter8

The CI pipeline gets triggered when new code is merged, and it is responsible for testing and building the new code. The end product is usually a newly build image pushed to an image repository.  When one wants to leverage Iter8, one should create a Candidate Deployment (using the new image) and an Iter8 Experiment CR. Subsequently, Iter8 will observe metrics collected from the Candidate version and check if all success criteria have passed. If so, the candidate version will be promoted. Here is an example Tekton task that creates these resources in Git.

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: start-experiment
  annotations:
    tekton.dev/displayName: "Start an Iter8 experiment"
spec:
  description: >-
    This task create a candidate and Iter8 experiment resource in
    the Env repo and make a PR from the changes

  params:
    - name: USER
      description: Github username
    - name: REPO
      description: Github repo name
    - name: BRANCH
      description: Base branch PR is opened against
    - name: GITHUB-TOKEN-SECRET
      description: Holds Github token

  steps:
  - name: start-experiment
    image: alpine/git:latest
    script: |
      #!/usr/bin/env sh
      apk add curl jq make
      git config --global user.email 'iter8@iter8.tools'
      git config --global user.name 'Iter8'
      git clone https://$(params.USER):${GITHUB_TOKEN}@github.com/$(params.USER)/$(params.REPO) --branch=$(params.BRANCH)
      [create a Deployment Candidate yaml]
      [create an Iter8 Experiment yaml]
      git add -A
      git commit -a -m 'start Iter8 experiment'
      git push -f origin $(params.BRANCH)
    env:
    - name: GITHUB_TOKEN
      valueFrom:
        secretKeyRef:
           name: $(params.GITHUB-TOKEN-SECRET)
           key: token
```

In a non-GitOps environment, the `git` commands will be replaced with `oc` commands to directly deploy into the cluster. An example Tekton Pipeline where the above Task can be added to is:

```yaml
  - name: start-experiment
    taskRef:
      name: start-experiment
    params:
    - name: USER
      value: [YOUR GIT USERNAME]
    - name: REPO
      value: [YOUR GIT REPO]
    - name: BRANCH
      value: master
    - name: GITHUB-TOKEN-SECRET
      value: github-token
```
