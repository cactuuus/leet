# Leet

A small CLI for practicing LeetCode problems locally, in your own editor, without copy-pasting
between the browser and your files. \
It provides quality-of-life features that allow you to solve problems without interacting with its web interface (which we all know and love 💛).

> ***Submission is not implemented yet*** - though I am getting there!

## Install

> Requires [**Go 1.26**](https://go.dev/doc/install) or higher (lower might work but was not tested with). \
> Make sure `~/go/bin` is on your `PATH`.

```bash
go install github.com/cactuuus/leet@latest
```

## Usage

### Load a problem

This fetches the problem's details (such as description, code snippets, and testcases) and scaffolds it into a local folder.

```bash
leet load 2135                       # uses your configured default languages
leet load daily                      # today's daily challenge
leet load 2135 --langs go,python3    # override languages to load
leet load 2135 --open                # also opens the problem folder afterward
```

For example, the command `leet load 22 -l go,python3` fetches problem 22 (Generate Parentheses) and creates:

```bash
~/leet-problems/22.generate-parentheses/
├── desc-22.html
├── 22.go
└── 22.py
```

### Open a problem

This opens the problem folder in your configured editor (or the whole problems directory if no number is given).

```bash
leet open 2135    # opens problem 2135 folder
leet open daily   # opens today's challenge folder
leet open         # opens the root problems directory
```

### Run problem tests

If you have set your dredentials in the config file, you can test your solution on leetcode. \
The testcases used are stored in `testcases-<number>.txt` inside of the problem folder, where each argument is separated by a newline, and each testcase is separated by a divider (`---`). This file is initially setup with the default 'example' testcases, but you can manually edit this file, adding or removing testcases as you please.\

```bash
leet test 2135 go         # runs testcases for problem 2135 using your Go solution
leet test daily python3   # runs testcases for today's challenge using your Python solution
```

### Templating

Leetcode's code snippets often miss some boilerplate code, meaning that these snippets are not always valid programs (for example, Go missing package declaration), or require you to manually import some packages that are usually already automatically included by leetcode. \
To fix this, each language comes with a template file that 'wraps' the code snippet. This allows to add any missing boilerplate code (or anything else you want) knowing that it would be ignored when submitting to leetcode (such as via the `leet test` command). \
These templates can be customized according to your preferences. If a custom template doesn't exist for a language, the default one will be used.

```bash
leet template make go         # creates a custom Go template, based on the default one
leet template open python3    # opens the custom Python template in your editor
leet template open            # opens the custom template folder in your editor
leet template delete java     # deletes the custom Java template.
leet template list            # lists all languages with a custom template
```

### Configuration

Config lives at `~/.config/leet/config.toml`. It is meant to be edited directly, with comments explaining each option. \
Here you can specify your LeetCode credentials (session and CSRF tokens). If specified, these will automatically be added to your requests, allowing you to run testcases and (eventually!) submit solutions.

```bash
leet config show  # show current configuration
leet config edit  # open the config file in your editor
leet config reset # reset config to defaults
```

## Project layout

```bash
main.go                   # entrypoint
├── cmd/                  # CLI commands (using Cobra)
└── internal/
    ├── auth/             # Credentials model
    ├── config/           # Configuration management
    ├── language/         # Language definitions and templates
    ├── leetcode/         # LeetCode API client
    |   └── cache/        # Cache used by the LeetCode API client
    ├── problem/          # Problem domain models
    └── scaffold/         # File management
```

## License

GPL v3 — see [LICENSE](./LICENSE).
