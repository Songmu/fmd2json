---
name: fmd2json
description: Convert Markdown with YAML frontmatter into JSON or NDJSON using the fmd2json CLI, then filter or reshape records with its built-in jq support. Use this skill whenever a user needs to extract frontmatter, index Markdown collections, turn posts or notes into machine-readable data, select records by metadata, or pipe Markdown metadata into another command—even when they ask for JSON, metadata extraction, or content inventory without naming fmd2json.
license: MIT
---

# fmd2json

Use `fmd2json` to turn Markdown documents into structured records without writing a custom parser. It preserves YAML frontmatter values and adds path and content fields that are useful for indexing, filtering, and pipelines.

## Workflow

1. Identify whether the input is file arguments, one Markdown document on stdin, or a newline-separated file list on stdin.
2. Confirm the command is available with `command -v fmd2json`. If it is missing, report that clearly rather than silently substituting a different parser.
3. Choose the simplest invocation that produces the requested shape.
4. Run the command with quoted paths. Preserve the path as supplied when the caller needs `dir`; do not `cd` into its parent or replace `docs/article.md` with `article.md`, because `dir` is derived from the argument text. Use `find ... -print0` with `xargs -0` when filenames may contain spaces or special characters; the no-argument file-list mode accepts newline-separated paths and cannot represent filenames containing newlines.
5. Inspect a small sample or validate the resulting JSON before reporting success. For NDJSON, validate each line independently or use a tool that understands streaming JSON.

## Input modes

| User's input | Invocation | Important behavior |
|---|---|---|
| One or more files | `fmd2json file.md docs/other.md` | Emits one JSON object per input file |
| One document on stdin | `fmd2json -` | `filename` is empty and `mtime` is omitted |
| Stdin document with a logical path | `fmd2json -filename docs/article.md -` | Derives `dir` and `filename` from the supplied path; `mtime` is omitted |
| A newline-separated list of paths | `find docs -name '*.md' -print \| fmd2json` | With no file arguments, stdin is interpreted as paths, not Markdown content |

Do not confuse the two stdin modes: `-` means “read one Markdown document,” while no arguments means “read file paths.”

## Output contract

Each input record produces JSON containing:

- `dir`: the input directory using `/` separators; omitted for the current directory or an unnamed stdin document
- `filename`: basename with a trailing `.md` removed
- `body`: content after the closing frontmatter delimiter, or the entire file when valid frontmatter is absent
- `mtime`: filesystem modification time in RFC 3339; omitted for stdin documents
- all non-conflicting YAML frontmatter properties

Multiple inputs produce newline-delimited JSON (NDJSON), not one JSON array. Preserve NDJSON for streaming workflows; wrap it only when the consumer explicitly requires an array.

For example, `fmd2json docs/article.md` emits `"dir":"docs"` and `"filename":"article"`. `fmd2json article.md` omits `dir` because the argument has no directory component. The same rule applies to the logical path passed with `-filename`.

Frontmatter is recognized only when the document begins with a `---` delimiter and has a closing `---` delimiter. Without valid opening and closing delimiters, the entire document is body text. If the delimiters are present but the YAML cannot be decoded, no frontmatter fields are added and only the content after the closing delimiter becomes `body`.

The generated fields `dir`, `filename`, `body`, and `mtime` take precedence over frontmatter keys with the same names. `fmd2json` writes a warning to stderr when a conflict occurs.

## Common commands

```bash
# Convert one file
fmd2json article.md

# Preserve the relative directory in the generated dir field
fmd2json docs/article.md

# Convert a safely expanded collection, including paths with spaces
find docs -type f -name '*.md' -print0 | xargs -0 fmd2json

# Read one Markdown document from stdin and preserve its logical path
cat article.md | fmd2json -filename docs/article.md -

# Select published records while keeping NDJSON objects
fmd2json --jq 'select(.draft != true)' docs/*.md

# Emit one raw title per line, falling back to the filename
fmd2json --jq '.title // .filename' --raw-output docs/*.md

# Reshape each record
fmd2json --jq '{slug: .filename, title: (.title // .filename), body: .body}' article.md
```

Use `-r` as shorthand for `--raw-output`. Raw output requires `--jq`; otherwise the command fails. Without raw output, string results remain JSON-encoded so each result stays safely on one line.

## Filtering guidance

Use the built-in `--jq` option for per-record selection and transformation:

```bash
# Records tagged "go"
fmd2json --jq 'select((.tags // []) | contains(["go"]))' docs/*.md

# Filenames for non-draft posts
fmd2json --jq 'select(.draft != true) | .filename' -r docs/*.md

# Compact metadata inventory without the body
fmd2json --jq 'del(.body)' docs/*.md
```

Default missing optional fields before applying array or string functions. For example, prefer `(.tags // [])` to `.tags` when some documents may not define tags.

## Error handling

Treat a nonzero exit as failure and surface the stderr message. Common causes include unreadable paths, invalid jq expressions, jq runtime errors, and using `-r` without `--jq`.

Warnings on stderr do not necessarily mean conversion failed. In particular, frontmatter key conflicts are warnings; the output remains valid and uses the generated values.
