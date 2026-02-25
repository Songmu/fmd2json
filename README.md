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

### `--jq` option

Filter and transform JSON output using [jq](https://jqlang.github.io/jq/) expressions (powered by [gojq](https://github.com/itchyny/gojq)):

```console
% fmd2json --jq '.filename' a.md b.md
"a"
"b"

% fmd2json --jq '{title: .filename, size: (.body | length)}' article.md
{"title":"article","size":42}
```

By default, string values are output as JSON-encoded strings (with quotes and escapes), which keeps each result on a single line even if the value contains newlines. This is safe for multi-file processing:

```console
% fmd2json --jq '.body' a.md b.md
"first line\nsecond line\n"
"another body\n"
```

Use `--raw-output` (or `-r`) to output raw strings without JSON encoding, similar to `jq -r`:

```console
% fmd2json --jq '.filename' -r a.md b.md
a
b
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
