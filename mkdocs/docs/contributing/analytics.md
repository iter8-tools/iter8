---
template: main.html
---

# Extending Iter8's Analytics Functions

Iter8's analytics functions are implemented in the [`iter8-analytics` repo](https://github.com/iter8-tools/iter8-analytics).

## Python virtual environment
Use a Python 3+ virtual environment to locally develop `iter8-analytics`. You can create and activate a virtual environment as follows.

```shell
git clone git@github.com:iter8-tools/iter8-analytics.git
cd iter8-analytics
python3 -m venv .venv
source .venv/bin/activate
```

## Running iter8-analytics locally
Create and activate a Python 3+ virtual environment as described above. The following instructions have been verified in a Python 3.9 virtual environment. Run them from the root folder of your `iter8-analytics` local repo.

```
1. pip install -r requirements.txt 
2. pip install -e .
3. cd iter8_analytics
4. python fastapi_app.py 
```

Navigate to http://localhost:8080/docs on your browser. You can interact with the iter8-analytics service and read its API documentation here. The iter8-analytics APIs are intended to work with metric databases, and use Kubernetes secrets for obtaining the required authentication information for querying the metric DBs.

### Running unit tests for iter8-analytics locally
```
1. pip install -r requirements.txt 
2. pip install -r test-requirements.txt
3. pip install -e .
4. coverage run --source=iter8_analytics --omit="*/__init__.py" -m pytest
```
You can see the coverage report by opening `htmlcov/index.html` in your browser.