# Plugin security: what smith protects and what it does not

An external plugin is a program that smith starts on your machine. This page says what smith does
about that, in terms of what you will see, and, just as plainly, what it does not do. For the steps
to trust, pin or untrust a plugin, see [Installing, enabling and configuring
plugins](managing-plugins.md#trusting-an-external-plugin).

## The short version

- **Built-in plugins** are the smith binary itself. They need no trust and keep the environment
  of the shell smith runs in, because releasing needs credentials from it.
- **External plugins** (declared in `smith.toml` under `[plugins.external.<name>]`) start only
  after you have trusted them, and start with a small environment: not your whole shell.
- **Trusting a plugin is not sandboxing it.** A trusted plugin runs with your privileges. It can read
  your files, use your network and start other programs. smith has no sandbox. Trust a plugin only
  as far as you would trust any program you run.

## What smith protects against

| Risk | What smith does |
|---|---|
| A repository you cloned names a plugin in `smith.toml`, and the first command you run in it would start that program | smith does not start an external plugin until you have trusted it. Opening the workspace with any command, a read-only one included, is refused with the command to review it: `smith plugin trust <name>` |
| A plugin that was fine yesterday is replaced or edited | A trust decision is for one workspace, one plugin name, one command and the files that command runs. If the command or a file changes, smith asks again |
| A plugin reads secrets it does not need (`NPM_TOKEN`, cloud keys, smith's own cache key) | An external plugin gets a [minimal environment](#the-environment-an-external-plugin-gets) plus the variables it declares. Nothing else in your environment reaches it |
| A plugin quietly asks for smith's own credentials | A plugin's own `plugin.json` cannot ask for the variables smith uses for its cache and releases (`SMITH_CACHE_KEY`, `SMITH_RELEASE_*` and the like). Only the workspace owner can, in `smith.toml` |
| A CI job runs whatever plugin a pull request adds | In CI nobody is there to ask, so an external plugin starts only when `smith.toml` pins its `sha256` and the file still matches |
| A tool or an editor opens the workspace and starts a plugin nobody looked at | `smith serve`, `smith mcp`, the editor extension and scripts never ask. They refuse with the same remedy |

You can always see what is going on: `smith plugin list` shows, for every external plugin the
workspace uses, where it comes from, whether it is trusted and pinned, and the environment
variables it may see (names only, never values). `smith doctor` reports a plugin that smith would
refuse to start.

## What smith does not protect against

- **A plugin you trusted.** Once you say yes, the plugin runs as you. If its code is hostile, or it
  becomes hostile later and you trust the new version, nothing stops it. Read the code of a plugin
  before you trust it, and prefer plugins whose code you can read.
- **Anything a plugin does after it starts.** The environment limit covers what smith hands the
  plugin when it starts it. A plugin can still read files, including files that hold credentials
  (`~/.npmrc`, `~/.aws`, a credentials store), and it can start programs that inherit what they need.
- **What a plugin runs.** smith checks the program that starts the plugin and, for an interpreter, the
  script it is given (`node plugin.mjs`). It does not look inside a launcher that starts something
  else, a package a plugin loads, or anything a plugin downloads while it runs.
- **The built-in plugins and smith itself.** They are part of the program you installed.
- **A grant that is wider than it should be.** A grant gives the plugin exactly what it says:
  `AWS_*` is every variable that starts with `AWS_`. Grant the names a plugin needs.
- **Network access, folders and programs.** smith does not restrict where a plugin may connect, which
  folders it may touch or what it may start. Declaring such needs in a manifest is planned; today the
  environment is the only thing a manifest can declare.
- **Who wrote a plugin.** Plugins are not signed and there is no registry that vouches for them.

## The environment an external plugin gets

An external plugin starts with:

| Group | Variables |
|---|---|
| Finding programs and a home | `PATH`, `HOME`, `USERPROFILE`, `TMPDIR`, `TEMP`, `TMP`, `TZ` |
| Language and terminal | `LANG`, `LANGUAGE`, every `LC_*`, `TERM` |
| Network | `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, `NO_PROXY`, `SSL_CERT_FILE`, `SSL_CERT_DIR` |
| Set by smith | `SMITH_PLUGIN` (always `1`) and `SMITH_VERBOSE` |
| Windows only | `SYSTEMROOT`, `SYSTEMDRIVE`, `WINDIR`, `COMSPEC`, `PATHEXT`, `APPDATA`, `LOCALAPPDATA`, `PROGRAMDATA`, `ALLUSERSPROFILE`, `PUBLIC`, `PROGRAMFILES`, `PROGRAMFILES(X86)`, `PROGRAMW6432`, `COMMONPROGRAMFILES`, `COMMONPROGRAMFILES(X86)`, `COMMONPROGRAMW6432`, `HOMEDRIVE`, `HOMEPATH`, `USERNAME`, `USERDOMAIN`, `COMPUTERNAME`, `NUMBER_OF_PROCESSORS`, `PROCESSOR_ARCHITECTURE`, `OS` |

plus the variables it **declares**:

- in its own `plugin.json`: `"env": ["NPM_TOKEN", "AWS_*"]`, the plugin author's statement of what
  the plugin needs;
- in `smith.toml`: `env = ["GITHUB_TOKEN"]` under `[plugins.external.<name>]`, your own grant for this
  workspace.

A name is exact; a name ending in `*` is a prefix. Variables that are not set are not passed. A
plugin that declares nothing gets the baseline only.

## Related

- [Installing, enabling and configuring plugins](managing-plugins.md): trust, pin and untrust, step by step
- [Troubleshooting plugins](troubleshooting-plugins.md): the messages you may see
- [`smith plugin trust`](../../manual/plugin-trust.md), [`smith plugin untrust`](../../manual/plugin-untrust.md),
  [`smith plugin list`](../../manual/plugin-list.md), [`smith doctor`](../../manual/doctor.md)
- [Declaring the environment a plugin needs](../creating/packaging-a-plugin.md#declaring-the-environment-a-plugin-needs)
- [Decision 0036](../../architecture/decisions/0036-external-plugins-are-trusted-and-get-a-minimal-environment.md)
