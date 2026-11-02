# ADR 0076: One composer for the stack of a project

- Status: accepted
- Date: 2026-11-02

## Context

A person who wants "a React app that is also a desktop app" must know four places today: the
Add wizard (kind, language, then the choices of that language), `smith tauri add` for an
existing Node app, Settings > Forms for Tauri, and `smith recipe new` for a saved stack. The
language filter of the Add wizard finds a language by the names of its choices ("react" finds
Node), and nothing finds `electron` because no plugin offers it. A person cannot ask "what can
I build for the desktop?" or "what does serverless mean here?".

The owner asks for one screen where you search, then compose the stack and the things you want
to install, with every item placed on a layer. The owner rules (ask-never-impose, ADR 0072) apply:
every decision is a presented choice with its consequence in plain words and the default marked;
a conflict is explained before anything is written; a run with no terminal refuses and names the
flags. Electron must be discoverable and addable, and so must Tauri.

## Decision

### The catalogue is the choices, plus the languages and the plugins that already exist

There is no hand list. `smith compose --json` builds one catalogue from three sources the session
already holds:

1. the **choices** every loaded plugin publishes (ADR 0045): frameworks, desktop shells, function
   hosts, triggers and runtimes, test, lint and style tools;
2. the **languages**: the loaded plugins with the `language` role;
3. the **workspace plugins** with the `structure` role (architecture) or the `deploy` role
   (CI and deploy): they are listed and found by search, but they are not part of `--with`.
   Their entry carries a `route` that says which command turns them on.

An item is: `id` (`language/name`), `kind` (`language`, `choice`, `plugin`), `name`, `title`,
`description`, `aliases`, `layer`, `axis`, `language`, `default`, `suggested`, `needs`,
`requires`, `conflicts`, `problem` (why it cannot be used here) and `route` (set when the item is
not composed with `--with`, for example "run smith tauri add web").

### Protocol: three optional members on Choice

Display name, search aliases and layer are plugin data, so a third-party plugin gets them too.
`Choice` grows `title` (a display name: "Electron"), `aliases` (words that find the choice:
"desktop", "serverless", "web") and `layer`. All three are optional; the protocol stays at
version 1. The schema, the TypeScript, Python and C# types and the contract page are regenerated
(`smith run generate`). Without a title the name is shown; without aliases only name, title and
description are searched; without a layer the layer comes from the axis.

### Layers

A layer is where an item sits in the stack, bottom to top:

| Layer | What sits there | Comes from |
|---|---|---|
| `language` | the language of the project | language plugins |
| `framework` | how the app or lib is written: react, vue, svelte, html | axis `framework` |
| `shell` | what wraps the app as a desktop program: tauri, electron | `layer = "shell"` |
| `host` | where the code runs in the cloud: azure-functions, aws-lambda, gcp-functions, and the trigger and runtime they ask | `layer = "host"` |
| `test`, `lint`, `style` | tools of the project | the axis of the same name |
| `architecture` | the layout rule of the code | plugins with the `structure` role |
| `ci` | pipelines and deploy | plugins with the `deploy` role |

A choice with no `layer` sits on the layer named like its axis. An axis with no layer of its own
(a new `database` axis) shows as a layer of that name, so a new plugin needs no host change. The
layer is for display and grouping. The rules that decide a conflict stay the axis rules of the
solver.

### The composer reuses the solver

`choiceselect.SolveIn` already says, in one sentence that names both choices, why a selection
fails and how to fix it. The composer calls it with the picks so far, so it has one source of
truth and the CLI, the TUI and VS Code say the same words. A pick that has an open question
(`NeedsAxisError`: azure-functions needs a trigger) is not an error: it becomes a **question**
with its options, the suggested one marked and what each means. The composer never fills a
needed axis that has no default by itself.

### Commands

- `smith compose --json [--search words] [--lang l] [--with a,b,c]` prints the catalogue (items,
  layers, aliases). With `--search` the items are filtered and ranked. With `--with` it also
  prints the **check**: the conflicts, the questions and, when the selection is complete, the
  **plan**: the exact `smith add <kind> <name> --lang <l> --with a,b,c` line, the files and
  directories that will be written (the plan `smith add --dry-run` shows), and the tools needed.
  The command writes nothing.
- `smith compose` in a terminal is the interactive screen (below). Without a terminal and without
  `--json` it refuses: "smith compose needs a terminal. Add --json to print the catalogue, and
  --with a,b,c to check a selection." The screen ends in the plan, and the person chooses: run it
  now, copy the command, or save it as a recipe (the `smith recipe new` machinery with the same
  stack).
