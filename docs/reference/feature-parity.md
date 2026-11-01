# Feature parity: CLI, TUI, VS Code, JetBrains

Every capability starts in the CLI. The TUI (`smith` in a terminal), the VS Code extension and the JetBrains plugin offer the same function in their own medium. "Terminal" means the surface opens `smith` in an integrated terminal, so it has the TUI's behaviour. A dash means the surface has no such function yet.

| Feature | CLI | TUI | VS Code | JetBrains |
|---|---|---|---|---|
| List projects, affected marks | `smith projects`, `affected` | Projects tab | Projects view | Projects tool window |
| Run a task on one project or all | `smith run`, `build`, `test` | `b` `t` `l` `c`, palette, per-project tasks in the palette | Run Task, Run All, Run Affected | Run task action |
| Add a project with choices and siblings | `smith add --with --fresh` | Add form, plan with choices and siblings | Add Project (quick picks, plan) | - |
| Refuse a taken project name | `smith add` | live check in the form | input box check | - |
| Remove a project | `smith remove` | pick list, starts on the selected project | - | - |
| Commit a change so it can be undone | `--commit` on adopt, new, add, remove, convert, upgrade | unticked box "Commit this change so it can be undone" (key `c`) | "Add and commit" buttons beside "Add project" | - |
| Undo the last change | `smith undo` | palette, after a confirmation | Undo | - |
| Derive, explain, why | `smith derive`, `smith why` | palette | Health view, Why command | - |
| Doctor and safe fixes | `smith doctor [--fix]` | Doctor tab | Health view | - |
| Release preview | `smith release --dry-run` | Release tab; the plan stays after a preview | Release view | - |
| Release channel notice | `smith release` | Release tab | Release view | - |
| Dependencies and one version | `smith deps` | Deps tab; `o` is hidden when `dependencies.one_version` is on | Dependencies view | - |
| Report | `smith report` | palette | Open Report | - |
| Agent file | `smith agent file` | palette | Agent File | - |
| Plugin search and install | `smith plugin search`, `install` | Settings tab | terminal | - |
| Plugin on or off | `smith plugin enable`, `disable` | Settings tab (toast warns what is hidden) | terminal | - |
| Settings (`smith config`) | `smith config set` | Settings tab | Settings command | - |
| Bring a folder in | `smith adopt`, `smith convert` | adopt wizard | terminal (`smith`) | - |
| Unreadable `smith.toml` | error with line and column | named in the header; adopt is not offered | folder row: "smith failed: smith.toml:3:9 ..." | - |

## Rules

- The `--commit` box is unticked in every surface. smith commits only when asked, or when `commit_operations = true` is set.
- The undo hint appears only for a committed change. Otherwise the surface says "Not committed: review with git status".
- A check that lives in the CLI (the npm scope of `smith config set workspace.scope`, a plugin list) reaches every surface at once.
- JetBrains has no gradle build in the development environment of this change. Its parity gaps are not closed yet.
