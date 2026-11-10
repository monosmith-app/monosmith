<!-- lore-master
parent: cache.md
-->
# smith cache setup

`smith cache setup` sets up the remote cache with a list of choices. Each choice says what its
options do. smith runs a command, writes a file or stores a secret only after you confirm, and
shows the exact command first. The background (what the key is, rotation, SAS token and key,
one prefix for each workspace) is in
[Set up the remote cache](../reference/tasks-and-cache.md#set-up-the-remote-cache).

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
   `az storage container generate-sas` and copy it, unseen. You choose how it is signed: as
   yourself (`--auth-mode login --as-user`; Azure limits it to 7 days, so you renew it weekly) or
   with the account key (`--auth-mode key`; it can last months, and rotating the account key voids
   it). Signing in to Azure through `az` directly, with no token, is not supported by the backend
   today.
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
| `--clear-after <clear-after>` | string |  | clear the clipboard after this long if it still holds the key; 0 leaves it |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--json` | boolean | off | print the questions as JSON (with --questions) |
| `--questions` | boolean | off | print the questions smith cache setup asks, with their answers, and change nothing (needs --json) |

<!-- smith:end cache setup -->

`--questions --json` prints the questions (`key`, `prompt`, `answers` with `label`, `value` and
`hint`, `free` for typed text, `after` for the answer that makes a question appear). An editor
can ask the same questions. The choices inside a cloud come from the cloud's own program at that
time, so those questions have no fixed answers.

## Needs a terminal

The command asks questions and reads single key presses, so it needs a terminal. In a pipe it
stops and names `smith config set cache.remote`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The setup ended, with or without a change. |
| 1 | Not a workspace, no terminal, or the setting could not be written. |
