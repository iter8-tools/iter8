---
template: overrides/main.html
hide:
- navigation
---

# Contributing
We are delighted that you are considering a contribution to Iter8!

Please discuss the change you wish to make using [issues](https://github.com/iter8-tools/iter8/issues), [discussions](https://github.com/iter8-tools/iter8/discussions), or the [Iter8 slack workspace](https://join.slack.com/t/iter8-tools/shared_invite/zt-awl2se8i-L0pZCpuHntpPejxzLicbmw) before submitting your PR.

## Locally building Iter8 docs
Iter8 documentation is built using [mkdocs for material](https://squidfunk.github.io/mkdocs-material/). Follow the instructions below to build Iter8 docs locally.

**Pre-requisite:** [Node.js 15+](https://nodejs.org/en/)

**Note:** Fork Iter8 repo, clone your fork, and run the commands below. This step is required to build Iter8 docs for the first time. This step is required after you change static assets like images. This step is *not* required if you only change .md files or `mkdocs.yml`.

```shell
cd iter8/mkdocs
git update-index --assume-unchanged package-lock.json
npm run build
```

## Locally serving Iter8 docs
**Pre-requisite:** Python 3+. Setup, activate, and use a Python 3+ virtual environment as follows.

```shell
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
mkdocs serve -s
```

Browse [http://localhost:8000](http://localhost:8000) to view your local Iter8 docs.

## Locally viewing live changes to Iter8 docs
1. The overall structure of the documentation, as reflected in the nav tabs of [https://iter8.tools](https://iter8.tools), is located in the `iter8/mkdocs/mkdocs.yml` file.

2. The markdown files for Iter8 docs are located under the `iter8/mkdocs/docs` folder.

You will see live updates to [http://localhost:8000](http://localhost:8000) as you update markdown files in the `iter8/mkdocs` folder.

## Contributing an Iter8 tutorial
All iter8 tutorials include e2e tests, either as [part of GitHub Actions workflows](https://github.com/iter8-tools/iter8/blob/master/.github/workflows/knative-e2e-tests.yaml) or as a standalone test script [like this one](https://github.com/iter8-tools/iter8/blob/master/samples/knative/mirroring/e2etest.sh) if they require more resources than what is available in GitHub Actions workflows. When contributing a tutorial, please include relevant e2e tests.

## Extending Iter8 in other ways
Documentation for contributing other Iter8 extensions such as new handler tasks, analytics capabilities, and observability features is coming soon.
