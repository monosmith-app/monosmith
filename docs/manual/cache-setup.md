<!-- lore-master
parent: cache.md
-->
# smith cache setup

`smith cache setup` sets up the remote cache with a list of choices. Each choice says what its
options do. smith runs a command, writes a file or stores a secret only after you confirm, and
shows the exact command first. The background (what the key is, rotation, SAS token and key,
one prefix for each workspace) is in
[Set up the remote cache](../reference/tasks-and-cache.md#set-up-the-remote-cache).

## Going back

You can go back to the question before the one on screen and change the answer. In the
interactive shell, press Esc, `b` or Left, or click **Back (b)**; the earlier answer is marked
again. In a list, `b` goes back only when no entry starts with `b`; Left and the button always
do. On the first question Esc closes the setup and nothing more is changed. Ctrl+C or the
**Cancel** button closes it from any screen. In a terminal, type `b` at a choice.

A listing with `az`, `aws` or `git` only reads, so smith goes back over it and runs it again
when you reach it again, and the list is current. A step that changed something is a point of no
return: a sign-in, the write to `smith.toml`, the key or token screen, and the pipeline
variables. smith does not go back over such a step and does not run it again. It says "That step already ran; nothing was
undone." and asks the same question again. A key or token is never shown again after you leave its
screen.

## The steps

1. **The kind of storage:** Azure Blob, S3 or an HTTPS server. Google Cloud Storage is not
   supported.
2. **Where.** For Azure, smith uses the `az` program with read-only commands, each shown before
   it runs: `az account list`, then `az storage account list --subscription ...`, then
   `az storage container list --account-name ... --auth-mode login`, then
   `az storage blob list --delimiter /` for the folders. You pick a base folder such as `smith`,
   and smith proposes a sub-folder named after the workspace. For S3 it is the same with `aws s3api
   list-buckets` and `list-objects-v2`. For HTTPS you type the address. When `az` or `aws` is
   missing, smith says so and you type the names. When nobody is signed in, smith offers the
   sign-in command (`az login`, `aws sso login`) as a choice and never runs it unasked.
3. **The setting.** smith shows the change to `smith.toml` (`[cache] remote = "azblob://account/container/smith/<workspace>"`)
   and writes it after you say yes, as `smith config set cache.remote` does: comments stay.
   Nothing is committed unless you pass `--commit`.
4. **The key.** The same screen as [`smith cache key`](cache-key.md), with one more choice for
   each provider that has a command smith can run: `h` stores the key as a GitHub repository
   secret (`gh secret set SMITH_CACHE_KEY`) and `l` as a masked GitLab variable
   (`glab variable set SMITH_CACHE_KEY --masked`). The command is shown, you confirm with `y`,
   and the key goes in through standard input, never as an argument.
5. **The SAS token (Azure).** smith offers to make the container's token with
   `az storage container generate-sas` and copy it, unseen. The token has read, write, delete,
   create and list permission (`rwdcl`); delete and list let `smith cache clear --remote` work. A
   token made by an older smith has no delete right: make a new one for the remote clear. You choose how it is signed: as
   yourself (`--auth-mode login --as-user`; Azure limits it to 7 days, so you renew it weekly) or
   with the account key (`--auth-mode key`; it can last months, and rotating the account key voids
   it). Signing in to Azure through `az` directly, with no token, is not supported by the backend
   today.
   **Azure Pipelines.** When `azure` is in `[ci] providers`, smith offers to set the secret
   variables in the pipeline itself, before the key or token screen opens. You say yes, and smith
   reads `git remote get-url origin` (the `dev.azure.com`, `ssh.dev.azure.com` and
   `<org>.visualstudio.com` forms) and runs `az pipelines list --organization ... --project ...
   --repository ... --repository-type tfsgit`. It needs `az` signed in, with the `azure-devops`
   extension. When several pipelines match, you choose one. smith reads the pipeline's variables:
   for a name that exists you choose "Replace it" or "Leave it", and nothing is overwritten
   unasked. The key screen (and the token screen) then has the button `a`, "set it in Azure
   Pipelines (az)". It shows the organisation, project, pipeline and the variable name, with "will
   be created" or "will be replaced" (never the value), and writes only after `y`. Each variable is
   a secret (`isSecret`). The value never goes on a command line: smith reads the pipeline's
   definition with `az devops invoke --area build --resource definitions`, adds the variable, and
   writes it back with `--http-method put --in-file <file>`. The file lives in a private temporary
   folder that only you can read, and smith removes it when the call ends. When Azure says the
   definition changed in between, smith reads it again once. When `az`, the extension or the
   sign-in is missing, smith names which one and shows the manual steps. `smith ci generate`
   already maps `SMITH_CACHE_KEY` and `AZURE_STORAGE_SAS_TOKEN` into the steps with `env:`.
6. **Each CI provider** in `[ci] providers` (GitHub, Azure Pipelines, GitLab, CircleCI,
   Bitbucket): the web steps for the secrets, which secrets (`SMITH_CACHE_KEY` and the backend's
   credential), the command-line form where there is one, and how pull requests and forks behave.

## Synopsis

```sh
smith cache setup [--commit] [--clear-after <duration>]
smith cache setup --questions --json
```

## Arguments and flags

<!-- smith:begin cache setup -->
Set up the remote cache step by step: where, the key, and each CI provider's secrets

Usage: `smith cache setup [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--clear-after <clear-after>` | duration (`60s`, `2m`, `0`) | `60s` | clear the clipboard after this long if it still holds the key; 0 leaves it |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--json` | boolean | off | print the questions as JSON (with --questions) |
| `--questions` | boolean | off | print the questions smith cache setup asks, with their answers, and change nothing (needs --json) |

<!-- smith:end cache setup -->

`--questions --json` prints the questions (`key`, `prompt`, `answers` with `label`, `value` and
`hint`, `free` for typed text, `after` for the answer that makes a question appear). An editor
can ask the same questions. The choices inside a cloud come from the cloud's own program at that
time, so those questions have no fixed answers.

## In the interactive shell

A bare `smith` opens the same setup as screens. Press Enter on `cache.remote` in the Settings tab,
or open the command palette (`:`) and choose "Set up the remote cache". The decisions, the
commands and the order are the ones above, because both read the same engine:

- Each choice is a list: move with the arrows, press Enter or a number, or click and double-click.
- A command is shown in full before it runs, with the buttons "Run it" (`y`) and "Type it by hand"
  (`n`). While `az` or `aws` answers, the screen shows a spinner and says what it waits for.
- The key screen has the buttons `[ Copy again  c ]`, `[ Generate another  g ]`, the user variable
  button `u`, one button for each provider that has a command (`h` GitHub, `l` GitLab, `a` Azure
  Pipelines, when you said yes to its offer) and
  `[ Done  q ]`. Click a button or press its key. The key is on the clipboard and never on the
  screen. It stays in memory only while the screen is open. After 60 seconds, or on Done, the
  clipboard is cleared if it still holds the key. A command that sends the key shows the command
  first and runs only after `y`, with the key on standard input.
- The steps of each provider in `[ci] providers` are one page each: Enter goes on.
- Esc goes back to the question before, and closes the setup on the first question. Nothing is written before the "Write it?" answer, so closing before it
  changes nothing.

The "Write it?" screen has the unticked box "Commit this change so it can be undone" (`c` ticks
it, or click it). smith commits only when you tick it, or when `[workspace] commit_operations` is
on and the work tree was clean, through the same operation log as `--commit`; `smith undo`
reverts it. Without the tick, smith.toml changes, comments stay, and nothing is committed.
Typing `cache.remote` by hand in the Settings tab (`smith config set`) shows no such box: no
Settings change does.

## Needs a terminal

The command asks questions and reads single key presses, so it needs a terminal. In a pipe it
stops and names `smith config set cache.remote`.

An editor cannot confirm the commands this setup runs, so the VS Code command "Cache: Set Up the
Remote Cache in a Terminal" opens a MonoSmith terminal and runs `smith cache setup` in it.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The setup ended, with or without a change. |
| 1 | Not a workspace, no terminal, or the setting could not be written. |
