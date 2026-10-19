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
| `doctor-failing.png`, `doctor-fix.png` | `smith doctor` and `smith doctor --fix` with a stale pipeline and a `go.work` missing a module | terminal output, rendered |
| `release-preview.png` | `smith release --dry-run` | terminal output, rendered |
| `graph-text.png` | `smith graph` | terminal output, rendered |
| `ci-yml.png` | lines of the generated `.github/workflows/ci.yml` | `sed -n '35,76p'`, rendered |
| `tui-menu.png`, `tui-commands.png`, `tui-doctor.png`, `tui-release.png`, `tui-settings.png` | the interactive shell's tabs | see below |

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

## The interactive shell

The shell needs a real terminal, so its screens were captured by building a
throwaway program that renders the same Bubble Tea model `smith` opens
(`tui.NewShell` with `tui.WorkspaceSources`, and a key sequence for each tab),
writing each `View()` with its colours, and rendering that text with the script
above. The header reads `smith dev` because the binary was a development build.
That program is not part of the repository.

Keep each image under about 300 KB.
