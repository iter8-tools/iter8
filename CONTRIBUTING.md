# Overview

Welcome! We are delighted that you want to contribute to Iter8! 💖

As you get started, you are in the best position to give us feedback on key areas including:

* Problems found during setup of Iter8
* Gaps in our quick start tutorial and other documentation
* Bugs in our test and automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a
bug report and let us know!


## Ways to contribute

We welcome many different types of contributions including:

* [Tutorials and other documentation](#iter8-docs)
* [Experiment charts](#iter8-hub)
* [Iter8 CLI features](#iter8-cli)
* CI, builds, and tests
* [Web design](#iter8-docs)
* Reviewing pull requests
* Communication, social media, blog posts

## Ask for help

The best ways to reach us with a question is to ask...

* On the original GitHub issue
* In the `#development` channel in the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
* During our [community meetings](https://iter8.tools/community/)

## Find an issue

Iter8 issues are managed [here](https://github.com/iter8-tools/iter8/issues).

Issued labeled **good first issue** have extra information to
help you make your first contribution. Issues labeled **help wanted** are issues
suitable for someone who has already submitted their first pull request and is good to move on to the second one.

Sometimes there won’t be any issues with these labels. That’s ok! There is
likely still something for you to work on. If you want to contribute but you
don’t know where to start or can't find a suitable issue, you can reach out to us over the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) for help finding something to work on.

Once you see an issue that you'd like to work on, please post a comment saying
that you want to work on it. Something like "I want to work on this" is fine.

## Pull request lifecycle

* Your PR is associated with one (and infrequently, with more than one) [GitHub issue](https://github.com/iter8-tools/iter8/issues). You can start the submission of your PR as soon as this issue has been created.
* Follow the [standard GitHub fork and pull request process](https://gist.github.com/Chaser324/ce0505fbed06b947d962) when creating and submitting your PR.
* The associated GitHub issue might need to go through design discussions and may not be ready for development. Your PR might require new tests; these new or existing tests may not yet be running successfully. At this stage, [keep your PR as a draft](https://github.blog/2019-02-14-introducing-draft-pull-requests/), to signal that it is not yet ready for review.
* Once design discussions are complete and tests pass, convert the draft PR into a regular PR to signal that it is ready for review. Additionally, post a message in the `#development` Slack channel of the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) with a link to your PR. This will expedite the review.
* You can expect an initial review within 1-2 days of submitting a PR, and follow up reviews (if any) to happen over 2-5 days.
* Use the `#development` Slack channel of [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) to ping/bump when the pull request is ready for further review or if it appears stalled.
* Iter8 releases happen frequently. Once your PR is merged, you can expect your contribution to show up *live* in a short amount of time at https://iter8.tools.

## Sign Your Commits

Licensing is important to open source projects. It provides some assurances that
the software will continue to be available based under the terms that the
author(s) desired. We require that contributors sign off on commits submitted to
our project's repositories. The [Developer Certificate of Origin
(DCO)](https://developercertificate.org/) is a way to certify that you wrote and
have the right to contribute the code you are submitting to the project.

Read [GitHub's documentation on signing your commits](https://docs.github.com/en/github/authenticating-to-github/managing-commit-signature-verification/signing-commits).

You sign-off by adding the following to your commit messages. Your sign-off must
match the Git user and email associated with the commit.

    This is my commit message

    Signed-off-by: Your Name <your.name@example.com>

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running:

    git commit --amend -s 


## Development environment setup

The Iter8 project consists of three main repos.

1. [Iter8](https://github.com/iter8-tools/iter8): source for Iter8 CLI
2. [Hub](https://github.com/iter8-tools/hub): source for Iter8 experiment charts
3. [Docs](https://github.com/iter8-tools/docs): source for Iter8 docs

### Iter8 CLI

This is the source for the Iter8 CLI.

#### Clone `iter8`

```shell
git clone https://github.com/iter8-tools/iter8.git
```

#### Build Iter8
```shell
make build
```

#### Install Iter8 locally
```shell
make clean install
iter8 version
```

#### Run tests and see coverage for Iter8
```shell
make tests
make coverage
make htmlcov
```

### Iter8 hub

This is the source for Iter8 experiment charts.

#### Clone `hub`

```shell
git clone https://github.com/iter8-tools/hub.git
```

#### Add tests
Add integration tests for Iter8 hub in the `.github/workflows/tests.yaml` file.

#### Versioning
Iter8 experiment charts are Helm charts under the covers, and are semantically versioned as per [Helm chart versioning specifications](https://helm.sh/docs/topics/charts/#charts-and-versioning). Every change to the chart must be accompanied by an increment to the version number of the chart. For most changes, this would mean an increment to the patch version (for example, the `version` field in `Chart.yaml` might need to be incremented from `0.1.0` to `0.1.1`).


### Iter8 docs
This is the source for Iter8 documentation. Uses [Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/).

#### Clone `docs`

```shell
git clone https://github.com/iter8-tools/docs.git
```

#### Locally serve docs
From the root of this repo:

```shell
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
mkdocs serve -s
```

You can now see your local docs at [http://localhost:8000](http://localhost:8000). You will also see live updates to [http://localhost:8000](http://localhost:8000) as you update the contents of the `docs` folder.

#### Add tests
Add end-to-end tests for Iter8 docs in the `.github/workflows/tests.yaml` file.


<!-- ## Pull Request Checklist

When you submit your pull request, or you push new commits to it, our automated
systems will run some checks on your new code. We require that your pull request
passes these checks, but we also have more criteria than just that before we can
accept and merge it. We recommend that you check the following things locally
before you submit your code:

**TODO** -->
<!-- list both the automated and any manual checks performed by reviewers, it
is very helpful when the validations are automated in a script for example in a
Makefile target. Below is an example of a checklist:

* It passes tests: run the following command to run all of the tests locally:
  `make build test lint`
* Impacted code has new or updated tests
* Documentation created/updated
* We use [Azure DevOps, GitHub Actions, CircleCI] to test all pull
  requests. We require that all tests succeed on a pull request before it is merged.

-->
