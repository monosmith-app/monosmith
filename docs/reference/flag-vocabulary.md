# Flag vocabulary

<!-- Generated from libs/actions/flagvocabulary/flag_vocabulary.json by `go run ./flaggen` in libs/actions. Do not edit: change the vocabulary and run it again. -->

One idea has one flag in every smith command. This page lists the ideas that several commands share, the flag that carries each, and the other spellings every such command also accepts. The rule and its reason are in [ADR 0084](../architecture/decisions/0084-one-concept-one-flag.md). To add a shared flag, add the concept to `libs/actions/flagvocabulary/flag_vocabulary.json` first.

| Meaning | Flag | Also accepted | Value | Commands |
|---|---|---|---|---|
| The project or projects a command acts on; repeat it or separate names with commas (in smith the workspace is the whole repository, so -w is only a synonym) | `-p, --project` | `-w`, `--workspace`, `--projects` | text, repeatable or comma separated | `smith build`, `smith deps repair`, `smith docker description`, `smith install`, `smith lint`, `smith pipeline`, `smith release assets`, `smith release pub`, `smith run`, `smith tauri icon`, `smith test`, `smith typecheck`, `smith uninstall` |
| A programming language, by the name its plugin has | `-l, --lang` | `--language` | text | `smith add`, `smith compose`, `smith new`, `smith plugin draft`, `smith slices check` |
| The file or folder a command writes its result to | `-o, --output` | `--out` | text | `smith graph`, `smith npm-prune`, `smith plugin draft`, `smith plugin pack`, `smith report` |
| Do not ask for confirmation: take the safe answer | `-y, --yes` | - | on or off | many commands |
| Change the manifests but do not run the package manager | `--no-install` | `--skip-install` | on or off | `smith convert`, `smith deps up`, `smith sync`, `smith up` |
| Show what the command would do and change nothing | `--dry-run` | - | on or off | many commands |
| Print the result as JSON | `--json` | - | on or off | many commands |
| Commit the change as one operation so smith undo can revert it | `--commit` | - | on or off | many commands |
| Build the optimised profile (cargo --release, dotnet -c Release); a tool without profiles ignores it | `--release` | - | on or off | `smith build`, `smith lint`, `smith run`, `smith test`, `smith typecheck` |

The synonyms behave exactly like the flag: they can be mixed on one command line, and `--help` shows the flag once.

## Deliberate exceptions

These flags look like a shared one but are not.

| Command | Flag | Why |
|---|---|---|
| `smith graph` | `--format` | text, json, dot, html or tui: a choice of drawing, of which JSON is one, not the JSON switch |
| `smith release` | `--format` | text or markdown: the layout of the report; --json is its own switch |
