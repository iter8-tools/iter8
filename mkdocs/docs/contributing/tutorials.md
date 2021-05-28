---
template: main.html
---

# Contribute Tutorials

## Test your tutorial
All iter8 tutorials include e2e tests, either as [part of GitHub Actions workflows](https://github.com/iter8-tools/iter8/blob/master/.github/workflows/knative-e2e-tests.yaml) or as a standalone test script [like this one](https://github.com/iter8-tools/iter8/blob/master/samples/knative/mirroring/e2etest.sh) if they require more resources than what is available in GitHub Actions workflows. When contributing a tutorial, please include relevant e2e tests.

## Locally serve Iter8 docs
**Pre-requisite:** Python 3+. 

Use a Python 3 virtual environment to locally serve Iter8 docs. Run the following commands from the top-level directory of the Iter8 repo.

```shell
cd mkdocs
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
mkdocs serve -s
```

Browse [http://localhost:8000](http://localhost:8000) to view your local Iter8 docs.

## Locally view live changes to Iter8 docs
1. The overall structure of the documentation, as reflected in the nav tabs of [https://iter8.tools](https://iter8.tools), is located in the `iter8/mkdocs/mkdocs.yml` file.

2. The markdown files for Iter8 docs are located under the `iter8/mkdocs/docs` folder.

You will see live updates to [http://localhost:8000](http://localhost:8000) as you update the above files.
