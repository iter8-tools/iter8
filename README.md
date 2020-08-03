# Usage

Install Hugo

```bash
brew install hugo
```

Clone repository and submodules (Hugo theme)

```bash
git clone --recurse-submodules https://github.com/iter8-tools/docs.git
```

Host locally

```bash
cd hugo-iter8-docs
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

Additional front matter fields may be added.

***

For example:

```md
---
date: 2016-04-09T16:50:16+02:00
title: Algorithms
weight: 3
---
```

In this example using the [hugo-theme-learn](https://themes.gohugo.io/hugo-theme-learn/) theme, the `date` is self-explanatory, the `title` is used to name the sidebar tab as well as the title of page, and the `weight` is used to order to different pages in the sidebar. 

### Add content

Below the front matter, directly add Markdown.

***

Image files should be stored in [static/images/](static/images/).

Images can be displayed using the following syntax:

```
![Alt Text](url)
```

***

For example:

```md
![iter8pic](images/iter8pic.png)
```

**Note**: the preceding `/` in the `url` component is important! Otherwise the image will not display.