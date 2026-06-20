# leet

A small CLI for practicing LeetCode problems locally, in your own editor, without copy-pasting
between the browser and your files.


> - This is a small, personal project, likely to be in-progress for a long time, and never fully completed nor published. For now, the core flow (fetch, scaffold, open) works, though it could be improved.
> - ***Submission is not implemented yet***.

## What it does

`leet` talks to LeetCode's (undocumented) API to fetch a problem's starter code and description,
then scaffolds it into a local folder so you can solve it in your own editor instead of the web
interface.

```
leet load 22 -l go,python3
```

creates:

```
~/leet-problems/22.generate-parentheses/
├── problem.html
├── 22.go
└── 22.py
```

## Install

```bash
go install github.com/cactuuus/leet@latest
```

Make sure `~/go/bin` is on your `PATH`.

## Usage

### Load a problem

```bash
leet load 2135                       # uses your configured default languages
leet load 2135 --langs go,python3    # override languages for this run
leet load 2135 --desc-only           # just the description, no code files
leet load daily                      # today's daily challenge
leet load 2135 --open                # also opens the folder afterward
leet load 2135 --force               # skip the overwrite prompt
```

### Open a problem folder (or the whole problems directory)

```bash
leet open 2135    # opens that problem's folder
leet open         # opens the root problems directory
```

### List supported languages
Note: this *should* match the languages that LeetCode itself supports.

```bash
leet languages
```

### Configuration

```bash
leet config show
leet config set-languages go,python3,typescript
leet config set-problems-dir ~/code/leetcode
leet config set-editor-cmd code     # or subl, nvim, etc.
```

Config lives at `~/.config/leet/config.toml`. \
Problems list is cached at `~/.cache/leet/problems.json` so `leet` doesn't refetch LeetCode's full
problem list on every run. It is only ever refreshed on a cache miss.

## How it works

- LeetCode exposes an internal GraphQL endpoint (`leetcode.com/graphql`) and a REST endpoint
  (`leetcode.com/api/problems/all/`) that the web app itself uses.
- Problem numbers aren't valid API identifiers — only slugs are (e.g. `two-sum`). `leet` caches a
  number → slug map locally and refreshes it automatically on a cache miss.
- Paid-only problems are detected and rejected with a clear error, since fetching them requires a
  LeetCode Premium session (not yet supported).

## Project layout

```bash
cmd/                  # CLI commands (Cobra)
internal/leetcode/    # API client + local problem cache
internal/scaffold/    # Handles problem folders/files on disk
internal/language/    # Known languages (pretty name, slug, and file extension)
internal/config/      # Handles configuration
```

## License

GPL v3 — see [LICENSE](./LICENSE).
