# Show smith in your shell prompt

`smith prompt` prints one line about the workspace you are in, and prints nothing anywhere
else. It reads what the last `smith doctor` run saved and asks git for the changed files, so
it answers in about 50 ms. It never runs doctor and never starts a plugin. See
[`smith prompt`](../manual/prompt.md) for what each part of the line means.

```text
smith 2.2.0 · 34 projects · × 3 ▲ 4 · actions, smith
```

Run `smith doctor` once, and again when you want fresh numbers: the failure and warning
counts are the ones of that run. Before the first run the line is only the version.

`smith prompt --json` prints the same facts for a custom segment:

```json
{"version":"2.2.0","projects":34,"doctor":{"failed":3,"warnings":4,"checkedAt":"2026-11-10T00:28:10Z"},"affected":["actions","smith"]}
```

## Oh My Posh

Add a `command` segment to your theme. The segment hides itself when the command prints
nothing, so it shows only inside a workspace.

```json
{
  "type": "command",
  "style": "plain",
  "foreground": "#ffaf00",
  "template": " {{ .Output }} ",
  "properties": {
    "shell": "bash",
    "command": "smith prompt"
  }
}
```

On Windows with PowerShell, use `"shell": "pwsh"`.

## Starship

Add a custom command module to `~/.config/starship.toml`. Starship hides the module when the
command prints nothing.

```toml
[custom.smith]
command = "smith prompt"
when = true
format = "[$output]($style) "
style = "yellow"
```

## powerlevel10k

Define a segment function in `~/.p10k.zsh`, then add `smith` to
`POWERLEVEL9K_LEFT_PROMPT_ELEMENTS` (or the right-hand list).

```zsh
function prompt_smith() {
  local line
  line=$(smith prompt 2>/dev/null) || return
  [[ -n $line ]] && p10k segment -f 3 -t "$line"
}
```

## Notes

- The line is local to the working tree. It lists the projects with changed or untracked
  files and every project that depends on them; it does not compare with a base branch (use
  [`smith affected`](../manual/affected.md) for that).
- The data comes from `.smith/prompt.json`, which a normal `smith doctor` run writes. `smith
  doctor --fix` does not write it.
- A large untracked tree makes `git status` slow, and so the prompt. Keep build output in
  `.gitignore`.
