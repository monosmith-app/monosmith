# ADR 0089: One text catalogue for every interface

- Status: accepted
- Date: 2026-11-06
- Follows: [0084](0084-one-concept-one-flag.md)

## Context

The terminal interface (TUI) keeps all its text in a catalogue: 1,900 keys in `apps/smith/tui/words/catalog/en/`, filled by name (`{name}`), with a guard test that fails an unused key, a missing key and raw wording. The two editor extensions do not read it.

- VS Code holds 90 command titles in `package.json`, about 80 prompt and message strings in TypeScript, and 49 titles in `guided-commands.config.ts`.
- JetBrains holds 68 `text=` entries in `plugin.xml` and several hundred literals in Kotlin handlers.

Only 21 of the 90 VS Code command titles equal a TUI title. The rest say the same thing in other words ("Remove a package…" in one place, another wording in the other). A reword or a translation has to be made three times, and the three can drift.

ADR 0084 solved the same problem for flags: one JSON vocabulary in `libs/actions/flagvocabulary`, and a generator (`flaggen`) that writes the Go, TypeScript and Kotlin clients.

## Options considered

- Each interface keeps its own text and a test compares them: not chosen. It finds drift after it happened and does not remove the triple work.
- A smith command serves the text at run time (the extensions ask the engine): not chosen. `package.json` and `plugin.xml` are read by the marketplaces and the IDE before any process runs, so static files are needed anyway.
- One source catalogue and a generator that writes each interface's native text files: chosen.

## Decision

1. **One source.** A new library holds the catalogue as JSON, one file per area, one folder per locale (`catalog/<locale>/`). English is the reference, and a locale that lacks a key falls back to English key by key, as the TUI does now.
2. **Generated output, in each interface's own format.**
   - Go (TUI and CLI): the embedded JSON the `words` package reads today.
   - VS Code: `package.nls.json` and `package.nls.<locale>.json` (the native localisation of `package.json`, with `%key%` references) and a typed TypeScript module for the prompts.
   - JetBrains: `messages.properties` and `messages_<locale>.properties` (a resource bundle, referenced from `plugin.xml` by key) and a Kotlin accessor with typed keys.
3. **Keys.** `<area>.<what>`. A text that only one interface can show carries that interface as the first part (`tui.` for key labels and mouse hints, `vscode.`, `jetbrains.`). Every other text is shared and written once. A generator step emits to each interface only the keys it uses.
4. **Placeholders.** The source writes `{name}`. The generator turns it into the target form (`{0}` for VS Code and for the JetBrains MessageFormat) and keeps the order a translation chooses. Plurals stay as two keys (`key.one`, `key.other`) that the code picks.
5. **Guard tests.** The existing guard for the TUI stays. Each generator run fails on a key a surface uses and the catalogue lacks, on a placeholder the call does not fill, and on a generated file that is out of date. A guard per interface forbids new raw user-facing literals in migrated code, as the TUI guard does.
6. **Generated files are committed**, like the flag clients, so a build never needs the generator and a review sees every text change.

## Plan (each step ships on its own)

1. Move the catalogue and the loader out of `apps/smith/tui/words` into the new library. The TUI reads the generated embed. No visible change.
2. VS Code `package.json`: 90 titles become `%key%` references with a `package.nls.json`. Mostly mechanical.
3. VS Code prompts and `guided-commands.config.ts` become keys, with a guard.
4. JetBrains `plugin.xml` and the Kotlin literals move to the bundle, with a guard. The Kotlin build cannot run offline here, so CI is the check.
5. Unify the wording where the same action has different words in the interfaces. Each difference is a product decision for the owner, shown as a choice (ask, never impose).

## Consequences

- A change, a correction or a new locale is made once and reaches every interface.
- Shared text drifts no more, and a translation is one set of files per locale.
- The build gains a generator step and generated files in each extension. A stale file fails CI, so nobody forgets to run it.
- Text that is only right in a terminal (keys, mouse hints) stays out of the editors by its `tui.` prefix.
- Step 5 changes visible wording in the editors and in the terminal. It is the only step that does, and it is done last and per decision.

## Owner decisions

- The library lives in `libs/words`, next to `libs/actions` and the flag vocabulary.
- The second locale is pt-BR, so the first generated bundles are tested with a real second language. The first release carries a small reviewed sample of pt-BR keys. English fills every other key by the key-by-key fallback.
- Step 5 (one wording across the interfaces) ships as its own later release, after steps 1 to 4, which change no visible text.

## Where this can be checked

Step 1 is built: the catalogue, the loader and the generator live in `libs/words` (`libs/words/words`, `libs/words/wordgen`), and the guard in `apps/smith/tui/wordsguard`. The TUI catalogue and its guards: `apps/smith/tui/words`. The flag precedent: `libs/actions/flagvocabulary` and `libs/actions/flaggen`.
