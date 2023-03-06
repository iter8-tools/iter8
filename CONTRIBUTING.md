# Overview

Welcome! We are delighted that you want to contribute to Iter8! 💖

As you get started, you are in the best position to give us feedback on key areas including:

* Problems found during setup of Iter8
* Gaps in our getting started tutorial and other documentation
* Bugs in our test and automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a
bug report and let us know!

## Find an issue

Iter8 issues are tracked [here](https://github.com/iter8-tools/iter8/issues). If the issue you wish to tackle is not on the repo, please create one.

The best ways to reach us with a question is to ask...

* On the original GitHub issue
* In the `#development` channel in the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
* During our [community meetings](https://iter8.tools/latest/community/community/)

## Development environment setup

The Iter8 project consists of the following repos.

1. [iter8-tools/iter8](https://github.com/iter8-tools/iter8): source for the Iter8 CLI, Iter8 service and Iter8 experiment chart
2. [iter8-tools/docs](https://github.com/iter8-tools/docs): source for Iter8 docs

### iter8-tools/iter8

[This](https://github.com/iter8-tools/iter8) is the GitHub repo for Iter8 CLI, Iter8 service, and Iter8 experiment chart.

#### Build Iter8

```shell
go build
```

#### Install Iter8 locally

```shell
go install
```

#### Run unit tests and see coverage information

```shell
make test
make coverage
make htmlcov
```

### iter8-tools/docs

[This](https://github.com/iter8-tools/docs) is the GitHub repo for Iter8 documentation.

#### Locally serve docs
From the root of this repo:

```shell
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
mkdocs serve -s
```

You should now see your local docs at [http://localhost:8000](http://localhost:8000). You should also see live updates to [http://localhost:8000](http://localhost:8000) as you update the contents of the `docs` folder.

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

