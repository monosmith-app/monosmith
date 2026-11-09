# The terminal hand-over in the shell

Principle: nothing the shell or `smith doctor` tells a person to run may need a terminal to
copy a command into. A command leaves the shell only when a child program is interactive by
nature. The catalogue (`libs/actions/catalog`) states that reason in `Terminal`, and a test
fails when an action needs a terminal without one.

## Where a command can leave the shell

| Command | Why it leaves, or how it runs in place |
| --- | --- |
| `dev`, `serve`, `mcp`, `lsp`, `agent` | By nature: a server or an interactive program that owns the terminal streams. Not offered in the menus. |
| `tools install` | Runs in place: a dialog shows the plan from `--dry-run --json`, the choice per tool, and the install output. It leaves only for a package manager password (`sudo`), and says so in the dialog. |
| `graph --format tui` | By nature: a full-screen browser of its own. |
| `compose`, `recipe new` (bare) | A questionnaire. The shell has its own Compose screen and recipe form, which run the same command with the answers. |
| `up` (bare) | An interactive picker. The shell opens the Dependencies tab, which shows the same candidates. |
| `new`, `adopt`, `convert` outside a workspace | Open gap: there is no run panel before a workspace exists, so the command runs in the terminal after the shell quits. `new` and `adopt` have a guided form and a wizard. `convert` has neither. |

Every other action runs in place: a confirmation dialog shows the plan (`--dry-run` or
`--check`) and the exact command, then the run panel streams the output. Commands that ask on
a terminal get `--yes` after the person agreed. This covers `upgrade`, `upgrade --defaults`,
`adopt` (in a workspace), `ci generate`, `kind`, `eject`, `derive` (`--reclaim`, `--replace`),
`release`, `sync`, `deps up`, `bootstrap`, `agent file`, `agent install`, `plugin install`,
`config set`, `commit`, `undo` and `cache prune`.

Open gaps in the choices: `sync` asks which version wins on a terminal. The shell shows the
plan and runs it with smith's pick; `--prefer highest|lowest` is not yet a button. The
release review (`release`) is a screen of its own, not the terminal prompt.

## Doctor

A check whose remedy names a smith command carries a `fix` made of steps whose first word is
`smith`. The host runs it with the smith that is running. The fix is never safe, so
`doctor --fix --yes` skips it and every interface asks. The review prints the plan of each
step (its plan flag) under the command.

| Check | Fix |
| --- | --- |
| The workspace uses the latest defaults | `smith upgrade --defaults` |
| The dependencies are at one version | `smith deps up --align` |
| A derived file is out of date | `smith derive` |
| A derived file was edited | asks: `--replace` (default), `--adopt` or `--eject` |
| The local cache is over its limit, or has unusable entries | `smith cache prune`, `smith cache verify --fix` |
| A lib nobody imports | asks: `smith kind <name> app` or `package` |
| A package nothing publishes | `smith kind <name> lib` |
| Python venv missing (plugin) | `smith ci setup` |
| CI pipelines out of date, agent files, tools, task groups | already a file or tool fix |

Still text only: findings of the plugins that say "run `smith ...`" and carry no fix, for
example the Rust, JVM and Playwright checks. A plugin adds a fix with steps `["smith", ...]`.

In the editors: the Doctor tab (`f`) and the Health view (the Fix row) run the same fix.
JetBrains: recorded only.
