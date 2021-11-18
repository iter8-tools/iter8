# Iter8

<img alt="Iter8" src="mkdocs/docs/images/favicon.png" width="100" align="left">

## Metrics Driven Experiments

[![GitHub stars](https://img.shields.io/github/stars/iter8-tools/iter8?style=social)](https://github.com/iter8-tools/iter8/stargazers)
[![Slack channel](https://img.shields.io/badge/Slack-Join-purple)](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
[![Community meetups](https://img.shields.io/badge/meet-Iter8%20community%20meetups-brightgreen)](https://iter8.tools/0.7/getting-started/help/#iter8-community-meetings)
[![GitHub issues](https://img.shields.io/github/issues/iter8-tools/iter8)](https://github.com/iter8tools/iter8/issues)

Iter8 is a metrics-driven experimentation platform that enables DevOps/SRE/MLOps/data science teams to maximize release velocity and business value of apps and ML models while protecting end-user experience.

Iter8 enables the following use-cases.

1. Load testing/service level objectives (SLOs) with built-in metrics
2. A/B(/n) testing with business metrics from any backend
3. SLOs with metrics from any backend
4. Mirroring
5. User segmentation
6. Session affinity
7. Gradual rollout

The traffic engineering use-cases (4 - 7 above) are achieved by using Iter8 along with a Kubernetes service-mesh or ingress.

<img alt="Iter8" src="images/ghbanner.png" align="center">

## Your First Experiment

1. Install Iter8 using [Go 1.16+](https://golang.org/) as follows.
```shell
GOBIN=/usr/local/bin/ go install github.com/iter8-tools/iter8@latest
```
`GOBIN` may be any folder in your `PATH`.

2. Download the `load-test` experiment folder from the [Iter8 hub](../user-guide/topics/iter8hub.md) as follows.

```shell
iter8 hub -e load-test
```

3. The `iter8 run` command reads the experiment specified in the `experiment.yaml` file, runs the experiment, and writes the result of the experiment into the `result.yaml` file. Run `load-test` as follows.

```shell
cd load-test
iter8 run
```

4. The experiment should complete in a few seconds. Upon completion, assert that the experiment completed without any failures, and SLOs are satisfied, as follows.

```shell
iter8 assert -c completed -c nofailure -c slos
```

5. Generate a report of the experiment including a summary of the experiment, SLOs, and metrics.

```shell
iter8 gen 
```

<details>
  <summary>Look inside a sample report</summary>

  ```
      -----------------------------|-----
                  Experiment summary|
      -----------------------------|-----
              Experiment completed |true
      -----------------------------|-----
                  Experiment failed |false
      -----------------------------|-----
          Number of completed tasks |2
      -----------------------------|-----



      -----------------------------|-----
                                SLOs|
      -----------------------------|-----
            built-in/error-rate <= 0|true
      -----------------------------|-----
              built-in/p95.0 <= 100|true
      -----------------------------|-----


      -----------------------------|-----
                            Metrics|
      -----------------------------|-----
                built-in/error-count|0
      -----------------------------|-----
                built-in/error-rate|0
      -----------------------------|-----
                built-in/max-latency|201.75 (msec)
      -----------------------------|-----
              built-in/mean-latency|17.02 (msec)
      -----------------------------|-----
                built-in/min-latency|3.80 (msec)
      -----------------------------|-----
                      built-in/p50.0|10.75 (msec)
      -----------------------------|-----
                      built-in/p75.0|12.12 (msec)
      -----------------------------|-----
                      built-in/p90.0|13.88 (msec)
      -----------------------------|-----
                      built-in/p95.0|15.60 (msec)
      -----------------------------|-----
                      built-in/p99.0|201.31 (msec)
      -----------------------------|-----
                      built-in/p99.9|201.71 (msec)
      -----------------------------|-----
              built-in/request-count|100
      -----------------------------|-----
            built-in/stddev-latency|37.81 (msec)
      -----------------------------|-----
  ```
</details>

Congratulations :tada: You completed your first Iter8 experiment.