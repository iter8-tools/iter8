---
template: main.html
---

# What is Iter8?

Iter8 enables **DevOps/SRE/MLOps/data science teams** to maximize release velocity and business value of apps and ML models while protecting end-user experience.

Use **Iter8 experiments** for load testing, SLO validation, canary releases, A/B(/n) testing different versions of apps and identifying a winner based on business metrics, chaos testing, and hybrids of the above.

***

## What is an Iter8 experiment?
Iter8 defines the concept of an experiment that automates various aspects of the application release process as shown below.[^1]

![Process automated by an Iter8 experiment](../images/whatisiter8.png)

***

## How is an Iter8 experiment specified?
Iter8 experiment is specified in the form of a YAML file as shown in the following example.

```yaml
name: my-exp
spec:
  iter8Version: "0.8.0"
  # versions of the app that are assessed in this experiment
  versions: ["v1"]
  tasks:
  # generate requests for the app URL and collect built-in metrics
  - task: collect-fortio-metrics
    with:
      versionInfo:
      - url: https://example.com
  # assess how app versions are performing relative to criteria
  - task: assess-versions
    with:
      criteria:
        requestCount: iter8-fortio/request-count
        objectives:
        - metric: iter8-fortio/error-rate
          upperLimit: 0
        - metric: iter8-fortio/p95
          upperLimit: 100
# result: null
#   Do not provide this section when you define an experiment.
#   This section is reserved for Iter8's internal usage. 
#   Iter8 will populate the result section during the course of the experiment.
```

The core of the specification is a sequence of tasks which are executed by Iter8 during the experiment. Iter8 provides a variety specialized tasks that can collect built-in metrics, query metrics from databases or app endpoints, assess how the versions are performing relative to the assessment criteria, and compute optional traffic splits between versions. Iter8 also provides a generic task for running a bash script, which can be used for sending notifications, checking readiness conditions for apps, applying traffic splits, automatically triggering GitHub action workflows, and creating pull requests based on the results of the experiment.

## How are experiments run?
Iter8 provides a command line utility for running experiments locally.

```shell
# this will run experiment.yaml
iter8 run
```

Iter8 experiments can also be run inside a Kubernetes cluster in the form of a job, as a step within a GitHub actions workflow, in any environment that can run a Docker image (such as a Tekton task), or in any environment that can run an executable built by the `go` compiler.

***

## Can Iter8 be used within [... my unique environment]?
Iter8 can be used with:

  * any app/serverless/ML framework
  * any metrics provider
  * any service mesh/ingress technology for managing traffic, and 
  * any CI/CD/GitOps process.

## How is Iter8 implemented?

Iter8 is implemented as `go` module.

[^1]: Tasks with squiggly and dashed boundaries are optional.