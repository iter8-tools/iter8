# Overview

Welcome! We are delighted that you want to contribute to Iter8! 💖

As you get started, you are in the best position to give us feedback on key areas including:

* Problems found during setup of Iter8
* Gaps in our getting started tutorial and other documentation
* Bugs in our test and automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a bug report and let us know!

## Ways to contribute

We welcome many types of contributions including:

* [CLI and Iter8 experiment charts](#iter8-toolsiter8)
* [Docs](#iter8-toolsdocs)
* CI, builds, and tests
* Reviewing pull requests

## Ask for help

The best ways to reach us with a question is to ask...

* On the original GitHub issue
* In the `#development` channel in the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
* During our [community meetings](https://iter8.tools/latest/community/community/)

## Find an issue

Iter8 issues are tracked [here](https://github.com/iter8-tools/iter8/issues).

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

The Iter8 project consists of the following repos.

1. [iter8-tools/iter8](https://github.com/iter8-tools/iter8): source for the Iter8 CLI
2. [iter8-tools/hub](https://github.com/iter8-tools/hub): source for Iter8 experiment and supplementary charts
3. [iter8-tools/docs](https://github.com/iter8-tools/docs): source for Iter8 docs
4. [iter8-tools/homebrew-iter8](https://github.com/iter8-tools/homebrew-iter8): Homebrew formula for the Iter8 CLI

### iter8-tools/iter8

This is the source repo for Iter8 CLI.

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

#### Run unit tests and see coverage information
```shell
make tests
make coverage
make htmlcov
```

#### Lint Iter8
```shell
make lint
```

#### Build and push Iter8 image

Define a name for your Docker image

```shell
IMG=[Docker image name]
```

Build and push Iter8 image to Docker

```shell
make dist
docker build -f Dockerfile.dev -t $IMG .
docker push $IMG
```

### iter8-tools/docs

This is the source repo for Iter8 documentation.

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
