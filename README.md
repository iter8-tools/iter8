# iter8

iter8 documentation is available at https://iter8.tools.

## Contributing to iter8 docs
We are delighted that you are contributing to iter8 docs!

Please discuss the change you wish to make using [issues](https://github.com/iter8-tools/iter8/issues), [discussions](https://github.com/iter8-tools/iter8/discussions), or in the [iter8 slack workspace](https://iter8-tools.slack.com) before submitting your PR.

### Locally build iter8 docs
iter8 documentation is built using [mkdocs for material](https://squidfunk.github.io/mkdocs-material/). Follow the instructions below to serve iter8 docs locally.

**Pre-requisite:** [Node.js 12+](https://nodejs.org/en/)

**Note:** Fork the iter8 repo, clone your fork, and `npm run build`. This step is usually required only for the first time you work with iter8 docs locally.

```shell
git clone git@github.com:<your-github-account>/iter8.git
cd iter8/mkdocs
npm run build
```

### Locally serve iter8 docs

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

### Make and view changes to iter8 docs locally

1. YAMLs, scripts and other code artifacts that are part of code-samples are located under the `iter8/samples` folder.

2. The overall structure of the documentation, as reflected in the nav tabs of https://iter8.tools, is located in the `iter8/mkdocs/mkdocs.yml` file. Do not alter this file unless you really need to.

3. The markdown files for iter8 docs are located under the `iter8/mkdocs/docs` folder.

You should see live updates in http://localhost:8000 as you change 3 (or 2) above.

### Create PR
Once you are satisfied with your changes, push changes to your fork and create a PR for the iter8 repo.
