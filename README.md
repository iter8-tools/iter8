# iter8
iter8 documentation is available at https://iter8.tools.

## Contributing to iter8
We are delighted that you are contributing to iter8!

Please discuss the change you wish to make using [issues](https://github.com/iter8-tools/iter8/issues), [discussions](https://github.com/iter8-tools/iter8/discussions), or in the [iter8 slack workspace](https://iter8-tools.slack.com) before submitting your PR.

### Locally building iter8 docs
iter8 documentation is built using [mkdocs for material](https://squidfunk.github.io/mkdocs-material/). Follow the instructions below to build iter8 docs locally.

**Pre-requisite:** [Node.js 12+](https://nodejs.org/en/)

**Note:** Fork the iter8 repo, clone your fork, and `npm run build`. This step is usually required only for the first time you work with iter8 docs locally.

```shell
git clone git@github.com:<your-github-account>/iter8.git
cd iter8/mkdocs
npm run build
```

### Locally serving iter8 docs

**Pre-requisite:** Python 3+. You may find it useful to setup and activate a Python 3+ virtual environment as follows. The `.venv` folder contains the virtual environment.

```shell
python3 -m venv .venv
source .venv/bin/activate
```

**Note:** the pip install below is usually required only for the first time you serve iter8 docs locally.

```shell
pip install -r requirements.txt
```

```shell
mkdocs serve
```

Browse http://localhost:8000 to view your local iter8 docs.

<!-- ### Process for updating code artifacts
YAMLs, scripts and other code artifacts that are part of code-samples are located under the `iter8/samples` folder. Changes to code artifacts are followed by a tagged release, so that versioned artifacts are available. -->

### Locally viewing live changes to iter8 docs

<!-- 1. While referring to code artifacts in docs (for example, a remote `kustomize` resource referenced in an experiment), use versioned artifacts. -->

1. The overall structure of the documentation, as reflected in the nav tabs of https://iter8.tools, is located in the `iter8/mkdocs/mkdocs.yml` file.

2. The markdown files for iter8 docs are located under the `iter8/mkdocs/docs` folder.

You will see live updates in http://localhost:8000 as you update markdown files in the `iter8/mkdocs` folder.

### Contributing an iter8 tutorial
1. When contributing a tutorial under the `Code samples` nav tab (like [this one](https://iter8-tools/http://localhost:8000/code-samples/iter8-knative/canary-progressive/), please include the relevant e2e test.

2. In iter8 tutorials, an artifact such as `experiment.yaml` could depend on another artifact such as a YAML app-manifest that is `kubectl` applied as part of the experiment's finish task. In such cases, split your contribution into two (or more) PRs. The first PR pushes artifacts that are dependencies. The second pushes the dependent artifacts, markdown files and tests.