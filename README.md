| Site | Status
|------|-------
| [iter8.tools](https://iter8.tools) (v1.0.0) | [![Netlify Status](https://api.netlify.com/api/v1/badges/3351aece-eb1e-4078-aab5-da6108a53ba7/deploy-status)](https://app.netlify.com/sites/iter8-v1-0-0/deploys)
| [v0-2-1.iter8.tools](https://v0-2-1.iter8.tools) (v0.2.1) | [![Netlify Status](https://api.netlify.com/api/v1/badges/c2832f30-9130-4c9b-8183-c110c890426e/deploy-status)](https://app.netlify.com/sites/iter8-v0-2-1/deploys)
| [preliminary.iter8.tools](https://preliminary.iter8.tools) | [![Netlify Status](https://api.netlify.com/api/v1/badges/b800a77c-9df7-49f0-9f55-010a0dbcdc70/deploy-status)](https://app.netlify.com/sites/preliminary-iter8-tools/deploys)

# iter8.tools

This repository contains the source code for [iter8.tools](https://iter8.tools) and
[preliminary.iter8.tools](https://preliminary.iter8.tools).

The main iter8 repository can be found [here](https://github.com/iter8-tools/iter8).

# Usage

Install Hugo

```bash
brew install hugo
```

Clone repository and submodules

```bash
git clone --recurse-submodules https://github.com/iter8-tools/docs.git
```

Host locally

```bash
cd docs
hugo serve
```

By default, Hugo will use [localhost:1313](localhost:1313).

# File structure

* [content/](content/): Contains all the Markdown files, which will be used to generate the documentation
* [static/](static/): Assets used in the documentation, including all the YAML files used by the tutorials.
* [archetypes](archetypes): Stores templates for [front matter](https://gohugo.io/content-management/front-matter/)
* [layouts/](layouts): Store templates for converting the Markdown files into HTML
* [themes/](themes): Contains the [Hugo theme](https://themes.gohugo.io/) which does the bulk of generation
* [resources/](resources): Caches files to speed up generation

Content creators will mostly be working with the [content/](content/), [data/](data/), and [static/](static/) directories.

For more information about these files, see [here](https://gohugo.io/getting-started/directory-structure/).

# Creating content

### New page

Create a new page by using the [hugo new](https://gohugo.io/commands/hugo_new/) command.

```bash
hugo new [path]
```

**Note**: You can also create a new page without using the command but the front matter, described below, will need to be manually inserted.

***

For example:

```bash
hugo new content/introduction/about.md
```

...would create a new _about_ page in the _introduction_ section.

### Front matter

This markdown file will have some code at the top of the page, known as [front matter](https://gohugo.io/content-management/front-matter/).

Front matter contains some meta data which is used for generation.

The following describes a number of useful front matter properties.

| Front matter property | Type | Description
|-----------------------|------|------------
| menuTitle | string | The name that will appear in the sidebar tab
| title | string | The name that will appear at the top of a page
| chapter | boolean | Change the way the page is rendered
| weight | integer | Used to order the page in the sidebar
| hidden | boolean | Whether the page should appear in the sidebar

For learn about other front matter properties, see [here](https://themes.gohugo.io//theme/hugo-theme-learn/en/cont/pages/#front-matter-configuration).

***

For example:

```md
---
title: Algorithms
weight: 3
---
```

### Add content

##### Text

Below the front matter, directly add Markdown.

##### Links to other pages

Use the `ref` or `relref` shortcodes to link to other pages.

```md
[alt text]({{< rel "[page name]" >}})
```

The `ref` and `relref` shortcodes will automatically search for pages based on their logical names or their relative paths. The advantage of using these short codes is that Hugo will check the validity of these links on build, allowing for greater maintainability.

For more information, see [here](https://gohugo.io/content-management/shortcodes/#ref-and-relref).
                                               
##### Images

Image files should be stored in [static/images/](static/images/). The [static/](static/) folder will form the base of the built files. Therefore, the image path, provided that the files are stored in in [static/images/](static/images/), will begin with `/images/`.

***

For example:

```md
![iter8 logo](/images/logo.png)
```

##### Files

Files should also be stored under the [static/](static/) folder.

Currently, files related to tutorials are stored under [static/tutorials](static/tutorials).

Files can be referenced using the `resourceAbsUrl` shortcode, which will construct a URL using the Hugo [baseURL](https://gohugo.io/getting-started/configuration/#all-configuration-settings).

For example:

```md
{{< resourceAbsUrl path="/images/logo.png" >}}
```

The `resourceAbsUrl` shortcode can also be used in the following ways to create links:

```md
[Click here to see see the file]({{< resourceAbsUrl path="/images/logo.png" >}})
```

or

```md
[{{< resourceAbsUrl path="/images/logo.png" >}}]({{< resourceAbsUrl path="/images/logo.png" >}})
```

# Link checking

We use the [Link Checker](https://github.com/marketplace/actions/link-checker) GitHub Action in order to verify that the links in our documentation are still valid.

If there is a link that is broken, then the workflow will fail. You will need to go to the logs and go to the `Link checker` step in order to see the report.

***

**Note**: The workflow can be configured to ignore certain links using the `-x` or `--exclude` [Liche](https://github.com/raviqqe/liche) option. See [here](https://github.com/marketplace/actions/link-checker#usage) and [here](https://github.com/raviqqe/liche#usage) for more information.