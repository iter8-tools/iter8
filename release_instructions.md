# Creating a release

## Prerequisites

- The user creating the release must have direct push access to repository (so they can create a release tag)
- Verify that all pull requests to be included in the release have been approved and merged
- Verify that latest release branch builds were successful

Define the following environment variables identifying the release:

```bash
export RELEASE_BRANCH=v1.0
export RELEASE=v1.0.0-rc2
export PREVIOUS_RELEASE=v1.0.0-rc1
```

## Create Release Branch **If Necessary**

If necessary, create a release branch. This should be an unusual step.

```bash
git checkout master
git fetch upstream master
git rebase upstream/master
git push origin master
git checkout -b $RELEASE_BRANCH
git push upstream $RELEASE_BRANCH
```

## Checkout Release Branch

Checkout out local release branch and make sure it matches upstram release branch:

```bash
git checkout $RELEASE_BRANCH
git fetch upstream
git rebase upstream/$RELEASE_BRANCH
```

Create a pull request on upstream release branch (v1.0) that contains:

- all the changes from upstream/master
- any changes needed to replace things with the (new) release tag

```bash
git checkout -b prepare-$RELEASE
```

cherry-pick all changes from master that should be included in release. We should be able to skip merge commits.

Change all the files with references to the `$PREVIOUS_RELEASE` tag (or `master`) to refer to the new `$RELEASE` tag.

Determine which files might need to change using `grep`:

```bash
grep -R --exclude-dir vendor --exclude Gopkg.lock $PREVIOUS_RELEASE *
grep -R --exclude-dir vendor --exclude Gopkg.lock master *

```

You might try the following strategies:

Using a script to change all instances of $PREVIOUS_RELEASE to $RELEASE:

```bash
for f in $(grep -R --exclude-dir vendor --exclude Gopkg.lock --exclude CHANGELOG $PREVIOUS_RELEASE * | cut -f1 -d:  | uniq); do sed -i.old "s#$PREVIOUS_RELEASE#$RELEASE#g" $f ; done
for f in $(find . -name "*.old" ! -path "vendor/*" -print); do rm $f; done
```

- `integrations/grafana/install_grafana_dashboard.sh`
- `install/helm/iter8-controller/values.yaml` to set `image.tag` to $RELEASE
- `install/install.sh`
- `test/e2e/install-iter8.sh`

By hand:

- `CHANGELOG` identify features of new version and separator
- `install/helm/iter8-controller/Chart.yaml` to set `version` to $RELEASE
- `README.md` (if created a new release branch)

By running `make build-default`:

- `install/iter8-controller*.yaml`

Commit and push changes:

```bash
git commit -a -m "update for release ${RELEASE}"
git push -u origin prepare-${RELEASE}
```

Create a pull request against ${RELEASE_BRANCH} on the upstream project.
After tests complete and approval, merge pull request.

### `iter8-analytics`

Checkout out local release branch and make sure it matches upstram release branch:

```bash
git checkout $RELEASE_BRANCH
git fetch upstream
git rebase upstream/$RELEASE_BRANCH
```

Create a pull request on upstream release branch (v1.0) that contains:

- all the changes from upstream/master
- any changes needed to replace things with the (new) release tag

```bash
git checkout -b prepare-$RELEASE
```

Cherry-pick all changes on master that should be included in the release.

Change all the files with references to the `$PREVIOUS_RELEASE` tag (or `master`) to refer to the new `$RELEASE` tag.

Determine which files might need to change using `grep`:

```bash
grep -R --exclude-dir vendor --exclude Gopkg.lock $PREVIOUS_RELEASE *
grep -R --exclude-dir vendor --exclude Gopkg.lock master *
```

You might try the following strategies:

Using a script to change all instances of $PREVIOUS_RELEASE to $RELEASE:

```bash
for f in $(grep -R --exclude-dir vendor --exclude Gopkg.lock --exclude CHANGELOG $PREVIOUS_RELEASE * | cut -f1 -d:  | uniq); do sed -i.old "s#$PREVIOUS_RELEASE#$RELEASE#g" $f ; done
for f in $(find . -name "*.old" ! -path "vendor/*" -print); do rm $f; done
```

- `install/kubernetes/helm/iter8-analytics/values.yaml` to set `image.tag` to $RELEASE
- `tests/e2e/install-iter8`

By hand:

- `CHANGELOG` identify features of new version and separator
- `install/kubernetes/helm/iter8-analytics/Chart.yaml` to set `version` to $RELEASE

By running `make build-default`:

- `install/kubernetes/iter8-analytics.yaml`

Push changes either by creating a new pull request, having it approved and merged or by pushing:

Commit changes:

```bash
git commit -a -m "update for release ${RELEASE}"
git push -u origin prepare-${RELEASE}
```

Create a pull request against ${RELEASE_BRANCH} on the upstream project.
After tests complete and approval, merge pull request.

## Create tag for each repository

On the upstream project on release branch:

```bash
git checkout $RELEASE_BRANCH
git pull upstream $RELEASE_BRANCH
git push origin $RELEASE_BRANCH
git tag $RELEASE
git push upstream --tags
```

