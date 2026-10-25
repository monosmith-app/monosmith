# ADR 0036: External plugins get a minimal environment and need to be trusted

- Status: accepted
- Date: 2026-10-25

## Context

An external plugin is a program smith starts with the user's privileges. Until now smith started
every plugin, built in or external, with the whole environment of the process (`NPM_TOKEN`, cloud
credentials, smith's own cache key) and started external plugins without asking anyone: a
`smith.toml` that came with a cloned repository could name a `command`, and the first command that
opened the workspace, a read-only one included, ran it. Built-in plugins are the smith binary
itself and legitimately read release credentials from the environment. The plugin system is
described in [the plugin system](../plugin-system.md) and the overall plan is in
[#565](https://github.com/monosmith-app/monosmith/issues/565); this record is its first step.

## Options considered

- **Leave the environment alone and only ask for trust.** Rejected: a trusted plugin would still
  receive every secret of the shell it happens to be started from, whether it needs them or not.
- **A sandbox.** Out of scope for this step: it is a different mechanism per operating system, and
  the first step must not claim a protection it does not give.
- **Trust keyed by plugin name alone, or kept in the repository.** Rejected: the name is chosen by
  whoever writes `smith.toml`, and a file that arrives with a clone cannot vouch for itself.
- **Trust every plugin the workspace declares in CI.** Rejected: a pull request can change
  `smith.toml`. In CI the workspace owner has to pin the plugin's digest instead.

## Decision

**Environment.** Built-in plugins keep today's environment. An external plugin starts with a
documented baseline (`PATH`, the home and temporary directories, the variables a Windows process
needs, locale and terminal variables, proxy and certificate variables, `SMITH_PLUGIN` and
`SMITH_VERBOSE`) plus the variables it declares it needs: `env` in its `plugin.json`, and `env` in
`[plugins.external.<name>]` of `smith.toml`, which is the workspace owner's own grant. An entry is
a name or a prefix ending in `*` (`AWS_*`). The variables smith itself uses for a credential or a
release (`SMITH_CACHE_KEY`, `SMITH_CACHE_TOKEN`, `SMITH_RELEASE_*`, other `SMITH_` variables that
look like secrets) can be given to a plugin only by `smith.toml`, never by a manifest or a
pattern. `smith plugin list` and `smith doctor` show, by name and never by value, what each
external plugin may see. `smith plugin check` is for a plugin's author, so it needs no trust, but
it uses the same minimal environment; `--env NAME` grants a variable for the run.

**Trust.** smith starts an external plugin only when it is trusted. A trust decision is kept in the
user's smith home (`$SMITH_HOME/trust.json`, or `~/.smith/trust.json`) for one workspace, one
plugin name, the command line, and the SHA-256 of the files the command runs (the program, and a
script it is given). If the command or a file changes, the plugin is untrusted again.

- In a terminal, the first use shows the plugin (source, command, digests, environment) and asks
  `Trust "<name>"? [y/N]`.
- Without a terminal (`smith serve`, the shell, `smith mcp`, an editor extension, a script), smith
  never asks. It refuses with a problem and the remedy `smith plugin trust <name>`.
- In CI (`CI`, `GITHUB_ACTIONS` or `TF_BUILD` set), no decision of a person is on the machine. A
  plugin starts only when `smith.toml` pins `sha256` for it and the file matches. On a person's
  machine a pin does not make a plugin trusted.
- `smith plugin trust <name> [--yes] [--pin]` records the decision; `--pin` also writes the digest
  to `smith.toml` with the same text-preserving edit `smith config set` uses. `smith plugin untrust
  <name>` removes it.
- `smith doctor` lists the external plugins the workspace declares with their source, whether each
  is trusted and pinned, and what it may see, and fails with the remedy on one smith would refuse.
- Only the plugins `smith.toml` declares are needed by the workspace. A plugin that is merely
  installed on the search path is never started and needs no trust.

New slices: `pluginenv` (the environment), `plugintrust` (digests, the record, the decision) and
`pluginadmit` (asking or refusing, and the status shown by `list` and `doctor`). `pluginhost`
depends on them, not the reverse.

## Consequences

A trusted plugin still runs with the user's privileges: there is no sandbox, and the documentation
says so. The digest covers the program and the script named on the command line, not what a
launcher script starts or what a plugin downloads while it runs. Existing workspaces that declare
an external plugin are asked once per plugin (or refused with a remedy, in CI, until the digest is
pinned), and a plugin that read a variable it never declared stops seeing it. Network, folder and
program restrictions, signed plugins and a registry are not part of this step.

## Sources

- Issue [#565](https://github.com/monosmith-app/monosmith/issues/565)
- [Using plugins: plugin security](../../plugins/using/plugin-security.md)
- [The plugin system](../plugin-system.md)
