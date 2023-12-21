These instructions are a guide to making a new major or minor (not patch) release. The challenge is that charts refer to the image version. But this version is only created when the release is published. Consequently the following sequence of steps is needed:

1. Modiyf only golang code (no version bump)
    1. Do not change `MajorMinor` or `Version` in `base/util.go`

Make new major/minor release

2. Bump version references in `/charts` changes and bump `/charts/iter8` chart version (no changes to `/testdata`) and bump Kustomize files and bump verifyUserExperience workflow
    1. The charts are modified to use the new image
    2. The chart versions should be bumped to match the major/minor version (this is required for the `iter8` chart) but is desirable for all

Merging the chart changes triggers a automatic chart releases

3. Version bump golang and `/testdata` and other workflows
    1. Bump `MajorMinor` or `Version` in `base/util.go`
    2. Bump explicit version references in remaining workflows
    3. Bump Dockerfile
    5. Changes to `/testdata` is only a version bump in charts (`iter8.tools/version`)

Make a release (new patch version)

***

At this point the documentation can be updated to refer to the new version. 
Some things to change in the docs

* `iter8.tools/version` in Kubernetes manifests samples
* `--version` for any `helm upgrade` and `helm template` commands
* `deleteiter8controller.md` and `installiter8controller.md`
* Reference to `values.yaml`
