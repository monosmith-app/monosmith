# Turborepo: install notes

## What was installed

- `turbo` 2.11.10, installed locally: `npm i turbo@latest` in the sample workspace. No global install.
- Windows is supported natively. WSL was not needed.

## Sample workspace

`.tools/samples/turboplay` is an npm workspace with the same 3 projects as the Nx sample (`lib-b`, `app-a`, `srv-c`).

```text
turboplay/
  package.json         workspaces: packages/*, packageManager npm@10.9.0
  turbo.json           "ui": "tui"; build dependsOn ^build; test dependsOn build; dev is persistent, cache off
  packages/
    lib-b/  app-a/  srv-c/     scripts: build, test, lint, dev (app-a lint fails on purpose)
```

## Commands used

- `turbo run build lint dev --concurrency=10 --force` (first run).
- `turbo run build lint dev --concurrency=10 --force --continue=always` (all later runs).
- Environment: `TURBO_TELEMETRY_DISABLED=1`. The command was started as `cmd.exe /c node_modules\.bin\turbo.cmd ...`.

## Things to know

- In the first run, the failing `app-a#lint` stopped every task, including the `dev` servers, in 1.3 seconds.
  The TUI then closed and turbo printed the plain logs and a summary. With `--continue=always` the TUI stayed open.
- The log pane showed some garbled characters at the left edge in several captures (for example `a>lnodeu-en"letai=0;...`).
  The task names in the left pane overflow into the log pane at 130 columns. I cannot say whether the real terminal shows the same.
