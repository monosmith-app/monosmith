# smith compose

`smith compose` is one place to search everything smith can put in a project and to compose the
stack: the language, the framework, the desktop shell (Tauri, Electron), the cloud host (Azure
Functions, AWS Lambda, Google Cloud Functions) and the test, lint and style tools. It ends in a
plan: the exact [`smith add`](add.md) line, the files it writes and the tools it needs. The
decision is in [ADR 0076](../architecture/decisions/0076-one-composer-for-the-stack-of-a-project.md).

It replaces nothing. `smith add`, `smith tauri add` and `smith recipe new` work as before; the
composer is the screen that shows them together.

## When to use it

- You know what you want in words ("desktop", "serverless", "react") but not the name of the choice.
- You want to see why two things do not go together before anything is written.
- You want the plan first: the command, the files and the tools.

For a project you already know how to describe, `smith add app web --lang node --with react` is
shorter.

## Synopsis

```sh
smith compose
smith compose --json [--search <words>] [--lang <language>] [--with <choice,...>] [--kind app|lib|package] [--name <name>]
```

## In a terminal

The screen has a search box, the catalogue grouped by layer and a line that tells the state of your
picks.

| Key | What it does |
|---|---|
| type | search by name, title, alias (`desktop`, `serverless`, `web`) or description |
| up, down | move in the list |
| enter | tick or untick the item under the cursor |
| tab | switch between the catalogue and your stack by layer |
| ctrl+n | go on: answer the open question, or name the project and see the plan |
| esc | clear the search, then leave the stack view, then close |

What the screen does for you:

- A **conflict** is shown as soon as you tick: it names both items and the fix, for example
  `react and electron are both on the framework axis: keep one`. The next step is refused until
  you fix it. The words come from the solver of `smith add --with`, so both always agree.
- An **open question** is asked, never answered for you. A function host needs a trigger and a
  runtime. The suggested answer is marked and starts selected, and you confirm it.
- A **default** is marked `(default)` and listed in the plan: `react` needs a lint tool and a test
  tool, so eslint and vitest are added, and the plan says so.
- A **workspace plugin** (an architecture or a CI plugin) is found by search but is not part of one
  project. The item says which command turns it on (`smith plugin enable <name>`).

The plan has three actions: `r` runs the command, `c` copies it, `s` saves the plugins of the
selection as a recipe (`smith recipe new --fixed ...`). A recipe fixes plugins, not choices:
another project made from it can pick another framework of the same plugin. `smith add --with`
already saves your exact selection as a stack in `smith.toml`. Nothing is written before you press
`r` or `s`, and `smith add` still asks to confirm.

The same screen opens in the [TUI](tui.md) from the command palette (Compose a stack…, key `A` in
the Projects view) and from Home when the workspace has no project. VS Code has the command
**MonoSmith: Compose a Stack…**, also offered as the first line of **Add Project**.

## Arguments and flags

<!-- smith:begin compose -->
Search everything smith can build a project from, and compose the stack

Usage: `smith compose [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON: the catalogue, and with --with the check and the plan |
| `--kind <kind>` | string |  | the kind of project the plan is for: app, lib or package |
| `-l, --lang <lang>` | string |  | show only this language's items, and check --with for it (also --language) |
| `--name <name>` | string |  | the project name the plan uses (<name> when not given) |
| `--search <search>` | string |  | find items by words: name, title, alias (desktop, serverless, web) or description |
| `--with <with>` | string, repeatable |  | check this selection and print its plan, comma separated (react,vitest) |

<!-- smith:end compose -->

## Without a terminal

`smith compose` alone refuses and names the flags:

```
smith: smith compose is a screen and needs a terminal. To read the catalogue without one, add --json
(smith compose --json --search desktop). To check a selection, add --with a,b,c
(smith compose --json --lang node --with react).
```

`--json` prints the catalogue:

```json
{
  "layers": [{ "id": "shell", "title": "Desktop shell", "about": "what wraps the app as a program you install on a computer" }],
  "items": [
    { "id": "node/electron", "kind": "choice", "name": "electron", "title": "Electron", "layer": "shell",
      "aliases": ["desktop", "windows", "mac", "linux", "native", "app", "chromium"], "axis": "framework", "language": "node",
      "conflicts": ["jest", "azure-functions", "aws-lambda", "gcp-functions", "tauri"], "score": 80, "matchedBy": "alias" }
  ]
}
```

The output also has `outcomes`: the answers to "What are you building?" (`web`, `api`, `function`, `desktop`,
`library`, `package`), from the whole catalogue whatever `--search` or `--lang` say. Each has `id`, `title`, `kind`
(app, lib or package), `languages`, `words`, `typical` and the `items` of its layer. The TUI and the editors list them
as the first question of an add.

With `--with`, the output also has `check` and `plan`:

| Member | Meaning |
|---|---|
| `check.complete` | the picks make a project |
| `check.conflict` | the picks cannot go together: the sentence names the items and the fix |
| `check.missing` | not wrong, but unfinished: the next thing to add |
| `check.question` | an axis the person must answer: `axis`, `options` with the suggested one marked |
| `check.selection`, `check.added` | the picks with the defaults, and the names added by default |
| `plan.command`, `plan.argv` | the exact `smith add` line |
| `plan.files` | what the add writes, as `smith add --dry-run` lists it, and what each choice puts in the project |
| `plan.tools` | the programs the selection needs (today the local tool of a cloud host) |
| `plan.recipeArgv` | the `smith recipe new` line that fixes the plugins |

Nothing is written in any mode.

## Layers

| Layer | What sits there |
|---|---|
| `language` | the language of the project |
| `framework` | how the app is written: react, vue, svelte, html |
| `shell` | what wraps it as a desktop program: Tauri, Electron |
| `host` | where the code runs in the cloud: Azure Functions, AWS Lambda, Google Cloud Functions, with their trigger and runtime |
| `library`, `style`, `test`, `lint` | tools of the project |
| `architecture` | the layout rule of the code (plugins with the `structure` role) |
| `ci` | pipelines and deploy (plugins with the `deploy` role) |

A plugin sets the layer, the title and the aliases of its choices
([Choice](../plugins/creating/protocol-cheat-sheet.md)). A choice with no layer sits on the layer
named like its axis, so a new plugin appears in the composer without a change to smith.

## See also

[`smith add`](add.md), [`smith choices`](choices.md), [`smith tauri add`](tauri.md),
[Feature parity](../reference/feature-parity.md).
