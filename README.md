# Iter8

<img alt="Iter8" src="mkdocs/docs/images/favicon.png" width="100" align="left">

## Metrics Driven Experiments

[![GitHub stars](https://img.shields.io/github/stars/iter8-tools/iter8?style=social)](https://github.com/iter8-tools/iter8/stargazers)
[![Slack channel](https://img.shields.io/badge/Slack-Join-purple)](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
[![Community meetups](https://img.shields.io/badge/meet-Iter8%20community%20meetups-brightgreen)](https://iter8.tools/0.7/getting-started/help/#iter8-community-meetings)
[![GitHub issues](https://img.shields.io/github/issues/iter8-tools/iter8)](https://github.com/iter8tools/iter8/issues)

### Kubernetes-friendly metrics-driven <strong>experiments</strong> and <strong>safe rollouts</strong>. 

## Use Cases

1.  Load testing with SLOs
2.  A/B(/n) testing with business reward metrics
3.  SLOs with metrics from any backend
4.  Traffic mirroring
5.  User segmentation
6.  Session affinity
7.  Gradual rollout

The traffic engineering use-cases (4 - 7 above) are achieved by using Iter8 along with a Kubernetes service mesh or ingress.

## Quick Start

### 1. Install Iter8
Install Iter8 using [Go 1.16+](https://golang.org/) as follows.
```shell
go install github.com/iter8-tools/iter8@latest
```
You can now run `iter8` (from your gopath bin/ directory)

## 2. Download experiment
Download the `load-test` experiment chart from Iter8 hub as follows.

```shell
iter8 hub -e load-test
```

## 3. Run experiment
Iter8 experiments are specified using the `experiment.yaml` file. The `iter8 run` command reads this file, runs the specified experiment, and writes the results of the experiment into the `result.yaml` file.

Run the experiment you downloaded above as follows.

```shell
cd load-test
iter8 run
```

## 4. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

## 5. Generate report
Generate a report of the experiment in HTML or text formats as follows.

### HTML Report

```shell
iter8 report -o html > report.html
# open report.html with a browser. In MacOS, you can use the command:
# open report.html
```

The HTML report looks as follows.

![HTML report](mkdocs/docs/getting-started/images/report.html.png)

### Text Report

```shell
iter8 report -o text
```

Congratulations! :tada: You completed your first Iter8 experiment.

## [Documentation](https://iter8.tools)

## [Contributing](https://iter8.tools/latest/contributing/overview/)
