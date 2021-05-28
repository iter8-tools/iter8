---
template: main.html
---

# Contributing Guide

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

* [Iter8 documentation / tutorials](../tutorials)
* New features
* Builds, CI
* Bug fixes
* Web design for https://iter8.tools
* Communications / social media / blog posts
* Reviewing pull requests

Not everything happens through a GitHub pull request. Please come to our
[meetings](#come-to-meetings) or [contact us](../../../getting-started/help) and let's discuss how we can work together. 

*** 

## Come to meetings!
Absolutely everyone is welcome to come to any of our meetings. You never need an
invite to join us. In fact, we want you to join us, even if you don’t have
anything you feel like you want to contribute. Just being there is enough!

You can find out more about our meetings [here](../../../getting-started/help). You don’t have to turn on your video. The first time you come, introducing yourself is more than enough.
Over time, we hope that you feel comfortable voicing your opinions, giving
feedback on others’ ideas, and even sharing your own ideas, and experiences.

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
don’t know where to start or can't find a suitable issue, you can reach out to us over the [Iter8 slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) for help finding something to work on.

Once you see an issue that you'd like to work on, please post a comment saying
that you want to work on it. Something like "I want to work on this" is fine.

***

## Ask for Help

The best ways to reach us with a question when contributing is to ask on:

* The original GitHub issue
* `#development` channel in the [Iter8 slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw)

<!-- ## Pull Request Lifecycle

**TODO** -->
<!-- This is an optional section but we encourage you to think about your 
pull request process and help set expectations for both contributors and 
reviewers.

Instead of a fixed template, use these questions below as an exercise to uncover
the unwritten rules and norms your project has for both reviewers and
contributors. Using your answers, write a description of what a
contributor can expect during their pull request.

* When should contributors start to submit a PR - when it’s ready for review or
  as a work-in-progress?
* How do contributors signal that a PR is ready for review or that it’s not
  complete and still a work-in-progress?
* When should the contributor should expect initial review? The follow-up
  reviews?
* When and how should the author ping/bump when the pull request is ready for
  further review or appears stalled?
* How to handle stuck pull requests that you can’t seem to get reviewed?
* How to handle follow-up issues and pull requests?
* What kind of pull requests do you prefer: small scope, incremental value or
  feature complete?
* What should contributors do if they no longer want to follow-through with the
  PR? For example, will maintainers potentially refactor and use the code?
  Will maintainers close a PR if the contributor hasn’t responded in a specific
  timeframe?
* Once a PR is merged, what is the process for it getting into the next release?
* When does a contribution show up “live”?

Here are some examples from other projects:
 
* https://porter.sh/src/CONTRIBUTING.md#the-life-of-a-pull-request

-->

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