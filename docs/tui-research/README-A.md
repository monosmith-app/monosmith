# TUI research, part A: monorepo and Git/Docker TUIs

Tools: lazygit 0.66.2, lazydocker 0.25.2, Nx 24.0.5, Turborepo 2.11.10. All ran natively on Windows. Each has `install-notes.md`, `findings.md` and `screens/`.

PNG files are colour renders drawn from the terminal screen, not real screenshots. The `.txt` files are the raw captures. The tools live in `.tools/` (git-ignored).

Comparison:
1. Layout: lazygit and lazydocker stack numbered panels beside a detail pane. Nx and Turborepo show a task list with an output pane.
2. Footer: lazygit, Nx and Turborepo follow the focus. lazydocker shows one static line.
3. Help: lazygit `?` has a filter box (closest to a palette). Turborepo `m` lists all keys. lazydocker `x` is an action menu. Nx help is thin.
4. Mouse: lazygit selects and switches tabs, with no hover. Turborepo selects a task and scrolls logs. Nx scrolls output only. lazydocker is unverified.
5. Confirmations: only lazygit has described menus, a Cancel row, "Disabled: reason" and an error popup with copy.
6. Progress: Nx is best (spinner, duration, "Continuous", `n/m`, cache column). Turborepo shows a glyph per task.
7. Failure: Nx marks a red cross with raw output. Turborepo stops everything unless `--continue` is set.
8. Command transparency: only lazygit has a Command log strip.
9. Buttons: none of the four has buttons. Everything is a key.
10. Pick for smith: lazygit for footer, menus and confirmations. Nx for the Run list. Turborepo for `m` and click-select.

Caveats: the first lazygit start ran once against the real monosmith repo (Enter only) and wrote state to its user config folder. Redraws lagged 3 to 9 seconds, and captures that could not be matched to a key were deleted.