- `smith choices` stays as the flat listing by axis; its rows gain the three new members.

### The screens

One data source (`smith compose --json`), three screens, no logic in the screens beyond layout.

- **TUI**: a "Compose a stack" overlay, reached from the first step of the Add wizard, from the
  palette and from Home. It has a search box, the list of items grouped by layer, a ticked set,
  a live line for the conflict (naming both items and the fix) and the final plan with three
  actions. The model is a pure Bubble Tea model in its own package (`tui/composer`), tested by
  sending keys.
- **VS Code**: a multi-step quick pick (search, pick, questions, plan, run or copy) built in new
  files from the same JSON.
- **JetBrains**: recorded in the feature-parity page; not built in this batch.

The old entry points keep working and point at the composer: the Add wizard's language step
offers "Compose a stack" first, `smith tauri add` says that `smith compose` shows the
alternatives, and the language filter reads its aliases from the catalogue.

### Electron and Tauri

**Electron** is a new Node toolchain plugin `electron`, beside `functions` and `vite`: the
choice `electron` on the `framework` axis, `layer = "shell"`, role `scaffold`. It writes a
minimal Electron and TypeScript app: a main process, a preload script that uses
`contextBridge`, and one renderer page; tasks `start`, `build`, `typecheck`, `lint`, `test`,
`package`. It is on the `framework` axis because it writes the whole project, so the axis rule
makes `electron` with `react` an error ("both on the framework axis: keep one"). Electron
with a React renderer is a later step (a `shell` axis and a renderer slot), and the ADR does
not promise it. `electron` conflicts with the three function hosts and with `tauri`.
The packaging tool and the sources are in the section below. No code signing in v2 (ADR 0011).

**Tauri** keeps `smith tauri add` for an app that exists. For a new app it is a choice `tauri`
on the `shell` axis with the `toolchain` role, `layer = "shell"`, that the `tauri` plugin
publishes. The stack that `smith add --with react,tauri` saves puts the plugin in its toolchain,
as the stack `react-tauri` does today. If the scratch check in the Consequences section fails,
the composer lists `tauri` and routes it to `smith tauri add` through the `route` member, and
says so in the plan.

### Tools: a hook, not a table

The plan lists the tools the selection needs. The composer reads them through one function type,
`ToolSource func(choice string) []Tool`, that has no implementation in this change except the one
that reads `packages/plugin-sdk/functionhost` for the three hosts (the data that exists today).
The work on installing missing tools (the `toolinstall` slice and a protocol member for tool
requirements) replaces that source with the protocol data, and the plan then also offers to
install them. Nothing else in the composer changes.

### Rules, as the owner stated them

- Nothing is chosen silently. A default is marked `(default)` and a suggestion `(suggested)`;
  each option says what it does in plain words.
- A conflict message names both items and the fix, and is shown while the person picks, before
  the plan. `smith add` still checks again before it writes.
- Without a terminal nothing is asked: the command refuses and names the flags.

## Electron: packaging and sources

Checked on 2026-11-02: Electron's own packaging guide names Electron Forge as the recommended
all-in-one tool, and `electron-builder` as the usual alternative. Forge's quick start also needs
a bundler plugin and writes a lot of configuration. The template uses **Electron Forge**
(`@electron-forge/cli`, `make` creates the distributable) when the sources confirm it is current,
else `electron-builder`. The choice made, with the URLs, is recorded in the Electron section of
the Node catalogue page once the template runs. Sources:

- Electron tutorial and process model: https://www.electronjs.org/docs/latest/tutorial/process-model
- Context isolation and `contextBridge`: https://www.electronjs.org/docs/latest/tutorial/context-isolation
- Application packaging: https://www.electronjs.org/docs/latest/tutorial/application-distribution
- Electron Forge: https://www.electronforge.io/

## Consequences

- One search covers every plugin, so a new plugin appears on its own once it sets `title`,
  `aliases` and `layer`.
- The protocol grows by three optional members; third-party plugins keep working.
- A plugin author who leaves `layer` out gets the axis as the layer: the catalogue never hides
  an item.
- `electron` cannot be combined with a React renderer in v2.
- The architecture and CI items are findable, but turning them on stays with `smith plugin` and
  a recipe. The composer says that in the item, instead of pretending it writes them.
- The first check of Tauri as a choice for a new app is run in a scratch workspace; the result is
  written in the Node catalogue page.
