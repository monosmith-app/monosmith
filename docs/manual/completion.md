# smith completion

`smith completion` sets up tab completion for smith in your shell, when you ask for it. Smith
never sets it up on its own: `install` shows the files and the lines it will write, and asks
first. `uninstall` removes exactly what `install` added.

## When to use it

Once per machine and shell, to complete commands, flags and project names with Tab. It
supports `bash`, `zsh`, `fish` and `powershell`.

## Synopsis

```sh
smith completion install [--shell zsh|bash|fish|powershell] [--yes] [--dry-run]
smith completion uninstall [--shell <shell>] [--yes]
smith completion status
smith completion bash | zsh | fish | powershell     # print the script, as before
```

## What install does

1. It finds your shell: the program that started smith, then `$PSModulePath` (outside
   Windows), then `$SHELL`, and on Windows PowerShell. `--shell` overrides this.
2. It prints each file and each line it will write, then asks: **Yes**, **No** or **Show me
   the script**. `--yes` skips the question, for scripts and CI. Without `--yes` and without a
   terminal, install fails and writes nothing. `--dry-run` prints the plan and writes nothing.
3. It writes the script to a file that smith owns, and adds **one line** to your profile that
   sources that file. Smith never pastes the script into your profile. Every file and every
   profile line sits between two marker lines, so `uninstall` finds it again:

   ```text
   # >>> smith completion >>>
   # Managed by "smith completion install". Remove it with "smith completion uninstall".
   ...
   # <<< smith completion <<<
   ```

The script files live in the per-user folder of smith, `completions` in the smith home
(`$SMITH_HOME`, or `~/.smith`), next to `tui.json`. Every profile line is guarded, so a deleted
script never breaks the start of your shell: `[ -f file ] && source file` in bash and zsh,
`if (Test-Path file) { . file }` in PowerShell.

A second run changes nothing and adds no second line. A profile you wrote is only appended to,
and only between the markers. Nothing else in it is touched, and smith makes no backup because
it changes nothing but its own line. When the script file comes from another smith version
(`smith completion status` says `stale`), `install` rewrites that file alone: the profile
already has its line, so it is not edited.

| Shell | Script file | Line goes into |
|---|---|---|
| bash | `~/.smith/completions/smith.bash` | `~/.bashrc` |
| zsh | `~/.smith/completions/smith.zsh` | `~/.zshrc` (`$ZDOTDIR` is honoured). The line loads `compinit` first when your `.zshrc` has not. |
| fish | `~/.config/fish/completions/smith.fish` (`$XDG_CONFIG_HOME` is honoured) | none: fish loads that folder on its own |
| powershell | `~/.smith/completions/smith.ps1` | `$PROFILE`. Windows PowerShell 5.1: `Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1`. PowerShell 7: `Documents\PowerShell\...`. When smith cannot tell which one runs it, the plan lists both. A `Documents` folder that OneDrive moved is followed. Elsewhere: `~/.config/powershell/Microsoft.PowerShell_profile.ps1`. If PowerShell blocks the profile, run `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`. |

A script file that already has the name smith would write, and holds no marker lines, is not
smith's: install refuses to overwrite it and says so. Bash on macOS reads `~/.bash_profile` for a
login shell; source `~/.bashrc` from it, as usual.

### Packages

When a package of smith (Homebrew, a Linux package) already ships the completion file for your
shell, in a system folder that the shell reads, `install` says so and does nothing. The files
that count are `.../share/bash-completion/completions/smith`, `.../etc/bash_completion.d/smith`,
`.../share/zsh/site-functions/_smith` and `.../share/fish/vendor_completions.d/smith.fish`, under
`/usr`, `/usr/local`, `/opt/homebrew`, Linuxbrew or `$HOMEBREW_PREFIX`.

## The one-time offer

After your first interactive run (the terminal interface, or `smith doctor` in a terminal), smith
asks once: **Set up tab completion?** It shows the files first. The answers:

| Answer | Effect |
|---|---|
| Yes | Writes the files, as `install` does. |
| No, do not ask again | Remembered. Smith never asks again; `smith completion install` still works. |
| Ask me later (Enter) | Asks again after two weeks. |
| Show me the script | Prints the script, then asks again. |

Smith never asks in CI, in a pipe, or in a command that serves other programs (`smith serve`,
`smith mcp`, `smith lsp`). The answer is kept for you only, in `tui.json` in the smith home
(`$SMITH_HOME`, or `~/.smith`), as `completionOffer` (`no`, `later` or `installed`) and
`completionAskAfter`. It never goes into `smith.toml`. `smith completion status` shows it.

The terminal interface and the editor extensions offer the same choice: see
[the terminal interface](tui.md#home) and [editors](../guides/editors.md).

## status

Prints one line for each shell: where the files are (with `stale` when the script file comes
from another smith version), that a package ships the file, or `not installed`, with
`<- your shell` on the shell found, and the answer you gave to the offer.
`--json` prints `{"shell": "zsh", "wanted": true, "answer": "", "stale": false}` for editors: `wanted` is true when
nothing is installed for the shell, no package ships the file and you did not say no.

## uninstall

Lists the files that hold a block, asks (`--yes` skips the question) and removes the profile
line and the script file, with the blank line install put before the line. A file that holds
nothing else is deleted. Without `--shell` it looks at every shell.

## Exit codes

0 on success, including "nothing to do". 1 when install cannot ask (no terminal and no `--yes`),
when a file in the way is not smith's, when the shell is not known or when a write fails.

## Files and environment

Writes only the files in the table, and `tui.json` in the smith home. Reads `SHELL`,
`PSModulePath`, `SMITH_HOME`, `XDG_CONFIG_HOME`, `ZDOTDIR` and `HOMEBREW_PREFIX`. No workspace
is needed.

## Examples

```sh
smith completion install                       # shows the plan, asks, writes
smith completion install --shell powershell --dry-run
smith completion install --yes                 # CI
smith completion status
smith completion uninstall
```

## Related

[Installing](../getting-started/installing.md#tab-completion), [`smith commands`](commands.md).
