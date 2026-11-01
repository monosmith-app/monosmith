# Feature parity: CLI, TUI, VS Code, JetBrains

Every capability starts in the CLI. The TUI (`smith` in a terminal), the VS Code extension and the JetBrains plugin offer the same function in their own medium. "Terminal" means the surface opens `smith` in an integrated terminal, so it has the TUI's behaviour. A dash means the surface has no such function yet.

| Feature | CLI | TUI | VS Code | JetBrains |
|---|---|---|---|---|
| List projects, affected marks | `smith projects`, `affected` | Projects tab | Projects view | Projects tool window |
| Run a task on one project or all | `smith run`, `build`, `test` | `b` `t` `l` `c`, palette, per-project tasks in the palette | Run Task, Run All, Run Affected | Run task action |
| Add a project with choices and siblings | `smith add --with --fresh --like` | Add form, plan with choices and siblings; the sibling question offers "Follow one project" and a list for `--like` | Add Project (quick picks, plan; "follow which project?" picks `--like`) | Tools > MonoSmith > Add Project (dialog, unticked commit box; no `--like` yet; unbuilt) |
| Language the Add form starts on | `smith add --lang` | the language most projects use, with the line "Preselected: N of the M projects here use X." | - (the quick pick lists kinds, no preselection) | - |
| Refuse a taken project name | `smith add` | live check in the form | input box check | - |
| Remove a project | `smith remove` | pick list, starts on the selected project | - | - |
| Commit a change so it can be undone | `--commit` on adopt, new, add, remove, convert, upgrade, derive, swap | unticked box "Commit this change so it can be undone" (key `c`, or a click on the box) in the confirmation of add, swap, derive and the others, also in the conversion review of the adopt wizard | "... and commit" buttons beside "Add project", "Swap" and each derive change | unticked box in the Add Project dialog (unbuilt) |
| Undo the last change | `smith undo` | palette, after a confirmation | Undo | Tools > MonoSmith > Undo (after a question; unbuilt) |
| Derive, explain | `smith derive [--dry-run]`, `smith derive --explain` | palette: "Check the derived files", "Preview the derived files" (`--dry-run`), "Update the derived files" | Health view; derive of every file offers "Preview" (`--dry-run`) | - (the Why action explains a file only) |
| Swap a choice of a project | `smith swap <project> <axis> <choice> [--dry-run] [--yes] [--commit]` | palette "Swap a choice": pick lists for project, decision and answer, the plan of `--dry-run` first, the commit box unticked | Swap a Choice (quick picks for project, axis and choice; the plan first) | - |
| Divergences from the defaults | `smith divergences` | palette "Divergences from the defaults" (override, ejected, drift, with the reason and the way back) | Show Divergences | - |
| Why a file or a setting is as it is | `smith why <file or setting>` | palette "Why is this here?": asks for a file, a folder or a setting | Why command: this file, a project, a task or a setting (asks for the setting path) | Why in the editor and project view menus, for a file (unbuilt) |
| Why the default of a setting is what it is | `smith why <setting>`, `smith config list --json` (`reason`) | Settings tab: a "Why:" line under the setting and in its editor | via Why > A setting (the extension has no editor of smith.toml) | - |
| Try a command in a copy | `smith try <command> [--checks] [--json] [--yes]` | palette "Try a command in a copy": asks for the command, shows the result in the Run view | Try a Command in a Copy (the page of changed files and checks; runs it for real on a yes) | - |
| Write the launcher | `smith bootstrap [--check]` | palette "Write the launcher" (after a confirmation) | Write the smith Launcher | - |
| Plugin draft from a sample | `smith plugin draft` | - (CLI only: a plugin author works at a terminal on a sample project; the catalog says so) | - | - |
| Doctor and safe fixes | `smith doctor [--fix]` | Doctor tab | Health view | Tools > MonoSmith > Doctor, Apply the Safe Fixes (Run window; unbuilt) |
| Release preview | `smith release --dry-run` | Release tab; the plan stays after a preview | Release view | Tools > MonoSmith > Release Preview (Run window; unbuilt) |
| Release channel notice | `smith release` | Release tab | Release view | - |
| Dependencies and one version | `smith deps` | Deps tab; `o` is hidden when `dependencies.one_version` is on | Dependencies view | - |
| Report | `smith report` | palette | Open Report | - |
| Agent file | `smith agent file` | palette | Agent File | - |
| Plugin search and install | `smith plugin search`, `install` | Settings tab | terminal | - |
| Plugin on or off | `smith plugin enable`, `disable [--force]` | Settings tab (the note counts the projects that disappear) | terminal | - |
| Settings (`smith config`) | `smith config set` | Settings tab | Settings command | - |
| Bring a folder in | `smith adopt`, `smith convert` | adopt wizard | terminal (`smith`) | - |
| Unreadable `smith.toml` | error with line and column | named in the header and wrapped in the Start screen; new and adopt are not offered | folder row: "smith failed: smith.toml:3:9 ..."; the adopt offer skips a folder with a smith.toml | tool window row: "smith failed: smith.toml:3:9 ... Try: ..." (unbuilt) |

## Rules

- The `--commit` box is unticked in every surface. smith commits only when asked, or when `commit_operations = true` is set.
- The undo hint appears only for a committed change. Otherwise the surface says "Not committed: review with git status".
- A check that lives in the CLI (the npm scope of `smith config set workspace.scope`, a plugin list) reaches every surface at once.
- `smith plugin disable` stops when projects use the plugin, and `--force` goes on. The TUI toggle is reversible, so it shows the count in its note instead of asking.
- The "did you mean" text of the CLI has no surface of its own: each surface lists the valid values in a pick list.
- JetBrains has no gradle build in the development environment of this change: the cells marked "unbuilt" are written and have unit tests for the pure parts, but nobody compiled them. The JetBrains gaps left are swap, derive, divergences, `--like`, why for a setting, try and bootstrap.
