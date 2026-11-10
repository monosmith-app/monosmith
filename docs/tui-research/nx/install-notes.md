# Nx: install notes

## What was installed

- `nx` 24.0.5, installed locally: `npm i nx@latest` in `.tools/npm` and in the sample workspace. No global install.
- Windows is supported natively. WSL was not needed.

## Sample workspace

`.tools/samples/nxplay` is an npm workspace with 3 projects. Nx infers the targets from each `package.json`.

```text
nxplay/
  nx.json              "tui": { "enabled": true }, build depends on ^build
  packages/
    lib-b/             build 3 s, test 2 s, lint ok, serve (continuous)
    app-a/             depends on lib-b; build 4 s, test 2.5 s, lint FAILS (exit 1), serve (continuous)
    srv-c/             depends on lib-b; build 2.5 s, test 1.5 s, lint ok, serve (continuous)
```

`serve` has `"nx": { "targets": { "serve": { "continuous": true } } }`. It never ends, so it tests the long-running case.

## Commands used

- `nx run-many -t build lint test --parallel=2 --skip-nx-cache --no-cloud`
- `nx run-many -t build lint serve --parallel=6 --skip-nx-cache --no-cloud` (the main capture run)
- Environment: `NX_TUI=true NX_DAEMON=false`.

## Things to know

- The first attempt, `nx run-many -t build lint test serve` without `NX_TUI=true`, printed plain text and no TUI,
  although `nx.json` had `"tui": { "enabled": true }`. With `NX_TUI=true` the TUI started. I did not find the cause.
- The TUI took about 2 seconds to draw its first frame.
- Some frames lagged behind my keys by several seconds (machine load). Captures that I could not match to a key were deleted.
- Two captures (`08-help-overlay-partial-render`, `03-output-pane-split-view`) show overlay text that is drawn only in part.
  This can be a limit of my terminal model, not an Nx defect. I could not tell which.
