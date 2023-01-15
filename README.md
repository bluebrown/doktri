# doktri

Yet another static site generator. This time with a slightly different spin.

- There is no `frontmatter`. I like to have *pure* markdown content in the sense
that it can be used anywhere without modification. `Frontmatter` makes this
difficult.
- Every directory and piece of content is a *TreeNode*. The tree can be
traversed from within the templates. Similar to the document object model (DOM)
in the browser.
- The go template engine is used to render templates. This allows to have
methods on the nodes such as `NextSibling` and `FirstChild.`

The result is a powerful templating experience that can get away without
`frontmatter`. For example, if you want to show related posts, you can group
your content in folders by category and list all sibling as the related posts.

## Everything is a Node

Every directory and file is a Node. Files are `Leafs`, meaning they don't have
children.

There are serval fields and methods attached to the nodes. Methods can be called
from within a template and used in pipelines. For example to render the markdown
content to html, you call the Content method and pass it to the render function.

```console
{{ .Content | render }}
```

## Documentation

Read the
[documentation](https://pkg.go.dev/github.com/bluebrown/doktri/internal/engine),
to learn more about the fields and methods you can use inside a template.

## Project Layout

The project should be structured as follows. `doktri init` will generate this
layout for you.

```bash
├── .theme
├── assets
├── docs
└── meta.yaml
```

The *.theme* dir contains the theme to use. The *assets* dir contains extra
assets that are copied to *dist/assets* after the assets of the theme have been
copied, in order to allow for extra assets not contained in the theme with
overwrite behavior. The *docs* dir contains the actual markdown files. These can
be nested into sub directories. The *meta.yaml* contains extra meta information
that can be used from within the templates.

## File Names

The markdown files should be prefixed with `yyyy-mm-dd-`. This allows to infer
the date of the content without using frontmatter. `doktri create` can be used
to create files with the right name format.

## Site Meta

You may want to access some meta data about your site. For example the title or
social links. For this purpose you can place a *meta.yaml* at the root of your
source directory. This file can contain arbitrary data which can be accessed
from within the template via the `meta` function. For example

```html
<head>
  <title>{{ meta.title }}</title>
</head>
```

## Theme

doktri requires some files in order to function. Primarily it needs 3 templates:
`base.html`, `dir.html` and `file.html`. These are looked up in the theme folder
at *templates/layouts*. Additionally it will parse any template in
*templates/includes* if that directory exists.

A typical theme might look like the below. By default is assumed to be at
*.theme* in the source dir.

```bash
├── assets
└── templates
    ├── includes
    └── layouts
        ├── base.html
        ├── dir.html
        └── file.html
```

Assets are minified, if possible, and then copied to the dist dir.

If you create a new project with `doktri init`, the [default
theme](https://github.com/bluebrown/doktri-theme-default) is fetched and added
to your project.

## Examples

Here are some examples that showcase why using this model is good.

### Bread Crumbs

Bread crumbs are a common component on webpages, especially content focused
ones. We can easily build out the breadcrumb links by traversing the tree
upwards until the root node is founds by calling the segment template
recursively.

```html
{{- define "bread-crumbs" }}
{{- if .IsRoot }}
<li><a href="{{ .Root.Path }}">{{ .Root.Title }}</a></li>
{{- else }}
{{- template "bread-crumbs" .Parent }}
<li role="separator" class="vr">/</li>
<li><a href="{{ .Path }}">{{ .Title }}</a></li>
{{- end }}
{{- end }}
```

Then we call this inside a `ul` on some content page.

```html
<ul class="bread-crumbs">{{ template "bread-crumbs" . }}</ul>
```

The result looks something like the below.

```html
<ul class=bread-crumbs>
  <li><a href="/">Home</a></li>
  <li role=separator class=vr>/</li>
  <li><a href="/kubernetes/">Kubernetes</a></li>
  <li role=separator class=vr>/</li>
  <li><a href="/kubernetes/container/">Container</a></li>
  <li role=separator class=vr>/</li>
  <li><a href="/kubernetes/container/runtime/">Runtime</a></li>
</ul>
```

## CLI

```console
NAME:
   doktri - a static site generator

USAGE:
   doktri [global options] command [command options] [arguments...]

COMMANDS:
   build, b   build the static html content
   serve, s   build and serve the static html content, with hot reload
   init, i    initialize a new project
   create, c  create a new post
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## Installation

### Binary Release

```bash
curl -fsSL github.com/bluebrown/releases/latest/download/doktri_linux_x86_64.tar.gz | tar -xzf - doktri
```

### Form Source

```bash
go install github.com/bluebrown/doktri/cmd/doktri@latest
```

### Container Image

```bash
docker run --rm -u "$(id -u):$(id -g)" -v "$PWD:/tmp" bluebrown/doktri build
```
