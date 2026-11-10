# Smith TUI: UX review and proposal

Sources: the research in this folder, a coverage review of what smith can do and how the TUI reaches it, and a first-time-user audit of the local build (Windows, 110-column terminal, mouse not tested).

Goal: a user who knows nothing about smith reaches what they need, with the fewest steps and the least vocabulary.

## What works

- Build and test: the command is shown, a spinner and an `n/m` bar run, cache hits show, and a toast sums up.
- A failed task is selected with its log at once.
- The Fix dialog has numbered options, a recommended default and a clear cancel.
- Every dialog has Esc, and the status says nothing changed.
- Undo after a committed change shows the plan, the files and the command.
- The empty folder names the next key.

## Findings, ranked

1. Home says "All good" while the Doctor tab shows a warning.
2. A failed run is not the Home headline and the next action is not "see why".
3. The dependency plan lists 5 packages and "and 4 more" when 3 are marked.
4. The palette matches names, not intent: "update", "outdated", "publish" and "ship" find nothing useful.
5. Undo is off by default, so the palette answers "Nothing to undo".
6. A key that cannot run gives no reason (palette, footer, help in an empty folder).
7. The key `b` means "this project" on Projects and "all" on Run.
8. The footer misses the key that matters: no "run again" on a failed task, and `n` and `q` stay shown while a text box is open.
9. Cancel looks like failure: the bar reads 100% and the task is red.
10. The TUI cannot finish a release. The text says "run smith release here".
11. The release empty state is a wall of text, repeated in the preview.
12. Doctor prints raw tool strings (build hash, repeated text).
13. The dependency wait has no spinner. Plugin toggles in Settings save at once with no undo.
14. Jargon on first screens (see the vocabulary table).
15. No task badge for background work. A dry run shows a green tick as if something shipped.

## Vocabulary (plain word first, smith word in brackets)

- Doctor: "Check my project"
- Typecheck, Lint, Verify all: "Check types", "Check style", "Run all checks"
- Compose a stack: "Pick frameworks and tools"
- Swap a choice: "Change my test tool or linter"
- Derived files: "Generated files smith keeps in sync"
- Divergences: "Where I differ from the defaults"
- Cache: "Speed-up storage: size, check, trim, delete"
- Upgrade workspace: "Update smith's own files" (not the same as Deps update)
- Align, one version: "Use the same version everywhere"
- Regenerate CI: "Update my automatic build files"
- Plugins, recipes, trust: "Language and tool support", "team presets", "approve a plugin"
- Write the launcher: "Make everyone use the same smith"
- Convert: no word today. Add "Nx, Turborepo, Lerna" to the "Bring into MonoSmith" entry.

## Abilities with no TUI path (verify each before acting)

`eject`, `sync`, `up`, `kind`, `tauri add`, `boundaries`, `slices check`, `ci check`, `agent install`, `release assets`, `docker` build and push, `plugin pack/draft/check`, `graph --blast-radius`. Editor and script tools (`schema`, `serve`, `lsp`, `completion`) need none.

Reached only by a three-step palette search: undo, remove a project, install missing tools, cache care, regenerate CI.

## Ideas from the research

- lazygit: `Disabled: reason` for a key that cannot run. A filterable key list. Confirm menus that describe each choice. A Command log line.
- Nx: running, done and pending groups. A `n/m (time)` counter. A "Continuous" label for dev servers. A pin.
- Turborepo: `m - More binds`. Click to select a task. An Interact mode.
- k9s: a title that holds state and count (`pods(demo)[6]`). A Cancel-first delete dialog. Row colour by health.
- Yazi: a panel that lists the next keys after a prefix. A task badge. A confirmation that lists what changes. A named empty state.
- Posting: a palette with a description per command. A footer that follows focus. Toasts for errors.
- btop: the hotkey letter inside the label. Reflow on a small terminal.
- Avoid: kill with no confirmation (btop), a blank task list (Yazi), a stale status badge (Posting), a static footer (lazydocker), fail-fast that stops everything (Turborepo).

## Proposed packages (each ships alone)

- A. Home tells the truth: a rollup headline of health, last run and updates. A failure becomes the first row with "See why".
- B. Plain words: synonyms in the palette, "did you mean" instead of "no command", the vocabulary table above.
- C. Safety net: undo asked once as a choice, a grey "cancelled" state, undo for plugin toggles.
- D. Keys and feedback: `Disabled: reason`, one meaning for `b` and `B`, a footer per mode, a spinner on waits, a task badge.
- E. Finish a release in the TUI: a confirmed "Release now" with a pre-flight line, and a short empty state.
- F. Clean details: the dependency plan lists only what is marked. Doctor shows a one-line cause with details folded.
- G. Missing paths: a TUI route for the abilities in the list above that users need.

## Not verified

Mouse and hover. The Customise wizard after "Start with the defaults". The end of the add-project wizard. The meaning of the `⎇ main ▲1` header. Wide terminals only. Release publishing and major upgrades.
