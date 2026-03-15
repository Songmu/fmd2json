---
name: fmd2json
description: Convert Markdown files with YAML frontmatter to JSON using fmd2json
license: MIT
compatibility:
  - claude
  - codex
  - agents
allowed_tools:
  - Bash
  - Read
---

# fmd2json

`fmd2json` reads Markdown files with YAML frontmatter and outputs newline-delimited JSON (ndjson). Each output object contains all frontmatter properties plus default properties.

## Output JSON Schema

Each line of output is a JSON object conforming to the following schema:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "fmd2json output",
  "description": "JSON output schema for fmd2json",
  "type": "object",
  "properties": {
    "filename": {
      "type": "string",
      "description": "file name with .md extension removed"
    },
    "body": {
      "type": "string",
      "description": "content after frontmatter delimiter"
    },
    "mtime": {
      "type": "string",
      "format": "date-time",
      "description": "file modification time in RFC 3339"
    }
  },
  "required": ["filename", "body"],
  "additionalProperties": true
}
```

`additionalProperties: true` means any YAML frontmatter key-value pairs are included as additional properties in the output object.

## Basic Usage

```bash
# Convert a single file
fmd2json article.md

# Convert multiple files (outputs one JSON object per line)
fmd2json a.md b.md

# Read from stdin using '-'
cat article.md | fmd2json -

# Specify filename when reading from stdin
cat article.md | fmd2json -filename article.md -

# Read a list of file paths from stdin (when no arguments given)
find . -name '*.md' | fmd2json
```

## `--jq` / `-r` Options

Use `--jq` to filter or transform each JSON output with a jq expression (powered by gojq):

```bash
# Extract a single field
fmd2json --jq '.filename' a.md b.md

# Transform the output object
fmd2json --jq '{title: .filename, size: (.body | length)}' article.md
```

By default, string values from `--jq` are JSON-encoded (quoted). Use `--raw-output` (or `-r`) to output raw strings without JSON encoding:

```bash
# Output raw string values (no quotes)
fmd2json --jq '.filename' -r a.md b.md

# Equivalent shorthand
fmd2json --jq '.title' --raw-output a.md
```

Note: `-r` / `--raw-output` requires `--jq` to be specified.

## Filtering with `select()`

Use jq's `select()` to filter records:

```bash
# Only output records where the 'draft' frontmatter field is true
fmd2json --jq 'select(.draft == true)' *.md

# Only output records with a specific tag
fmd2json --jq 'select(.tags | arrays | contains(["go"]))' *.md

# Output filenames of published posts
fmd2json --jq 'select(.draft != true) | .filename' -r *.md
```

## Input Modes Summary

| Invocation | Input source |
|---|---|
| `fmd2json file.md ...` | File arguments |
| `fmd2json -` | stdin (single document) |
| `fmd2json -filename name.md -` | stdin with explicit filename |
| `fmd2json` (no args) | File paths read from stdin, one per line |
