# smith report

`smith report` writes one self-contained HTML page about the workspace. Attach it to a pull
request, or open it in a browser.

The page has five sections.

| Section | Shows |
|---|---|
| Workspace map | An inline dependency graph, and the same facts as a table. |
| Health | The results of [`smith doctor`](doctor.md): failures and warnings first, passes in a closed list. |
| Derived files | The files [`smith derive`](derive.md) would change, and how many are current. |
| Stack and choices | The stack, the choices and the conventions of each project. |
| Release plan | What a merge of the current commit to the release branch would release. |

The page has no script and makes no request. Every status is a word (Pass, Warning, Fail),
so colour is never the only signal. Tables have captions and header cells, the sections have
a skip link, and the page follows the light or dark setting of the browser. It holds no time
and no machine path, so the same workspace gives the same page.

The release plan reads the commits since the last tag of each package. It asks no registry
and does not check the public API, so it is a preview. [`smith release --dry-run`](release.md)
checks more.

## When to use it

- To show a reviewer the shape and the health of the workspace in one file.
- To attach what a merge would release to the pull request.

For the graph alone, use [`smith graph --format html`](graph.md).

## Synopsis

```sh
smith report [-o <file>] [--force]
```

## Arguments and flags

<!-- smith:begin report -->
Write one HTML page: the workspace map, health, derived-file drift, choices and release plan

Usage: `smith report [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--force` | boolean | off | replace the file when it exists with other content (the diff is printed) |
| `-o, --output <output>` | string |  | write the page to a file |

<!-- smith:end report -->

Without `-o` the page goes to the standard output.

## Where the page goes

A file that exists with other content is not replaced silently: smith names the file, shows a diff and asks for `--force`.

Without `-o`, the page goes to the standard output. When the standard output is a terminal, it goes to
`smith-report.html` instead, because a page of HTML on a terminal is noise.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The page was written, whatever it reports. |
| 1 | The workspace cannot be opened, or the file cannot be written. |

## Files and environment

Reads the workspace like `smith doctor` does, and the git history for the release plan.
Writes only the file named by `-o`.

## Examples

```sh
smith report -o smith-report.html
```

In GitHub Actions, upload the file as an artifact:

```yaml
- run: smith report -o smith-report.html
- uses: actions/upload-artifact@v4
  with:
    name: smith-report
    path: smith-report.html
```

## Related

[`smith graph`](graph.md), [`smith doctor`](doctor.md), [`smith derive`](derive.md),
[`smith release`](release.md).
