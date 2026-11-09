# words

The text catalogue of every smith interface ([ADR 0089](../../docs/architecture/decisions/0089-one-text-catalogue-for-every-interface.md)).
One source holds every sentence, label and hint. Each interface reads it in its own format.

## Layout

| Path | What it is |
|---|---|
| `catalog/<locale>/*.json` | The texts: one file per area, a flat map from key to text. English (`en`) is the reference. Package `catalog` embeds them. |
| `surface/<name>.json` | The keys one interface uses: `tui`, `vscode`, `jetbrains`. Package `surface` reads them. |
| `wordgen/` | The generator `go run ./wordgen`. It writes the files of the editors from the catalogue. |
| `words/` | Package `words` (import `github.com/monosmith-app/monosmith/libs/words/words`): the loader, `Say`, `Count`, `Reference` and `LoadCatalog`. The TUI and the CLI use it. The module root holds only `doc.go`, as the slice rules need. |

## Keys and placeholders

A key is `<area>.<what>`. A text that one interface only can show starts with that interface
(`tui.`, `vscode.`, `jetbrains.`). A text takes `{name}` placeholders, filled by name. A plural is
two keys, `key.one` and `key.other`. A locale that lacks a key shows the English text, key by key.

## Surface manifest

`surface/<name>.json` has two lists. Both can be empty.

```json
{
  "keys": ["cmd.add.title", "cmd.remove.title"],
  "prefixes": ["vscode."]
}
```

- `keys`: exact catalogue keys the surface uses. A key the catalogue lacks is an error.
- `prefixes`: a prefix claims every catalogue key that starts with it.

A key is valid when at least one surface claims it. A key that no surface claims fails the guard
(`apps/smith/tui/wordsguard`). The `tui` manifest stays empty: the TUI claims a key by asking for it in
code, which the guard scans. `vscode` claims `vscode.` and `jetbrains` claims `jetbrains.`.
To use a shared key in an editor, add it to that editor's `keys`.

## Generator

Run it from `libs/words`:

```
go run ./wordgen            # write every file
go run ./wordgen -check     # write nothing; fail when a file is out of date (a Go test runs this)
go run ./wordgen -surface vscode
```

| Argument | Default (relative to `-root`, default `../..`) |
|---|---|
| `-words` | `libs/words` |
| `-surface` | `vscode,jetbrains` |
| `-vscode-module` | `apps/vscode/src/words` (`smith-words.config.ts`) |
| `-vscode-nls` | `apps/vscode` (`package.nls.json`, `package.nls.<locale>.json`) |
| `-jetbrains-resources` | `apps/jetbrains/src/main/resources` (`messages.properties`, `messages_<locale>.properties`) |
| `-jetbrains-kotlin` | `apps/jetbrains/src/main/kotlin/dev/monosmith/jetbrains/words` (`SmithWordsConfig.kt`) |

Rules the generator follows:

- A placeholder `{name}` becomes `{0}`, `{1}`... in the order of first appearance in the English text.
  A translation keeps the same positions, so it may reorder words. The `.properties` file has a
  comment above each such key that lists the order.
- A translation that uses a placeholder its English text lacks is an error.
- `package.nls.<locale>.json` and `messages_<locale>.properties` hold only the keys that locale
  translates. VS Code and `ResourceBundle` fall back to the base file. The TypeScript module holds
  complete tables (English fills the gaps).
- In a `.properties` text with placeholders, an apostrophe is doubled (MessageFormat). A locale gets files
  only when it has at least one own key that the surface claims.
- With no key claimed, the files are still valid (an empty object, a bundle with only a comment, an
  object with only `say`). In check mode a missing file of a surface with no key is not stale.
- The generated files are committed. A stale file fails `smith run test -p words`.

## Guards

- `apps/smith/tui/wordsguard`: key exists, placeholders filled, key claimed by a surface, no raw
  wording in a migrated TUI package.
- `words/surface_claims_validator_test.go`: an exact key in a manifest exists in the catalogue.
- `wordgen`: the generated files of the repository are up to date.
