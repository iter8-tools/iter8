# Creating a release

## Prerequisites

- The user creating the release must have push access to repository (so they can create a release tag).
- Verify that all pull requests to be included in the release have been approved and merged.
- Verify that latest release branch builds were successful.

## Checkout Release Branches

Define the following environment variables identifying the release (values here are illustrative):

```bash
export RELEASE_BRANCH=v0.2
export RELEASE=v0.2.0
```

## Retag Image and push to Docker Hub

We do this first so that tests triggered by next step will succeed.

### iter8-analytics

```bash
docker pull iter8/iter8-controller:${RELEASE_BRANCH}
docker tag iter8/iter8-controller:${RELEASE_BRANCH} iter8/iter8-controller:${RELEASE}
docker push iter8/iter8-controller:${RELEASE}
```

### iter8-controller

```bash
docker pull iter8/iter8-controller:${RELEASE_BRANCH}
docker tag iter8/iter8-controller:${RELEASE_BRANCH} iter8/iter8-controller:${RELEASE}
docker push iter8/iter8-controller:${RELEASE}
```

## Update source to reflect release version

### `iter8-analytics`

On a fork of the iter8-analytics project, check out and update the release branch:

```bash
git fetch upstream
git checkout ${RELEASE_BRANCH}
git merge upstream/${RELEASE_BRANCH}
```

Create a branch on which to make updates:

```bash
git checkout -b prepareRelease-${RELEASE}
```

Update the following files:

- `install/kubernetes/helm/iter8-analytics/Chart.yaml` to set `version` to $RELEASE
- `install/kubernetes/helm/iter8-analytics/values.yaml` to set `image.tag` to $RELEASE

Update the default install yaml:

```bash
make build-default
```

Push changes:

```bash
git add install/kubernetes/helm/iter8-analytics/Chart.yaml \
        install/kubernetes/helm/iter8-analytics/values.yaml \
        install/kubernetes/iter8-analytics.yaml
git commit -m "update version for release ${RELEASE}"
git push origin prepareRelease-${RELEASE}
```

Create a pull request against ${RELEASE_BRANCH} on the upstream project.
After tests complete and approval, merge pull request.

### `iter-controller`

On a fork of the iter8-analytics project, check out and update the release branch:

```bash
git fetch upstream
git checkout ${RELEASE_BRANCH}
git merge upstream/${RELEASE_BRANCH}
```

Create a branch on which to make updates:

```bash
git checkout -b prepareRelease-${RELEASE}
```

Update the following files:

- `install/helm/iter8-controller/Chart.yaml` to set `version` to $RELEASE
- `install/helm/iter8-controller/values.yaml` to set `image.tag` to $RELEASE
- `install/install.sh` to change all instances of $RELEASE_BRANCH to $RELEASE

Update the default install yaml files:

```bash
make build-default
```

Push changes:

```bash
git add install/helm/iter8-controller/Chart.yaml \
        install/helm/iter8-controller/values.yaml \
        install/iter8-controller.yaml \
        install/iter8-controller-telemetry-v2.yaml \
        install/install.sh
git commit -m "update version for release ${RELEASE}"
git push origin prepareRelease-${RELEASE}
```

Create a pull request against ${RELEASE_BRANCH} on the upstream project.
After tests complete and approval, merge pull request.

## Create git releases

### `iter8-tools/iter8-analytics`

#### Create tag

On the upstream project on release branch:

```bash
git checkout $RELEASE_BRANCH
git pull
git tag $RELEASE
git push origin --tags
```

#### Create release

Go to the list of project [releases](https://github.com/iter8-tools/iter8-analytics/releases)

Use "_Draft new release_" to create a new release:

- Select the tag  (`$RELEASE`) created in the previous step
- It may be necessary to pick the Target as $RELEASE_BRANCH
- Use tag (`$RELEASE`) as the release name
- Add release notes describing the changes
- Use "_Attach binaries by dropping them here or selecting them_" to add assets to the release
  - use the list of assets from the previous release as a guide
  - manually build the helm chart archive `tar cf iter8-analytics-helm-chart.tar iter8-analytics`
- publish the release

**Note**: You can save a draft release at any time and continue to work on it later.

**Note**: If you forget an asset, you can edit the release later to add or replace it.

### `iter8-tools/iter8-controller`

#### Create tag

On the upstream project on release branch:

```bash
git checkout $RELEASE_BRANCH
git pull
git tag $RELEASE
git push origin --tags
```

#### Create release

Go to the list of project [releases](https://github.com/iter8-tools/iter8-controller/releases)

Go to the list of project [releases](https://github.com/iter8-tools/iter8-analytics/releases)

Use "_Draft new release_" to create a new release:

- Select the tag  (`$RELEASE`) created in the previous step
- It may be necessary to pick the Target as $RELEASE_BRANCH
- Use tag (`$RELEASE`) as the release name
- Add release notes describing the changes
- Use "_Attach binaries by dropping them here or selecting them_" to add assets to the release
  - use the list of assets from the previous release as a guide
  - manually build the helm chart archive `tar cf iter8-controller-helm-chart.tar iter8-controller`
- publish the release

### `iter8-tools/docs`

#### Update release branch

On a fork of the `docs` project, check out and update the release branch:

```bash
git fetch upstream
git checkout ${RELEASE_BRANCH}
git merge upstream/${RELEASE_BRANCH}
```

Create a branch on which to make updates:

```bash
git checkout -b prepareRelease-${RELEASE}
```

- Update all references to the release ($RELEASE_BRANCH --> $RELEASE)
- `git add` all updated files
- commit and push the changes, submit a pull request against ${RELEASE_BRANCH} on the upstream project
- get approval and merge pull request

### Update master branch

On a fork of the `docs` project, checkout and update the master branch:

```bash
git checkout master
git merge upstream/master
```

Create a branch on which to make updates:

```bash
git checkout -b prepareRelease-${RELEASE}-master
```

- Update all references in `README.md` to the release (last release --> $RELEASE)
- add entry to `releases.md` for previous release
- commit and push changes, submit a pull request against master on the upstream project
- get approval and merge pull request

#### Create tag and release

#### Create tag

On the upstream project on release branch:

```bash
git checkout $RELEASE_BRANCH
git pull
git tag $RELEASE
git push origin --tags
```

#### Create release

As in previous projects; combine notes about changes.

## Undo changes

Submit PRs to all projects changing references to $RELEASE back to $RELEASE_BRANCH on the release branches of each project.

## Rectifying Mistakes

If you make a mistake (miss something) that needs to be added to the release, it is necessary to delete the release and the tag. To delete a tag:

```bash
git tag -d ${RELEASE}
git push --delete origin ${RELEASE}
```
