# Creating a release

## Prerequisites

- The user creating the release must have push access to repository (so they can update files and create a release tag)
- Verify that all pull requests to be included in the release have been approved and merged.

## Checkout Release Branches

Define the following environment variables identifying the release (values here are illustrative):

```bash
export RELEASE_BRANCH=v0.1
export RELEASE=v0.1.1
```

In each project checkout and update the release branch:

```bash
git checkout ${RELEASE_BRANCH}
git pull
```

## Test the release

Run the tutorial using the release branch. Clean up before hand. It may be necessary to build the images.

### Cleanup Any Existing Deployment(s)

- Delete any `Experiment` objects
- Delete the application namespace, `bookinfo-iter8`
- Delete the `iter8` namespace
- Delete the `experiments.iter8.tools` CRD
- Delete the iter8 dashboard in grafana

### Build Images and Deploy

**Note**: If not already merged, [pull request 153](https://github.com/iter8-tools/iter8-controller/pull/153) updates the Makefile to support the `TELEMETRY` option used below.

```bash
export RELEASE_VERSION=v0.1.1
export RELEASE_CANDIDATE=cr1
export TELEMETRY_VERSION=v2

CONTROLLER_IMG=iter8/iter8-comtroller:${RELEASE}
CANDIDATE_CONTROLLER_IMG=${CONTROLLER_IMG}-${RELEASE_CANDIDATE}

ANALYTICS_IMG=iter8/iter8-analytics:${RELEASE}
CANDIDATE_ANALYTICS_IMG=${ANALYTICS_IMG}-${RELEASE_CANDIDATE}

pushd iter8-analytics
IMG=${CANDIDATE_ANALYTICS_IMG} make docker-build
IMG=${CANDIDATE_ANALYTICS_IMG} make docker-push
IMG=${CANDIDATE_ANALYTICS_IMG} make deploy
popd

pushd iter8-controller
IMG=${CANDIDATE_CONTROLLER_IMG} make docker-build
IMG=${CANDIDATE_CONTROLLER_IMG} make docker-push
IMG=${CANDIDATE_CONTROLLER_IMG} TELEMETRY=${TELEMETRY_VERSION} make deploy
popd
```

### Run Tutorials

When running the tutorials use the files in the release branch directly instead of those in the (still unupdated) documentation. These files are in `iter8-controller/docs/tutorials/istio/bookinfo`.

**Note**: should verify latest Istio version, grafana config, probably with kiali, on Redhat, etc.

## Update source to reflect release version

### `iter8-analytics`

Check out the release branch:

```bash
git checkout ${RELEASE_BRANCH}
```

Update the following files:

- `install/kubernetes/helm/iter8-analytics/Chart.yaml` to set `version` to the release version
- `install/kubernetes/helm/iter8-analytics/values.yaml` to set `image.tag` to the release version

Update the default install yaml:

```bash
make build-default
```

Push changes to the release branch:

```bash
git add install/kubernetes/helm/iter8-analytics/Chart.yaml \
        install/kubernetes/helm/iter8-analytics/values.yaml \
        install/kubernetes/iter8-analytics.yaml
git commit -m "update version for release ${RELEASE}"
git push origin ${RELEASE_BRANCH}
```

### `iter-controller`

Check out the release branch:

```bash
git checkout ${RELEASE_BRANCH}
```

Update the following files:

- `install/helm/iter8-controller/Chart.yaml` to set `version` to the release version
- `install/helm/iter8-controller/values.yaml` to set `image.tag` to the release version

```bash
make build-default
```

Push changes to the release branch:

```bash
git add install/helm/iter8-controller/Chart.yaml \
        install/helm/iter8-controller/values.yaml \
        install/iter8-controller.yaml \
        install/iter8-controller-telemetry-v2.yaml
git commit -m "update version for release ${RELEASE}"
git push origin ${RELEASE_BRANCH}
```

## Create git releases

### `iter8-tools/iter8-analytics`

#### Retag Image and push to Docker Hub

```bash
docker pull $CANDIDATE_ANALYTICS_IMG
docker tag $CANDIDATE_ANALYTICS_IMG $ANALYTICS_IMG
docker push $ANALYTICS_IMG
```

#### Create tag

```bash
git tag $RELEASE
git push origin --tags
```

#### Create release

Go to the list of project [releases](https://github.com/iter8-tools/iter8-analytics/releases)

Use "_Draft new release_" to create a new release:

- Select the tag  (`$RELEASE`) created in the previous step
- Use tag (`$RELEASE`) as the release name
- Add release notes describing the changes
- Use "_Attach binaries by dropping them here or selecting them_" to add assets to the release
  - use the list of assets from the previous release as a guide
  - manually build the helm chart archive `tar tf iter8-analytics-helm-chart.tar iter8-analytics`
- publish the release

**Note** that you can save a draft release at any time and continue to work on it later.

**Note** that if you forget an asset, you can edit the release later to add or replace it.

### `iter8-tools/iter8-controller`

#### Retag Image and push to Docker Hub

```bash
docker pull $CANDIDATE_CONTROLLER_IMG
docker tag $CANDIDATE_CONTROLLER_IMG $CONTROLLER_IMG
docker push $CONTROLLER_IMG
```

#### Create tag

```bash
git tag $RELEASE
git push origin --tags
```

#### Create release

Go to the list of project [releases](https://github.com/iter8-tools/iter8-controller/releases)

Create similar to how done for analytics engine.

### `iter8-tools/docs`

#### Update release branch

Checkout the release branch:

```bash
git checkout ${RELEASE_BRANCH}
```

- Update all references to the release
- `git add` all updated files
- commit and push the release branch

#### Create tag and release

As above for the other projects. Combine the notes from the releases created above.

#### Update master branch

Checkout master branch and update all references to the release in _README.md_.

Update _releases.md_ with the new release.

Commit and push the changes

sanity check links including in commands

**Note**: If you make a mistake (miss something) that needs to be added to the release, it is necessary to delete the release and the tag. To delete a tag:

```bash
git tag -d ${RELEASE}
git push --delete origin ${RELEASE}
```
