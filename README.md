| Site | Status
|------|-------
| [iter8.tools](https://iter8.tools) (v0.2.1) | [![Netlify Status](https://api.netlify.com/api/v1/badges/5e3faba2-d2ae-4252-b829-b9cb639bc5df/deploy-status)](https://app.netlify.com/sites/iter8/deploys)
| [preliminary.iter8.tools](https://preliminary.iter8.tools) (v1.0.0 preview) | [![Netlify Status](https://api.netlify.com/api/v1/badges/8e53cd9b-0cf4-4b3b-8db6-dee596b99bf1/deploy-status)](https://app.netlify.com/sites/preliminary-iter8/deploys)

# iter8.tools

This repository contains the source code for [iter8.tools](https://iter8.tools) and
[preliminary.iter8.tools](https://preliminary.iter8.tools).

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
* [data/](data/): Contains all the JSON, YAML, or TOML, which contains configuration files and data for dynamically generated content
* [static/](static/): Other assets used in the documentation, such as images
* [archetypes](archetypes): Stores templates for [front matter](https://gohugo.io/content-management/front-matter/)
* [layouts/](layouts): Store templates for converting the Markdown files into HTML
* [themes/](themes): Contains the [Hugo theme](https://themes.gohugo.io/) which does the bulk of generation
* [resources/](resources): Caches files to speed up generation
* [public/](public/): Outputted HTML and CSS files

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
                                               
##### Images

Image files should be stored in [static/images/](static/images/).

Images can be displayed using the following syntax:

```md
![alt text]({{< resourceAbsUrl path="[image path]" >}})
```

**Notes**: the [static/](static/) folder will form the base of the built files. Therefore, the image path, provided that the files are stored in in [static/images/](static/images/), will begin with "images/".

***

For example:

```md
![iter8 logo]({{< resourceAbsUrl path="images/logo.png" >}})
```

##### Files

Files should also be stored under the [static/](static/) folder.

Currently, files related to tutorials are stored under [static/tutorials](static/tutorials).