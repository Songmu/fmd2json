fmd2json
=======

[![Test Status](https://github.com/Songmu/fmd2json/actions/workflows/test.yaml/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/fmd2json/branch/main/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/fmd2json)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/fmd2json)][PkgGoDev]

[actions]: https://github.com/Songmu/fmd2json/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/fmd2json
[license]: https://github.com/Songmu/fmd2json/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/fmd2json

fmd2json short description

## Synopsis

Convert Markdown files with YAML frontmatter to JSON.

```console
% cat article.md
---
title: Hello
tags:
  - go
  - cli
---
This is the body.

% fmd2json article.md
{"title":"Hello","tags":["go","cli"],"filename":"article","body":"This is the body.\n","mtime":"2025-01-01T00:00:00+09:00"}
```

## Description

`fmd2json` reads Markdown files with YAML frontmatter and outputs JSON. Frontmatter properties are included as-is, along with default properties:

- `filename` — file name without `.md` extension
- `body` — content after frontmatter (or entire content if no frontmatter)
- `mtime` — file modification time in RFC 3339 format

Multiple files produce newline-delimited JSON (ndjson):

```console
% fmd2json a.md b.md
{"filename":"a",...}
{"filename":"b",...}
```

Read from stdin with `-`:

```console
% cat article.md | fmd2json -
{"filename":"","body":"...","..."}
```

Read a file list from stdin (when no arguments given):

```console
% find . -name '*.md' | fmd2json
{"filename":"article1",...}
{"filename":"article2",...}
```

## Installation

```console
# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/fmd2json/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/fmd2json/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/fmd2json/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/Songmu/fmd2json/cmd/fmd2json@latest
```

## Author

[Songmu](https://github.com/Songmu)
