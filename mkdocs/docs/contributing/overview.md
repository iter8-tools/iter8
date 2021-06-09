---
template: main.html
---

# Overview

Welcome! We are delighted that you want to contribute to Iter8! 💖

As you get started, you are in the best position to give us feedback on areas of
our project that we need help with including:

* Problems found during setup of Iter8
* Gaps in our quick start guide or other tutorials and documentation
* Bugs in our test and automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a
bug report and let us know!

***

## Ways to Contribute

We welcome many different types of contributions including:

* [Iter8 documentation/tutorials](tutorials.md)
* New features
    * [New K8s stack/service mesh/ingress](newk8sstack.md)
    * Iter8 code samples for OpenShift
    * [Analytics](analytics.md)
    * Tasks
* Builds, CI
* Bug fixes
* Web design for https://iter8.tools
* Communications/social media/blog posts
* Reviewing pull requests

Not everything happens through a GitHub pull request. Please come to our
[community meetings](#come-to-weekly-community-meetings) or [contact us](../getting-started/help.md) and let's discuss how we can work together. 

*** 

## Come to Weekly Community Meetings!
- Everyone is welcome to join our weekly community meetings. You never need an
invite to attend. In fact, we want you to join us, even if you don’t have
anything you feel like you want to contribute. Just being there is enough!
- Our community meetings are on Wednesdays at 1PM EST/EDT. You can find the video
conference link [here](https://meet.google.com/ocg-bzoa-gqe). You can find the
agenda [here](https://drive.google.com/drive/folders/1khHMD7JKt-kNAkbLgiIfwDbJYAhHMkyg?usp=sharing) as well as the meeting notes [here](https://drive.google.com/drive/folders/1rVrheoO-nfUI7JdPEgCqM39rl59KrPBk?usp=sharing).
- If you have a topic you would like to discuss, please put it on the agenda as 
well as a time estimate. Community meetings will be recorded
and publicly available on our [YouTube channel](https://www.youtube.com/channel/UCVybpnQAhr1o-QRPHBNdUgg).

***

## Find an Issue

Iter8 issues are managed centrally [here](https://github.com/iter8-tools/iter8/issues).

We have good first issues for new contributors and help wanted issues suitable
for any contributor. Issued labeled **good first issue** have extra information to
help you make your first contribution. Issues labeled **help wanted** are issues
suitable for someone who isn't a core maintainer and is good to move onto after
your first pull request.

Sometimes there won’t be any issues with these labels. That’s ok! There is
likely still something for you to work on. If you want to contribute but you
don’t know where to start or can't find a suitable issue, you can reach out to us over the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) for help finding something to work on.

Once you see an issue that you'd like to work on, please post a comment saying
that you want to work on it. Something like "I want to work on this" is fine.

***

## Ask for Help

The best ways to reach us with a question when contributing is to ask on:

* The original GitHub issue
* `#development` channel in the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)
* Bring your questions to our [community meetings](#come-to-weekly-community-meetings)

## Pull Request Lifecycle

* Your PR is associated with one (and infrequently, with more than one) [GitHub issue](https://github.com/iter8-tools/iter8/issues). You can start the submission of your PR as soon as this issue has been created.
* Follow the [standard GitHub fork and pull request process](https://gist.github.com/Chaser324/ce0505fbed06b947d962) when creating and submitting your PR.
* The associated GitHub issue might need to go through design discussions and may not be ready for development. Your PR might require new tests; these new or existing tests may not yet be running successfully. At this stage, [keep your PR as a draft](https://github.blog/2019-02-14-introducing-draft-pull-requests/), to signal that it is not yet ready for review.
* Once design discussions are complete and tests pass, convert the draft PR into a regular PR to signal that it is ready for review. Additionally, post a message in the `#development` Slack channel of the [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) with a link to your PR. This will expedite the review.
* You can expect an initial review within 1-2 days of submitting a PR, and follow up reviews (if any) to happen over 2-5 days.
* Use the `#development` Slack channel of [Iter8 Slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) to ping/bump when the pull request is ready for further review or if it appears stalled.
* Iter8 releases happen frequently. Once your PR is merged, you can expect your contribution to show up *live* in a short amount of time at https://iter8.tools.


<!-- ## Development Environment Setup

**TODO** -->
<!-- Provide enough information so that someone can find your project on 
the weekend and get set up, build the code, test it and submit a pull request 
successfully without having to ask any questions. If there is a one-off tool
they need to install, of common error people run into, or useful script they
should run, document it here. 

Document any necessary tools, for example VS Code and recommended extensions.
You don’t have to document the beginner’s guide to these tools, but how they
are used within the scope of your project.

* How to get the source code
* How to get any dependencies
* How to build the source code
* How to run the project locally
* How to test the source code, unit and "integration" or "end-to-end"
* How to generate and preview the documentation locally
* Links to new user documentation videos and examples to get people started and
  understanding how to use the project

-->
***

## Sign Your Commits

Licensing is important to open source projects. It provides some assurances that
the software will continue to be available based under the terms that the
author(s) desired. We require that contributors sign off on commits submitted to
our project's repositories. The [Developer Certificate of Origin
(DCO)](https://developercertificate.org/) is a way to certify that you wrote and
have the right to contribute the code you are submitting to the project.

Read [GitHub's documentation on signing your commits](https://docs.github.com/en/github/authenticating-to-github/managing-commit-signature-verification/signing-commits).

You sign-off by adding the following to your commit messages. Your sign-off must
match the git user and email associated with the commit.

    This is my commit message

    Signed-off-by: Your Name <your.name@example.com>

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running 

    git commit --amend -s 

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
* We use [Azure DevOps, GitHub Actions, CircleCI]  to test all pull
  requests. We require that all tests succeed on a pull request before it is merged.

-->