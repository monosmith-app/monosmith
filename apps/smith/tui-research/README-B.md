# TUI research, part B: dashboards, files and API clients

Tools: k9s v0.51.0, btop4win 1.0.5, Yazi 26.9.1, Posting 2.11.1. All ran natively on Windows. Each has `install-notes.md`, `findings.md` (btop, yazi and posting; k9s too) and `screens/`.

PNG files are colour renders drawn from the terminal screen. Nerd Font and emoji glyphs show as boxes. The `.txt` files are the raw captures. The btop and Posting screens show real process names, a user name and an IP address: do not publish them. The folders `btop/first-run-mixed` and `yazi/aborted-first-run` hold unreliable runs.

Comparison:
1. Layout: k9s is one table with a header. btop has four live boxes. Yazi has three columns. Posting has a URL bar and three panes.
2. Navigation: k9s uses `:` then Enter and Esc. btop has box toggles 1 to 4 and presets. Yazi has vim keys and prefix panels. Posting has focus areas and a `Ctrl+O` jump.
3. Key hints: k9s has a header per view. btop puts hotkey letters in titles. Yazi has a panel after a prefix. Posting has a context footer.
4. Help: k9s `?`, btop F1 or `h`, Yazi `~`, Posting F1 for the focused widget.
5. Palette: only Posting has a true palette (`Ctrl+P`). k9s has `:` with autocomplete.
6. Confirmations: k9s has a Cancel-first dialog. Yazi lists the paths with `[Y]es` and `(N)o`. btop has none for kill.
7. Progress: k9s has a live table and logs. btop refreshes every 1500 ms. Yazi has a task badge. Posting has a latency strip.
8. Errors: k9s uses a modal. Yazi is inline. Posting uses a toast and a status pill.
9. Mouse: click selects in all four. Wheel and hover were not verified.
10. Best source for smith: k9s for Doctor and Deps, Posting for palette and footer, Yazi for the task badge and empty state, btop for hotkey letters and reflow.

Caveats: a `k` key sent to btop with a process selected may have killed a process. The agent started Docker Desktop (still running) and a temporary `kind` cluster (deleted).