This triggers travis job that builds an image for the tag. You can verify by inspecting travis, docker hub (verify image with tag $RELEASE created) and git (verify release $RELEASE created)

# Rectifying Mistakes

If you make a mistake (or miss something) that needs to be added to the release, it is necessary to delete the release and the tag. To delete a tag:

```bash
git tag -d ${RELEASE}
git push --delete upstream ${RELEASE}
```

You can then make the changes and push the changes. When the tag is re-created, the travis job will re-create the release.

## If you want to permanently delete a release

To permanently delete a release, additional steps are needed:

- Select the release from <https://github.com/iter8-tools/iter8-analytics/releases> and **Delete**
- Delete the image from <https://hub.docker.com/>, identify the tag and delete

# TO BE UPDATED

### `iter8-tools/docs`

#### Update release branch

On a fork of the `docs` project, check out and update the release branch:

```bash
git fetch upstream
git checkout ${RELEASE_BRANCH}
git rebase upstream/${RELEASE_BRANCH}
git push origin ${RELEASE_BRANCH}
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

## Documentation

### Netlify set up

Create a [Netlify](https://app.netlify.com/) account 

### Create a new site

Log in to Netlify

***

Go to "New site from Git"

***

Pick the "GitHub" Git provider

***

Pick the `iter8-tools` organization and the `docs` repository

Pick an appropriate branch to deploy

Set up a build command. If, in the Hugo [configuration file](https://gohugo.io/getting-started/configuration) (i.e. `config.toml`), the [baseURL](https://gohugo.io/getting-started/configuration#all-configuration-settings) is not set (i.e. it is `"\"` or `""`), then use the `-b` or `--baseURL` [option](https://gohugo.io/getting-started/usage/) to assign a base URL to the `hugo` build command.

For example:

```bash
hugo -b iter8.tools
```

```bash
hugo -b preliminary.iter8.tools
```

```bash
hugo -b v0-2-1.iter8.tools
```

***

Deploy site!

### Set up a custom domain

When you first deploy a site, Netlify will assign it a random name and URL.

To set up a custom domain, go to the site overview, and there is an option to "Set up a custom domain".

When you add a domain of a new namespace, Netlify will ask you if you are the namespace owner. If you are, then after adding the domain, you will need to go its options and `Set up Netlify DNS`.

### New documentation release

The `master` branch contains the build of the next unstable release and corresponds to [preliminary.iter8.tools](preliminary.iter8.tools).

The latest branch corresponds to [iter8.tools](iter8.tools). 

Since `v0.2.1`, we have begun using the [iter8.tools](iter8.tools) site. Previous versions can still be found in the [v0.0](https://github.com/iter8-tools/docs/tree/v0.0), [v0.1](https://github.com/iter8-tools/docs/tree/v0.1), and [v0.2](https://github.com/iter8-tools/docs/tree/v0.2) branches.

***

When a new version is created, then the following changes must be made:
the old latest site (originally [iter8.tools](iter8.tools)) must point to a new archival site (e.g. [v0-2-1.iter8.tools](v0-2-1.iter8.tools)); the new version must have its own dedicated branch; and the new latest site (originally [preliminary.iter8.tools](preliminary.iter8.tools)) must point to [iter8.tools](iter8.tools).

To make these changes, follow these steps:

Changes for the old latest site:

1. Go to `Domain management` of [iter8.tools](iter8.tools). Remove the `iter8.tools` and `www.iter8.tools` domains. Add a custom archival domain (e.g. [v0-2-1.iter8.tools](v0-2-1.iter8.tools)) with the appropriate version number.
2. Go to `Build & deploy`. Change the Hugo build command to use the the new domain via the `-b` or `--baseURL` option (e.g. `hugo -b v0-2-1.iter8.tools
`).

Changes for the new latest site:

1. Go to the `master` branch and edit the configuration (i.e. `config.toml`). Change the `versionNumber`, `versionName`, and `editURL` appropriately. The `versionNumber` is used in conjunction with the `{{< versionNumber >}}` shortcode to generate URLs, pointing to resources released in other repositories under the [iter8-tools](https://github.com/iter8-tools) organization. The `versionName` is a human-readable version of the `versionNumber` which is displayed in the sidebar. The `editURL` is required for a feature on each page that allows you to easily change a file and create a pull request.
2. Change the [content/releases/_index.md](https://github.com/iter8-tools/docs/blob/master/content/releases/_index.md) to include the new version as well as update the `preview`, `stable`, and `deprecated` version categories.
3. Create a new branch using the format `release-<release version>`. For example: `release-0.2.1` or `release-1.0.0`.
4. Follow the [Create a new site](#create-a-new-site) instructions and create a new site with an archival domain name (e.g. [v0-2-1.iter8.tools](v0-2-1.iter8.tools)) with the appropriate version number, using the new branch. Ensure that the build command also uses the Hugo `-b` or `--baseURL`, or else some links will not be generated correctly.

Changes for the preview site:

1. Change the site table and the top of the [README.md](https://github.com/iter8-tools/docs/blob/master/README.md) so that it states the correct preview and stable sites and uses the correct Netlify badges.