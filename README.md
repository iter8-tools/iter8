# Iter8

<img alt="Iter8" src="mkdocs/docs/images/favicon.png" width="100" align="left">

## Kubernetes Release Engineering

[![GitHub stars](https://img.shields.io/github/stars/iter8-tools/iter8?style=social)](https://github.com/iter8-tools/iter8/stargazers)
[![Slack channel](https://img.shields.io/badge/Slack-Join-purple)](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
[![Community meetups](https://img.shields.io/badge/meet-Iter8%20community%20meetups-brightgreen)](https://iter8.tools/0.7/getting-started/help/#iter8-community-meetings)
[![GitHub issues](https://img.shields.io/github/issues/iter8-tools/iter8)](https://github.com/iter8tools/iter8/issues)

> Safely rollout new versions of apps and ML models. Maximize business value with each release.


## Use Cases

1.  Load testing with SLOs
2.  A/B(/n) testing for improving business value with each release of app/ML model
3.  Safe rollout for multi-cluster and edge
4.  Traffic mirroring experiments

The traffic mirroring use-case is achieved by using Iter8 along with a Kubernetes service mesh or ingress that supports mirroring.

## Quick Start

### 1. Install Iter8
#### Using Brew
```shell
brew tap iter8-tools/iter8
brew install iter8
```

#### Using Go 1.16+
```shell
go install github.com/iter8-tools/iter8@latest
```
You can now run `iter8` (from your gopath bin/ directory)

#### Using pre-compiled binary
Pre-compiled Iter8 binaries for many platforms are available [here](https://github.com/iter8-tools/iter8/releases). Uncompress the iter8-X-Y.tar.gz archive for your platform, and move the `iter8` binary to any folder in your PATH.

## 2. Download experiment chart
Download the `load-test` experiment chart from Iter8 hub as follows.

```shell
iter8 hub -e load-test
```

This creates a local folder called `load-test` containing the chart.

## 3. Run experiment
The `iter8 run` command generates the `experiment.yaml` file from an experiment chart, runs the experiment, and writes the results of the experiment into the `result.yaml` file. Run the load test experiment as follows.

```shell
cd load-test
iter8 run --set url=https://example.com
```

## 4. Assert outcomes
Assert that the experiment completed without any failures and SLOs are satisfied.

```shell
iter8 assert -c completed -c nofailure -c slos
```

The `iter8 assert` subcommand asserts if experiment result satisfies the specified conditions. 
If assert conditions are satisfied, it exits with code `0`, and exits with code `1` otherwise. Assertions are especially useful within CI/CD/GitOps pipelines.

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

## Documentation
Iter8 documentation is available at https://iter8.tools.

## Contributing
We are delighted that you want to contribute to Iter8! 💖

As you get started, you are in the best position to give us feedback on areas of
our project that we need help with including:

* Problems found during setup of Iter8
* Gaps in our quick start tutorial and other documentation
* Bugs in our test and automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a
bug report and let us know!

See [here](https://iter8.tools/latest/contributing/overview/) for information about ways to contribute, Iter8 community meetings, finding an issue, asking for help, pull-request lifecycle, and more.
