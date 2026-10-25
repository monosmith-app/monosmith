# Screenshots

Every image here was captured from a real run of `smith` in a scratch
workspace (`acme`: a Go app and lib, two Node projects and a Python package). None is
mocked up.

| Image | Shows | How it was made |
|---|---|---|
| `graph-html.png` | the project graph page | `smith graph --format html -o graph.html`, opened in a browser and screenshotted |
| `run-cache-hit.png` | `smith test` replaying every task from the cache | terminal output, rendered by `render-terminal.mjs` |
| `why-miss.png` | `smith why-miss` after a change to `ui` | terminal output, rendered |
| `explain.png` | `smith explain web:test` | terminal output, rendered |
| `doctor.png` | `smith doctor` on a healthy workspace (the middle of the list cut to `...`) | terminal output, rendered |
| `doctor-failing.png`, `doctor-fix.png` | `smith doctor` and `smith doctor --fix` with a stale pipeline and a `go.work` missing a module (the passed checks cut to `...` in the first) | terminal output, rendered; re-captured October 2026 with smith 1.62.0 because doctor now lists the passed checks after the failed ones |
| `release-preview.png` | `smith release --dry-run` | terminal output, rendered |
| `graph-text.png` | `smith graph` | terminal output, rendered |
| `ci-yml.png` | lines of the generated `.github/workflows/ci.yml` | `sed -n '35,76p'`, rendered |

## Rendering terminal output

`render-terminal.mjs` turns captured text (with or without ANSI colours) into
a PNG that looks like a terminal window:

```sh
smith test > run.txt 2>&1
node docs/assets/render-terminal.mjs run.txt run-cache-hit.png --cmd "smith test" --title acme --width 820
```

It needs the `playwright` package in the directory you run it from
(`npm i -D playwright`). With no downloaded Chromium, set
`RENDER_CHANNEL=chrome` (or `msedge`) to use an installed browser. An output
name ending in `.html` writes the page instead.

Keep each image under about 300 KB.

## Last checked

October 2026, smith 1.62.0. `doctor.png`, `doctor-failing.png`, `doctor-fix.png`,
`explain.png`, `why-miss.png`, `release-preview.png` (now captured with
`--format markdown`) and `ci-yml.png` (the install step is pinned to the version that
generated the file) were captured again because the output they show had changed.
`run-cache-hit.png`, `graph-text.png` and `graph-html.png` still match what smith prints.
